package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestAttachmentHandler_GetAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	attachmentID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		expectedAttachment := &models.Attachment{
			ID:        attachmentID,
			BlockID:   blockID,
			MinioKey:  "test-key",
			AttachURL: "https://example.com/file",
		}

		mockUsecase.EXPECT().
			GetAttachment(gomock.Any(), noteID, blockID, userID).
			Return(expectedAttachment, nil)

		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.Attachment
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ID != attachmentID {
			t.Errorf("expected ID %v, got %v", attachmentID, response.ID)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes//blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes/invalid/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", "invalid")
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks//attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/invalid/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", "invalid")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetAttachment(gomock.Any(), noteID, blockID, userID).
			Return(nil, attachments.ErrForbidden)

		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetAttachment(gomock.Any(), noteID, blockID, userID).
			Return(nil, attachments.ErrAttachmentNotFound)

		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetAttachment(gomock.Any(), noteID, blockID, userID).
			Return(nil, errors.New("db error"))

		req := httptest.NewRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestAttachmentHandler_UploadAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	attachmentID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		expectedAttachment := &models.Attachment{
			ID:        attachmentID,
			BlockID:   blockID,
			MinioKey:  "test-key",
			AttachURL: "https://example.com/file",
		}

		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, "test.jpg", int64(11), "image/jpeg", gomock.Any()).
			Return(expectedAttachment, nil)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", "test.jpg")
		if err != nil {
			t.Fatalf("failed to create form file: %s", err)
		}
		// JPEG signature
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/notes//blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/notes/invalid/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", "invalid")
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks//attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/invalid/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", "invalid")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("no file in request", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		// не добавляем файл
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("invalid mime type", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.exe")
		// EXE signature
		part.Write([]byte("MZ\x90\x00\x03\x00\x00\x00\x04\x00\x00\x00\xFF\xFF\x00\x00"))
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, attachments.ErrForbidden)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("block already has attachment", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, attachments.ErrBlockAlreadyHasAttach)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status 409, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, attachments.ErrNoteNotFound)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, attachments.ErrBlockNotFound)

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAttachment(gomock.Any(), noteID, blockID, userID, gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write([]byte{0xFF, 0xD8, 0xFF, 0xE0})
		writer.Close()

		req := httptest.NewRequest("POST", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UploadAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
func TestAttachmentHandler_DeleteAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAttachment(gomock.Any(), noteID, blockID, userID).
			Return(nil)

		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/notes//blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/notes/invalid/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", "invalid")
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks//attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/invalid/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", "invalid")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAttachment(gomock.Any(), noteID, blockID, userID).
			Return(attachments.ErrForbidden)

		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAttachment(gomock.Any(), noteID, blockID, userID).
			Return(attachments.ErrAttachmentNotFound)

		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAttachment(gomock.Any(), noteID, blockID, userID).
			Return(errors.New("db error"))

		req := httptest.NewRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
