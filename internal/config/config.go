package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	AppPort     int
	AppEnv      string
}

func Load() (*Config, error) {
	_ = godotenv.Load()
	port := 8080
	if v := os.Getenv("APP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}
	return &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppPort:     port,
		AppEnv:      os.Getenv("APP_ENV"),
	}, nil
}
