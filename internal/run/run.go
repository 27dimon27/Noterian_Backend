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
	defer func() {
		if err := database.Close(); err != nil {
			log.Error("failed to close database connection in main", "error", err)
		}
	}()

	log.Info("Connected to database successfully")

	minioService, err := minio.NewMinIOService(cfg.MinIO)
	if err != nil {
		log.Error("Failed to create MinIO service", "error", err)
		return err
	}

	ctx := context.Background()
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AttachmentsBucket); err != nil {
		log.Error("Failed to create attachments bucket", "error", err)
		return err
	}
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.AttachmentsBucket))
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AvatarsBucket); err != nil {
		log.Error("Failed to create avatars bucket", "error", err)
		return err
	}
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.AvatarsBucket))
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.HeadersBucket); err != nil {
		log.Error("Failed to create headers bucket", "error", err)
		return err
	}
	log.Info(fmt.Sprintf("Created %s bucket successfully", cfg.MinIO.HeadersBucket))

	log.Info("Connected to MinIO successfully")

	attachmentsConn, err := grpc.NewClient(cfg.Services.AttachmentsAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to attachments service", "error", err)
		return err
	}
	defer func() {
		if err := attachmentsConn.Close(); err != nil {
			log.Error("failed to close attachments gRPC connection", "error", err)
		}
	}()

	notesConn, err := grpc.NewClient(cfg.Services.NotesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to notes service", "error", err)
		return err
	}
	defer func() {
		if err := notesConn.Close(); err != nil {
			log.Error("failed to close notes gRPC connection", "error", err)
		}
	}()

	profilesConn, err := grpc.NewClient(cfg.Services.ProfilesAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Error("Failed to connect to profiles service", "error", err)
		return err
	}
	defer func() {
		if err := profilesConn.Close(); err != nil {
			log.Error("failed to close profiles gRPC connection", "error", err)
		}
	}()

	log.Info("Connected to gRPC services successfully")

	addr := ":" + cfg.Server.Port
	srvRouter, err := router.New(cfg, log, database, minioService, attachmentsConn, notesConn, profilesConn)
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
