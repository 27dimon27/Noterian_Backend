package handler

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
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
	logger      *slog.Logger
}

func NewNoteHandler(noteUsecase NoteUsecase, logger *slog.Logger) *NoteHandler {
	return &NoteHandler{
		noteUsecase: noteUsecase,
		logger:      logger,
	}
}

func (h *NoteHandler) GetNotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	notes, err := h.noteUsecase.GetNotes(r.Context(), userID)
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToNotesResponse(notes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	note, blocks, blockFormattings, err := h.noteUsecase.GetNote(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Note not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	if note == nil {
		h.logger.Warn("Note not found")
		write.JSONErrorResponse(w, http.StatusNotFound, notes.ErrNoteNotFound)
		return
	}

	response := dto.ToNoteResponse(note, blocks, blockFormattings)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in CreateNote", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	var noteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &noteCreationRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	noteCreationRequest.UserID = userID

	note := dto.FromNoteRequestDTO(noteCreationRequest)

	createdNote, err := h.noteUsecase.CreateNote(r.Context(), note)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrInvalidNoteData):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("failed to close request body in UpdateNote", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	var noteUpdateRequest dto.NoteRequest

	if err := body.GetBody(r, &noteUpdateRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteData)
		return
	}

	note := dto.FromNoteRequestDTO(noteUpdateRequest)

	updatedNote, err := h.noteUsecase.UpdateNote(r.Context(), noteID, userID, note)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Note not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrInvalidNoteData):
			h.logger.Warn("Invalid note data")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
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
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	err = h.noteUsecase.DeleteNote(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Note not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in CreateBlock", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	var blockCreationRequest dto.BlockRequest

	if err := body.GetBody(r, &blockCreationRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	block := dto.FromBlockRequestDTO(blockCreationRequest)

	createdBlock, err := h.noteUsecase.CreateBlock(r.Context(), noteID, userID, block)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Note not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidBlockType), errors.Is(err, notes.ErrInvalidPosition):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockDTO(*createdBlock)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in UpdateBlockContent", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var updateBlockContentRequest dto.UpdateBlockContentRequest

	if err := body.GetBody(r, &updateBlockContentRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	updatedBlock, err := h.noteUsecase.UpdateBlockContent(r.Context(), blockID, noteID, userID, updateBlockContentRequest.Content)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			h.logger.Warn("Requested info was not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToBlockDTO(*updatedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in MoveBlock", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var moveBlockRequest dto.MoveBlockRequest

	if err := body.GetBody(r, &moveBlockRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockData)
		return
	}

	movedBlock, err := h.noteUsecase.MoveBlock(r.Context(), blockID, noteID, userID, moveBlockRequest.NewPosition)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidPosition):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
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
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	err = h.noteUsecase.DeleteBlock(r.Context(), blockID, noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) UpdateBlockFormatting(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in UpdateBlockFormatting", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidBlockID)
		return
	}

	var formattingRequest dto.FormattingRange

	if err := body.GetBody(r, &formattingRequest); err != nil {
		h.logger.Warn("Error during reading body")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidFormatting)
		return
	}

	formattingRange := dto.FromFormattingRangeDTO(formattingRequest)

	updatedFormatting, err := h.noteUsecase.UpdateBlockFormatting(r.Context(), blockID, noteID, userID, formattingRange)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound), errors.Is(err, notes.ErrBlockNotFound), errors.Is(err, notes.ErrBlockTypeNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrInvalidBlockType), errors.Is(err, notes.ErrInvalidFormattingForImageBlock), errors.Is(err, notes.ErrFormattingNotSupported), errors.Is(err, notes.ErrInvalidFormattingRange):
			h.logger.Warn("Bad request from user")
			write.JSONErrorResponse(w, http.StatusBadRequest, err)
		default:
			h.logger.Error("Internal server error", "error", err)
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
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	subnotes, err := h.noteUsecase.GetSubnotes(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToNotesResponse(subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateSubnote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn("Body is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error("Failed to close request body in CreateSubnote", "error", err)
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid userID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	positionStr := r.URL.Query().Get("position")
	var position int
	var hasPosition bool
	if positionStr != "" {
		position, err = strconv.Atoi(positionStr)
		if err != nil {
			h.logger.Warn("Invalid position in url")
			write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidPosition)
			return
		}
		hasPosition = true
	}

	var subnoteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &subnoteCreationRequest); err != nil {
		h.logger.Warn("Error during reading body")
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
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToSubnoteDTO(*createdNote, blockID)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) DeleteSubnote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	subnoteIDStr := r.PathValue("subnoteId")
	if subnoteIDStr == "" {
		h.logger.Warn("SubnoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	subnoteID, err := uuid.Parse(subnoteIDStr)
	if err != nil {
		h.logger.Warn("Invalid subnoteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	err = h.noteUsecase.DeleteSubnote(r.Context(), noteID, subnoteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) GetNotePDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	pdfBuffer, err := h.noteUsecase.GenerateNotePDF(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, notes.ErrNoteNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"note-"+noteID.String()+".pdf\"")
	w.Header().Set("Content-Length", strconv.Itoa(pdfBuffer.Len()))

	_, err = w.Write(pdfBuffer.Bytes())
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
}

func (h *NoteHandler) GetPublicNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.ErrInvalidNoteID)
		return
	}

	note, err := h.noteUsecase.GetPublicNote(r.Context(), noteID)
	if err != nil {
		if errors.Is(err, notes.ErrNoteNotFound) {
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
			return
		}
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToPublicNoteResponse(*note)

	write.JSONResponse(w, http.StatusOK, response)
}
