package server

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/signal"
	"reflect"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Config defines the configuration for a Gin-based HTTP server.
type Config struct {
	// Port to listen on (e.g., "8080")
	Port string
	// AccessLog enables Gin's built-in request logging middleware
	AccessLog bool `yaml:"accessLog"`
	// Production enables Gin's release mode
	Production bool
	// ReadTimeout for the HTTP server (default 15s)
	ReadTimeout time.Duration
	// WriteTimeout for the HTTP server (default 60s)
	WriteTimeout time.Duration
	// IdleTimeout for the HTTP server (default 120s)
	IdleTimeout time.Duration
}

// Define a buffer pool for efficient buffer reuse
var bufferPool = &sync.Pool{
	New: func() any {
		return new(bytes.Buffer)
	},
}

// Setup creates and configures a new Gin router with standard middleware.
// Returns an HTTP server and the router for further route configuration.
func Setup(serverConfig Config) (*http.Server, *gin.Engine) {

	log.Info().Msg("Starting HTTP server on port " + serverConfig.Port)

	if serverConfig.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if serverConfig.AccessLog {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

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
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// The context is used to inform the server it has 30 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %s", err)
	}

	log.Info().Msg("Server exiting")
}

// HandleRequestBody reads the request body and unmarshals it into out based on
// contentType. Supported content types are "application/json" and
// "application/x-protobuf". The out parameter must be a non-nil pointer; for
// protobuf content types it must also implement proto.Message.
func HandleRequestBody(c *gin.Context, contentType string, out any) error {

	buf, done := requestBodyBuffer(c)
	if done {
		return fmt.Errorf("failed to read request body")
	}
	defer bufferPool.Put(buf)

	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("out must be a non-nil pointer")
	}

	switch contentType {
	case "application/json":
		msg, ok := out.(proto.Message)
		if !ok {
			log.Error().Msg("out does not implement proto.Message")
			c.Status(http.StatusBadRequest)
			return fmt.Errorf("out does not implement proto.Message")
		}
		unmarshaler := protojson.UnmarshalOptions{}
		if err := unmarshaler.Unmarshal(buf.Bytes(), msg); err != nil {
			log.Error().Err(err).Msg("Failed to decode JSON")
			c.Status(http.StatusBadRequest)
			return err
		}
	case "application/x-protobuf":
		msg, ok := out.(proto.Message)
		if !ok {
			log.Error().Msg("out does not implement proto.Message")
			c.Status(http.StatusBadRequest)
			return fmt.Errorf("out does not implement proto.Message")
		}
		if err := proto.Unmarshal(buf.Bytes(), msg); err != nil {
			log.Error().Err(err).Msg("Failed to decode Proto")
			c.Status(http.StatusBadRequest)
			return err
		}
	default:
		log.Error().Msg("Unsupported Content-Type")
		c.Status(http.StatusUnsupportedMediaType)
		return fmt.Errorf("unsupported Content-Type")
	}
	return nil
}

func requestBodyBuffer(c *gin.Context) (*bytes.Buffer, bool) {
	buf := bufferPool.Get().(*bytes.Buffer)
	buf.Reset()

	if _, err := io.Copy(buf, c.Request.Body); err != nil {
		bufferPool.Put(buf)
		log.Error().Err(err).Msg("Failed to read request body")
		c.Status(http.StatusInternalServerError)
		return nil, true
	}
	return buf, false
}
