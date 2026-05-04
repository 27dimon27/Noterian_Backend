package grpc

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	profilesGrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/grpc/gen"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ProfileGrpcServer struct {
	profilesGrpc.UnimplementedProfileServiceServer
	profileUsecase ProfileUsecase
}

type ProfileUsecase interface {
	GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	UpdateProfile(ctx context.Context, userID uuid.UUID, username string) (*models.Profile, error)
	DeleteProfile(ctx context.Context, userID uuid.UUID) error
	GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error)
	UploadAvatar(ctx context.Context, profileID uuid.UUID, fileReader io.Reader, fileName string, fileSize int64, mimeType string) (*models.Avatar, error)
	DeleteAvatar(ctx context.Context, profileID uuid.UUID) error
	ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error)
}

func NewProfileGrpcServer(profileUsecase ProfileUsecase) *ProfileGrpcServer {
	return &ProfileGrpcServer{
		profileUsecase: profileUsecase,
	}
}

func getUserIDFromContext(ctx context.Context) (uuid.UUID, error) {
	userID, ok := ctx.Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		return uuid.Nil, status.Error(codes.Unauthenticated, profiles.ErrInvalidUserID.Error())
	}
	return userID, nil
}

func (s *ProfileGrpcServer) GetProfile(ctx context.Context, req *profilesGrpc.GetProfileRequest) (*profilesGrpc.Profile, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	profile, err := s.profileUsecase.GetProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoProfile(profile), nil
}

func (s *ProfileGrpcServer) UpdateProfile(ctx context.Context, req *profilesGrpc.UpdateProfileRequest) (*profilesGrpc.Profile, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedProfile, err := s.profileUsecase.UpdateProfile(ctx, userID, req.GetUsername())
	if err != nil {
		if errors.Is(err, profiles.ErrInvalidProfileData) {
			return nil, status.Error(codes.InvalidArgument, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoProfile(updatedProfile), nil
}

func (s *ProfileGrpcServer) DeleteProfile(ctx context.Context, req *profilesGrpc.DeleteProfileRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.profileUsecase.DeleteProfile(ctx, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}

func (s *ProfileGrpcServer) ChangePassword(ctx context.Context, req *profilesGrpc.ChangePasswordRequest) (*profilesGrpc.Profile, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	updatedProfile, err := s.profileUsecase.ChangePassword(ctx, userID, req.GetOldPassword(), req.GetNewPassword())
	if err != nil {
		if errors.Is(err, profiles.ErrUserNotExist) || errors.Is(err, profiles.ErrWrongPassword) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoProfile(updatedProfile), nil
}

func (s *ProfileGrpcServer) GetAvatar(ctx context.Context, req *profilesGrpc.GetAvatarRequest) (*profilesGrpc.Avatar, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	avatar, err := s.profileUsecase.GetAvatar(ctx, userID)
	if err != nil {
		if errors.Is(err, profiles.ErrAvatarNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return ToProtoAvatar(avatar), nil
}

func (s *ProfileGrpcServer) UploadAvatar(stream profilesGrpc.ProfileService_UploadAvatarServer) error {
	ctx := stream.Context()

	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	var metadata *FileMetadata
	var buffer bytes.Buffer

	for {
		req, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Error(codes.Internal, err.Error())
		}

		switch data := req.Data.(type) {
		case *profilesGrpc.UploadAvatarRequest_Metadata:
			metadata = FromProtoFileMetadata(data.Metadata)
		case *profilesGrpc.UploadAvatarRequest_Chunk:
			if _, err := buffer.Write(data.Chunk); err != nil {
				return status.Error(codes.Internal, err.Error())
			}
		}
	}

	if metadata == nil {
		return status.Error(codes.InvalidArgument, "metadata is required")
	}

	avatar, err := s.profileUsecase.UploadAvatar(
		ctx,
		userID,
		&buffer,
		metadata.FileName,
		metadata.FileSize,
		metadata.MimeType,
	)
	if err != nil {
		if errors.Is(err, profiles.ErrInvalidMimeType) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		if errors.Is(err, profiles.ErrFileTooLarge) {
			return status.Error(codes.ResourceExhausted, err.Error())
		}
		return status.Error(codes.Internal, err.Error())
	}

	return stream.SendAndClose(ToProtoAvatar(avatar))
}

func (s *ProfileGrpcServer) DeleteAvatar(ctx context.Context, req *profilesGrpc.DeleteAvatarRequest) (*emptypb.Empty, error) {
	userID, err := getUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}

	err = s.profileUsecase.DeleteAvatar(ctx, userID)
	if err != nil {
		if errors.Is(err, profiles.ErrAvatarNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &emptypb.Empty{}, nil
}
