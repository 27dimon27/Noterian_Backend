package grpc

import (
	"context"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type ProfileUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username string) (*models.Profile, error)
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
		switch {
		case errors.Is(err, profiles.ErrUsernameExists):
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case errors.Is(err, profiles.ErrInvalidProfileData):
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
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
	user, err := s.profileUsecase.SigninUser(ctx, req.GetUsername())
	if err != nil {
		switch {
		case errors.Is(err, profiles.ErrUserNotExist):
			return nil, status.Error(codes.NotFound, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
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
