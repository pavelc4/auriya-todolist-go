// @title Auriya Todolist API
// @version 1.0
// @description This is a sample server for a todolist application.
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080
// @BasePath /api
// @schemes http https

// @securityDefinitions.apiKey BearerAuth
// @in header
// @name Authorization
package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pavelc4/auriya-todolist-go/internal/cache"
	"github.com/pavelc4/auriya-todolist-go/internal/config"
	"github.com/pavelc4/auriya-todolist-go/internal/database"
	"github.com/pavelc4/auriya-todolist-go/internal/http/repository"
	"github.com/pavelc4/auriya-todolist-go/internal/http/router"
	"github.com/pavelc4/auriya-todolist-go/internal/http/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is empty")
	}

	ctx := context.Background()
	db, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Close()

	// Initialize cache service
	cacheSvc := cache.NewService(5*time.Minute, 10*time.Minute)

	googleConf := cfg.GoogleOAuthConfig
	githubConf := cfg.GitHubOAuthConfig
	userRepo := repository.NewUserRepository(db, cacheSvc)
	jwtService := service.NewJWTService(os.Getenv("JWT_SECRET"))

	r := router.New(db, googleConf, githubConf, userRepo, jwtService, cacheSvc)

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.AppPort),
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("HTTP listening on :%d", cfg.AppPort)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	// Graceful shutdown
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-errCh:
		log.Printf("server error: %v", err)
	case sig := <-stop:
		log.Printf("shutdown signal: %v", sig)
	}

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		log.Printf("server shutdown error: %v", err)
	}
	log.Println("server exited")
}
