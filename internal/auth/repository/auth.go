package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserRepository interface {
	CreateUser(login, password string) (*models.Account, error)
	ValidateUser(login, password string) (*models.Account, error)
}

type userRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateUser(login, password string) (*models.Account, error) {
	var exists bool
	err := r.db.QueryRow("SELECT EXISTS(SELECT 1 FROM accounts WHERE username = $1)", login).Scan(&exists)
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

	_, err = r.db.Exec(
		"INSERT INTO accounts (id, username, password, token_version, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)",
		user.ID, user.Username, user.Password, user.TokenVersion, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) ValidateUser(login, password string) (*models.Account, error) {
	user := &models.Account{}

	err := r.db.QueryRow(
		"SELECT id, username, password, token_version, created_at, updated_at FROM accounts WHERE username = $1",
		login,
	).Scan(&user.ID, &user.Username, &user.Password, &user.TokenVersion, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, auth.ErrUserNotExist
		}
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, auth.ErrUserNotExist
	}

	return user, nil
}
