package grpc

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	authGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthGrpcClient клиент для gRPC сервера auth
type AuthGrpcClient struct {
	client authGrpc.AuthServiceClient
	conn   *grpc.ClientConn
}

// NewAuthGrpcClient создает новый gRPC клиент
func NewAuthGrpcClient(addr string, opts ...grpc.DialOption) (*AuthGrpcClient, error) {
	if opts == nil {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &AuthGrpcClient{
		client: authGrpc.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает соединение
func (c *AuthGrpcClient) Close() error {
	return c.conn.Close()
}

// AddUserIDToContext добавляет userID в метаданные gRPC
// (для методов, где нужна авторизация)
func (c *AuthGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	return metadata.NewOutgoingContext(ctx, md)
}

// SignupUser регистрация пользователя
func (c *AuthGrpcClient) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	req := &authGrpc.SignupRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.client.SignupUser(ctx, req)
	if err != nil {
		// Конвертация gRPC ошибок в бизнес-ошибки
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, auth.ErrUserExist
			case codes.InvalidArgument:
				if st.Message() == "username is required" || st.Message() == "Невалидное имя пользователя" {
					return nil, auth.ErrInvalidUsername
				}
				if st.Message() == "password is required" || st.Message() == "Невалидный пароль" {
					return nil, auth.ErrInvalidPassword
				}
				return nil, auth.ErrInvalidInput
			case codes.Internal:
				return nil, auth.ErrInternal
			}
		}
		return nil, err
	}

	return FromProtoUserResponse(resp), nil
}

// SigninUser вход пользователя
func (c *AuthGrpcClient) SigninUser(ctx context.Context, username, password string) (*models.Profile, error) {
	req := &authGrpc.SigninRequest{
		Username: username,
		Password: password,
	}

	resp, err := c.client.SigninUser(ctx, req)
	if err != nil {
		// Конвертация gRPC ошибок в бизнес-ошибки
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Unauthenticated:
				return nil, auth.ErrBadCredentials
			case codes.InvalidArgument:
				return nil, auth.ErrInvalidInput
			case codes.Internal:
				return nil, auth.ErrInternal
			}
		}
		return nil, err
	}

	return FromProtoUserResponse(resp), nil
}

// LogoutUser выход пользователя
func (c *AuthGrpcClient) LogoutUser(ctx context.Context, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.LogoutUser(ctxWithUserID, &authGrpc.LogoutRequest{})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.Internal:
				return auth.ErrInternal
			}
		}
		return err
	}

	return nil
}
