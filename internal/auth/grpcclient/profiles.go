package grpcclient

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	profilesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/google/uuid"
)

type ProfileRepositoryClient struct {
	client profilesgrpc.ProfileServiceClient
}

func NewProfileRepositoryClient(client profilesgrpc.ProfileServiceClient) *ProfileRepositoryClient {
	return &ProfileRepositoryClient{client: client}
}

func (c *ProfileRepositoryClient) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	resp, err := c.client.SignupUser(ctx, &profilesgrpc.SignupUserRequest{
		Username: username,
		Password: password,
	})
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(resp.GetId())
	if err != nil {
		return nil, err
	}

	return &models.Profile{
		ID:        userID,
		Username:  resp.GetUsername(),
		Avatar:    resp.GetAvatarUrl(),
		CreatedAt: resp.GetCreatedAt().AsTime(),
		UpdatedAt: resp.GetUpdatedAt().AsTime(),
	}, nil
}

func (c *ProfileRepositoryClient) SigninUser(ctx context.Context, username string) (*models.Profile, error) {
	resp, err := c.client.SigninUser(ctx, &profilesgrpc.SigninUserRequest{
		Username: username,
	})
	if err != nil {
		return nil, err
	}

	userID, err := uuid.Parse(resp.GetId())
	if err != nil {
		return nil, err
	}

	return &models.Profile{
		ID:        userID,
		Username:  resp.GetUsername(),
		Avatar:    resp.GetAvatarUrl(),
		Password:  []byte(resp.GetPassword()),
		CreatedAt: resp.GetCreatedAt().AsTime(),
		UpdatedAt: resp.GetUpdatedAt().AsTime(),
	}, nil
}
