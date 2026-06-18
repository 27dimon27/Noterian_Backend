package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, types.AppErrorInterface)
	GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, types.AppErrorInterface)
	GetPublicNote(ctx context.Context, noteID uuid.UUID) (*models.Note, types.AppErrorInterface)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, types.AppErrorInterface)
	UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, types.AppErrorInterface)
	DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
	CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, types.AppErrorInterface)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, types.AppErrorInterface)
	MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, types.AppErrorInterface)
	DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, types.AppErrorInterface)
	GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, types.AppErrorInterface)
	CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note, hasPosition bool, position int) (*models.Note, uuid.UUID, types.AppErrorInterface)
	DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) types.AppErrorInterface
	GenerateNotePDF(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*bytes.Buffer, types.AppErrorInterface)
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
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNotes", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	notes, customErr := h.noteUsecase.GetNotes(r.Context(), userID)
	if customErr != nil {
		h.logger.Error(customErr.Error())
		write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		return
	}

	response := dto.ToNotesResponse(notes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	note, blocks, blockFormattings, customErr := h.noteUsecase.GetNote(r.Context(), noteID, userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	if note == nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "GetNote", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	response := dto.ToNoteResponse(note, blocks, blockFormattings)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateNote", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateNote", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateNote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	var noteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &noteCreationRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateNote", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	noteCreationRequest.UserID = userID

	note := dto.FromNoteRequestDTO(noteCreationRequest)

	createdNote, customErr := h.noteUsecase.CreateNote(r.Context(), note)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrInvalidNoteData):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToNoteDTO(createdNote)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *NoteHandler) UpdateNote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	var noteUpdateRequest dto.NoteRequest

	if err := body.GetBody(r, &noteUpdateRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateNote", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	note := dto.FromNoteRequestDTO(noteUpdateRequest)

	updatedNote, customErr := h.noteUsecase.UpdateNote(r.Context(), noteID, userID, note)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrInvalidNoteData):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToNoteDTO(updatedNote)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) DeleteNote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteNote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteNote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteNote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	if customErr := h.noteUsecase.DeleteNote(r.Context(), noteID, userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) CreateBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	var blockCreationRequest dto.BlockRequest

	if err := body.GetBody(r, &blockCreationRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateBlock", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	block := dto.FromBlockRequestDTO(blockCreationRequest)

	createdBlock, customErr := h.noteUsecase.CreateBlock(r.Context(), noteID, userID, block)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrInvalidBlockType), errors.Is(customErr.Unwrap(), notes.ErrInvalidPosition):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToBlockDTO(*createdBlock)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *NoteHandler) UpdateBlockContent(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", notes.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidBlockID)
		return
	}

	var updateBlockContentRequest dto.UpdateBlockContentRequest

	if err := body.GetBody(r, &updateBlockContentRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockContent", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	updatedBlock, customErr := h.noteUsecase.UpdateBlockContent(r.Context(), blockID, noteID, userID, updateBlockContentRequest.Content)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound), errors.Is(customErr.Unwrap(), notes.ErrBlockNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToBlockDTO(*updatedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) MoveBlock(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", notes.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidBlockID)
		return
	}

	var moveBlockRequest dto.MoveBlockRequest

	if err := body.GetBody(r, &moveBlockRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "MoveBlock", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	movedBlock, customErr := h.noteUsecase.MoveBlock(r.Context(), blockID, noteID, userID, moveBlockRequest.NewPosition)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound), errors.Is(customErr.Unwrap(), notes.ErrBlockNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrInvalidPosition):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToBlockDTO(*movedBlock)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) DeleteBlock(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteBlock", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteBlock", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteBlock", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteBlock", notes.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteBlock", notes.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidBlockID)
		return
	}

	if customErr := h.noteUsecase.DeleteBlock(r.Context(), blockID, noteID, userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound), errors.Is(customErr.Unwrap(), notes.ErrBlockNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) UpdateBlockFormatting(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", notes.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidBlockID)
		return
	}

	var formattingRequest dto.FormattingRange

	if err := body.GetBody(r, &formattingRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UpdateBlockFormatting", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	formattingRange := dto.FromFormattingRangeDTO(formattingRequest)

	updatedFormatting, customErr := h.noteUsecase.UpdateBlockFormatting(r.Context(), blockID, noteID, userID, formattingRange)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound), errors.Is(customErr.Unwrap(), notes.ErrBlockNotFound), errors.Is(customErr.Unwrap(), notes.ErrBlockTypeNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(err, notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(err, notes.ErrInvalidBlockType), errors.Is(err, notes.ErrInvalidFormattingForImageBlock), errors.Is(err, notes.ErrFormattingNotSupported), errors.Is(err, notes.ErrInvalidFormattingRange):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusBadRequest, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToBlockFormattingDTO(*updatedFormatting)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) GetSubnotes(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetSubnotes", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetSubnotes", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetSubnotes", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	subnotes, customErr := h.noteUsecase.GetSubnotes(r.Context(), noteID, userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToNotesResponse(subnotes)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) CreateSubnote(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", notes.ErrBodyRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrBodyRequired)
		return
	}
	defer func() {
		if err := r.Body.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", err))
		}
	}()

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	positionStr := r.URL.Query().Get("position")
	var position int
	var hasPosition bool
	if positionStr != "" {
		position, err = strconv.Atoi(positionStr)
		if err != nil {
			h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", notes.ErrInvalidPosition))
			write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidPosition)
			return
		}
		hasPosition = true
	}

	var subnoteCreationRequest dto.NoteRequest

	if err := body.GetBody(r, &subnoteCreationRequest); err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "CreateSubnote", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}

	subnoteCreationRequest.UserID = userID
	subnoteCreationRequest.ParentID = &noteID

	note := dto.FromNoteRequestDTO(subnoteCreationRequest)

	createdNote, blockID, customErr := h.noteUsecase.CreateSubnote(r.Context(), noteID, userID, note, hasPosition, position)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToSubnoteDTO(*createdNote, blockID)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *NoteHandler) DeleteSubnote(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteSubnote", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteSubnote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteSubnote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	subnoteIDStr := r.PathValue("subnoteId")
	if subnoteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteSubnote", notes.ErrSubnoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrSubnoteIDRequired)
		return
	}

	subnoteID, err := uuid.Parse(subnoteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteSubnote", notes.ErrInvalidSubnoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidSubnoteID)
		return
	}

	if customErr := h.noteUsecase.DeleteSubnote(r.Context(), noteID, subnoteID, userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *NoteHandler) GetNotePDF(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNotePDF", notes.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, notes.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNotePDF", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetNotePDF", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	pdfBuffer, customErr := h.noteUsecase.GenerateNotePDF(r.Context(), noteID, userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"note-"+noteID.String()+".pdf\"")
	w.Header().Set("Content-Length", strconv.Itoa(pdfBuffer.Len()))

	_, err = w.Write(pdfBuffer.Bytes())
	if err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "GetNotePDF", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, notes.PublicMsgErrInternalServer)
		return
	}
}

func (h *NoteHandler) GetPublicNote(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetPublicNote", notes.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetPublicNote", notes.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, notes.PublicMsgErrInvalidNoteID)
		return
	}

	note, customErr := h.noteUsecase.GetPublicNote(r.Context(), noteID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), notes.ErrNoteNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToPublicNoteResponse(*note)

	write.JSONResponse(w, http.StatusOK, response)
}
