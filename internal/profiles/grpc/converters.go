package grpc

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ========== PROFILE CONVERTERS ==========

// ToProtoProfile конвертирует models.Profile в protobuf Profile
func ToProtoProfile(profile *models.Profile) *profilesGrpc.Profile {
	if profile == nil {
		return nil
	}

	return &profilesGrpc.Profile{
		Id:        profile.ID.String(),
		Username:  profile.Username,
		CreatedAt: timestamppb.New(profile.CreatedAt),
		UpdatedAt: timestamppb.New(profile.UpdatedAt),
	}
}

// FromProtoProfile конвертирует protobuf Profile в models.Profile
func FromProtoProfile(proto *profilesGrpc.Profile) *models.Profile {
	if proto == nil {
		return nil
	}

	return &models.Profile{
		ID:        uuid.MustParse(proto.GetId()),
		Username:  proto.GetUsername(),
		CreatedAt: proto.GetCreatedAt().AsTime(),
		UpdatedAt: proto.GetUpdatedAt().AsTime(),
	}
}

// ========== AVATAR CONVERTERS ==========

// ToProtoAvatar конвертирует models.Avatar в protobuf Avatar
func ToProtoAvatar(avatar *models.Avatar) *profilesGrpc.Avatar {
	if avatar == nil {
		return nil
	}

	return &profilesGrpc.Avatar{
		Id:           avatar.ID.String(),
		ProfileId:    avatar.ProfileID.String(),
		MinioKey:     avatar.MinioKey,
		AvatarUrl:    avatar.AvatarURL,
		UrlExpiresAt: timestamppb.New(avatar.URLExpiresAt),
		CreatedAt:    timestamppb.New(avatar.CreatedAt),
		UpdatedAt:    timestamppb.New(avatar.UpdatedAt),
	}
}

// FromProtoAvatar конвертирует protobuf Avatar в models.Avatar
func FromProtoAvatar(proto *profilesGrpc.Avatar) *models.Avatar {
	if proto == nil {
		return nil
	}

	return &models.Avatar{
		ID:           uuid.MustParse(proto.GetId()),
		ProfileID:    uuid.MustParse(proto.GetProfileId()),
		MinioKey:     proto.GetMinioKey(),
		AvatarURL:    proto.GetAvatarUrl(),
		URLExpiresAt: proto.GetUrlExpiresAt().AsTime(),
		CreatedAt:    proto.GetCreatedAt().AsTime(),
		UpdatedAt:    proto.GetUpdatedAt().AsTime(),
	}
}

// ========== FILE METADATA CONVERTERS ==========

// FileMetadata структура для метаданных файла
type FileMetadata struct {
	FileName string
	FileSize int64
	MimeType string
}

// ToProtoFileMetadata конвертирует FileMetadata в protobuf FileMetadata
func ToProtoFileMetadata(metadata *FileMetadata) *profilesGrpc.FileMetadata {
	if metadata == nil {
		return nil
	}

	return &profilesGrpc.FileMetadata{
		FileName: metadata.FileName,
		FileSize: metadata.FileSize,
		MimeType: metadata.MimeType,
	}
}

// FromProtoFileMetadata конвертирует protobuf FileMetadata в FileMetadata
func FromProtoFileMetadata(proto *profilesGrpc.FileMetadata) *FileMetadata {
	if proto == nil {
		return nil
	}

	return &FileMetadata{
		FileName: proto.GetFileName(),
		FileSize: proto.GetFileSize(),
		MimeType: proto.GetMimeType(),
	}
}
