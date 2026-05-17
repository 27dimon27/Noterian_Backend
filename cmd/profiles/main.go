package main

import (
	"net"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	profilesgrpcserver "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/grpc"
	profilesrepository "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository"
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

	repo := profilesrepository.NewProfileRepository(database, minioService, cfg.MinIO.AvatarsBucket)
	profileUsecase, err := profilesUsecase.NewProfileUsecase(repo)
	if err != nil {
		log.Error("Failed to create profile usecase", "error", err)
		return
	}
	server := profilesgrpcserver.NewServer(profileUsecase)

	lis, err := net.Listen("tcp", ":"+cfg.Services.ProfilesPort)
	if err != nil {
		log.Error("Failed to listen", "error", err)
		return
	}

	s := grpc.NewServer()
	profilesgrpc.RegisterProfileServiceServer(s, server)

	log.Info("Starting profiles gRPC server on port " + cfg.Services.ProfilesPort)
	if err := s.Serve(lis); err != nil {
		log.Error("Failed to serve", "error", err)
	}
}
