package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pipeline-arch/app/internal/api"
	"github.com/pipeline-arch/app/internal/config"
	"github.com/pipeline-arch/app/pkg/logger"
	"github.com/pipeline-arch/app/pkg/metrics"
	"github.com/rs/zerolog"
)

// @title Pipeline Architecture API
// @version 1.0.0
// @description Production-grade REST API for pipeline orchestration platform
// @host 0.0.0.0:8080
// @BasePath /api/v1

func main() {
	// Initialize configuration
	cfg, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load config: %v", err))
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info().Str("environment", cfg.Environment).Msg("starting application")

	// Initialize metrics
	m := metrics.New("pipeline-arch", cfg.MetricsPort)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize repository (PostgreSQL)
	// repo, err := repository.New(ctx, cfg.DatabaseURL)
	// if err != nil {
	// 	log.Fatal().Err(err).Msg("failed to connect to database")
	// }
	// defer repo.Close()

	// Initialize service layer
	// svc := services.New(repo)

	// Initialize handlers
	handlers := api.NewHandlers(cfg, m, log)

	// Setup router
	router := setupRouter(handlers, log)

	// Initialize metrics server
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", m.Handler())
		srv := &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.MetricsPort),
			Handler: mux,
		}
		log.Info().Int("port", cfg.MetricsPort).Msg("metrics server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error().Err(err).Msg("metrics server failed")
		}
	}()

	// Create main server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler: router,
	}

	// Start server in goroutine
	go func() {
		log.Info().Int("port", cfg.Port).Str("host", cfg.Host).Msg("HTTP server starting")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("HTTP server failed")
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("shutting down servers...")

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Error().Err(err).Msg("HTTP server forced to shutdown")
	}

	log.Info().Msg("servers stopped")
}

func setupRouter(h *api.Handlers, log *zerolog.Logger) *chi.Mux {
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))
	r.Use(middleware.Heartbeat("/healthz"))
	r.Use(middleware.Heartbeat("/readyz"))

	// Logging middleware
	r.Use(httplog.RequestLogger(
		httplog.NewLogger("pipeline-arch", httplog.Options{
			JSON: true,
		}),
	))

	// CORS
	cors := cors.New(cors.Options{
		AllowedOrigins:   []string{"*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	})
	r.Use(cors.Handler)

	// Routes
	r.Get("/", h.Index)
	r.Get("/healthz", h.Healthz)
	r.Get("/readyz", h.Readyz)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// User routes
		r.Route("/users", func(r chi.Router) {
			r.Get("/", h.ListUsers)
			r.Post("/", h.CreateUser)
			r.Get("/{id}", h.GetUser)
			r.Put("/{id}", h.UpdateUser)
			r.Delete("/{id}", h.DeleteUser)
		})

		// Health check with detailed status
		r.Get("/status", h.Status)
	})

	return r
}