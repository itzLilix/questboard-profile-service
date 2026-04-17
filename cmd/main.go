package main

import (
	"os"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/itzLilix/questboard-profile-service/internal/config"
	"github.com/itzLilix/questboard-profile-service/internal/database"
	"github.com/itzLilix/questboard-profile-service/internal/handlers"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/repositories"
	"github.com/itzLilix/questboard-profile-service/internal/usecases"
	"github.com/itzLilix/questboard-shared/hash"
	"github.com/itzLilix/questboard-shared/jwt"
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

	cfg := config.Load()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type"},
	}))
	app.Use(middleware.Logger(log.Logger))

	conn, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("successfully connected to database")
	defer conn.Close()

	err = database.RunMigrations(cfg.MigrateURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations ran successfully")

	tokenProvider := jwt.NewTokenProvider([]byte(cfg.JWTSecret), cfg.AccessTTL, cfg.RefreshTTL)
	passwordHasher := hash.NewPasswordHasher()

	authRepo := repositories.NewAuthRepository(conn)
	authService := usecases.NewAuthUseCase(authRepo, tokenProvider, passwordHasher)
	authHandler := handlers.NewAuthHandler(authService, log.Logger, cfg)

	authHandler.RegisterRoutes(app)

	log.Fatal().Err(app.Listen(":" + cfg.ServerPort)).Msg("server stopped")
}
