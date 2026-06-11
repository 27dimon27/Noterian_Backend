package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_handler_attachments.go -package=mocks

type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader, hasPosition bool, position int) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error
	GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, error)
	UploadHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Header, error)
	DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error
}

type AttachmentHandler struct {
	attachmentUsecase AttachmentUsecase
	logger            *slog.Logger
}

func NewAttachmentHandler(attachmentUsecase AttachmentUsecase, logger *slog.Logger) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentUsecase: attachmentUsecase,
		logger:            logger,
	}
}

func (h *AttachmentHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidBlockID)
		return
	}

	attachment, err := h.attachmentUsecase.GetAttachment(r.Context(), noteID, blockID, userID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound), errors.Is(err, attachments.ErrAttachmentNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
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

	r.Body = http.MaxBytesReader(w, r.Body, attachments.MAX_VIDEO_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn("Too large file")
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.ErrFileTooLarge)
		} else {
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error("Failed to close file in UploadAttachment", "error", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	mimeType := http.DetectContentType(buffer)

	maxSize, contentType, err := getMaxSizeByMimeType(mimeType)
	if err != nil {
		h.logger.Warn("Invalid MIME-type of file")
		write.JSONErrorResponse(w, http.StatusBadRequest, err)
		return
	}

	if fileHeader.Size > maxSize {
		h.logger.Warn("Too large file")
		write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.ErrSpecificFileTooLarge[contentType])
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	attachment, err := h.attachmentUsecase.UploadAttachment(
		r.Context(),
		noteID,
		userID,
		fileHeader.Filename,
		fileHeader.Size,
		mimeType,
		fileToUpload,
		hasPosition,
		position,
	)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound), errors.Is(err, attachments.ErrAttachmentNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, attachments.ErrBlockAlreadyHasAttach):
			h.logger.Warn("Block already has attach")
			write.JSONErrorResponse(w, http.StatusConflict, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	if attachment == nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn("BlockID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn("Invalid blockID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidBlockID)
		return
	}

	if err := h.attachmentUsecase.DeleteAttachment(r.Context(), noteID, blockID, userID); err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			h.logger.Warn("Access denied")
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound), errors.Is(err, attachments.ErrAttachmentNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *AttachmentHandler) GetHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	header, err := h.attachmentUsecase.GetHeader(r.Context(), noteID, userID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrHeaderNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToHeaderDTO(*header)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *AttachmentHandler) UploadHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, attachments.MAX_IMAGE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn("Too large file")
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.ErrFileTooLarge)
		} else {
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error("Failed to close file in UploadHeader", "error", err)
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	mimeType := http.DetectContentType(buffer)

	if !attachments.AllowedMimeTypesForImage[mimeType] {
		h.logger.Warn("Invalid MIME-type of file")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidMimeType)
		return
	}

	if fileHeader.Size > attachments.MAX_IMAGE_SIZE {
		h.logger.Warn("Too large file")
		write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.ErrSpecificFileTooLarge["IMAGE"])
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	header, err := h.attachmentUsecase.UploadHeader(
		r.Context(),
		noteID,
		userID,
		fileHeader.Filename,
		fileHeader.Size,
		mimeType,
		fileToUpload,
	)
	if err != nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	if header == nil {
		h.logger.Error("Internal server error", "error", err)
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToHeaderDTO(*header)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *AttachmentHandler) DeleteHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn("Invalid userID in context")
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn("NoteID is required")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn("Invalid noteID in url")
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	if err := h.attachmentUsecase.DeleteHeader(r.Context(), noteID, userID); err != nil {
		switch {
		case errors.Is(err, attachments.ErrHeaderNotFound):
			h.logger.Warn("Requested info not found")
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			h.logger.Error("Internal server error", "error", err)
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func getMaxSizeByMimeType(mimeType string) (int64, string, error) {
	if attachments.AllowedMimeTypesForImage[mimeType] {
		return attachments.MAX_IMAGE_SIZE, "IMAGE", nil
	}
	if attachments.AllowedMimeTypesForGIF[mimeType] {
		return attachments.MAX_GIF_SIZE, "GIF", nil
	}
	if attachments.AllowedMimeTypesForAudio[mimeType] {
		return attachments.MAX_AUDIO_SIZE, "AUDIO", nil
	}
	if attachments.AllowedMimeTypesForVideo[mimeType] {
		return attachments.MAX_VIDEO_SIZE, "VIDEO", nil
	}
	return 0, "", attachments.ErrInvalidMimeType
}
