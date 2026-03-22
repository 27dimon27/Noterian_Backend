package accounts

import (
	"errors"
)

var (
	ErrUserNotExist  = errors.New("Пользователь не найден")
	ErrInvalidUserID = errors.New("Невалидный UserID")
)
