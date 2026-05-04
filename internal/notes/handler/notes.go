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

type NoteClient interface {
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, map[string][]models.Note, error)
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	CreateNote(ctx context.Context, userID uuid.UUID, title string, parentID *uuid.UUID) (*models.Note, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	UpdateNote(ctx context.Context, noteID, userID uuid.UUID, title string, parentID *uuid.UUID) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID, userID uuid.UUID) error

	GetSubnotes(ctx context.Context, noteID, userID uuid.UUID) ([]models.Note, error)
	CreateSubnote(ctx context.Context, parentNoteID, userID uuid.UUID, title string) (*models.Note, error)
	DeleteSubnote(ctx context.Context, noteID, subnoteID, userID uuid.UUID) error

	CreateBlock(ctx context.Context, noteID, userID uuid.UUID, blockTypeID, position int, content string) (*models.Block, error)
	GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID, noteID, userID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, blockID, noteID, userID uuid.UUID, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) error

	UpdateBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	ResetBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetBlockFormatting(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.BlockFormatting, error)
	GetBlocksWithFormatting(ctx context.Context, noteID uuid.UUID) ([]models.Block, map[string]models.BlockFormatting, error)
}

type NoteHandler struct {
	noteClient NoteClient
}

func NewNoteHandler(noteClient NoteClient) *NoteHandler {
	return &NoteHandler{
		noteClient: noteClient,
	}
}

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	notes, subnotes, err := h.noteClient.GetNotes(r.Context(), userID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNotesResponse(notes, subnotes)

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
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	note, err := h.noteClient.GetNote(r.Context(), noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, notes.ErrNoteNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	blocks, blockFormattings, err := h.noteClient.GetBlocksWithFormatting(r.Context(), noteID)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
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

	createdNote, err := h.noteClient.CreateNote(r.Context(), userID, noteCreationRequest.Title, noteCreationRequest.ParentID)
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

	updatedNote, err := h.noteClient.UpdateNote(r.Context(), noteID, userID, noteUpdateRequest.Title, noteUpdateRequest.ParentID)
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

	if err := h.noteClient.DeleteNote(r.Context(), noteID, userID); err != nil {
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var blockCreationRequest dto.BlockRequest

	if err := body.GetBody(r, &blockCreationRequest); err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	createdBlock, err := h.noteClient.CreateBlock(r.Context(), noteID, userID, blockCreationRequest.BlockTypeID, blockCreationRequest.Position, blockCreationRequest.Content)
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	block, err := h.noteClient.GetBlock(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	updatedBlock, err := h.noteClient.UpdateBlockContent(r.Context(), blockID, noteID, userID, updateBlockContentRequest.Content)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	movedBlock, err := h.noteClient.MoveBlock(r.Context(), blockID, noteID, userID, moveBlockRequest.NewPosition)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	if err := h.noteClient.DeleteBlock(r.Context(), blockID, noteID, userID); err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	updatedFormatting, err := h.noteClient.UpdateBlockFormatting(r.Context(), blockID, noteID, userID, formattingRange)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		if errors.Is(err, notes.ErrForbidden) {
			write.JSONErrorResponse(w, http.StatusForbidden, err)
			return
		}
		if errors.Is(err, notes.ErrInvalidBlockType) || errors.Is(err, notes.ErrInvalidFormattingRange) {
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
			return
		}
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToBlockFormattingDTO(*updatedFormatting)
	write.JSONResponse(w, http.StatusOK, response)
}

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

	updatedFormatting, err := h.noteClient.ResetBlockFormatting(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	formatting, err := h.noteClient.GetBlockFormatting(r.Context(), blockID, noteID, userID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) || errors.Is(err, notes.ErrBlockNotFound) {
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

	subnotes, err := h.noteClient.GetSubnotes(r.Context(), noteID, userID)
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

	createdNote, err := h.noteClient.CreateSubnote(r.Context(), noteID, userID, subnoteCreationRequest.Title)
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

	err = h.noteClient.DeleteSubnote(r.Context(), noteID, subnoteID, userID)
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
