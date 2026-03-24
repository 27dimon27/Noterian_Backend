package usecase

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/handler"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type AccountRepository interface {
	GetAccount(ctx context.Context, userID uuid.UUID) (*models.Account, error)
}

type accountUsecase struct {
	accountRepo AccountRepository
}

func NewAccountUsecase(accountRepo AccountRepository) handler.AccountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
	}
}

func (u *accountUsecase) GetAccount(ctx context.Context, userID uuid.UUID) (*models.Account, error) {
	account, err := u.accountRepo.GetAccount(ctx, userID)
	if err != nil {
		return nil, err
	}

	return account, nil
}
