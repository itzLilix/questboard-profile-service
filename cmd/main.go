package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/itzLilix/QuestBoard/backend/internal/handlers"
	"github.com/itzLilix/QuestBoard/backend/internal/middleware"
	"github.com/itzLilix/QuestBoard/backend/internal/repositories"
	"github.com/itzLilix/QuestBoard/backend/internal/useCases"
	"github.com/itzLilix/QuestBoard/backend/pkg/database"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger()

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading .env file")
	}

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type"},
	}))
	app.Use(middleware.Logger(log.Logger))

	dbURL := os.Getenv("POSTGRES_URL")
	conn, err := database.Connect(dbURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("successfully connected to database")
	defer conn.Close()

	err = database.RunMigrations(os.Getenv("MIGRATE_URL"))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations ran successfully")

	authRepo := repositories.NewAuthRepository(conn)
	authService := useCases.NewAuthUseCase(authRepo)
	authHandler := handlers.NewAuthHandler(authService, log.Logger)

	authHandler.RegisterRoutes(app)

	log.Fatal().Err(app.Listen(":3000")).Msg("server stopped")
}
