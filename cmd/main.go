package main

import (
	"os"
	"time"

	sq "github.com/Masterminds/squirrel"
	swaggo "github.com/gofiber/contrib/v3/swaggo"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/itzLilix/questboard-profile-service/internal/config"
	"github.com/itzLilix/questboard-profile-service/internal/handlers"
	"github.com/itzLilix/questboard-profile-service/internal/infrastructure"
	"github.com/itzLilix/questboard-profile-service/internal/middleware"
	"github.com/itzLilix/questboard-profile-service/internal/usecase"
	"github.com/itzLilix/questboard-shared/images"
	"github.com/itzLilix/questboard-shared/jwt"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// @title          Session Zero Profile Service
// @version        1.0
// @description    Profile, auth and user catalog API for Session Zero
// @host           localhost:3000
// @BasePath       /
// @securityDefinitions.apikey  CookieAuth
// @in             cookie
// @name           access_token
func main() {
	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		With().Timestamp().Logger()

	err := godotenv.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("error loading .env file")
	}

	cfg := config.Load()
	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowCredentials: true,
		AllowHeaders:     []string{"Content-Type"},
	}))
	app.Use(middleware.Logger(log.Logger))

	app.Get("/uploads/*", static.New(cfg.UploadDir))
	if cfg.Env != config.ProdEnv {
   		app.Get("/swagger/*", swaggo.HandlerDefault)
	}

	conn, err := infrastructure.Connect(cfg.DatabaseURL, int32(cfg.MinPoolSize), int32(cfg.MaxPoolSize))
	if err != nil {
		log.Fatal().Err(err).Msg("failed to connect to database")
	}
	log.Info().Msg("successfully connected to database")
	defer conn.Close()

	err = infrastructure.RunMigrations(cfg.MigrateURL)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to run migrations")
	}
	log.Info().Msg("migrations ran successfully")

	tokenProvider := jwt.NewProvider([]byte(cfg.JWTSecret), cfg.AccessTTL)
	refreshTokens := infrastructure.NewRefreshTokenManager(cfg.RefreshTTL)
	passwordHasher := infrastructure.NewPasswordHasher()
	rbacMiddleware := middleware.NewRBACMiddleware(tokenProvider, log.Logger)
	internalOnly := middleware.InternalOnly(cfg.InternalToken)

	authRepo := infrastructure.NewAuthRepository(conn, psql)
	authService := usecase.NewAuthUsecase(authRepo, tokenProvider, refreshTokens, passwordHasher)
	authHandler := handlers.NewAuthHandler(authService, log.Logger, cfg)

	imageUploader := images.NewLocalImageUploader(cfg.UploadDir, cfg.PublicBaseURL, cfg.MaxUploadSize)
	usersRepo := infrastructure.NewUsersRepository(conn, psql)
	usersService := usecase.NewUsersUsecase(usersRepo, imageUploader)
	usersHandler := handlers.NewUsersHandler(usersService, log.Logger, rbacMiddleware, internalOnly)

	catalogRepo := infrastructure.NewCatalogRepository(conn, psql)
	catalogUsecase := usecase.NewCatalogUsecase(catalogRepo)
	catalogHandler := handlers.NewCatalogHandler(catalogUsecase, log.Logger, rbacMiddleware)

	v1 := app.Group("/v1")
	authHandler.RegisterRoutes(v1)
	usersHandler.RegisterRoutes(v1)
	usersHandler.RegisterInternalRoutes(v1)
	catalogHandler.RegisterRoutes(v1)

	log.Fatal().Err(app.Listen(":" + cfg.ServerPort)).Msg("server stopped")
}
