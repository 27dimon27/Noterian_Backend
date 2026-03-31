package profiles

import (
	"errors"
)

var (
	ErrUserNotExist      = errors.New("Пользователь не найден")
	ErrInvalidUserID     = errors.New("Невалидный UserID")
	ErrInvalidProfileData = errors.New("Невалидные данные профиля")
)
