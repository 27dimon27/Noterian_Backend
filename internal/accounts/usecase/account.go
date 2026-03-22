package usecase

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/accounts/repository"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type AccountUsecase interface {
	GetAccount(userID uuid.UUID) (*models.Account, error)
}

type accountUsecase struct {
	accountRepo repository.AccountRepository
}

func NewAccountUsecase(accountRepo repository.AccountRepository) AccountUsecase {
	return &accountUsecase{
		accountRepo: accountRepo,
	}
}

func (u *accountUsecase) GetAccount(userID uuid.UUID) (*models.Account, error) {
	account, err := u.accountRepo.GetAccount(userID)
	if err != nil {
		return nil, err
	}

	return account, nil
}
