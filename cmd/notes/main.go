package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/grpcserver"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	notesGrpcServer "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc"
	notesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpc/gen"
	notesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	notesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/db"
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

	noteRepo := notesRepo.NewNoteRepository(database)
	noteUsecase := notesUsecase.NewNoteUsecase(noteRepo)

	noteServer := notesGrpcServer.NewNoteGrpcServer(noteUsecase)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.UnaryAuthInterceptor(cfg.JWT.Secret)),
		grpc.StreamInterceptor(grpcserver.StreamAuthInterceptor(cfg.JWT.Secret)),
	)
	notesGrpc.RegisterNoteServiceServer(grpcServer, noteServer)

	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to listen on notes port", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdownCancel()

	go func() {
		log.Info("notes service started", "address", address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("notes grpc server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	log.Info("shutting down notes service")
	grpcServer.GracefulStop()
	log.Info("notes service stopped")
	time.Sleep(100 * time.Millisecond)
}
