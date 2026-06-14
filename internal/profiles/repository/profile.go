package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:generate mockgen -source=profile.go -destination=mocks/mock_repository_profile.go -package=mocks

type MinIOService interface {
	UploadFile(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, bucketName, key string) error
	GeneratePresignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}

type profileRepository struct {
	db           *sql.DB
	minio        MinIOService
	avatarBucket string
	logger       *slog.Logger
}

func NewProfileRepository(db *sql.DB, minio MinIOService, avatarBucket string, logger *slog.Logger) *profileRepository {
	return &profileRepository{
		db:           db,
		minio:        minio,
		avatarBucket: avatarBucket,
		logger:       logger,
	}
}

func (r *profileRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, types.AppErrorInterface) {
	user := &models.Profile{}

	if err := r.db.QueryRowContext(ctx, GET_PROFILE_BY_USER_ID, userID).Scan(
		&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetProfile",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetProfile",
		}
	}

	return user, nil
}

func (r *profileRepository) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	user := &models.Profile{}

	if err := r.db.QueryRowContext(ctx, GET_PROFILE_BY_USERNAME, username).Scan(
		&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetProfileByUsername",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetProfileByUsername",
		}
	}

	return user, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, types.AppErrorInterface) {
	updatedProfile := &models.Profile{}

	if err := r.db.QueryRowContext(ctx, UPDATE_PROFILE_BY_USER_ID, userID, profile.Username).Scan(
		&updatedProfile.ID, &updatedProfile.Username, &updatedProfile.CreatedAt, &updatedProfile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "UpdateProfile",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateProfile",
		}
	}

	return updatedProfile, nil
}

func (r *profileRepository) DeleteProfile(ctx context.Context, userID uuid.UUID) types.AppErrorInterface {
	var id uuid.UUID

	if err := r.db.QueryRowContext(ctx, DELETE_PROFILE_BY_USER_ID, userID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteProfile",
			}
		}
		return &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteProfile",
		}
	}

	return nil
}

func (r *profileRepository) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, types.AppErrorInterface) {
	avatar := &models.Avatar{}

	if err := r.db.QueryRowContext(ctx, GET_AVATAR_BY_PROFILE_ID, profileID).Scan(
		&avatar.ID,
		&avatar.ProfileID,
		&avatar.MinioKey,
		&avatar.AvatarURL,
		&avatar.URLExpiresAt,
		&avatar.CreatedAt,
		&avatar.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrAvatarNotFound,
				PublicMsg:  profiles.PublicMsgErrAvatarNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetAvatar",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetAvatar",
		}
	}

	if time.Now().After(avatar.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.avatarBucket, avatar.MinioKey, profiles.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  profiles.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetAvatar",
			}
		}

		newExpiry := time.Now().Add(profiles.PRESIGNED_URL_EXPIRY)

		if err = r.updateAvatarURL(ctx, avatar.ID, newURL, newExpiry); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  profiles.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetAvatar",
			}
		}

		avatar.AvatarURL = newURL
		avatar.URLExpiresAt = newExpiry
		avatar.UpdatedAt = time.Now()
	}

	return avatar, nil
}

func (r *profileRepository) UploadAvatar(
	ctx context.Context,
	profileID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Avatar, types.AppErrorInterface) {
	if err := r.DeleteAvatar(ctx, profileID); err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAvatar",
		}
	}

	avatarID := uuid.New()
	minioKey := avatarID.String()

	if err := r.minio.UploadFile(ctx, r.avatarBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAvatar",
		}
	}

	presignedURL, generateErr := r.minio.GeneratePresignedURL(ctx, r.avatarBucket, minioKey, profiles.PRESIGNED_URL_EXPIRY)
	if generateErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.avatarBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", generateErr, delErr),
				PublicMsg:  profiles.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadAvatar",
			}
		}
		return nil, &types.AppError{
			Err:        generateErr,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAvatar",
		}
	}

	now := time.Now()
	avatar := &models.Avatar{
		ID:           avatarID,
		ProfileID:    profileID,
		MinioKey:     minioKey,
		AvatarURL:    presignedURL,
		URLExpiresAt: now.Add(profiles.PRESIGNED_URL_EXPIRY),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if queryErr := r.db.QueryRowContext(
		ctx,
		CREATE_AVATAR,
		avatar.ID,
		avatar.ProfileID,
		avatar.MinioKey,
		avatar.AvatarURL,
		avatar.URLExpiresAt,
	).Scan(
		&avatar.ID,
		&avatar.ProfileID,
		&avatar.MinioKey,
		&avatar.AvatarURL,
		&avatar.URLExpiresAt,
		&avatar.CreatedAt,
		&avatar.UpdatedAt,
	); queryErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.avatarBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", queryErr, delErr),
				PublicMsg:  profiles.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadAvatar",
			}
		}
		return nil, &types.AppError{
			Err:        queryErr,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAvatar",
		}
	}

	return avatar, nil
}

func (r *profileRepository) DeleteAvatar(ctx context.Context, profileID uuid.UUID) types.AppErrorInterface {
	var minioKey string

	if err := r.db.QueryRowContext(ctx, DELETE_AVATAR_BY_ID, profileID).Scan(&minioKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.AppError{
				Err:        profiles.ErrAvatarNotFound,
				PublicMsg:  profiles.PublicMsgErrAvatarNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteAvatar",
			}
		}
		return &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteAvatar",
		}
	}

	if err := r.minio.DeleteFile(ctx, r.avatarBucket, minioKey); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteAvatar",
		}
	}

	return nil
}

func (r *profileRepository) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, types.AppErrorInterface) {
	updatedProfile := &models.Profile{}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "ChangePassword",
		}
	}

	if err = r.db.QueryRowContext(ctx, CHANGE_PASSWORD_BY_USER_ID, userID, hashPassword).Scan(
		&updatedProfile.ID, &updatedProfile.Username, &updatedProfile.CreatedAt, &updatedProfile.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "ChangePassword",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "ChangePassword",
		}
	}

	return updatedProfile, nil
}

func (r *profileRepository) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, types.AppErrorInterface) {
	var password []byte

	if err := r.db.QueryRowContext(ctx, GET_PASSWORD_BY_USER_ID, userID).Scan(&password); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetPassword",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetPassword",
		}
	}

	return password, nil
}

func (r *profileRepository) SignupUser(ctx context.Context, username, password string) (*models.Profile, types.AppErrorInterface) {
	var exists bool

	if err := r.db.QueryRowContext(ctx, CHECK_USER_EXISTS, username).Scan(&exists); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "SignupUser",
		}
	}

	if exists {
		return nil, &types.AppError{
			Err:        profiles.ErrUsernameExists,
			PublicMsg:  profiles.PublicMsgErrUsernameExists,
			StatusCode: 409,
			Layer:      "repo",
			Op:         "SignupUser",
		}
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "SignupUser",
		}
	}

	user := &models.Profile{
		ID:           uuid.New(),
		Username:     username,
		Password:     hashPassword,
		TokenVersion: 1,
	}

	if _, err = r.db.ExecContext(ctx, CREATE_USER, user.ID, user.Username, user.Password, user.TokenVersion); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "SignupUser",
		}
	}

	return user, nil
}

func (r *profileRepository) SigninUser(ctx context.Context, username string) (*models.Profile, types.AppErrorInterface) {
	user := &models.Profile{}

	if err := r.db.QueryRowContext(ctx, GET_USER_BY_USERNAME, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.TokenVersion, &user.CreatedAt, &user.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        profiles.ErrUserNotExist,
				PublicMsg:  profiles.PublicMsgErrUserNotExist,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "SigninUser",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "SigninUser",
		}
	}

	return user, nil
}

func (r *profileRepository) updateAvatarURL(ctx context.Context, avatarID uuid.UUID, url string, expiresAt time.Time) types.AppErrorInterface {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	if err := r.db.QueryRowContext(ctx, UPDATE_AVATAR_URL, avatarID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  profiles.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "updateAvatarURL",
		}
	}

	return nil
}
