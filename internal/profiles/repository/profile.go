package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
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
}

func NewProfileRepository(db *sql.DB, minio MinIOService, avatarBucket string) *profileRepository {
	return &profileRepository{
		db:           db,
		minio:        minio,
		avatarBucket: avatarBucket,
	}
}

func (r *profileRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	user := &models.Profile{}

	err := r.db.QueryRowContext(ctx, GET_PROFILE_BY_USER_ID, userID).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}

func (r *profileRepository) GetProfileByUsername(ctx context.Context, username string) (*models.Profile, error) {
	user := &models.Profile{}

	err := r.db.QueryRowContext(ctx, GET_PROFILE_BY_USERNAME, username).Scan(&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}

func (r *profileRepository) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	updatedProfile := &models.Profile{}

	err := r.db.QueryRowContext(ctx, UPDATE_PROFILE_BY_USER_ID, userID, profile.Username).Scan(&updatedProfile.ID, &updatedProfile.Username, &updatedProfile.CreatedAt, &updatedProfile.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return updatedProfile, nil
}

func (r *profileRepository) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, DELETE_PROFILE_BY_USER_ID, userID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return profiles.ErrUserNotExist
		}
		return err
	}

	return nil
}

func (r *profileRepository) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error) {
	avatar := &models.Avatar{}

	err := r.db.QueryRowContext(ctx, GET_AVATAR_BY_PROFILE_ID, profileID).Scan(
		&avatar.ID,
		&avatar.ProfileID,
		&avatar.MinioKey,
		&avatar.AvatarURL,
		&avatar.URLExpiresAt,
		&avatar.CreatedAt,
		&avatar.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrAvatarNotFound
		}
		return nil, err
	}

	if time.Now().After(avatar.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.avatarBucket, avatar.MinioKey, profiles.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, err
		}

		newExpiry := time.Now().Add(profiles.PRESIGNED_URL_EXPIRY)

		err = r.updateAvatarURL(ctx, avatar.ID, newURL, newExpiry)
		if err != nil {
			return nil, err
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
) (*models.Avatar, error) {
	err := r.DeleteAvatar(ctx, profileID)
	if err != nil && !errors.Is(err, profiles.ErrAvatarNotFound) {
		return nil, err
	}

	avatarID := uuid.New()
	minioKey := avatarID.String()

	if err := r.minio.UploadFile(ctx, r.avatarBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, profiles.ErrFailedToUpload
	}

	presignedURL, err := r.minio.GeneratePresignedURL(ctx, r.avatarBucket, minioKey, profiles.PRESIGNED_URL_EXPIRY)
	if err != nil {
		_ = r.minio.DeleteFile(ctx, r.avatarBucket, minioKey)
		return nil, profiles.ErrFailedToGenerateURL
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

	err = r.db.QueryRowContext(
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
	)
	if err != nil {
		_ = r.minio.DeleteFile(ctx, r.avatarBucket, minioKey)
		return nil, err
	}

	return avatar, nil
}

func (r *profileRepository) DeleteAvatar(ctx context.Context, profileID uuid.UUID) error {
	var minioKey string

	err := r.db.QueryRowContext(ctx, DELETE_AVATAR_BY_ID, profileID).Scan(&minioKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return profiles.ErrAvatarNotFound
		}
		return err
	}

	if err := r.minio.DeleteFile(ctx, r.avatarBucket, minioKey); err != nil {
		return err
	}

	return nil
}

func (r *profileRepository) ChangePassword(ctx context.Context, userID uuid.UUID, newPassword string) (*models.Profile, error) {
	updatedProfile := &models.Profile{}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, CHANGE_PASSWORD_BY_USER_ID, userID, hashPassword).Scan(
		&updatedProfile.ID, &updatedProfile.Username, &updatedProfile.CreatedAt, &updatedProfile.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return updatedProfile, nil
}

func (r *profileRepository) GetPassword(ctx context.Context, userID uuid.UUID) ([]byte, error) {
	var password []byte

	err := r.db.QueryRowContext(ctx, GET_PASSWORD_BY_USER_ID, userID).Scan(&password)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return password, nil
}

func (r *profileRepository) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, CHECK_USER_EXISTS, username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, profiles.ErrUsernameExists
	}

	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.Profile{
		ID:           uuid.New(),
		Username:     username,
		Password:     hashPassword,
		TokenVersion: 1,
	}

	_, err = r.db.ExecContext(ctx, CREATE_USER, user.ID, user.Username, user.Password, user.TokenVersion)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *profileRepository) SigninUser(ctx context.Context, username string) (*models.Profile, error) {
	user := &models.Profile{}

	err := r.db.QueryRowContext(ctx, GET_USER_BY_USERNAME, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.TokenVersion, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}

func (r *profileRepository) updateAvatarURL(ctx context.Context, avatarID uuid.UUID, url string, expiresAt time.Time) error {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	err := r.db.QueryRowContext(ctx, UPDATE_AVATAR_URL, avatarID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}
