package repository

import (
	"database/sql"
	"embed"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage/queries"
	"github.com/google/uuid"
)

//go:embed queries/*.sql
var queriesFS embed.FS

type accountRepository struct {
	db      *sql.DB
	queries map[string]string
}

func NewAccountRepository(db *sql.DB) (usecase.AccountRepository, error) {
	queries, err := queries.LoadQueries(queriesFS, "queries")
	if err != nil {
		return nil, err
	}

	return &accountRepository{
		db:      db,
		queries: queries,
	}, nil
}

func (r *accountRepository) GetAccount(userID uuid.UUID) (*models.Account, error) {
	user := &models.Account{}

	err := r.db.QueryRow(r.queries["get_account_by_user_id"], userID).Scan(&user.ID, &user.Username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, accounts.ErrUserNotExist
		}
		return nil, err
	}

	return user, nil
}
