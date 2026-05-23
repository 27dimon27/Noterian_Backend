package attachments

import (
	"errors"
	"fmt"
	"time"
)

var (
	ErrAttachmentNotFound    = errors.New("Вложение не найдено")
	ErrHeaderNotFound        = errors.New("Шапка не найдена")
	ErrBlockAlreadyHasAttach = errors.New("Блок уже содержит вложение")
	ErrInvalidMimeType       = errors.New("Неподдерживаемый MIME-тип файла")
	ErrFileTooLarge          = errors.New("Слишком большой файл")
	ErrFailedToUpload        = errors.New("Не удалось загрузить файл")
	ErrFailedToDelete        = errors.New("Не удалось удалить файл")
	ErrBlockNotFound         = errors.New("Блок не найден")
	ErrNoteNotFound          = errors.New("Заметка не найдена")
	ErrNoteIDRequired        = errors.New("NoteID обязателен")
	ErrInvalidNoteID         = errors.New("Невалидный NoteID")
	ErrBlockIDRequired       = errors.New("BlockID обязателен")
	ErrInvalidBlockID        = errors.New("Невалидный BlockID")
	ErrInvalidUserID         = errors.New("Невалидный UserID")
	ErrForbidden             = errors.New("Доступ запрещен")
	ErrFailedToGenerateURL   = errors.New("Не удалось сгенерировать ссылку")
	ErrInvalidPosition       = errors.New("Невалидная позиция")
	ErrSpecificFileTooLarge  = map[string]error{
		"IMAGE": fmt.Errorf("Слишком большой файл фотографии, максимальный размер - %d МБ", MAX_IMAGE_SIZE/MB_CONST),
		"GIF":   fmt.Errorf("Слишком большой файл GIF, максимальный размер - %d МБ", MAX_GIF_SIZE/MB_CONST),
		"AUDIO": fmt.Errorf("Слишком большой аудиофайл, максимальный размер - %d МБ", MAX_AUDIO_SIZE/MB_CONST),
		"VIDEO": fmt.Errorf("Слишком большой файл видео, максимальный размер - %d МБ", MAX_VIDEO_SIZE/MB_CONST),
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
