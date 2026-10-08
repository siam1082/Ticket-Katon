package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"ticket-katon-backend/internal/config"
	"ticket-katon-backend/internal/database"
	httpHandler "ticket-katon-backend/internal/handler/http"
	"ticket-katon-backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Init Database Connections
	pgPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Postgres connection failed: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := database.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("[FATAL] Redis connection failed: %v", err)
	}
	defer redisClient.Close()

	// 2. Init Service & Handler
	healthSvc := service.NewHealthService(pgPool, redisClient)
	healthHdl := httpHandler.NewHealthHandler(healthSvc)

	// 3. Router Setup
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/health", healthHdl.HealthCheck)

	// 4. Server with Graceful Shutdown
	srv := &http.Server{
		Addr:         ":" + cfg.Port,
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	serverErrors := make(chan error, 1)
	go func() {
		log.Printf("[SERVER] Ticket-Katon API running on port %s", cfg.Port)
		serverErrors <- srv.ListenAndServe()
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErrors:
		if !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[FATAL] Server closed with error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("[SERVER] Received signal %v. Shutting down gracefully...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[FATAL] Graceful shutdown failed: %v", err)
		}
		log.Println("[SERVER] Server exited successfully.")
	}
}