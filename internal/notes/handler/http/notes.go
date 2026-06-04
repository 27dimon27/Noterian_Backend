package handler

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/body"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/google/uuid"
)

//go:generate mockgen -source=notes.go -destination=mocks/mock_handler_notes.go -package=mocks

type NoteUsecase interface {
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error)
	GetPublicNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, error)
	CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note, hasPosition bool, position int) (*models.Note, uuid.UUID, error)
	DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) error
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error
	GenerateNotePDF(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*bytes.Buffer, error)
}

type NoteHandler struct {
	noteUsecase NoteUsecase
}

func NewNoteHandler(noteUsecase NoteUsecase) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
	}
}

// GetNotes godoc
// @Summary      Список заметок пользователя
// @Description  Возвращает все заметки текущего пользователя.
// @Tags         notes
// @Produce      json
// @Success      200  {object}  dto.NotesResponse
// @Failure      401  {object}  map[string]string  "Неавторизован"
// @Failure      500  {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /notes [get]
func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	notes, err := h.noteUsecase.GetNotes(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNotesResponse(notes)

	write.JSONResponse(w, http.StatusOK, response)
}

// GetNote godoc
// @Summary      Получить заметку
// @Description  Возвращает заметку c блоками и форматированиями.
// @Tags         notes
// @Produce      json
// @Param        noteId  path      string  true  "UUID заметки"
// @Success      200     {object}  dto.NoteResponse
// @Failure      400     {object}  map[string]string  "Некорректный noteId"
// @Failure      401     {object}  map[string]string  "Неавторизован"
// @Failure      403     {object}  map[string]string  "Доступ запрещён"
// @Failure      404     {object}  map[string]string  "Заметка не найдена"
// @Failure      500     {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /notes/{noteId} [get]
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	note, blocks, blockFormattings, err := h.noteUsecase.GetNote(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	if note == nil {
		write.JSONErrorResponse(w, http.StatusNotFound, notes.ErrNoteNotFound)
		return
	}

	response := dto.ToNoteResponse(note, blocks, blockFormattings)

	write.JSONResponse(w, http.StatusOK, response)
}

// CreateNote godoc
// @Summary      Создать заметку
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        request  body      dto.NoteRequest  true  "Параметры заметки"
// @Success      201      {object}  dto.Note
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes [post]
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in CreateNote: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var noteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &noteCreationRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	noteCreationRequest.UserID = userID

	note := dto.FromNoteRequestDTO(noteCreationRequest)

	createdNote, err := h.noteUsecase.CreateNote(r.Context(), note)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrInvalidNoteData):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusCreated, response)
}

// UpdateNote godoc
// @Summary      Обновить заметку
// @Tags         notes
// @Accept       json
// @Produce      json
// @Param        noteId   path      string           true  "UUID заметки"
// @Param        request  body      dto.NoteRequest  true  "Новые данные заметки"
// @Success      200      {object}  dto.Note
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка не найдена"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId} [put]
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in UpdateNote: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	var noteUpdateRequest dto.NoteRequest

	if err := body.GetBody(r, &noteUpdateRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	note := dto.FromNoteRequestDTO(noteUpdateRequest)

	updatedNote, err := h.noteUsecase.UpdateNote(r.Context(), noteID, userID, note)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrInvalidNoteData):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToNoteDTO(updatedNote)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteNote godoc
// @Summary      Удалить заметку
// @Tags         notes
// @Produce      json
// @Param        noteId  path  string  true  "UUID заметки"
// @Success      204     "Заметка удалена"
// @Failure      400     {object}  map[string]string  "Некорректный noteId"
// @Failure      401     {object}  map[string]string  "Неавторизован"
// @Failure      403     {object}  map[string]string  "Доступ запрещён"
// @Failure      404     {object}  map[string]string  "Заметка не найдена"
// @Failure      500     {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId} [delete]
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	err = h.noteUsecase.DeleteNote(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

// CreateBlock godoc
// @Summary      Создать блок заметки
// @Tags         blocks
// @Accept       json
// @Produce      json
// @Param        noteId   path      string             true  "UUID заметки"
// @Param        request  body      dto.BlockRequest   true  "Параметры блока"
// @Success      201      {object}  dto.Block
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка не найдена"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/blocks [post]
func (h *NoteHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in CreateBlock: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	var blockCreationRequest dto.BlockRequest

	if err := body.GetBody(r, &blockCreationRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	block := dto.FromBlockRequestDTO(blockCreationRequest)

	createdBlock, err := h.noteUsecase.CreateBlock(r.Context(), noteID, userID, block)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidBlockType), errors.Is(err, notes.ErrInvalidPosition):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockDTO(*createdBlock)

	write.JSONResponse(w, http.StatusCreated, response)
}

// UpdateBlockContent godoc
// @Summary      Обновить содержимое блока
// @Tags         blocks
// @Accept       json
// @Produce      json
// @Param        noteId   path      string                         true  "UUID заметки"
// @Param        blockId  path      string                         true  "UUID блока"
// @Param        request  body      dto.UpdateBlockContentRequest  true  "Новое содержимое"
// @Success      200      {object}  dto.Block
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка или блок не найдены"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/blocks/{blockId}/content [put]
func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in UpdateBlockContent: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var updateBlockContentRequest dto.UpdateBlockContentRequest

	if err := body.GetBody(r, &updateBlockContentRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	updatedBlock, err := h.noteUsecase.UpdateBlockContent(r.Context(), blockID, noteID, userID, updateBlockContentRequest.Content)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockDTO(*updatedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

// MoveBlock godoc
// @Summary      Переместить блок
// @Tags         blocks
// @Accept       json
// @Produce      json
// @Param        noteId   path      string                true  "UUID заметки"
// @Param        blockId  path      string                true  "UUID блока"
// @Param        request  body      dto.MoveBlockRequest  true  "Новая позиция блока"
// @Success      200      {object}  dto.Block
// @Failure      400      {object}  map[string]string  "Некорректные данные"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка или блок не найдены"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/blocks/{blockId}/move [put]
func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in MoveBlock: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var moveBlockRequest dto.MoveBlockRequest

	if err := body.GetBody(r, &moveBlockRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	movedBlock, err := h.noteUsecase.MoveBlock(r.Context(), blockID, noteID, userID, moveBlockRequest.NewPosition)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidPosition):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockDTO(*movedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteBlock godoc
// @Summary      Удалить блок
// @Tags         blocks
// @Produce      json
// @Param        noteId   path  string  true  "UUID заметки"
// @Param        blockId  path  string  true  "UUID блока"
// @Success      204      "Блок удалён"
// @Failure      400      {object}  map[string]string  "Некорректный noteId/blockId"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка или блок не найдены"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/blocks/{blockId} [delete]
func (h *NoteHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	err = h.noteUsecase.DeleteBlock(r.Context(), blockID, noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

// UpdateBlockFormatting godoc
// @Summary      Обновить форматирование блока
// @Tags         blocks
// @Accept       json
// @Produce      json
// @Param        noteId   path      string               true  "UUID заметки"
// @Param        blockId  path      string               true  "UUID блока"
// @Param        request  body      dto.FormattingRange  true  "Диапазон форматирования"
// @Success      200      {object}  dto.BlockFormatting
// @Failure      400      {object}  map[string]string  "Некорректные данные форматирования"
// @Failure      401      {object}  map[string]string  "Неавторизован"
// @Failure      403      {object}  map[string]string  "Доступ запрещён"
// @Failure      404      {object}  map[string]string  "Заметка, блок или тип блока не найдены"
// @Failure      500      {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/blocks/{blockId}/formatting [put]
func (h *NoteHandler) UpdateBlockFormatting(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in UpdateBlockFormatting: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var formattingRequest dto.FormattingRange

	if err := body.GetBody(r, &formattingRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidFormatting)
		return
	}

	formattingRange := dto.FromFormattingRangeDTO(formattingRequest)

	updatedFormatting, err := h.noteUsecase.UpdateBlockFormatting(r.Context(), blockID, noteID, userID, formattingRange)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound), errors.Is(err, notes.ErrBlockTypeNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidBlockType), errors.Is(err, notes.ErrInvalidFormattingForImageBlock), errors.Is(err, notes.ErrFormattingNotSupported), errors.Is(err, notes.ErrInvalidFormattingRange):
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockFormattingDTO(*updatedFormatting)

	write.JSONResponse(w, http.StatusOK, response)
}

// GetSubnotes godoc
// @Summary      Список подзаметок
// @Tags         subnotes
// @Produce      json
// @Param        noteId  path      string  true  "UUID родительской заметки"
// @Success      200     {array}   dto.Note
// @Failure      400     {object}  map[string]string  "Некорректный noteId"
// @Failure      401     {object}  map[string]string  "Неавторизован"
// @Failure      403     {object}  map[string]string  "Доступ запрещён"
// @Failure      404     {object}  map[string]string  "Заметка не найдена"
// @Failure      500     {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /notes/{noteId}/subnote [get]
func (h *NoteHandler) GetSubnotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	subnotes, err := h.noteUsecase.GetSubnotes(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToNotesResponse(subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

// CreateSubnote godoc
// @Summary      Создать подзаметку
// @Tags         subnotes
// @Accept       json
// @Produce      json
// @Param        noteId    path      string           true   "UUID родительской заметки"
// @Param        position  query     int              false  "Позиция вставки блока-ссылки в родительскую заметку"
// @Param        request   body      dto.NoteRequest  true   "Параметры подзаметки"
// @Success      200       {object}  dto.Subnote
// @Failure      400       {object}  map[string]string  "Некорректные данные"
// @Failure      401       {object}  map[string]string  "Неавторизован"
// @Failure      403       {object}  map[string]string  "Доступ запрещён"
// @Failure      404       {object}  map[string]string  "Родительская заметка не найдена"
// @Failure      500       {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/subnote [post]
func (h *NoteHandler) CreateSubnote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			log.Printf("failed to close request body in CreateSubnote: %v", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	positionStr := r.URL.Query().Get("position")
	var position int
	var hasPosition bool
	if positionStr != "" {
		position, err = strconv.Atoi(positionStr)
		if err != nil {
			write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidPosition)
			return
		}
		hasPosition = true
	}

	var subnoteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &subnoteCreationRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	subnoteCreationRequest.UserID = userID
	subnoteCreationRequest.ParentID = &noteID

	note := dto.FromNoteRequestDTO(subnoteCreationRequest)

	createdNote, blockID, err := h.noteUsecase.CreateSubnote(r.Context(), noteID, userID, note, hasPosition, position)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToSubnoteDTO(*createdNote, blockID)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteSubnote godoc
// @Summary      Удалить подзаметку
// @Tags         subnotes
// @Produce      json
// @Param        noteId     path  string  true  "UUID родительской заметки"
// @Param        subnoteId  path  string  true  "UUID подзаметки"
// @Success      204        "Подзаметка удалена"
// @Failure      400        {object}  map[string]string  "Некорректные идентификаторы"
// @Failure      401        {object}  map[string]string  "Неавторизован"
// @Failure      403        {object}  map[string]string  "Доступ запрещён"
// @Failure      404        {object}  map[string]string  "Заметка не найдена"
// @Failure      500        {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Security     CsrfToken
// @Router       /notes/{noteId}/subnote/{subnoteId} [delete]
func (h *NoteHandler) DeleteSubnote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	subnoteIDStr := r.PathValue("subnoteId")
	if subnoteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	subnoteID, err := uuid.Parse(subnoteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	err = h.noteUsecase.DeleteSubnote(r.Context(), noteID, subnoteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

// GetNotePDF godoc
// @Summary      Скачать заметку в PDF
// @Description  Возвращает PDF-файл заметки со всем её содержимым (текст, форматирование, аттачи, подзаметки).
// @Tags         notes
// @Produce      application/pdf
// @Param        noteId  path      string  true  "UUID заметки"
// @Success      200     {file}    application/pdf
// @Failure      400     {object}  map[string]string  "Некорректный noteId"
// @Failure      401     {object}  map[string]string  "Неавторизован"
// @Failure      403     {object}  map[string]string  "Доступ запрещён"
// @Failure      404     {object}  map[string]string  "Заметка не найдена"
// @Failure      500     {object}  map[string]string  "Внутренняя ошибка сервера"
// @Security     ApiKeyAuth
// @Router       /notes/{noteId}/pdf [get]
func (h *NoteHandler) GetNotePDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	pdfBuffer, err := h.noteUsecase.GenerateNotePDF(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrNoteNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"note-"+noteID.String()+".pdf\"")
	w.Header().Set("Content-Length", strconv.Itoa(pdfBuffer.Len()))

	_, err = w.Write(pdfBuffer.Bytes())
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
}

// GetPublicNote godoc
// @Summary      Получить публичные метаданные заметки
// @Description  Возвращает минимальные метаданные (id, title, icon) публичной заметки без авторизации. Используется для генерации meta-тегов для share-инга.
// @Tags         notes
// @Produce      json
// @Param        noteId  path      string  true  "UUID заметки"
// @Success      200     {object}  dto.PublicNoteResponse
// @Failure      400     {object}  map[string]string  "Некорректный noteId"
// @Failure      404     {object}  map[string]string  "Заметка не найдена"
// @Failure      500     {object}  map[string]string  "Внутренняя ошибка сервера"
// @Router       /public/notes/{noteId} [get]
func (h *NoteHandler) GetPublicNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	note, err := h.noteUsecase.GetPublicNote(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToPublicNoteResponse(*note)

	write.JSONResponse(w, http.StatusOK, response)
}
