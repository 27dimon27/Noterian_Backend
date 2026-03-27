package notes

import "errors"

var (
	ErrNoteIDRequired = errors.New("NoteID обязателен")
	ErrInvalidUserID  = errors.New("Невалидный UserID")
	ErrInvalidNoteID  = errors.New("Невалидный NoteID")
	ErrNoteNotFound   = errors.New("Заметка не найдена")
	ErrinvalidUUID    = errors.New("Невалидный UUID")
)
