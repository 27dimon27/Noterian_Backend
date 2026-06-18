package csrf

import "errors"

var (
	ErrCSRFTokenMissing          = errors.New("CSRF-token is missing")
	PublicMsgErrCSRFTokenMissing = "CSRF-токен отсутствует"

	ErrCSRFTokenInvalid          = errors.New("invalid CSRF-token")
	PublicMsgErrCSRFTokenInvalid = "Невалидный CSRF-токен"

	ErrFailedToGenerateCSRFToken = errors.New("Не удалось сгенерировать CSRF-токен")

	ErrInternalServer          = errors.New("internal server error")
	PublicMsgErrInternalServer = "Неизвестная ошибка сервера"
)
