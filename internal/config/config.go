package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"golang.org/x/oauth2"
	oauth2github "golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"
)

type Config struct {
	GoogleOAuthConfig *oauth2.Config
	GitHubOAuthConfig *oauth2.Config
	DatabaseURL       string
	AppPort           int
	AppEnv            string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

	port := 8080
	if v := os.Getenv("APP_PORT"); v != "" {
		if p, err := strconv.Atoi(v); err == nil {
			port = p
		}
	}

	cfg := &Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		AppPort:     port,
		AppEnv:      os.Getenv("APP_ENV"),
		GoogleOAuthConfig: &oauth2.Config{
			ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
			ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GOOGLE_REDIRECT_URL"),
			Scopes:       []string{"openid", "profile", "email"},
			Endpoint:     google.Endpoint,
		},
		GitHubOAuthConfig: &oauth2.Config{
			ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
			ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
			RedirectURL:  os.Getenv("GITHUB_REDIRECT_URL"),
			Scopes:       []string{"user.:email"},
			Endpoint:     oauth2github.Endpoint,
		},
	}

	return cfg, nil
}
