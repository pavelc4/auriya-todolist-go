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

	googleConf := cfg.GoogleOAuthConfig
	githubConf := cfg.GitHubOAuthConfig
	userRepo := repository.NewUserRepository(db)
	jwtService := service.NewJWTService(os.Getenv("JWT_SECRET"))

	r := router.New(db, googleConf, githubConf, userRepo, jwtService)

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
