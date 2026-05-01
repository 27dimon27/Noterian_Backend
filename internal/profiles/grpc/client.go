package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// ProfileGrpcClient клиент для gRPC сервера профилей
type ProfileGrpcClient struct {
	client profilesGrpc.ProfileServiceClient
	conn   *grpc.ClientConn
}

// NewProfileGrpcClient создает новый gRPC клиент
func NewProfileGrpcClient(addr string, opts ...grpc.DialOption) (*ProfileGrpcClient, error) {
	if opts == nil {
		opts = []grpc.DialOption{grpc.WithInsecure()}
	}

	conn, err := grpc.Dial(addr, opts...)
	if err != nil {
		return nil, err
	}

	return &ProfileGrpcClient{
		client: profilesGrpc.NewProfileServiceClient(conn),
		conn:   conn,
	}, nil
}

// Close закрывает соединение
func (c *ProfileGrpcClient) Close() error {
	return c.conn.Close()
}

// addUserIDToContext добавляет userID в метаданные gRPC
func (c *ProfileGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	return metadata.NewOutgoingContext(ctx, md)
}

// ========== PROFILE METHODS ==========

// GetProfile получение профиля
func (c *ProfileGrpcClient) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetProfile(ctxWithUserID, &profilesGrpc.GetProfileRequest{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoProfile(resp), nil
}

// UpdateProfile обновление профиля
func (c *ProfileGrpcClient) UpdateProfile(ctx context.Context, userID uuid.UUID, username string) (*models.Profile, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.UpdateProfile(ctxWithUserID, &profilesGrpc.UpdateProfileRequest{
		Username: username,
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoProfile(resp), nil
}

// DeleteProfile удаление профиля
func (c *ProfileGrpcClient) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteProfile(ctxWithUserID, &profilesGrpc.DeleteProfileRequest{})
	return c.handleError(err)
}

// ChangePassword смена пароля
func (c *ProfileGrpcClient) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.ChangePassword(ctxWithUserID, &profilesGrpc.ChangePasswordRequest{
		OldPassword: oldPassword,
		NewPassword: newPassword,
	})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoProfile(resp), nil
}

// ========== AVATAR METHODS ==========

// GetAvatar получение аватара
func (c *ProfileGrpcClient) GetAvatar(ctx context.Context, userID uuid.UUID) (*models.Avatar, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetAvatar(ctxWithUserID, &profilesGrpc.GetAvatarRequest{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoAvatar(resp), nil
}

// UploadAvatar загрузка аватара (streaming)
func (c *ProfileGrpcClient) UploadAvatar(ctx context.Context, userID uuid.UUID, fileReader io.Reader, fileName string, fileSize int64, mimeType string) (*models.Avatar, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	stream, err := c.client.UploadAvatar(ctxWithUserID)
	if err != nil {
		return nil, err
	}

	// Отправляем метаданные
	err = stream.Send(&profilesGrpc.UploadAvatarRequest{
		Data: &profilesGrpc.UploadAvatarRequest_Metadata{
			Metadata: &profilesGrpc.FileMetadata{
				FileName: fileName,
				FileSize: fileSize,
				MimeType: mimeType,
			},
		},
	})
	if err != nil {
		return nil, err
	}

	// Отправляем файл чанками по 64KB
	buf := make([]byte, 64*1024)
	for {
		n, err := fileReader.Read(buf)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		err = stream.Send(&profilesGrpc.UploadAvatarRequest{
			Data: &profilesGrpc.UploadAvatarRequest_Chunk{
				Chunk: buf[:n],
			},
		})
		if err != nil {
			return nil, err
		}
	}

	resp, err := stream.CloseAndRecv()
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoAvatar(resp), nil
}

// DeleteAvatar удаление аватара
func (c *ProfileGrpcClient) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteAvatar(ctxWithUserID, &profilesGrpc.DeleteAvatarRequest{})
	return c.handleError(err)
}

// handleError обрабатывает gRPC ошибки
func (c *ProfileGrpcClient) handleError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return profiles.ErrUserNotExist
	case codes.InvalidArgument:
		return profiles.ErrInvalidProfileData
	case codes.ResourceExhausted:
		return profiles.ErrFileTooLarge
	case codes.Unauthenticated:
		return profiles.ErrInvalidUserID
	default:
		return err
	}
}
