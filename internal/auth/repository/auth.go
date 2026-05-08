package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type authRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *authRepository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) SignupUser(ctx context.Context, username, password string) (*models.Profile, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, CHECK_USER_EXISTS, username).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, auth.ErrUserExist
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

func (r *authRepository) SigninUser(ctx context.Context, username string) (*models.Profile, error) {
	user := &models.Profile{}

	err := r.db.QueryRowContext(ctx, GET_USER_BY_USERNAME, username).Scan(
		&user.ID, &user.Username, &user.Password, &user.TokenVersion, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}
