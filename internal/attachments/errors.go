package attachments

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAttachmentNotFound          = errors.New("attachment not found")
	PublicMsgErrAttachmentNotFound = "Вложение не найдено"

	ErrHeaderNotFound          = errors.New("header not found")
	PublicMsgErrHeaderNotFound = "Шапка не найдена"

	ErrBlockAlreadyHasAttach          = errors.New("block already has attach")
	PublicMsgErrBlockAlreadyHasAttach = "Блок уже содержит вложение"

	ErrInvalidMimeType          = errors.New("unsupported MIME-type of file")
	PublicMsgErrInvalidMimeType = "Неподдерживаемый MIME-тип файла"

	ErrFileTooLarge          = errors.New("too large file")
	PublicMsgErrFileTooLarge = "Слишком большой файл"

	ErrFailedToUpload = errors.New("Не удалось загрузить файл")
	ErrFailedToDelete = errors.New("Не удалось удалить файл")

	ErrBlockNotFound          = errors.New("block not found")
	PublicMsgErrBlockNotFound = "Блок не найден"

	ErrNoteNotFound          = errors.New("note not found")
	PublicMsgErrNoteNotFound = "Заметка не найдена"

	ErrNoteIDRequired          = errors.New("noteID is required")
	PublicMsgErrNoteIDRequired = "NoteID обязателен"

	ErrInvalidNoteID          = errors.New("invalid noteID")
	PublicMsgErrInvalidNoteID = "Невалидный NoteID"

	ErrBlockIDRequired          = errors.New("blockID is required")
	PublicMsgErrBlockIDRequired = "BlockID обязателен"

	ErrInvalidBlockID          = errors.New("invalid blockID")
	PublicMsgErrInvalidBlockID = "Невалидный BlockID"

	ErrInvalidUserID          = errors.New("invalid userID")
	PublicMsgErrInvalidUserID = "Невалидный UserID"

	ErrForbidden          = errors.New("access denied")
	PublicMsgErrForbidden = "Доступ запрещен"

	ErrFailedToGenerateURL          = errors.New("failed to generate the link")
	PublicMsgErrFailedToGenerateURL = "Не удалось сгенерировать ссылку"

	ErrInvalidPosition          = errors.New("invalid position")
	PublicMsgErrInvalidPosition = "Невалидная позиция"

	ErrInternalServer          = errors.New("internal server error")
	PublicMsgErrInternalServer = "Неизвестная ошибка сервера"

	PublicMsgErrSpecificFileTooLarge = map[string]string{
		"IMAGE": fmt.Sprintf("Слишком большой файл фотографии, максимальный размер - %d МБ", MAX_IMAGE_SIZE/MB_CONST),
		"GIF":   fmt.Sprintf("Слишком большой файл GIF, максимальный размер - %d МБ", MAX_GIF_SIZE/MB_CONST),
		"AUDIO": fmt.Sprintf("Слишком большой аудиофайл, максимальный размер - %d МБ", MAX_AUDIO_SIZE/MB_CONST),
		"VIDEO": fmt.Sprintf("Слишком большой файл видео, максимальный размер - %d МБ", MAX_VIDEO_SIZE/MB_CONST),
	}
)

const (
	MB_CONST             = 1024 * 1024
	MAX_IMAGE_SIZE       = 5 * 1024 * 1024
	MAX_GIF_SIZE         = 15 * 1024 * 1024
	MAX_AUDIO_SIZE       = 30 * 1024 * 1024
	MAX_VIDEO_SIZE       = 50 * 1024 * 1024
	PRESIGNED_URL_EXPIRY = 30 * time.Minute
)

var AllowedMimeTypesForImage = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}

var AllowedMimeTypesForGIF = map[string]bool{
	"image/gif": true,
}

var AllowedMimeTypesForAudio = map[string]bool{
	"audio/mpeg":  true,
	"audio/mp4":   true,
	"audio/ogg":   true,
	"audio/wav":   true,
	"audio/webm":  true,
	"audio/flac":  true,
	"audio/x-m4a": true,
	"audio/aac":   true,
	"audio/opus":  true,
}

var AllowedMimeTypesForVideo = map[string]bool{
	"video/mp4":       true,
	"video/webm":      true,
	"video/ogg":       true,
	"video/quicktime": true,
	"video/x-msvideo": true,
	"video/mpeg":      true,
}
