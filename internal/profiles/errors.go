package profiles

import (
	"errors"
	"time"
)

var (
	ErrUserNotExist          = errors.New("profile not found")
	PublicMsgErrUserNotExist = "Пользователь не найден"

	ErrInvalidUserID       = errors.New("Невалидный UserID")
	ErrInvalidProfileData  = errors.New("Невалидные данные профиля")
	ErrFileTooLarge        = errors.New("Слишком большой файл")
	ErrInvalidMimeType     = errors.New("Неподдерживаемый MIME-тип файла")
	ErrFailedToUpload      = errors.New("Не удалось загрузить файл")
	ErrFailedToGenerateURL = errors.New("Не удалось сгенерировать ссылку")

	ErrAvatarNotFound          = errors.New("avatar not found")
	PublicMsgErrAvatarNotFound = "Аватар не найден"

	ErrWrongPassword       = errors.New("Неверный пароль")
	ErrBodyRequired        = errors.New("Тело запроса обязательно")
	ErrInvalidPasswordData = errors.New("Невалидные данные пароля")

	ErrUsernameExists          = errors.New("username already exists")
	PublicMsgErrUsernameExists = "Пользователь с таким именем уже существует"

	ErrInternalServer          = errors.New("internal server error")
	PublicMsgErrInternalServer = "Неизвестная ошибка сервера"
)

const (
	MAX_FILE_SIZE        = 5 * 1024 * 1024
	PRESIGNED_URL_EXPIRY = 30 * time.Minute
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}
