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
)

//go:generate mockgen -source=profile.go -destination=mocks/mock_repository_minio.go -package=mocks

type MinIOService interface {
	UploadFile(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, bucketName, key string) error
	GeneratePresignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}

type profileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) *profileRepository {
	return &profileRepository{
		db: db,
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
