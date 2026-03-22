package handler

import (
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
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

type NotesResponse struct {
	Notes []models.Note `json:"notes"`
	Total int           `json:"total"`
}

type NoteResponse struct {
	Note   *models.Note   `json:"note"`
	Blocks []models.Block `json:"blocks"`
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
		helpers.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidUserID)
		return
	}

	notes, err := h.noteUsecase.GetNotesByUserID(userUUID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := NotesResponse{
		Notes: notes,
		Total: len(notes),
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

	blocks, err := h.noteUsecase.GetBlocksByNoteID(noteID)
	if err != nil {
		helpers.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := NoteResponse{
		Note:   note,
		Blocks: blocks,
	}

	helpers.JSONResponse(w, http.StatusOK, response)
}
