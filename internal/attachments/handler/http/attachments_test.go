package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func withUserID(r *http.Request, userID uuid.UUID) *http.Request {
	return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, userID))
}

// pngBytes returns 512+ bytes that begin with a PNG magic header so
// http.DetectContentType returns "image/png".
func pngBytes(t *testing.T) []byte {
	t.Helper()
	header := []byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}
	payload := bytes.Repeat([]byte{0}, 600)
	return append(header, payload...)
}

func buildMultipartRequest(t *testing.T, fieldName, filename string, content []byte) *http.Request {
	t.Helper()
	body := &bytes.Buffer{}
	mw := multipart.NewWriter(body)
	fw, err := mw.CreateFormFile(fieldName, filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	if _, err := fw.Write(content); err != nil {
		t.Fatalf("write: %v", err)
	}
	if err := mw.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", body)
	req.Header.Set("Content-Type", mw.FormDataContentType())
	return req
}

func TestGetAttachmentHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", "not-a-uuid")
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", "not-a-uuid")
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, attachments.ErrForbidden)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, attachments.ErrAttachmentNotFound)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		expected := &models.Attachment{ID: uuid.New(), BlockID: blockID, AttachURL: "https://example.com/file"}
		uc.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(expected, nil)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.GetAttachment(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "https://example.com/file") {
			t.Errorf("expected URL in body, got %s", w.Body.String())
		}
	})
}

func TestUploadAttachmentHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := buildMultipartRequest(t, "file", "x.png", pngBytes(t))
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "x.png", pngBytes(t)), userID)
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "x.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", "not-uuid")
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "x.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		req.URL.RawQuery = "position=notanumber"
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid mime", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "f.txt", bytes.Repeat([]byte("plain text content "), 50)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing file form field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "other", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("file too large for image", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		// Image > MAX_IMAGE_SIZE but <= MAX_VIDEO_SIZE so MaxBytesReader doesn't trip
		big := append([]byte{0x89, 'P', 'N', 'G', 0x0D, 0x0A, 0x1A, 0x0A}, bytes.Repeat([]byte{0}, attachments.MAX_IMAGE_SIZE+1024)...)
		req := withUserID(buildMultipartRequest(t, "file", "f.png", big), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 413, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).Return(nil, attachments.ErrForbidden)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).Return(nil, attachments.ErrNoteNotFound)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("block already has attachment", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).Return(nil, attachments.ErrBlockAlreadyHasAttach)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", w.Code)
		}
	})

	t.Run("internal error from usecase", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).Return(nil, errors.New("boom"))

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success with explicit position", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		expected := &models.Attachment{ID: uuid.New(), BlockID: uuid.New(), AttachURL: "https://example.com/file"}
		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), true, 3).Return(expected, nil)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		req.URL.RawQuery = "position=3"
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d (%s)", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "https://example.com/file") {
			t.Errorf("expected URL in body, got %s", w.Body.String())
		}
	})

	t.Run("success default position", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		expected := &models.Attachment{ID: uuid.New(), BlockID: uuid.New(), AttachURL: "https://example.com/file"}
		uc.EXPECT().UploadAttachment(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any(), false, 0).Return(expected, nil)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d (%s)", w.Code, w.Body.String())
		}
	})

	t.Run("malformed multipart body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not a real multipart body"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=nonexistent")
		req = withUserID(req, userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("oversized exceeds max video", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		body := &bytes.Buffer{}
		mw := multipart.NewWriter(body)
		fw, err := mw.CreateFormFile("file", "v.mp4")
		if err != nil {
			t.Fatalf("CreateFormFile: %v", err)
		}
		if _, err := fw.Write(bytes.Repeat([]byte("A"), attachments.MAX_VIDEO_SIZE+1024)); err != nil {
			t.Fatalf("write: %v", err)
		}
		if err := mw.Close(); err != nil {
			t.Fatalf("close: %v", err)
		}
		req := httptest.NewRequest(http.MethodPost, "/", body)
		req.Header.Set("Content-Type", mw.FormDataContentType())
		req = withUserID(req, userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadAttachment(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 413, got %d", w.Code)
		}
	})
}

func TestDeleteAttachmentHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", "not-uuid")
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", "not-uuid")
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(attachments.ErrForbidden)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(attachments.ErrBlockNotFound)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(nil)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()
		h.DeleteAttachment(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})
}

func TestGetHeaderHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", "not-uuid")
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(nil, attachments.ErrHeaderNotFound)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(nil, errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		expected := &models.Header{ID: uuid.New(), NoteID: noteID, HeaderURL: "https://example.com/header"}
		uc.EXPECT().GetHeader(gomock.Any(), noteID, userID).Return(expected, nil)

		req := withUserID(httptest.NewRequest(http.MethodGet, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.GetHeader(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d (%s)", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "https://example.com/header") {
			t.Errorf("expected URL in body, got %s", w.Body.String())
		}
	})
}

func TestUploadHeaderHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := buildMultipartRequest(t, "file", "x.png", pngBytes(t))
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "x.png", pngBytes(t)), userID)
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "x.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", "not-uuid")
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("malformed multipart body", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader("not a real multipart body"))
		req.Header.Set("Content-Type", "multipart/form-data; boundary=nonexistent")
		req = withUserID(req, userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("file too large via MaxBytesReader", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		big := bytes.Repeat([]byte("A"), attachments.MAX_IMAGE_SIZE+1024)
		req := withUserID(buildMultipartRequest(t, "file", "x.png", big), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusRequestEntityTooLarge {
			t.Errorf("expected 413, got %d", w.Code)
		}
	})

	t.Run("missing file form field", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "other", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("invalid mime", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(buildMultipartRequest(t, "file", "f.txt", bytes.Repeat([]byte("plain text content "), 50)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("usecase error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().UploadHeader(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).Return(nil, errors.New("boom"))

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		expected := &models.Header{ID: uuid.New(), NoteID: noteID, HeaderURL: "https://example.com/header"}
		uc.EXPECT().UploadHeader(gomock.Any(), noteID, userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).Return(expected, nil)

		req := withUserID(buildMultipartRequest(t, "file", "f.png", pngBytes(t)), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.UploadHeader(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d (%s)", w.Code, w.Body.String())
		}
		if !strings.Contains(w.Body.String(), "https://example.com/header") {
			t.Errorf("expected URL in body, got %s", w.Body.String())
		}
	})
}

func TestDeleteHeaderHandler(t *testing.T) {
	userID := uuid.New()
	noteID := uuid.New()

	t.Run("unauthorized", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := httptest.NewRequest(http.MethodDelete, "/", nil)
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", "not-uuid")
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(attachments.ErrHeaderNotFound)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(errors.New("boom"))

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})

	t.Run("success", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		uc := mocks.NewMockAttachmentUsecase(ctrl)
		h := NewAttachmentHandler(uc)

		uc.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(nil)

		req := withUserID(httptest.NewRequest(http.MethodDelete, "/", nil), userID)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()
		h.DeleteHeader(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected 204, got %d", w.Code)
		}
	})
}

func TestGetMaxSizeByMimeType(t *testing.T) {
	cases := []struct {
		mime         string
		expectedSize int64
		expectedKind string
		expectErr    bool
	}{
		{"image/png", attachments.MAX_IMAGE_SIZE, "IMAGE", false},
		{"image/jpeg", attachments.MAX_IMAGE_SIZE, "IMAGE", false},
		{"image/webp", attachments.MAX_IMAGE_SIZE, "IMAGE", false},
		{"image/gif", attachments.MAX_GIF_SIZE, "GIF", false},
		{"audio/mpeg", attachments.MAX_AUDIO_SIZE, "AUDIO", false},
		{"audio/wav", attachments.MAX_AUDIO_SIZE, "AUDIO", false},
		{"video/mp4", attachments.MAX_VIDEO_SIZE, "VIDEO", false},
		{"video/webm", attachments.MAX_VIDEO_SIZE, "VIDEO", false},
		{"application/pdf", 0, "", true},
		{"", 0, "", true},
	}

	for _, c := range cases {
		size, kind, err := getMaxSizeByMimeType(c.mime)
		if c.expectErr {
			if err == nil {
				t.Errorf("%s: expected err, got nil", c.mime)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s: unexpected err %v", c.mime, err)
		}
		if size != c.expectedSize {
			t.Errorf("%s: expected size %d, got %d", c.mime, c.expectedSize, size)
		}
		if kind != c.expectedKind {
			t.Errorf("%s: expected kind %s, got %s", c.mime, c.expectedKind, kind)
		}
	}
}

// Compile-time guard ensuring multipart import is used even if all tests above
// were stripped — keeps the file robust against unused-import errors.
var _ io.Reader = (*multipart.Part)(nil)
