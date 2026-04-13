package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort   	string
	DatabaseURL     string
	MigrateURL  	string
	JWTSecret    	string    
	AccessTTL 		time.Duration
	RefreshTTL		time.Duration
}

func Load() *Config {
	JWTSecret := getEnv("JWT_SECRET", "")
	if JWTSecret == "" {
		panic("JWT_SECRET is required")
	}

	return &Config{
		ServerPort:    	getEnv("SERVER_PORT", "3000"),
		DatabaseURL:         	getEnv("POSTGRES_URL", ""),
		MigrateURL:    	getEnv("MIGRATE_URL", ""),
		JWTSecret:     	JWTSecret,
		AccessTTL: 		time.Minute * 15,
		RefreshTTL: 	time.Hour * 24 * 30,
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}

	parsed, err := strconv.Atoi(value)
	if err != nil || parsed <= 0 {
		return defaultValue
	}
	return parsed
}
