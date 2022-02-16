package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/ishank838/go-url-shortner/api/database"
	"github.com/ishank838/go-url-shortner/api/routes"
	"github.com/joho/godotenv"
)

func setupRoutes(app *fiber.App) {
	app.Get("/:url", routes.ResolveUrl)
	app.Post("/api/", routes.ShortenUrl)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("error loading env: %v", err)
	}

	//app := fiber.New(logger.New())
	app := fiber.New()

	setupRoutes(app)
	setupDb()
	log.Fatal(app.Listen(os.Getenv("APP_PORT")))
}

func setupDb() {
	redisAddress := os.Getenv("DB_ADDRESS")
	dbPass := os.Getenv("DB_PASS")
	err := database.InitRedis(redisAddress, dbPass)
	if err != nil {
		log.Fatal(err)
	}
}
