package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort    string
	DatabaseURL   string
	MigrateURL    string
	MinPoolSize   int64
	MaxPoolSize   int64
	JWTSecret     string
	AccessTTL     time.Duration
	RefreshTTL    time.Duration
	UploadDir     string
	PublicBaseURL string
	MaxUploadSize int64
	InternalToken string
}

func Load() *Config {
	JWTSecret := getEnv("JWT_SECRET", "")
	if JWTSecret == "" {
		panic("JWT_SECRET is required")
	}

	internalToken := getEnv("INTERNAL_TOKEN", "")
	if internalToken == "" {
		panic("INTERNAL_TOKEN is required")
	}

	return &Config{
		ServerPort:    getEnv("SERVER_PORT", "3000"),
		DatabaseURL:   getEnv("POSTGRES_URL", ""),
		MigrateURL:    getEnv("MIGRATE_URL", ""),
		JWTSecret:     JWTSecret,
		AccessTTL:     time.Minute * 15,
		RefreshTTL:    time.Hour * 24 * 30,
		MinPoolSize:   getEnvInt64("MIN_POOL_SIZE", 5),
		MaxPoolSize:   getEnvInt64("MAX_POOL_SIZE", 25),
		UploadDir:     getEnv("UPLOAD_DIR", "./uploads"),
		PublicBaseURL: getEnv("PUBLIC_BASE_URL", "http://localhost:3000"),
		MaxUploadSize: getEnvInt64("MAX_UPLOAD_SIZE", 5*1024*1024),
		InternalToken: internalToken,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
