package notes

import "errors"

var (
	ErrNoteIDRequired = errors.New("NoteID обязателен")
	ErrInvalidUserID  = errors.New("Невалидный UserID")
	ErrInvalidNoteID  = errors.New("Невалидный NoteID")

	ErrInvalidNoteData          = errors.New("invalid note data")
	PublicMsgErrInvalidNoteData = "Невалидные данные заметки"

	ErrNoteNotFound          = errors.New("note not found")
	PublicMsgErrNoteNotFound = "Заметка не найдена"

	ErrForbidden          = errors.New("access denied")
	PublicMsgErrForbidden = "Доступ запрещен"

	ErrInvalidUUID      = errors.New("Невалидный UUID")
	ErrBodyRequired     = errors.New("Тело запроса обязательно")
	ErrBlockIDRequired  = errors.New("BlockID обязателен")
	ErrInvalidBlockID   = errors.New("Невалидный BlockID")
	ErrInvalidBlockData = errors.New("Невалидные данные блока")

	ErrBlockNotFound          = errors.New("block not found")
	PublicMsgErrBlockNotFound = "Блок не найден"

	ErrInvalidBlockType          = errors.New("invalid block type")
	PublicMsgErrInvalidBlockType = "Невалидный тип блока"

	ErrInvalidBlockContent = errors.New("Невалидное содержимое блока")

	ErrInvalidPosition          = errors.New("invalid block position")
	PublicMsgErrInvalidPosition = "Невалидная позиция блока"

	ErrInvalidFormatting = errors.New("Невалидное форматирование блока")

	ErrInvalidFormattingRange          = errors.New("invalid formatting range")
	PublicMsgErrInvalidFormattingRange = "Невалидный диапазон форматирования"

	ErrInvalidFormattingForImageBlock          = errors.New("invalid formatting for image block")
	PublicMsgErrInvalidFormattingForImageBlock = "Для блока с изображением допустимо только выравнивание"

	ErrFormattingNotSupported          = errors.New("formatting is not supported for that block type")
	PublicMsgErrFormattingNotSupported = "Форматирование не поддерживается для данного типа блока"

	ErrBlockTypeNotFound          = errors.New("block type not found")
	PublicMsgErrBlockTypeNotFound = "Тип блока не найден"

	ErrInternalServer          = errors.New("internal server error")
	PublicMsgErrInternalServer = "Неизвестная ошибка сервера"
)
