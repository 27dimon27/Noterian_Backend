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
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

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
}

type NoteHandler struct {
	noteUsecase NoteUsecase
}

func NewNoteHandler(noteUsecase NoteUsecase) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
	}
}

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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

	blocks, err := h.noteUsecase.GetBlocks(r.Context(), noteID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteResponse(note, blocks)

	write.JSONResponse(w, http.StatusOK, response)
}

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

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
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

	var noteUpdateRequest dto.NoteRequest

	if err := body.GetBody(r, &noteUpdateRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	note := dto.FromNoteRequestDTO(noteUpdateRequest)

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

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
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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

func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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

func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
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
