package notes

import "errors"

var (
	ErrNoteIDRequired                 = errors.New("NoteID обязателен")
	ErrInvalidUserID                  = errors.New("Невалидный UserID")
	ErrInvalidNoteID                  = errors.New("Невалидный NoteID")
	ErrInvalidNoteData                = errors.New("Невалидные данные заметки")
	ErrNoteNotFound                   = errors.New("Заметка не найдена")
	ErrForbidden                      = errors.New("Доступ запрещен")
	ErrInvalidUUID                    = errors.New("Невалидный UUID")
	ErrBodyRequired                   = errors.New("Тело запроса обязательно")
	ErrBlockIDRequired                = errors.New("BlockID обязателен")
	ErrInvalidBlockID                 = errors.New("Невалидный BlockID")
	ErrInvalidBlockData               = errors.New("Невалидные данные блока")
	ErrBlockNotFound                  = errors.New("Блок не найден")
	ErrInvalidBlockType               = errors.New("Невалидный тип блока")
	ErrInvalidBlockContent            = errors.New("Невалидное содержимое блока")
	ErrInvalidPosition                = errors.New("Невалидная позиция блока")
	ErrInvalidFormatting              = errors.New("Невалидное форматирование блока")
	ErrInvalidFormattingRange         = errors.New("Невалидный диапазон форматирования")
	ErrInvalidFormattingForImageBlock = errors.New("Для блока с изображением допустимо только выравнивание")
	ErrFormattingNotSupported         = errors.New("Форматирование не поддерживается для данного типа блока")
)
