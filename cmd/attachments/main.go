package main

import (
	"net"

	attachmentsgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpc/gen"
	attachmentsgrpcserver "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/grpc"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository"
	attachmentsUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	notesrepository "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/minio"
	"google.golang.org/grpc"
)

func main() {
	log := logger.Init()

	cfg := config.Load()

	database, err := db.NewPostgresConnection(cfg.DB)
	if err != nil {
		log.Error("Failed to connect to database", "error", err)
		return
	}
	defer database.Close()

	log.Info("Connected to database successfully")

	minioService, err := minio.NewMinIOService(cfg.MinIO)
	if err != nil {
		log.Error("Failed to connect to MinIO", "error", err)
		return
	}

	log.Info("Connected to MinIO successfully")

	repo := repository.NewAttachmentRepository(database, minioService, cfg.MinIO.AttachmentsBucket)
	noteRepo := notesrepository.NewNoteRepository(database)
	attachmentUsecase := attachmentsUsecase.NewAttachmentUsecase(repo, noteRepo)
	server := attachmentsgrpcserver.NewServer(attachmentUsecase)

	lis, err := net.Listen("tcp", ":"+cfg.Services.AttachmentsPort)
	if err != nil {
		log.Error("Failed to listen", "error", err)
		return
	}

	s := grpc.NewServer()
	attachmentsgrpc.RegisterAttachmentServiceServer(s, server)

	log.Info("Starting attachments gRPC server on port " + cfg.Services.AttachmentsPort)
	if err := s.Serve(lis); err != nil {
		log.Error("Failed to serve", "error", err)
	}
}
