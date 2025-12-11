package main

import (
	"log"

	"achievements-uas/database"
	"achievements-uas/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("[WARN] .env file not found, using system environment")
	}

	// Connect PostgreSQL
	if err := database.ConnectPostgres(); err != nil {
		log.Fatal("[FATAL] Failed to connect PostgreSQL:", err)
	}

	// Connect MongoDB
	if err := database.ConnectMongo(); err != nil {
		log.Fatal("[FATAL] Failed to connect MongoDB:", err)
	}

	// Init Fiber App
	app := fiber.New()

	// Register all routes in one file
	routes.RegisterRoutes(app)

	// Start Server
	log.Println("Server running on :3000")
	if err := app.Listen(":3000"); err != nil {
		log.Fatal("[FATAL] Failed to start server:", err)
	}
}
