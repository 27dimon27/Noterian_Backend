package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	_ "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/docs"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/router"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
)

// @title WHITECROWSOFT API
// @version 1.0
// @description API for note-taking application

// @host localhost:8000
// @BasePath /

// @securityDefinitions.apikey ApiKeyAuth
// @in cookie
// @name token
// @description JWT token stored in cookie

// @securitydefinitions.apikey CsrfToken
// @in header
// @name X-CSRF-Token

// @accept json
// @produce json

func main() {
	logg := logger.Init()
	cfg := config.Load()

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		logg.Error("failed to connect to database", "error", err)
		log.Fatalf("Error to run application: %v", err)
	}
	defer database.Close()

	logg.Info("connected to database successfully")

	minioService, err := minio.NewMinIOService(cfg.MinIO)
	if err != nil {
		logg.Error("failed to initialize minio", "error", err)
		log.Fatalf("Error to run application: %v", err)
	}

	ctx := context.Background()
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AttachmentsBucket); err != nil {
		logg.Error("failed to create attachments bucket", "error", err)
		log.Fatalf("Error to run application: %v", err)
	}
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AvatarsBucket); err != nil {
		logg.Error("failed to create avatars bucket", "error", err)
		log.Fatalf("Error to run application: %v", err)
	}

	logg.Info("connected to S3 successfully")

	srvRouter, err := router.New(cfg, database, minioService)
	if err != nil {
		logg.Error("failed to init router", "error", err)
		log.Fatalf("Error to run application: %v", err)
	}

	srv := &http.Server{
		Handler:      srvRouter,
		Addr:         ":" + cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	shutdownCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		logg.Info("server started", "host", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logg.Error("server failed", "error", err)
			serverErrors <- err
		}
	}()

	select {
	case <-shutdownCtx.Done():
		logg.Info("shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logg.Error("server forced to shutdown", "error", err)
			log.Fatalf("Error to run application: %v", err)
		}

		logg.Info("server stopped gracefully")
	case err := <-serverErrors:
		log.Fatalf("Error to run application: %v", err)
	}
}
