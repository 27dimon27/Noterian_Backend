package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/pkg/helpers/write"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_handler_attachments.go -package=mocks

type AttachmentClient interface {
	GetAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID, fileReader io.Reader, fileName string, fileSize int64, mimeType string) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) error
}

type AttachmentHandler struct {
	attachmentClient AttachmentClient
}

func NewAttachmentHandler(attachmentClient AttachmentClient) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentClient: attachmentClient,
	}
}

func (h *AttachmentHandler) GetAttachment(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidBlockID)
		return
	}

	attachment, err := h.attachmentClient.GetAttachment(r.Context(), noteID, blockID, userID)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound), errors.Is(err, attachments.ErrAttachmentNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
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
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidBlockID)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, attachments.MAX_FILE_SIZE)

	if err := r.ParseMultipartForm(0); err != nil {
		var maxBytesError *http.MaxBytesError
		if errors.As(err, &maxBytesError) {
			write.JSONErrorResponse(w, http.StatusRequestEntityTooLarge, attachments.ErrFileTooLarge)
		} else {
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	file, fileHeader, err := r.FormFile("file")
	if err != nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil && err != io.EOF {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	fileToUpload := io.MultiReader(bytes.NewReader(buffer), file)

	mimeType := http.DetectContentType(buffer)

	if !attachments.AllowedMimeTypes[mimeType] {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidMimeType)
		return
	}

	attachment, err := h.attachmentClient.UploadAttachment(
		r.Context(),
		noteID,
		blockID,
		userID,
		fileToUpload,
		fileHeader.Filename,
		fileHeader.Size,
		mimeType,
	)
	if err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case errors.Is(err, attachments.ErrBlockAlreadyHasAttach):
			write.JSONErrorResponse(w, http.StatusConflict, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToAttachmentDTO(*attachment)
	write.JSONResponse(w, http.StatusCreated, response)
}

func (h *AttachmentHandler) DeleteAttachment(w http.ResponseWriter, r *http.Request) {
	noteIDStr := r.PathValue("noteId")
	if noteIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrNoteIDRequired)
		return
	}

	noteID, err := uuid.Parse(noteIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidNoteID)
		return
	}

	blockIDStr := r.PathValue("blockId")
	if blockIDStr == "" {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrBlockIDRequired)
		return
	}

	blockID, err := uuid.Parse(blockIDStr)
	if err != nil {
		write.JSONErrorResponse(w, http.StatusBadRequest, attachments.ErrInvalidBlockID)
		return
	}

	userID, ok := r.Context().Value(types.UserIDKey).(uuid.UUID)
	if !ok {
		write.JSONErrorResponse(w, http.StatusUnauthorized, attachments.ErrInvalidUserID)
		return
	}

	if err := h.attachmentClient.DeleteAttachment(r.Context(), noteID, blockID, userID); err != nil {
		switch {
		case errors.Is(err, attachments.ErrForbidden):
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case errors.Is(err, attachments.ErrNoteNotFound), errors.Is(err, attachments.ErrBlockNotFound), errors.Is(err, attachments.ErrAttachmentNotFound):
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
