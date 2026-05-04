package grpc

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type ProfileGrpcClient struct {
	client profilesGrpc.ProfileServiceClient
	conn   *grpc.ClientConn
}

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

// func (c *ProfileGrpcClient) Close() error {
// 	return c.conn.Close()
// }

func (c *ProfileGrpcClient) addUserIDToContext(ctx context.Context, userID uuid.UUID) context.Context {
	md := metadata.Pairs("user-id", userID.String())
	if token, ok := ctx.Value(types.JWTTokenKey).(string); ok && token != "" {
		md = metadata.Join(md, metadata.Pairs("authorization", "Bearer "+token, "token", token))
	}
	if existing, ok := metadata.FromOutgoingContext(ctx); ok {
		md = metadata.Join(existing, md)
	}
	return metadata.NewOutgoingContext(ctx, md)
}

func (c *ProfileGrpcClient) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetProfile(ctxWithUserID, &profilesGrpc.GetProfileRequest{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoProfile(resp), nil
}

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

func (c *ProfileGrpcClient) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteProfile(ctxWithUserID, &profilesGrpc.DeleteProfileRequest{})
	return c.handleError(err)
}

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

func (c *ProfileGrpcClient) GetAvatar(ctx context.Context, userID uuid.UUID) (*models.Avatar, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	resp, err := c.client.GetAvatar(ctxWithUserID, &profilesGrpc.GetAvatarRequest{})
	if err != nil {
		return nil, c.handleError(err)
	}

	return FromProtoAvatar(resp), nil
}

func (c *ProfileGrpcClient) UploadAvatar(ctx context.Context, userID uuid.UUID, fileReader io.Reader, fileName string, fileSize int64, mimeType string) (*models.Avatar, error) {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	stream, err := c.client.UploadAvatar(ctxWithUserID)
	if err != nil {
		return nil, err
	}

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

func (c *ProfileGrpcClient) DeleteAvatar(ctx context.Context, userID uuid.UUID) error {
	ctxWithUserID := c.addUserIDToContext(ctx, userID)

	_, err := c.client.DeleteAvatar(ctxWithUserID, &profilesGrpc.DeleteAvatarRequest{})
	return c.handleError(err)
}

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
