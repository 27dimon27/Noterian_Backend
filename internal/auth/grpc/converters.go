package grpc

import (
	"time"

	authGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

// ToProtoUserResponse конвертирует models.Profile в protobuf UserResponse
func ToProtoUserResponse(profile *models.Profile) *authGrpc.UserResponse {
	if profile == nil {
		return nil
	}

	return &authGrpc.UserResponse{
		Id:       profile.ID.String(),
		Username: profile.Username,
	}
}

// FromProtoSignupRequest конвертирует protobuf SignupRequest в параметры usecase
func FromProtoSignupRequest(req *authGrpc.SignupRequest) (username, password string) {
	return req.GetUsername(), req.GetPassword()
}

// FromProtoSigninRequest конвертирует protobuf SigninRequest в параметры usecase
func FromProtoSigninRequest(req *authGrpc.SigninRequest) (username, password string) {
	return req.GetUsername(), req.GetPassword()
}

// ToProtoProfile конвертирует models.Profile в protobuf Profile (если понадобится)
func ToProtoProfile(profile *models.Profile) *authGrpc.UserResponse {
	return ToProtoUserResponse(profile)
}

// FromProtoUserResponse конвертирует protobuf UserResponse в models.Profile
func FromProtoUserResponse(resp *authGrpc.UserResponse) *models.Profile {
	if resp == nil {
		return nil
	}

	id, _ := uuid.Parse(resp.GetId())

	return &models.Profile{
		ID:        id,
		Username:  resp.GetUsername(),
		CreatedAt: time.Time{}, // В auth ответе нет created_at
		UpdatedAt: time.Time{},
	}
}
