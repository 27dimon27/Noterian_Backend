package repository

import (
	"database/sql"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type accountRepository struct {
	db *sql.DB
}

func NewAccountRepository(db *sql.DB) (usecase.AccountRepository, error) {
	return &accountRepository{
		db: db,
	}, nil
}

func (r *accountRepository) GetAccount(userID uuid.UUID) (*models.Account, error) {
	user := &models.Account{}

	err := r.db.QueryRow(GET_ACCOUNT_BY_USER_ID, userID).Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, accounts.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}
