package main

import (
	"net"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	attachmentsclient "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpcclient"
	notesgrpcserver "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/grpc"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
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

	attachmentsClient, err := attachmentsclient.NewAttachmentsServiceClient(cfg.Services.AttachmentsAddr)
	if err != nil {
		log.Error("Failed to create attachments service client", "error", err)
		return
	}
	defer attachmentsClient.Close()

	repo := repository.NewNoteRepository(database)
	noteUsecase := notesUsecase.NewNoteUsecase(repo, attachmentsClient)
	server := notesgrpcserver.NewServer(noteUsecase)

	lis, err := net.Listen("tcp", ":"+cfg.Services.NotesPort)
	if err != nil {
		log.Error("Failed to listen", "error", err)
		return
	}

	s := grpc.NewServer()
	notesgrpc.RegisterNoteServiceServer(s, server)

	log.Info("Starting notes gRPC server on port " + cfg.Services.NotesPort)
	if err := s.Serve(lis); err != nil {
		log.Error("Failed to serve", "error", err)
	}
}
