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
	profilesGrpcServer "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	profilesRepo "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
	profilesUsecase "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"
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
	if err := minioService.CreateBucketIfNotExists(ctx, cfg.MinIO.AvatarsBucket); err != nil {
		log.Error("failed to create avatars bucket", "error", err)
		os.Exit(1)
	}

	log.Info("connected to S3 successfully")

	profileRepo := profilesRepo.NewProfileRepository(database, minioService, cfg.MinIO.AvatarsBucket)
	profileUC, err := profilesUsecase.NewProfileUsecase(profileRepo)
	if err != nil {
		log.Error("failed to initialize profile usecase", "error", err)
		os.Exit(1)
	}

	profileServer := profilesGrpcServer.NewProfileGrpcServer(profileUC)
	grpcServer := grpc.NewServer(
		grpc.UnaryInterceptor(grpcserver.UnaryAuthInterceptor(cfg.JWT.Secret)),
		grpc.StreamInterceptor(grpcserver.StreamAuthInterceptor(cfg.JWT.Secret)),
	)
	profilesGrpc.RegisterProfileServiceServer(grpcServer, profileServer)

	address := fmt.Sprintf(":%s", cfg.Server.Port)
	listener, err := net.Listen("tcp", address)
	if err != nil {
		log.Error("failed to listen on profiles port", "error", err)
		os.Exit(1)
	}

	shutdownCtx, shutdownCancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer shutdownCancel()

	go func() {
		log.Info("profiles service started", "address", address)
		if err := grpcServer.Serve(listener); err != nil {
			log.Error("profiles grpc server failed", "error", err)
			os.Exit(1)
		}
	}()

	<-shutdownCtx.Done()
	log.Info("shutting down profiles service")
	grpcServer.GracefulStop()
	log.Info("profiles service stopped")
	time.Sleep(100 * time.Millisecond)
}
