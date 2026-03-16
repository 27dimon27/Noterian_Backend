package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/jwt"
	"github.com/google/uuid"
)

type NoteHandler struct {
	noteUsecase usecase.NoteUsecase
}

func NewNoteHandler(noteUsecase usecase.NoteUsecase) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
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
		helpers.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	notes, err := h.noteUsecase.GetNotesByUserID(userUUID)
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
		helpers.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	note, err := h.noteUsecase.GetNoteByID(noteID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if note == nil {
		helpers.JSONErrorResponse(w, http.StatusNotFound, notes.ErrNoteNotFound)
		return
	}

	blocks, err := h.noteUsecase.GetBlocksWithStatesByNoteID(noteID)
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
