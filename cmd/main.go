package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/itzLilix/QuestBoard/backend/internal/auth"
	"github.com/itzLilix/QuestBoard/backend/internal/games"
	"github.com/itzLilix/QuestBoard/backend/pkg/database"
	"github.com/joho/godotenv"
)


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file", err)
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type"},
	}))
	
	dbURL := os.Getenv("POSTGRES_URL")
	conn, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}
	fmt.Println("Successfully connected to database")
	defer conn.Close()

	err = database.RunMigrations(os.Getenv("MIGRATE_URL"))
	if err != nil {
		log.Fatal("Failed to run migrations: ", err)
	}
	fmt.Println("Migrations ran successfully")

	authRepo := auth.NewRepository(conn)
	authService := auth.NewService(authRepo)
	authHandler := auth.NewHandler(authService)
	
	gamesHandler := games.NewHandler(authService)
	
	authHandler.RegisterRoutes(app)
	gamesHandler.RegisterRoutes(app)

	log.Fatal(app.Listen(":3000"))
}