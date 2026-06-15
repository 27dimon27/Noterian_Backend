package auth

import (
	"errors"
)

var (
	ErrInvalidInput     = errors.New("Невалидный ввод")
	ErrInternal         = errors.New("Неизвестная ошибка сервера")
	ErrUnauthorized     = errors.New("Неавторизован")
	ErrMethodNotAllowed = errors.New("Неверный метод")

	ErrBadCredentials          = errors.New("invalid credentials")
	PublicMsgErrBadCredentials = "Неверный логин или пароль"

	ErrInvalidUsername          = errors.New("invalid username")
	PublicMsgErrInvalidUsername = "Невалидное имя пользователя"

	ErrInvalidPassword          = errors.New("invalid password")
	PublicMsgErrInvalidPassword = "Невалидный пароль"

	ErrUserExist     = errors.New("Пользователь с таким именем уже существует")
	ErrUserNotExist  = errors.New("Пользователь не найден")
	ErrTokenCreation = errors.New("Ошибка при создании пользователя")
	ErrInvalidUserID = errors.New("Невалидный ID пользователя")

	ErrInternalServer          = errors.New("internal server error")
	PublicMsgErrInternalServer = "Неизвестная ошибка сервера"
)
