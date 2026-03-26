package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"

	"github.com/hnam/notafly/internal/config"
	"github.com/hnam/notafly/internal/handler"
	"github.com/hnam/notafly/internal/middleware"
	"github.com/hnam/notafly/internal/service"
	"go.uber.org/zap"
)

var (
	meetLink   string
	duration   int
	audioPath  string
	noAnalysis bool
	port       string
)

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "notafly",
	Short: "Google Meet bot — join, record, transcribe, and analyze meetings",
}

var joinCmd = &cobra.Command{
	Use:   "join",
	Short: "Join a Google Meet, record audio, and analyze",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, logger := mustInit()
		defer logger.Sync()

		if meetLink == "" {
			meetLink = cfg.MeetLink
		}
		if meetLink == "" {
			return fmt.Errorf("--meet-link is required (or set MEET_LINK env)")
		}
		if duration == 0 {
			duration = cfg.RecordingDuration
		}

		meetSvc := service.NewMeetService(cfg, logger)
		recorderSvc := service.NewRecorderService(cfg, logger)
		transcriberSvc := service.NewTranscriberService(cfg, logger)

		ctx := context.Background()

		// Temp dir for audio
		tmpDir, err := os.MkdirTemp("", "notafly-*")
		if err != nil {
			return fmt.Errorf("create temp dir: %w", err)
		}
		outPath := filepath.Join(tmpDir, "output.wav")

		// Join meeting
		fmt.Println("Joining meeting...")
		_, cancelBrowser, err := meetSvc.JoinMeeting(ctx, meetLink)
		if err != nil {
			return fmt.Errorf("join meeting: %w", err)
		}
		defer cancelBrowser()
		fmt.Println("Meeting joined.")

		// Record
		fmt.Printf("Recording for %d seconds...\n", duration)
		if err := recorderSvc.Record(ctx, outPath, duration); err != nil {
			return fmt.Errorf("record: %w", err)
		}
		fmt.Printf("Recording saved: %s\n", outPath)

		// Transcribe + analyze
		if !noAnalysis {
			fmt.Println("Transcribing and analyzing...")
			result, err := transcriberSvc.TranscribeAndAnalyze(ctx, outPath, recorderSvc)
			if err != nil {
				return fmt.Errorf("transcribe: %w", err)
			}
			fmt.Printf("\nAbstract Summary: %s\n", result.Minutes.AbstractSummary)
			fmt.Printf("Key Points: %s\n", result.Minutes.KeyPoints)
			fmt.Printf("Action Items: %s\n", result.Minutes.ActionItems)
			fmt.Printf("Sentiment: %s\n", result.Minutes.Sentiment)
		}

		return nil
	},
}

var transcribeCmd = &cobra.Command{
	Use:   "transcribe",
	Short: "Transcribe an audio file and optionally analyze",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, logger := mustInit()
		defer logger.Sync()

		if audioPath == "" {
			return fmt.Errorf("--audio is required")
		}

		recorderSvc := service.NewRecorderService(cfg, logger)
		transcriberSvc := service.NewTranscriberService(cfg, logger)

		ctx := context.Background()

		if noAnalysis {
			fmt.Println("Transcribing (no analysis)...")
			text, err := transcriberSvc.Transcribe(ctx, audioPath)
			if err != nil {
				return fmt.Errorf("transcribe: %w", err)
			}
			fmt.Printf("\nTranscription:\n%s\n", text)
			return nil
		}

		fmt.Println("Transcribing and analyzing...")
		result, err := transcriberSvc.TranscribeAndAnalyze(ctx, audioPath, recorderSvc)
		if err != nil {
			return fmt.Errorf("transcribe: %w", err)
		}
		fmt.Printf("\nTranscription: %s\n", result.Text)
		fmt.Printf("\nAbstract Summary: %s\n", result.Minutes.AbstractSummary)
		fmt.Printf("Key Points: %s\n", result.Minutes.KeyPoints)
		fmt.Printf("Action Items: %s\n", result.Minutes.ActionItems)
		fmt.Printf("Sentiment: %s\n", result.Minutes.Sentiment)

		return nil
	},
}

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the HTTP API server",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, logger := mustInit()
		defer logger.Sync()

		if port != "" {
			cfg.Port = port
		}

		meetSvc := service.NewMeetService(cfg, logger)
		recorderSvc := service.NewRecorderService(cfg, logger)
		transcriberSvc := service.NewTranscriberService(cfg, logger)
		h := handler.New(cfg, logger, meetSvc, recorderSvc, transcriberSvc)

		r := gin.New()
		r.Use(middleware.JSONRecovery(logger))
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

		srv := &http.Server{
			Addr:    fmt.Sprintf(":%s", cfg.Port),
			Handler: r,
		}

		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()

		go func() {
			logger.Info("server starting", zap.String("port", cfg.Port))
			fmt.Printf("Notafly server running on http://localhost:%s\n", cfg.Port)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				logger.Fatal("server failed", zap.Error(err))
			}
		}()

		<-ctx.Done()
		fmt.Println("\nShutting down...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	},
}

func init() {
	joinCmd.Flags().StringVarP(&meetLink, "meet-link", "m", "", "Google Meet link")
	joinCmd.Flags().IntVarP(&duration, "duration", "d", 0, "Recording duration in seconds (default: from env)")
	joinCmd.Flags().BoolVarP(&noAnalysis, "no-analysis", "n", false, "Skip analysis phase")

	transcribeCmd.Flags().StringVarP(&audioPath, "audio", "a", "", "Path to audio file")
	transcribeCmd.Flags().BoolVarP(&noAnalysis, "no-analysis", "n", false, "Skip analysis phase")

	serveCmd.Flags().StringVarP(&port, "port", "p", "", "Server port (default: from env or 8080)")

	rootCmd.AddCommand(joinCmd, transcribeCmd, serveCmd)
}

func mustInit() (*config.Config, *zap.Logger) {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("config error", zap.Error(err))
	}

	return cfg, logger
}
