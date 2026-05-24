package grpcclient

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/profiles/grpc/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ProfilesServiceClient interface {
	SignupUser(ctx context.Context, username, password string) (profile *profilesgrpc.ProfileResponse, err error)
	SigninUser(ctx context.Context, username string) (profile *profilesgrpc.ProfileResponse, err error)
	Close() error
}

type profilesServiceClient struct {
	client profilesgrpc.ProfileServiceClient
	conn   *grpc.ClientConn
}

func NewProfilesServiceClient(addr string) (ProfilesServiceClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return &profilesServiceClient{
		client: profilesgrpc.NewProfileServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *profilesServiceClient) SignupUser(ctx context.Context, username, password string) (*profilesgrpc.ProfileResponse, error) {
	resp, err := c.client.SignupUser(ctx, &profilesgrpc.SignupUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.AlreadyExists:
				return nil, auth.ErrUserExist
			case codes.InvalidArgument:
				return nil, auth.ErrBadCredentials
			}
		}
		return nil, err
	}

	return resp, nil
}

func (c *profilesServiceClient) SigninUser(ctx context.Context, username string) (*profilesgrpc.ProfileResponse, error) {
	resp, err := c.client.SigninUser(ctx, &profilesgrpc.SigninUserRequest{
		Username: username,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, auth.ErrUserNotExist
			case codes.InvalidArgument:
				return nil, auth.ErrBadCredentials
			}
		}
		return nil, err
	}

	return resp, nil
}

func (c *profilesServiceClient) Close() error {
	return c.conn.Close()
}
