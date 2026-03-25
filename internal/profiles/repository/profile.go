package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase"
	"github.com/google/uuid"
)

type profileRepository struct {
	db *sql.DB
}

func NewProfileRepository(db *sql.DB) (usecase.ProfileRepository, error) {
	return &profileRepository{
		db: db,
	}, nil
}

func (r *profileRepository) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	user := &models.Profile{}

	err := r.db.QueryRowContext(ctx, GET_PROFILE_BY_USER_ID, userID).Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, profiles.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}
