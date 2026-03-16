package auth

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

var Validate = validator.New()

var (
	ErrInvalidInput     = errors.New("Невалидный ввод")
	ErrInternal         = errors.New("Неизвестная ошибка сервера")
	ErrUnauthorized     = errors.New("Неавторизован")
	ErrMethodNotAllowed = errors.New("Неверный метод")
	ErrBadCredentials   = errors.New("Неверный логин или пароль")
	ErrInvalidLogin     = errors.New("Невалидный логин")
	ErrInvalidPassword  = errors.New("Невалидный пароль")
	ErrUserExist        = errors.New("Пользователь с таким логином уже существует")
	ErrUserNotExist     = errors.New("Пользователь не найден")
	ErrTokenCreation    = errors.New("Ошибка при создании пользователя")
)
