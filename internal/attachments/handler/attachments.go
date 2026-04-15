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

type AttachmentUsecase interface {
	GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error
}

type AttachmentHandler struct {
	attachmentUsecase AttachmentUsecase
}

func NewAttachmentHandler(attachmentUsecase AttachmentUsecase) *AttachmentHandler {
	return &AttachmentHandler{
		attachmentUsecase: attachmentUsecase,
	}
}

// GetAttachment godoc
// @Summary Получение вложения
// @Tags attachments
// @Accept json
// @Produce json
// @Param noteId path string true "ID заметки"
// @Param blockId path string true "ID блока"
// @Success 200 {object} dto.Attachment "Информация о вложении"
// @Failure 400 {object} map[string]string "Невалидный NoteID или BlockID"
// @Failure 401 {object} map[string]string "Невалидный UserID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заметка, блок или вложение не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /notes/{noteId}/blocks/{blockId}/attachments [get]
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

	attachment, err := h.attachmentUsecase.GetAttachment(r.Context(), noteID, blockID, userID)
	if err != nil {
		switch err {
		case attachments.ErrForbidden:
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusOK, response)
}

// UploadAttachment godoc
// @Summary Загрузка вложения
// @Tags attachments
// @Accept mpfd
// @Produce json
// @Param noteId path string true "ID заметки"
// @Param blockId path string true "ID блока"
// @Param file formData file true "Файл для загрузки (JPEG, PNG, WEBP, макс. 100MB)"
// @Success 201 {object} dto.Attachment "Информация о загруженном вложении"
// @Failure 400 {object} map[string]string "Невалидный NoteID/BlockID или неподдерживаемый MIME-тип"
// @Failure 401 {object} map[string]string "Невалидный UserID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заметка или блок не найдены"
// @Failure 409 {object} map[string]string "Блок уже содержит вложение"
// @Failure 413 {object} map[string]string "Файл слишком большой"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /notes/{noteId}/blocks/{blockId}/attachments [post]
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

	attachment, err := h.attachmentUsecase.UploadAttachment(
		r.Context(),
		noteID,
		blockID,
		userID,
		fileHeader.Filename,
		fileHeader.Size,
		mimeType,
		fileToUpload,
	)
	if err != nil {
		switch err {
		case attachments.ErrForbidden:
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		case attachments.ErrBlockAlreadyHasAttach:
			write.JSONErrorResponse(w, http.StatusConflict, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	if attachment == nil {
		write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		return
	}

	response := dto.ToAttachmentDTO(*attachment)

	write.JSONResponse(w, http.StatusCreated, response)
}

// DeleteAttachment godoc
// @Summary Удаление вложения
// @Tags attachments
// @Accept json
// @Produce json
// @Param noteId path string true "ID заметки"
// @Param blockId path string true "ID блока"
// @Success 204 "Вложение успешно удалено"
// @Failure 400 {object} map[string]string "Невалидный NoteID или BlockID"
// @Failure 401 {object} map[string]string "Невалидный UserID"
// @Failure 403 {object} map[string]string "Доступ запрещен"
// @Failure 404 {object} map[string]string "Заметка, блок или вложение не найдены"
// @Failure 500 {object} map[string]string "Внутренняя ошибка сервера"
// @Security ApiKeyAuth
// @Router /notes/{noteId}/blocks/{blockId}/attachments [delete]
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

	if err := h.attachmentUsecase.DeleteAttachment(r.Context(), noteID, blockID, userID); err != nil {
		switch err {
		case attachments.ErrForbidden:
			write.JSONErrorResponse(w, http.StatusForbidden, err)
		case attachments.ErrNoteNotFound, attachments.ErrBlockNotFound, attachments.ErrAttachmentNotFound:
			write.JSONErrorResponse(w, http.StatusNotFound, err)
		default:
			write.JSONErrorResponse(w, http.StatusInternalServerError, err)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
