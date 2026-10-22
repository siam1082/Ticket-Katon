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

	"github.com/go-chi/cors"
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"

	"ticket-katon-backend/internal/config"
	"ticket-katon-backend/internal/database"
	"ticket-katon-backend/internal/domain"
	httpHandler "ticket-katon-backend/internal/handler/http"
	"ticket-katon-backend/internal/middleware"
	"ticket-katon-backend/internal/pkg/token"
	"ticket-katon-backend/internal/repository/postgres"
	"ticket-katon-backend/internal/service"
)

func main() {
	cfg := config.LoadConfig()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 1. Databases
	pgPool, err := database.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("[FATAL] Postgres init failed: %v", err)
	}
	defer pgPool.Close()

	redisClient, err := database.NewRedisClient(ctx, cfg.RedisURL)
	if err != nil {
		log.Fatalf("[FATAL] Redis init failed: %v", err)
	}
	defer redisClient.Close()

	// 2. Repositories & Token Maker
	tokenMaker := token.NewTokenMaker(cfg.JWTSecret)
	userRepo := postgres.NewPostgresUserRepository(pgPool)
	tripRepo := postgres.NewPostgresTripRepository(pgPool)
	bookingRepo := postgres.NewBookingRepository(pgPool)
	adminRepo := postgres.NewAdminRepository(pgPool)

	// 3. Services & Workers
	healthSvc := service.NewHealthService(pgPool, redisClient)
	authSvc := service.NewAuthService(userRepo, tokenMaker)
	seatLockSvc := service.NewSeatLockService(redisClient)
	tripSvc := service.NewTripService(tripRepo, seatLockSvc)
	bookingSvc := service.NewBookingService(bookingRepo, redisClient)
	adminSvc := service.NewAdminService(adminRepo)

	// Start pending booking expiry worker
	bookingSvc.StartPendingExpiryWorker(ctx)

	// 4. Handlers
	healthHdl := httpHandler.NewHealthHandler(healthSvc)
	authHdl := httpHandler.NewAuthHandler(authSvc)
	tripHdl := httpHandler.NewTripHandler(tripSvc)
	seatLockHdl := httpHandler.NewSeatLockHandler(seatLockSvc)
	bookingHdl := httpHandler.NewBookingHandler(bookingSvc)
	adminTripHdl := httpHandler.NewAdminTripHandler(adminSvc)
	paymentHdl := httpHandler.NewPaymentHandler(bookingSvc)
	adminHdl := httpHandler.NewAdminHandler()

	// 5. Router Setup
	r := chi.NewRouter()
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "http://localhost:5173", "http://127.0.0.1:3000", "http://127.0.0.1:5173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(chiMiddleware.RequestID)
	r.Use(chiMiddleware.RealIP)
	r.Use(chiMiddleware.Logger)
	r.Use(chiMiddleware.Recoverer)
	r.Use(chiMiddleware.Timeout(30 * time.Second))

	// Health Check
	r.Get("/health", healthHdl.HealthCheck)

	// API Routes
	r.Route("/api/v1", func(api chi.Router) {
		// Auth Routes
		api.Route("/auth", func(auth chi.Router) {
			auth.Post("/register", authHdl.Register)
			auth.Post("/login", authHdl.Login)

			auth.Group(func(protected chi.Router) {
				protected.Use(middleware.AuthMiddleware(tokenMaker))
				protected.Get("/profile", authHdl.Profile)
			})
		})

		// Trip & Seat Routes (Public)
		api.Route("/trips", func(trips chi.Router) {
			trips.Get("/search", tripHdl.SearchTrips)
			trips.Get("/{tripID}/seats", tripHdl.GetTripSeats)
		})

		// Seat Lock Routes (Protected with JWT)
		api.Route("/seats", func(seats chi.Router) {
			seats.Use(middleware.AuthMiddleware(tokenMaker))
			seats.Post("/hold", seatLockHdl.Hold)
			seats.Post("/release", seatLockHdl.Release)
		})

		// Booking Routes (Protected with JWT)
		api.Route("/bookings", func(b chi.Router) {
			b.Use(middleware.AuthMiddleware(tokenMaker))
			b.Post("/confirm", bookingHdl.Confirm)
			b.Get("/my", bookingHdl.MyBookings)
			b.Get("/{pnr}", bookingHdl.GetByPNR)
		})

		// Server-Side Payment Routes (Protected with JWT)
		api.Route("/payments", func(p chi.Router) {
			p.Use(middleware.AuthMiddleware(tokenMaker))
			p.Post("/initiate", paymentHdl.Initiate)
			p.Get("/{paymentID}", paymentHdl.Get)
			p.Post("/{paymentID}/simulate", paymentHdl.Simulate)
		})

		// Admin & Counter Sales Routes (Protected with JWT & RBAC)
		api.Route("/admin", func(admin chi.Router) {
			admin.Use(middleware.AuthMiddleware(tokenMaker))

			// Counter Sales & Trip Schedules (Operator Admin & Super Admin)
			admin.Group(func(sales chi.Router) {
				sales.Use(middleware.RequireRole(domain.RoleOperatorAdmin, domain.RoleSuperAdmin))
				sales.Post("/tickets/issue", adminHdl.IssueCounterTicket)

				// Updated Dynamic Admin Trip Routes
				sales.Get("/vehicles", adminTripHdl.ListVehicles)
				sales.Get("/routes", adminTripHdl.ListRoutes)
				sales.Get("/trips", adminTripHdl.ListTrips)
				sales.Post("/trips", adminTripHdl.CreateTrip)
			})

			// Super Admin Only (Platform DevOps / Software Engineers)
			admin.Group(func(super chi.Router) {
				super.Use(middleware.RequireRole(domain.RoleSuperAdmin))
			})
		})
	})

	// 6. Graceful Server
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
			log.Fatalf("[FATAL] Server error: %v", err)
		}
	case sig := <-shutdown:
		log.Printf("[SERVER] Shutting down cleanly on signal %v...", sig)
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Fatalf("[FATAL] Graceful shutdown failed: %v", err)
		}
		log.Println("[SERVER] Server exited successfully.")
	}
}