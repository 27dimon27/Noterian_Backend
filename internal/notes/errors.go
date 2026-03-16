package notes

import "errors"

var (
	ErrNoteIDRequired = errors.New("NoteID обязателен")
	ErrInvalidNoteID  = errors.New("Невалидный NoteID")
	ErrNoteNotFound   = errors.New("Заметка не найдена")
)
