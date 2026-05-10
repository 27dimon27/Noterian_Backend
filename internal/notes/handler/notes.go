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
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, map[string][]models.Note, error)
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
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

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	notes, subnotes, err := h.noteUsecase.GetNotes(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNotesResponse(notes, subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

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

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
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

func (h *NoteHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
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

func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
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

func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
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

func (h *NoteHandler) UpdateBlockFormatting(w http.ResponseWriter, r *http.Request) {
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

	response := dto.ToSubnotesDTO(subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateSubnote(w http.ResponseWriter, r *http.Request) {
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

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusOK, response)
}

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
