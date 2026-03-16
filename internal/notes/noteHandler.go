package notes

import (
	"errors"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/storage"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

var (
	ErrNoteIDRequired = errors.New("note id is required")
	ErrInvalidNoteID  = errors.New("invalid note id")
	ErrNoteNotFound   = errors.New("note not found")
)

type NoteHandler struct {
	noteRepo *storage.NoteRepository
}

func NewNoteHandler(noteRepo *storage.NoteRepository) *NoteHandler {
	return &NoteHandler{
		noteRepo: noteRepo,
	}
}

func (h *NoteHandler) GetAllNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(string)
	if !ok {
		helpers.JSONErrorResponse(w, http.StatusUnauthorized, jwt.ErrNoUserID)
		return
	}

	userUUID, err := uuid.Parse(userID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, ErrInvalidNoteID)
		return
	}

	notes, err := h.noteRepo.GetNotesByUserID(userUUID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"notes": notes,
		"total": len(notes),
	}

	helpers.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("id")
	if noteIDStr == "" {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, ErrInvalidNoteID)
		return
	}

	note, err := h.noteRepo.GetNoteByID(noteID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if note == nil {
		helpers.JSONErrorResponse(w, http.StatusNotFound, ErrNoteNotFound)
		return
	}

	blocks, err := h.noteRepo.GetBlocksWithStatesByNoteID(noteID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := map[string]interface{}{
		"note":   note,
		"blocks": blocks,
	}

	helpers.JSONResponse(w, http.StatusOK, response)
}
