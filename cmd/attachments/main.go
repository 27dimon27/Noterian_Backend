package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	attachmentsGrpcServer "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc"
	attachmentsGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	attachmentsRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository"
	attachmentsUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/grpcserver"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	notesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
	"google.golang.org/grpc"
)

func main() {
	log := logger.Init()
	cfg := config.Load()

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer database.Close()

	log.Info("connected to database successfully")

	minioService, err := minio.NewMinIOService(cfg.MinIO)
	if err != nil {
		log.Error("failed to initialize minio", "error", err)
		os.Exit(1)
	}

	ctx := context.Background()
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AttachmentsBucket); err != nil {
		log.Error("failed to create attachments bucket", "error", err)
		os.Exit(1)
	}

	log.Info("connected to S3 successfully")

	noteRepo := notesRepo.NewNoteRepository(database)
	attachmentRepo := attachmentsRepo.NewAttachmentRepository(database, minioService, cfg.MinIO.AttachmentsBucket)
	attachmentUsecase := attachmentsUsecase.NewAttachmentUsecase(attachmentRepo, noteRepo)

	attachmentServer := attachmentsGrpcServer.NewAttachmentGrpcServer(attachmentUsecase)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.UnaryAuthInterceptor(cfg.JWT.Secret)),
		grpc.StreamInterceptor(grpcserver.StreamAuthInterceptor(cfg.JWT.Secret)),
	)
	attachmentsGrpc.RegisterAttachmentServiceServer(grpcServer, attachmentServer)

	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to listen on attachments port", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdownCancel()

	go func() {
		log.Info("attachments service started", "address", address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("attachments grpc server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	log.Info("shutting down attachments service")
	grpcServer.GracefulStop()
	log.Info("attachments service stopped")
	time.Sleep(100 * time.Millisecond)
}
