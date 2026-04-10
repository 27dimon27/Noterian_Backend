package run

import (
	"context"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/router"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
)

func Run() error {
	log := logger.Init()

	cfg := config.Load()

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		return err
	}
	defer database.Close()

	log.Info("Connected to database successfully")

	minioService, err := minio.NewMinIOService(cfg.MinIO)
	if err != nil {
		return err
	}

	ctx := context.Background()
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AttachmentsBucket); err != nil {
		return err
	}

	log.Info("Connected to MinIO successfully")

	addr := ":" + cfg.Server.Port
	srvRouter, err := router.New(cfg, database, minioService)
	if err != nil {
		log.Error("Failed to init router", "error", err)
		return err
	}

	srv := &http.Server{
		Handler:      srvRouter,
		Addr:         addr,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	serverErrors := make(chan error, 1)

	go func() {
		log.Info("Server started", "host", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("Server failed", "error", err)
			serverErrors <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Info("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("Server forced to shutdown", "error", err)
			return err
		}

		log.Info("Server stopped gracefully")
		return nil
	case err := <-serverErrors:
		return err
	}
}
