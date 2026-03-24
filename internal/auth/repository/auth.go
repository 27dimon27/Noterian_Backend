package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) (usecase.UserRepository, error) {
	return &userRepository{
		db: db,
	}, nil
}

func (r *userRepository) CreateUser(login, password string) (*models.Account, error) {
	var exists bool
	err := r.db.QueryRow(CHECK_USER_EXISTS, login).Scan(&exists)
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

	user := &models.Account{
		ID:           uuid.New(),
		Username:     login,
		Password:     hashPassword,
		TokenVersion: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	_, err = r.db.Exec(CREATE_USER, user.ID, user.Username, user.Password, user.TokenVersion, user.CreatedAt, user.UpdatedAt)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetUserByLogin(login string) (*models.Account, error) {
	user := &models.Account{}

	err := r.db.QueryRow(GET_USER_BY_LOGIN, login).Scan(
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
