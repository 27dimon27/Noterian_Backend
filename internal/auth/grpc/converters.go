package grpc

import (
	"time"

	authGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

func ToProtoUserResponse(profile *models.Profile) *authGrpc.UserResponse {
	if profile == nil {
		return nil
	}

	return &authGrpc.UserResponse{
		Id:       profile.ID.String(),
		Username: profile.Username,
	}
}

func FromProtoSignupRequest(req *authGrpc.SignupRequest) (username, password string) {
	return req.GetUsername(), req.GetPassword()
}

func FromProtoSigninRequest(req *authGrpc.SigninRequest) (username, password string) {
	return req.GetUsername(), req.GetPassword()
}

func ToProtoProfile(profile *models.Profile) *authGrpc.UserResponse {
	return ToProtoUserResponse(profile)
}

func FromProtoUserResponse(resp *authGrpc.UserResponse) *models.Profile {
	if resp == nil {
		return nil
	}

	id, _ := uuid.Parse(resp.GetId())

	return &models.Profile{
		ID:        id,
		Username:  resp.GetUsername(),
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	}
}
