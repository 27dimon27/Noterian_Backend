package auth

import (
	"errors"
)

var (
	ErrInvalidInput     = errors.New("Невалидный ввод")
	ErrInternal         = errors.New("Неизвестная ошибка сервера")
	ErrUnauthorized     = errors.New("Неавторизован")
	ErrMethodNotAllowed = errors.New("Неверный метод")
	ErrBadCredentials   = errors.New("Неверный логин или пароль")
	ErrInvalidUsername  = errors.New("Невалидное имя пользователя")
	ErrInvalidPassword  = errors.New("Невалидный пароль")
	ErrUserExist        = errors.New("Пользователь с таким именем уже существует")
	ErrUserNotExist     = errors.New("Пользователь не найден")
	ErrTokenCreation    = errors.New("Ошибка при создании пользователя")
	ErrInvalidUserID    = errors.New("Невалидный ID пользователя")
)
