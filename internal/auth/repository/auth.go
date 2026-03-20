package repository

import (
	"database/sql"
	"embed"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/queries"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

//go:embed queries/*.sql
var queriesFS embed.FS

type UserRepository interface {
	CreateUser(login, password string) (*models.Account, error)
	GetUserByLogin(login string) (*models.Account, error)
}

type userRepository struct {
	db      *sql.DB
	queries map[string]string
}

func NewUserRepository(db *sql.DB) (UserRepository, error) {
	queries, err := queries.LoadQueries(queriesFS, "queries")
	if err != nil {
		return nil, err
	}

	return &userRepository{
		db:      db,
		queries: queries,
	}, nil
}

func (r *userRepository) CreateUser(login, password string) (*models.Account, error) {
	var exists bool
	err := r.db.QueryRow(r.queries["check_user_exists"], login).Scan(&exists)
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

	_, err = r.db.Exec(r.queries["create_user"],
		user.ID, user.Username, user.Password, user.TokenVersion, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepository) GetUserByLogin(login string) (*models.Account, error) {
	user := &models.Account{}

	err := r.db.QueryRow(r.queries["get_user_by_login"], login).Scan(
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
