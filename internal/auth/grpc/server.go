package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	authGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthGrpcServer struct {
	authGrpc.UnimplementedAuthServiceServer
	authUsecase AuthUsecase
}

type AuthUsecase interface {
	SignupUser(ctx context.Context, username, password string) (*models.Profile, error)
	SigninUser(ctx context.Context, username, password string) (*models.Profile, error)
}

type Profile interface {
	GetID() uuid.UUID
	GetUsername() string
}

func NewAuthGrpcServer(authUsecase AuthUsecase) *AuthGrpcServer {
	return &AuthGrpcServer{
		authUsecase: authUsecase,
	}
}

func (s *AuthGrpcServer) SignupUser(ctx context.Context, req *authGrpc.SignupRequest) (*authGrpc.UserResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	profile, err := s.authUsecase.SignupUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		switch err {
		case auth.ErrUserExist:
			return nil, status.Error(codes.AlreadyExists, err.Error())
		case auth.ErrInvalidUsername:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		case auth.ErrInvalidPassword:
			return nil, status.Error(codes.InvalidArgument, err.Error())
		default:
			return nil, status.Error(codes.Internal, auth.ErrInternal.Error())
		}
	}

	return ToProtoUserResponse(profile), nil
}

func (s *AuthGrpcServer) SigninUser(ctx context.Context, req *authGrpc.SigninRequest) (*authGrpc.UserResponse, error) {
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	profile, err := s.authUsecase.SigninUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		switch err {
		case auth.ErrBadCredentials, auth.ErrUserNotExist:
			return nil, status.Error(codes.Unauthenticated, auth.ErrBadCredentials.Error())
		default:
			return nil, status.Error(codes.Internal, auth.ErrInternal.Error())
		}
	}

	return ToProtoUserResponse(profile), nil
}
