package notes

import "errors"

var (
	ErrNoteIDRequired      = errors.New("NoteID обязателен")
	ErrInvalidUserID       = errors.New("Невалидный UserID")
	ErrInvalidNoteID       = errors.New("Невалидный NoteID")
	ErrInvalidNoteData     = errors.New("Невалидные данные заметки")
	ErrNoteNotFound        = errors.New("Заметка не найдена")
	ErrForbidden           = errors.New("Доступ запрещен")
	ErrInvalidUUID         = errors.New("Невалидный UUID")
	ErrBadRequest          = errors.New("Неверный запрос")
	ErrBlockIDRequired     = errors.New("BlockID обязателен")
	ErrInvalidBlockID      = errors.New("Невалидный BlockID")
	ErrInvalidBlockData    = errors.New("Невалидные данные блока")
	ErrBlockNotFound       = errors.New("Блок не найден")
	ErrInvalidBlockType    = errors.New("Невалидный тип блока")
	ErrInvalidBlockContent = errors.New("Невалидное содержимое блока")
	ErrInvalidPosition     = errors.New("Невалидная позиция блока")
)
