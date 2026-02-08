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

type ConfigSchema struct {
	Port       string
	AccessLog  bool `yaml:"accessLog"`
	Production bool
}

// Define a buffer pool for efficient buffer reuse
var bufferPool = &sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func Setup(serverConfig ConfigSchema) (*http.Server, *gin.Engine) {

	log.Info().Msg("Starting HTTP server on port " + serverConfig.Port)

	if serverConfig.Production {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()
	if serverConfig.AccessLog {
		router.Use(gin.Logger())
	}
	router.Use(gin.Recovery())

	srv := &http.Server{
		Addr:    ":" + serverConfig.Port,
		Handler: router,
	}

	return srv, router
}

func Start(srv *http.Server) {
	// Initializing the server in a goroutine so that
	// it won't block the graceful shutdown handling below
	go func() {
		_ = srv.ListenAndServe()
	}()

	// Wait for interrupt signal to gracefully shutdown the server with
	// a timeout of 5 seconds.
	quit := make(chan os.Signal, 1)
	// kill (no param) default send syscall.SIGTERM
	// kill -2 is syscall.SIGINT
	// kill -9 is syscall.SIGKILL but can't be catch, so don't need add it
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Info().Msg("Shutting down server...")

	// The context is used to inform the server it has 5 seconds to finish
	// the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal().Msgf("Server forced to shutdown: %s", err)
	}

	log.Info().Msg("Server exiting")
}

func HandleRequestBody(c *gin.Context, contentType string, out interface{}) error {

	buf, done := requestBodyBuffer(c)
	if done {
		return fmt.Errorf("failed to read request body")
	}

	val := reflect.ValueOf(out)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("out must be a non-nil pointer")
	}

	switch contentType {
	case "application/json":
		unmarshaler := protojson.UnmarshalOptions{}
		if err := unmarshaler.Unmarshal(buf.Bytes(), out.(proto.Message)); err != nil {
			log.Error().Err(err).Msg("Failed to decode JSON")
			c.Status(http.StatusBadRequest)
			return err
		}
	case "application/x-protobuf":
		if err := proto.Unmarshal(buf.Bytes(), out.(proto.Message)); err != nil {
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
	defer bufferPool.Put(buf)

	if _, err := io.Copy(buf, c.Request.Body); err != nil {
		log.Error().Err(err).Msg("Failed to read request body")
		c.Status(http.StatusInternalServerError)
		return nil, true
	}
	return buf, false
}
