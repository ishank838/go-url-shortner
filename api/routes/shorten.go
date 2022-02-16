package routes

import (
	"log"
	"os"
	"strconv"
	"time"

	"github.com/asaskevich/govalidator"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/ishank838/go-url-shortner/api/database"
	"github.com/ishank838/go-url-shortner/api/helpers"
)

type request struct {
	Url         string        `json:"url"`
	CustomShort string        `json:"custom_short"`
	Expiry      time.Duration `json:"expiry"`
}

type response struct {
	Url           string        `json:"url"`
	CustomSort    string        `json:"short_url"`
	Expiry        time.Duration `json:"expiry"`
	XRateLimiting int           `json:"rate_limit"`
	XRateReset    time.Duration `json:"rate_limit_reset"`
}

func ShortenUrl(c *fiber.Ctx) error {

	req := request{}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "cannot parse json"})
	}
	//rate limiting
	value, err := database.GetValue(database.Ctx, c.IP())
	if err == redis.Nil {
		_ = database.SetValue(database.Ctx, c.IP(), os.Getenv("API_QUOTA"), time.Second*60*30)
		//log.Println(err)
	} else if err != nil {
		log.Println("error at shortenUrl:", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "ip not registered"})
	} else {
		valInt, err := strconv.Atoi(value)
		if err != nil {
			log.Println("error in converting value to int", err)
		}
		if valInt <= 0 {
			limit, _ := database.GetTTl(database.Ctx, c.IP())
			return c.Status(fiber.StatusServiceUnavailable).JSON(
				fiber.Map{"error": "Rate limit exceeded",
					"rate_limit_reset": limit / time.Nanosecond / time.Minute},
			)
		}
	}

	//check url
	if !govalidator.IsURL(req.Url) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "not a valid url"})
	}

	//check domain error
	if !helpers.RemoveDomainError(req.Url) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "invalid url"})
	}

	//enforce http
	req.Url = helpers.EnforceHttp(req.Url)

	//implement custom url
	var id string
	if req.CustomShort == "" {
		id = uuid.New().String()[:6]
	} else {
		id = req.CustomShort
	}

	val, _ := database.GetValue(database.Ctx, id)
	if val != "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"error": "Url custom short is already in use"})
	}

	if req.Expiry == 0 {
		req.Expiry = 24
	}

	err = database.SetValue(database.Ctx, id, req.Url, 0)
	if err != nil {
		log.Println("error in shorten setting url", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "unable to connect to server"})
	}

	resp := response{
		Url:           req.Url,
		CustomSort:    "",
		Expiry:        req.Expiry,
		XRateLimiting: 10,
		XRateReset:    30,
	}

	//Decrement limit
	err = database.Decrement(database.Ctx, c.IP())
	if err != nil {
		log.Println("error at shorten", err)
	}

	newVal, err := database.GetValue(database.Ctx, c.IP())
	if err != nil {
		log.Println("error at shorten get value", err)
	}
	resp.XRateLimiting, _ = strconv.Atoi(newVal)

	ttl, err := database.GetTTl(database.Ctx, c.IP())
	if err != nil {
		log.Println("error at shorten gettl:", err)
	}
	resp.XRateReset = ttl / time.Nanosecond / time.Minute

	resp.CustomSort = os.Getenv("DOMAIN") + "/" + id
	return c.Status(fiber.StatusOK).JSON(resp)
}
