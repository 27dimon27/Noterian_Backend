package grpcserver

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProfileRepository interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username string) (*models.Profile, error)
}

type Server struct {
	profilesgrpc.UnimplementedProfileServiceServer
	repo ProfileRepository
}

func NewServer(repo ProfileRepository) *Server {
	return &Server{repo: repo}
}

func (s *Server) SignupUser(ctx context.Context, req *profilesgrpc.SignupUserRequest) (*profilesgrpc.ProfileResponse, error) {
	user, err := s.repo.SignupUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &profilesgrpc.ProfileResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		AvatarUrl: user.Avatar,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *Server) SigninUser(ctx context.Context, req *profilesgrpc.SigninUserRequest) (*profilesgrpc.ProfileResponse, error) {
	user, err := s.repo.SigninUser(ctx, req.GetUsername())
	if err != nil {
		return nil, err
	}

	return &profilesgrpc.ProfileResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		AvatarUrl: user.Avatar,
		Password:  string(user.Password),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
