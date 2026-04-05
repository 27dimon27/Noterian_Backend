package attachments

import "errors"

var (
	ErrAttachmentNotFound    = errors.New("Вложнение не найдено")
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
)

const (
	MAX_FILE_SIZE = 100 * 1024 * 1024
)

var AllowedMimeTypes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/webp": true,
}
