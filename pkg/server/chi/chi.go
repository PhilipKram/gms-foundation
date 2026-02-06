package chi

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// ConfigSchema defines the configuration for a Chi-based HTTP server.
type ConfigSchema struct {
	// Port to listen on (e.g., "8080")
	Port string
	// AccessLog enables request logging middleware
	AccessLog bool
	// Production mode affects middleware behavior
	Production bool
	// ReadTimeout for HTTP server
	ReadTimeout time.Duration
	// WriteTimeout for HTTP server
	WriteTimeout time.Duration
	// IdleTimeout for HTTP server
	IdleTimeout time.Duration
}

// Setup creates and configures a new Chi router with standard middleware.
// Returns an HTTP server and the router for further route configuration.
func Setup(serverConfig ConfigSchema) (*http.Server, *chi.Mux) {
	log.Info().Msg("Starting HTTP server on port " + serverConfig.Port)

	router := chi.NewRouter()

	// Add standard middleware
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)

	if serverConfig.AccessLog {
		router.Use(middleware.Logger)
	}

	// Set timeouts with defaults
	readTimeout := serverConfig.ReadTimeout
	if readTimeout == 0 {
		readTimeout = 15 * time.Second
	}

	writeTimeout := serverConfig.WriteTimeout
	if writeTimeout == 0 {
		writeTimeout = 60 * time.Second
	}

	idleTimeout := serverConfig.IdleTimeout
	if idleTimeout == 0 {
		idleTimeout = 120 * time.Second
	}

	srv := &http.Server{
		Addr:         ":" + serverConfig.Port,
		Handler:      router,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}

	return srv, router
}

// Start begins serving HTTP requests and handles graceful shutdown.
// This function blocks until a shutdown signal (SIGINT or SIGTERM) is received.
// The server will attempt to gracefully shutdown with a 30-second timeout.
func Start(srv *http.Server) {
	// Start server in a goroutine
	go func() {
		log.Info().Msgf("Server listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	sig := <-quit

	log.Info().Msgf("Received signal %v, shutting down server...", sig)

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server forced to shutdown")
	}

	log.Info().Msg("Server exiting")
}
