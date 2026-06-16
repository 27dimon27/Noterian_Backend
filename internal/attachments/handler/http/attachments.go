package handler

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_handler_attachments.go -package=mocks

type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, types.AppErrorInterface)
	UploadAttachment(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader, hasPosition bool, position int) (*models.Attachment, types.AppErrorInterface)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
	GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, types.AppErrorInterface)
	UploadHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Header, types.AppErrorInterface)
	DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface
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
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAttachment", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAttachment", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAttachment", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAttachment", attachments.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetAttachment", attachments.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidBlockID)
		return
	}

	attachment, customErr := h.attachmentUsecase.GetAttachment(r.Context(), noteID, blockID, userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), attachments.ErrForbidden):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), attachments.ErrNoteNotFound), errors.Is(customErr.Unwrap(), attachments.ErrBlockNotFound), errors.Is(customErr.Unwrap(), attachments.ErrAttachmentNotFound):
			h.logger.Warn(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *AttachmentHandler) UploadAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	positionStr := r.URL.Query().Get("position")
	var position int
	var hasPosition bool
	if positionStr != "" {
		position, err = strconv.Atoi(positionStr)
		if err != nil {
			h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrInvalidPosition))
			write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidPosition)
			return
		}
		hasPosition = true
	}

	r.Body = http.MaxBytesReader(w, r.Body, attachments.MAX_VIDEO_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrFileTooLarge))
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.PublicMsgErrFileTooLarge)
		} else {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", err))
			write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", err))
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}

	mimeType := http.DetectContentType(buffer)

	maxSize, contentType, err := getMaxSizeByMimeType(mimeType)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrInvalidMimeType))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidMimeType)
		return
	}

	if fileHeader.Size > maxSize {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", attachments.ErrFileTooLarge))
		write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.PublicMsgErrSpecificFileTooLarge[contentType])
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	attachment, customErr := h.attachmentUsecase.UploadAttachment(
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
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), attachments.ErrForbidden):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), attachments.ErrNoteNotFound), errors.Is(customErr.Unwrap(), attachments.ErrBlockNotFound), errors.Is(customErr.Unwrap(), attachments.ErrAttachmentNotFound):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), attachments.ErrBlockAlreadyHasAttach):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusConflict, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	if attachment == nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadAttachment", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAttachment", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAttachment", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAttachment", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAttachment", attachments.ErrBlockIDRequired))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteAttachment", attachments.ErrInvalidBlockID))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidBlockID)
		return
	}

	if customErr := h.attachmentUsecase.DeleteAttachment(r.Context(), noteID, blockID, userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), attachments.ErrForbidden):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusForbidden, customErr.PublicMessage())
		case errors.Is(customErr.Unwrap(), attachments.ErrNoteNotFound), errors.Is(customErr.Unwrap(), attachments.ErrBlockNotFound), errors.Is(customErr.Unwrap(), attachments.ErrAttachmentNotFound):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	write.JSONResponse(w, http.StatusNoContent, nil)
}

func (h *AttachmentHandler) GetHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetHeader", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetHeader", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "GetHeader", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	header, customErr := h.attachmentUsecase.GetHeader(r.Context(), noteID, userID)
	if customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), attachments.ErrHeaderNotFound):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		}
		return
	}

	response := dto.ToHeaderDTO(*header)

	write.JSONResponse(w, http.StatusOK, response)
}

func (h *AttachmentHandler) UploadHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, attachments.MAX_IMAGE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrFileTooLarge))
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.PublicMsgErrFileTooLarge)
		} else {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", err))
			write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", err))
		}
	}()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}

	mimeType := http.DetectContentType(buffer)

	if !attachments.AllowedMimeTypesForImage[mimeType] {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrInvalidMimeType))
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.PublicMsgErrInvalidMimeType)
		return
	}

	if fileHeader.Size > attachments.MAX_IMAGE_SIZE {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", attachments.ErrFileTooLarge))
		write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.PublicMsgErrSpecificFileTooLarge["IMAGE"])
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	header, customErr := h.attachmentUsecase.UploadHeader(
		r.Context(),
		noteID,
		userID,
		fileHeader.Filename,
		fileHeader.Size,
		mimeType,
		fileToUpload,
	)
	if customErr != nil {
		h.logger.Error(customErr.Error())
		write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
		return
	}

	if header == nil {
		h.logger.Error(fmt.Sprintf("[%s:%s] %v", "handler", "UploadHeader", err))
		write.JSONErrorResponse(w, http.StatusInternalServerError, attachments.PublicMsgErrInternalServer)
		return
	}

	response := dto.ToHeaderDTO(*header)

	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *AttachmentHandler) DeleteHeader(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteHeader", attachments.ErrInvalidUserID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteHeader", attachments.ErrNoteIDRequired))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		h.logger.Warn(fmt.Sprintf("[%s:%s] %v", "handler", "DeleteHeader", attachments.ErrInvalidNoteID))
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.PublicMsgErrInvalidNoteID)
		return
	}

	if customErr := h.attachmentUsecase.DeleteHeader(r.Context(), noteID, userID); customErr != nil {
		switch {
		case errors.Is(customErr.Unwrap(), attachments.ErrHeaderNotFound):
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusNotFound, customErr.PublicMessage())
		default:
			h.logger.Error(customErr.Error())
			write.JSONErrorResponse(w, http.StatusInternalServerError, customErr.PublicMessage())
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
