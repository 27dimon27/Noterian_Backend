package csrf

import "errors"

var (
	ErrCSRFTokenMissing          = errors.New("CSRF-токен отсутствует")
	ErrCSRFTokenInvalid          = errors.New("Невалидный CSRF-токен")
	ErrFailedToGenerateCSRFToken = errors.New("Не удалось сгенерировать CSRF-токен")
)
