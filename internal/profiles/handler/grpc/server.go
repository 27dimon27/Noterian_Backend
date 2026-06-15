package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/profiles/grpc/gen"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProfileUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, types.AppErrorInterface)
	SigninUser(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface)
}

type Server struct {
	profilesgrpc.UnimplementedProfileServiceServer
	profileUsecase ProfileUsecase
}

func NewServer(profileUsecase ProfileUsecase) *Server {
	return &Server{profileUsecase: profileUsecase}
}

func (s *Server) SignupUser(ctx context.Context, req *profilesgrpc.SignupUserRequest) (*profilesgrpc.ProfileResponse, error) {
	user, err := s.profileUsecase.SignupUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err.Unwrap()
	}

	return &profilesgrpc.ProfileResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		Avatar:    user.Avatar,
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}

func (s *Server) SigninUser(ctx context.Context, req *profilesgrpc.SigninUserRequest) (*profilesgrpc.ProfileResponse, error) {
	user, err := s.profileUsecase.SigninUser(ctx, req.GetUsername())
	if err != nil {
		return nil, err.Unwrap()
	}

	return &profilesgrpc.ProfileResponse{
		Id:        user.ID.String(),
		Username:  user.Username,
		Avatar:    user.Avatar,
		Password:  string(user.Password),
		CreatedAt: timestamppb.New(user.CreatedAt),
		UpdatedAt: timestamppb.New(user.UpdatedAt),
	}, nil
}
