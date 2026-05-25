package run

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/router"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.AttachmentsBucket))
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AvatarsBucket); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.AvatarsBucket))
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.HeadersBucket); err != nil {
		return err
	}
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.HeadersBucket))

	log.Info("Connected to MinIO successfully")

	attachmentsConn, err := grpc.Dial(cfg.Services.AttachmentsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to attachments service", "error", err)
		return err
	}
	defer attachmentsConn.Close()

	notesConn, err := grpc.Dial(cfg.Services.NotesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to notes service", "error", err)
		return err
	}
	defer notesConn.Close()

	profilesConn, err := grpc.Dial(cfg.Services.ProfilesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to profiles service", "error", err)
		return err
	}
	defer profilesConn.Close()

	log.Info("Connected to gRPC services successfully")

	addr := ":" + cfg.Server.Port
	srvRouter, err := router.New(cfg, database, minioService, attachmentsConn, notesConn, profilesConn)
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
