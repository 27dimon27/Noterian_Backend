package grpcclient

import (
	"context"

	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/profiles/grpc/gen"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

//go:generate mockgen -source=profiles.go -destination=mocks/mock_grpcclient_profiles.go -package=mocks

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
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}

	return &profilesServiceClient{
		client: profilesgrpc.NewProfileServiceClient(conn),
		conn:   conn,
	}, nil
}

func (c *profilesServiceClient) SignupUser(ctx context.Context, username, password string) (*profilesgrpc.ProfileResponse, error) {
	return c.client.SignupUser(ctx, &profilesgrpc.SignupUserRequest{
		Username: username,
		Password: password,
	})
}

func (c *profilesServiceClient) SigninUser(ctx context.Context, username string) (*profilesgrpc.ProfileResponse, error) {
	return c.client.SigninUser(ctx, &profilesgrpc.SigninUserRequest{
		Username: username,
	})
}

func (c *profilesServiceClient) Close() error {
	return c.conn.Close()
}
