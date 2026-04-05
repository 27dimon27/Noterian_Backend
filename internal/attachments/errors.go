package attachments

import "errors"

var (
	ErrAttachmentNotFound    = errors.New("attachment not found")
	ErrBlockAlreadyHasAttach = errors.New("block already has attachment")
	ErrInvalidMimeType       = errors.New("invalid file mime type")
	ErrFileTooLarge          = errors.New("file size exceeds 100MB limit")
	ErrFailedToUpload        = errors.New("failed to upload file to MinIO")
	ErrFailedToDelete        = errors.New("failed to delete file from MinIO")
	ErrBlockNotFound         = errors.New("block not found")
	ErrNoteNotFound          = errors.New("note not found")
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
