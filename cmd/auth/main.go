package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	authGrpcServer "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc"
	authGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc/gen"
	authRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/repository"
	authUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
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

	userRepo := authRepo.NewUserRepository(database)
	authUC, err := authUsecase.NewAuthUsecase(userRepo, cfg.JWT)
	if err != nil {
		log.Error("failed to initialize auth usecase", "error", err)
		os.Exit(1)
	}

	authServer := authGrpcServer.NewAuthGrpcServer(authUC)
	grpcServer := grpc.NewServer()
	authGrpc.RegisterAuthServiceServer(grpcServer, authServer)

	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to listen on auth port", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdownCancel()

	go func() {
		log.Info("auth service started", "address", address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("auth grpc server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	log.Info("shutting down auth service")
	grpcServer.GracefulStop()
	log.Info("auth service stopped")
	time.Sleep(100 * time.Millisecond)
}
