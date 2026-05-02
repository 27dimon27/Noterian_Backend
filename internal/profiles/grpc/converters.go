package grpc

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/google/uuid"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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

type FileMetadata struct {
	FileName string
	FileSize int64
	MimeType string
}

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
