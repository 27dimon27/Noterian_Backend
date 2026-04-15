package handler

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "success",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				attachment := &models.Attachment{
					ID:           uuid.New(),
					BlockID:      blockID,
					MinioKey:     "test-key",
					AttachURL:    "http://example.com/test",
					URLExpiresAt: time.Now().Add(time.Hour),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockUsecase.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(attachment, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no user id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - missing note id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes//blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid note id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/invalid/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "invalid")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - missing block id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks//attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", "")
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid block id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/invalid/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", "invalid")
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				mockUsecase.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, attachments.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				mockUsecase.EXPECT().GetAttachment(gomock.Any(), noteID, blockID, userID).Return(nil, attachments.ErrAttachmentNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req := tt.setupRequest()
			w := httptest.NewRecorder()
			handler.GetAttachment(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAttachmentHandler_UploadAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name           string
		fileContent    []byte
		fileName       string
		contentType    string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success with PNG",
			fileContent: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52},
			fileName:    "test.png",
			contentType: "image/png",
			setupMock: func() {
				attachment := &models.Attachment{
					ID:           uuid.New(),
					BlockID:      blockID,
					MinioKey:     "test-key",
					AttachURL:    "http://example.com/test",
					URLExpiresAt: time.Now().Add(time.Hour),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockUsecase.EXPECT().UploadAttachment(
					gomock.Any(),
					noteID,
					blockID,
					userID,
					"test.png",
					gomock.Any(),
					"image/png",
					gomock.Any(),
				).Return(attachment, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "success with JPEG",
			fileContent: []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01},
			fileName:    "test.jpg",
			contentType: "image/jpeg",
			setupMock: func() {
				attachment := &models.Attachment{
					ID:           uuid.New(),
					BlockID:      blockID,
					MinioKey:     "test-key",
					AttachURL:    "http://example.com/test",
					URLExpiresAt: time.Now().Add(time.Hour),
					CreatedAt:    time.Now(),
					UpdatedAt:    time.Now(),
				}
				mockUsecase.EXPECT().UploadAttachment(
					gomock.Any(),
					noteID,
					blockID,
					userID,
					"test.jpg",
					gomock.Any(),
					"image/jpeg",
					gomock.Any(),
				).Return(attachment, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:        "conflict - block already has attachment",
			fileContent: []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			fileName:    "test.png",
			contentType: "image/png",
			setupMock: func() {
				mockUsecase.EXPECT().UploadAttachment(
					gomock.Any(),
					noteID,
					blockID,
					userID,
					"test.png",
					gomock.Any(),
					"image/png",
					gomock.Any(),
				).Return(nil, attachments.ErrBlockAlreadyHasAttach)
			},
			expectedStatus: http.StatusConflict,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.fileName)
			require.NoError(t, err)

			_, err = part.Write(tt.fileContent)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
			req = req.WithContext(ctx)

			req.SetPathValue("noteId", noteID.String())
			req.SetPathValue("blockId", blockID.String())

			w := httptest.NewRecorder()

			handler.UploadAttachment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAttachmentHandler_UploadAttachment_InvalidFiles(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name           string
		fileContent    []byte
		fileName       string
		expectedStatus int
	}{
		{
			name:           "invalid mime type - text file",
			fileContent:    []byte("This is a text file"),
			fileName:       "test.txt",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid mime type - PDF",
			fileContent:    []byte("%PDF-1.4"),
			fileName:       "test.pdf",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)

			part, err := writer.CreateFormFile("file", tt.fileName)
			require.NoError(t, err)

			_, err = part.Write(tt.fileContent)
			require.NoError(t, err)

			err = writer.Close()
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
			req = req.WithContext(ctx)

			req.SetPathValue("noteId", noteID.String())
			req.SetPathValue("blockId", blockID.String())

			w := httptest.NewRecorder()

			handler.UploadAttachment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAttachmentHandler_UploadAttachment_ValidationErrors(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
	}{
		{
			name: "unauthorized - no user id",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.png")
				part.Write([]byte{0x89, 0x50, 0x4E, 0x47})
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - missing note id",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.png")
				part.Write([]byte{0x89, 0x50, 0x4E, 0x47})
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/notes//blocks/"+blockID.String()+"/attachment", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid note id",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.png")
				part.Write([]byte{0x89, 0x50, 0x4E, 0x47})
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/notes/invalid/blocks/"+blockID.String()+"/attachment", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "invalid")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - missing block id",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "test.png")
				part.Write([]byte{0x89, 0x50, 0x4E, 0x47})
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks//attachment", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", "")
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - no file",
			setupRequest: func() *http.Request {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()

				req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := tt.setupRequest()
			w := httptest.NewRecorder()
			handler.UploadAttachment(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAttachmentHandler_DeleteAttachment(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAttachmentUsecase(ctrl)
	handler := NewAttachmentHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "success",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				mockUsecase.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "unauthorized - no user id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "bad request - missing note id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes//blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - invalid note id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/invalid/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", "invalid")
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "bad request - missing block id",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks//attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", "")
				return req
			},
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "forbidden",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				mockUsecase.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(attachments.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name: "not found",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/attachment", nil)
				ctx := context.WithValue(req.Context(), types.UserIDKey, userID)
				req = req.WithContext(ctx)
				req.SetPathValue("noteId", noteID.String())
				req.SetPathValue("blockId", blockID.String())
				return req
			},
			setupMock: func() {
				mockUsecase.EXPECT().DeleteAttachment(gomock.Any(), noteID, blockID, userID).Return(attachments.ErrAttachmentNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			req := tt.setupRequest()
			w := httptest.NewRecorder()
			handler.DeleteAttachment(w, req)
			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
