package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/handler"
	"github.com/hnam/notafly/internal/middleware"
	"github.com/hnam/notafly/internal/service"
	"go.uber.org/zap"
)

func main() {
	// Logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer logger.Sync()

	// Config
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("failed to load config", zap.Error(err))
	}

	// Services
	meetSvc := service.NewMeetService(cfg, logger)

	// Handler
	h := handler.New(cfg, logger, meetSvc)

	// Router
	router := setupRouter(h, logger)

	// Server with graceful shutdown
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		logger.Info("server starting", zap.String("port", cfg.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("server failed", zap.Error(err))
		}
	}()

	<-ctx.Done()
	logger.Info("shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("server forced shutdown", zap.Error(err))
	}

	logger.Info("server stopped")
}

func setupRouter(h *handler.Handler, logger *zap.Logger) *gin.Engine {
	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.CORS())
	r.Use(middleware.RequestLogger(logger))

	r.GET("/health", h.Health)

	v1 := r.Group("/api/v1")
	{
		meet := v1.Group("/meet")
		{
			meet.POST("/join", h.JoinMeet)
			meet.POST("/full", h.FullPipeline)
		}
		v1.POST("/transcribe", h.Transcribe)
	}

	return r
}
