package routes

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/ishank838/go-url-shortner/api/database"
)

func ResolveUrl(c *fiber.Ctx) error {
	url := c.Params("url")

	value, err := database.GetValue(database.Ctx, url)
	if err != nil {
		log.Println(err)
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "short not found in database"})
	}

	database.Increment(database.Ctx, "counter")
	c.Redirect(value, 301)
	return nil
}
