package handler

import (
	"context"
	"errors"
	"net/http"

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
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	ResetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetBlocksWithFormatting(ctx context.Context, noteID uuid.UUID) ([]models.Block, map[string]models.BlockFormatting, error)
	GetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, error)
	CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) error
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
// @Summary Получение всех заметок для пользователя
// @Tags notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} dto.NotesResponse "List of notes retrieved successfully"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes [get]
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
// @Summary Получение заметки по ID
// @Tags notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Success 200 {object} dto.NoteResponse "Note retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid note ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId} [get]
func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	note, err := h.noteUsecase.GetNote(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if note == nil {
		write.JSONErrorResponse(w, http.StatusNotFound, notes.ErrNoteNotFound)
		return
	}

	blocks, blockFormattings, err := h.noteUsecase.GetBlocksWithFormatting(r.Context(), noteID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteResponse(note, blocks, blockFormattings)

	write.JSONResponse(w, http.StatusOK, response)
}

// CreateNote godoc
// @Summary Создание заметки
// @Tags notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param request body dto.NoteRequest true "Note creation data"
// @Success 201 {object} dto.Note "Note created successfully"
// @Failure 400 {object} map[string]string "Invalid request body or note data"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes [post]
func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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
		if errors.Is(err, notes.ErrInvalidNoteData) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusCreated, response)
}

// UpdateNote godoc
// @Summary Обновление заметки
// @Tags notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param request body dto.NoteRequest true "Note update data"
// @Success 200 {object} dto.Note "Note updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body, note ID, or note data"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId} [put]
func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var noteUpdateRequest dto.NoteRequest

	if err := body.GetBody(r, &noteUpdateRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	note := dto.FromNoteRequestDTO(noteUpdateRequest)

	note.UserID = userID

	updatedNote, err := h.noteUsecase.UpdateNote(r.Context(), noteID, note, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidNoteData) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteDTO(updatedNote)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteNote godoc
// @Summary Удаление заметки
// @Tags notes
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Success 204 "Note deleted successfully (no content)"
// @Failure 400 {object} map[string]string "Invalid note ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId} [delete]
func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	if err := h.noteUsecase.DeleteNote(r.Context(), noteID, userID); err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateBlock godoc
// @Summary Создание блока в заметке
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param request body dto.BlockRequest true "Block creation data"
// @Success 201 {object} dto.Block "Block created successfully"
// @Failure 400 {object} map[string]string "Invalid request body, note ID, block data, or block type"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks [post]
func (h *NoteHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
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
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidBlockType) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockDTO(*createdBlock)

	write.JSONResponse(w, http.StatusCreated, response)
}

// GetBlock godoc
// @Summary Получение блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Success 200 {object} dto.Block "Block retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid note ID or block ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId} [get]
func (h *NoteHandler) GetBlock(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	block, err := h.noteUsecase.GetBlock(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockDTO(*block)

	write.JSONResponse(w, http.StatusOK, response)
}

// UpdateBlockContent godoc
// @Summary Обновление контента блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Param request body dto.UpdateBlockContentRequest true "New block content"
// @Success 200 {object} dto.Block "Block content updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body, note ID, block ID, or block data"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId}/content [put]
func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var updateBlockContentRequest dto.UpdateBlockContentRequest

	if err := body.GetBody(r, &updateBlockContentRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	updatedBlock, err := h.noteUsecase.UpdateBlockContent(r.Context(), blockID, noteID, userID, updateBlockContentRequest.Content)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockDTO(*updatedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

// MoveBlock godoc
// @Summary Перемещение блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Param request body dto.MoveBlockRequest true "New position data"
// @Success 200 {object} dto.Block "Block moved successfully"
// @Failure 400 {object} map[string]string "Invalid request body, note ID, block ID, or position"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId}/move [put]
func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var moveBlockRequest dto.MoveBlockRequest

	if err := body.GetBody(r, &moveBlockRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	movedBlock, err := h.noteUsecase.MoveBlock(r.Context(), blockID, noteID, userID, moveBlockRequest.NewPosition)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidPosition) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockDTO(*movedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

// DeleteBlock godoc
// @Summary Удаление блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Success 204 "Block deleted successfully (no content)"
// @Failure 400 {object} map[string]string "Invalid note ID or block ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId} [delete]
func (h *NoteHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	if err := h.noteUsecase.DeleteBlock(r.Context(), blockID, noteID, userID); err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UpdateBlockFormatting godoc
// @Summary Обновление форматирования блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Param request body dto.FormattingRange true "Formatting range and styles to apply"
// @Success 200 {object} dto.BlockFormatting "Formatting updated successfully"
// @Failure 400 {object} map[string]string "Invalid request body, note ID, block ID, formatting data, or formatting not supported for this block type"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId}/formatting [put]
func (h *NoteHandler) UpdateBlockFormatting(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer r.Body.Close()

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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
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
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidBlockType) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidFormattingForImageBlock) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, notes.ErrFormattingNotSupported) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidFormattingRange) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockFormattingDTO(*updatedFormatting)

	write.JSONResponse(w, http.StatusOK, response)
}

// ResetBlockFormatting godoc
// @Summary Сброс форматирования блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Success 200 {object} dto.BlockFormatting "All formatting reset successfully"
// @Failure 400 {object} map[string]string "Invalid note ID or block ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId}/formatting [delete]
func (h *NoteHandler) ResetBlockFormatting(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	updatedFormatting, err := h.noteUsecase.ResetBlockFormatting(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockFormattingDTO(*updatedFormatting)

	write.JSONResponse(w, http.StatusOK, response)
}

// GetBlockFormatting godoc
// @Summary Получение форматирования блока
// @Tags blocks
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param noteId path string true "Note ID (UUID format)"
// @Param blockId path string true "Block ID (UUID format)"
// @Success 200 {object} dto.BlockFormatting "Block formatting retrieved successfully"
// @Failure 400 {object} map[string]string "Invalid note ID or block ID format"
// @Failure 401 {object} map[string]string "Unauthorized - missing or invalid token"
// @Failure 403 {object} map[string]string "Forbidden - user does not have access to this note"
// @Failure 404 {object} map[string]string "Note or block not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /notes/{noteId}/blocks/{blockId}/formatting [get]
func (h *NoteHandler) GetBlockFormatting(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	formatting, err := h.noteUsecase.GetBlockFormatting(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockFormattingDTO(*formatting)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetSubnotes(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	subnotes, err := h.noteUsecase.GetSubnotes(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToSubnotesDTO(subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateSubnote(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var subnoteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &subnoteCreationRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	subnoteCreationRequest.UserID = userID
	subnoteCreationRequest.ParentID = &noteID

	note := dto.FromNoteRequestDTO(subnoteCreationRequest)

	createdNote, err := h.noteUsecase.CreateSubnote(r.Context(), noteID, userID, note)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) DeleteSubnote(w http.ResponseWriter, r *http.Request) {
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

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	err = h.noteUsecase.DeleteSubnote(r.Context(), noteID, subnoteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
