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
	GetNotesByUserID(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	GetBlocksByNoteID(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
}

type NoteHandler struct {
	noteUsecase NoteUsecase
}

func NewNoteHandler(noteUsecase NoteUsecase) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
	}
}

func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

	notes, err := h.noteUsecase.GetNotesByUserID(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNotesResponse(notes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("id")
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

	note, err := h.noteUsecase.GetNoteByID(r.Context(), noteID, userID)
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

	blocks, err := h.noteUsecase.GetBlocksByNoteID(r.Context(), noteID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNoteResponse(note, blocks)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		write.JSONErrorResponse(w, http.StatusMethodNotAllowed, notes.ErrMethodNotAllowed)
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

	noteIDStr := r.PathValue("id")
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
	noteIDStr := r.PathValue("id")
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
