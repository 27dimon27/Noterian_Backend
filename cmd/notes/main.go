package main

import (
	"net"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	notesgrpcserver "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpcserver"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
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

	repo := repository.NewNoteRepository(database)
	server := notesgrpcserver.NewServer(repo)

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
