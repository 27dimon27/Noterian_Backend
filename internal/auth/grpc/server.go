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

// AuthGrpcServer реализует gRPC сервер для auth сервиса
type AuthGrpcServer struct {
	authGrpc.UnimplementedAuthServiceServer
	authUsecase AuthUsecase
}

// AuthUsecase интерфейс бизнес-логики для gRPC сервера
type AuthUsecase interface {
	CreateUser(ctx context.Context, username, password string) (*models.Profile, error)
	ValidateUser(ctx context.Context, username, password string) (*models.Profile, error)
}

// Profile интерфейс для профиля (чтобы не зависеть от models)
type Profile interface {
	GetID() uuid.UUID
	GetUsername() string
}

// NewAuthGrpcServer создает новый gRPC сервер
func NewAuthGrpcServer(authUsecase AuthUsecase) *AuthGrpcServer {
	return &AuthGrpcServer{
		authUsecase: authUsecase,
	}
}

// SignupUser регистрация нового пользователя
func (s *AuthGrpcServer) SignupUser(ctx context.Context, req *authGrpc.SignupRequest) (*authGrpc.UserResponse, error) {
	// Валидация входных данных
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Вызов usecase
	profile, err := s.authUsecase.CreateUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		// Конвертация ошибок usecase в gRPC статусы
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

	// Конвертация ответа
	return ToProtoUserResponse(profile), nil
}

// SigninUser вход пользователя
func (s *AuthGrpcServer) SigninUser(ctx context.Context, req *authGrpc.SigninRequest) (*authGrpc.UserResponse, error) {
	// Валидация входных данных
	if req.GetUsername() == "" {
		return nil, status.Error(codes.InvalidArgument, "username is required")
	}
	if req.GetPassword() == "" {
		return nil, status.Error(codes.InvalidArgument, "password is required")
	}

	// Вызов usecase
	profile, err := s.authUsecase.ValidateUser(ctx, req.GetUsername(), req.GetPassword())
	if err != nil {
		// Конвертация ошибок usecase в gRPC статусы
		switch err {
		case auth.ErrBadCredentials, auth.ErrUserNotExist:
			return nil, status.Error(codes.Unauthenticated, auth.ErrBadCredentials.Error())
		default:
			return nil, status.Error(codes.Internal, auth.ErrInternal.Error())
		}
	}

	// Конвертация ответа
	return ToProtoUserResponse(profile), nil
}

// LogoutUser выход пользователя
func (s *AuthGrpcServer) LogoutUser(ctx context.Context, req *authGrpc.LogoutRequest) (*authGrpc.LogoutResponse, error) {
	// В auth нет бизнес-логики для logout, просто возвращаем успех
	// Очисткой cookie занимается HTTP handler
	return &authGrpc.LogoutResponse{}, nil
}
