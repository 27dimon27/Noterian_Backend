package notes

import "errors"

var (
	ErrNoteIDRequired   = errors.New("NoteID обязателен")
	ErrInvalidUserID    = errors.New("Невалидный UserID")
	ErrInvalidNoteID    = errors.New("Невалидный NoteID")
	ErrInvalidNoteData  = errors.New("Невалидные данные заметки")
	ErrNoteNotFound     = errors.New("Заметка не найдена")
	ErrForbidden        = errors.New("Доступ запрещен")
	ErrInvalidUUID      = errors.New("Невалидный UUID")
	ErrMethodNotAllowed = errors.New("Неверный метод")
)
