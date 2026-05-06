package handler

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockAttachmentUsecase struct {
	getAttachmentFunc    func(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error)
	uploadAttachmentFunc func(ctx context.Context, noteID, blockID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	deleteAttachmentFunc func(ctx context.Context, noteID, blockID, userID uuid.UUID) error
}

func (m *mockAttachmentUsecase) GetAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
	if m.getAttachmentFunc != nil {
		return m.getAttachmentFunc(ctx, noteID, blockID, userID)
	}
	return nil, nil
}

func (m *mockAttachmentUsecase) UploadAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error) {
	if m.uploadAttachmentFunc != nil {
		return m.uploadAttachmentFunc(ctx, noteID, blockID, userID, fileName, fileSize, mimeType, fileReader)
	}
	return nil, nil
}

func (m *mockAttachmentUsecase) DeleteAttachment(ctx context.Context, noteID, blockID, userID uuid.UUID) error {
	if m.deleteAttachmentFunc != nil {
		return m.deleteAttachmentFunc(ctx, noteID, blockID, userID)
	}
	return nil
}

func createMultipartFormData(t *testing.T, fileName string, fileContent []byte) (bytes.Buffer, string) {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	part, err := writer.CreateFormFile("file", fileName)
	assert.NoError(t, err)

	_, err = part.Write(fileContent)
	assert.NoError(t, err)

	err = writer.Close()
	assert.NoError(t, err)

	return buf, writer.FormDataContentType()
}

func createImageData() []byte {
	// Minimal valid JPEG data
	return []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10, 0x4A, 0x46, 0x49, 0x46, 0x00, 0x01}
}

func TestAttachmentHandler_GetAttachment(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(r *http.Request) *http.Request
		noteID         string
		blockID        string
		mockFunc       func(usecase *mockAttachmentUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.getAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
					return &models.Attachment{
						ID:           uuid.New(),
						BlockID:      blockID,
						MinioKey:     "test-key",
						AttachURL:    "http://example.com/test",
						URLExpiresAt: time.Now().Add(time.Hour),
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - missing userID",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			noteID:         uuid.New().String(),
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name: "bad request - missing noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "",
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "NoteID обязателен",
		},
		{
			name: "bad request - invalid noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "invalid-uuid",
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидный NoteID",
		},
		{
			name: "bad request - missing blockID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         uuid.New().String(),
			blockID:        "",
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "BlockID обязателен",
		},
		{
			name: "bad request - invalid blockID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         uuid.New().String(),
			blockID:        "invalid-uuid",
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидный BlockID",
		},
		{
			name: "forbidden",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.getAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
					return nil, attachments.ErrForbidden
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Доступ запрещен",
		},
		{
			name: "not found - note not found",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.getAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
					return nil, attachments.ErrNoteNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Заметка не найдена",
		},
		{
			name: "internal server error",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.getAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) (*models.Attachment, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockAttachmentUsecase{}
			tt.mockFunc(mockUsecase)
			handler := NewAttachmentHandler(mockUsecase)

			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.noteID+"/blocks/"+tt.blockID+"/attachments", nil)
			req = tt.setupContext(req)
			req.SetPathValue("noteId", tt.noteID)
			req.SetPathValue("blockId", tt.blockID)

			w := httptest.NewRecorder()
			handler.GetAttachment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestAttachmentHandler_UploadAttachment(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(r *http.Request) *http.Request
		noteID         string
		blockID        string
		fileContent    []byte
		fileName       string
		mockFunc       func(usecase *mockAttachmentUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success - image upload",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:      uuid.New().String(),
			blockID:     uuid.New().String(),
			fileContent: createImageData(),
			fileName:    "test.jpg",
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.uploadAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error) {
					return &models.Attachment{
						ID:           uuid.New(),
						BlockID:      blockID,
						MinioKey:     "test-key",
						AttachURL:    "http://example.com/test",
						URLExpiresAt: time.Now().Add(time.Hour),
						CreatedAt:    time.Now(),
						UpdatedAt:    time.Now(),
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "unauthorized - missing userID",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			noteID:         uuid.New().String(),
			blockID:        uuid.New().String(),
			fileContent:    createImageData(),
			fileName:       "test.jpg",
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name: "bad request - missing noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "",
			blockID:        uuid.New().String(),
			fileContent:    createImageData(),
			fileName:       "test.jpg",
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "NoteID обязателен",
		},
		{
			name: "bad request - invalid noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "invalid-uuid",
			blockID:        uuid.New().String(),
			fileContent:    createImageData(),
			fileName:       "test.jpg",
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидный NoteID",
		},
		{
			name: "conflict - block already has attachment",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:      uuid.New().String(),
			blockID:     uuid.New().String(),
			fileContent: createImageData(),
			fileName:    "test.jpg",
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.uploadAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error) {
					return nil, attachments.ErrBlockAlreadyHasAttach
				}
			},
			expectedStatus: http.StatusConflict,
			expectedBody:   "Блок уже содержит вложение",
		},
		{
			name: "forbidden",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:      uuid.New().String(),
			blockID:     uuid.New().String(),
			fileContent: createImageData(),
			fileName:    "test.jpg",
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.uploadAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error) {
					return nil, attachments.ErrForbidden
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Доступ запрещен",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockAttachmentUsecase{}
			tt.mockFunc(mockUsecase)
			handler := NewAttachmentHandler(mockUsecase)

			body, contentType := createMultipartFormData(t, tt.fileName, tt.fileContent)

			req := httptest.NewRequest(http.MethodPost, "/notes/"+tt.noteID+"/blocks/"+tt.blockID+"/attachments", &body)
			req.Header.Set("Content-Type", contentType)
			req = tt.setupContext(req)
			req.SetPathValue("noteId", tt.noteID)
			req.SetPathValue("blockId", tt.blockID)

			w := httptest.NewRecorder()
			handler.UploadAttachment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestAttachmentHandler_DeleteAttachment(t *testing.T) {
	tests := []struct {
		name           string
		setupContext   func(r *http.Request) *http.Request
		noteID         string
		blockID        string
		mockFunc       func(usecase *mockAttachmentUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name: "success",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.deleteAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name: "bad request - missing noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "",
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "NoteID обязателен",
		},
		{
			name: "bad request - invalid noteID",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:         "invalid-uuid",
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидный NoteID",
		},
		{
			name: "unauthorized - missing userID",
			setupContext: func(r *http.Request) *http.Request {
				return r
			},
			noteID:         uuid.New().String(),
			blockID:        uuid.New().String(),
			mockFunc:       func(usecase *mockAttachmentUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name: "forbidden",
			setupContext: func(r *http.Request) *http.Request {
				return r.WithContext(context.WithValue(r.Context(), types.UserIDKey, uuid.New()))
			},
			noteID:  uuid.New().String(),
			blockID: uuid.New().String(),
			mockFunc: func(usecase *mockAttachmentUsecase) {
				usecase.deleteAttachmentFunc = func(ctx context.Context, noteID, blockID, userID uuid.UUID) error {
					return attachments.ErrForbidden
				}
			},
			expectedStatus: http.StatusForbidden,
			expectedBody:   "Доступ запрещен",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockAttachmentUsecase{}
			tt.mockFunc(mockUsecase)
			handler := NewAttachmentHandler(mockUsecase)

			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.noteID+"/blocks/"+tt.blockID+"/attachments", nil)
			req = tt.setupContext(req)
			req.SetPathValue("noteId", tt.noteID)
			req.SetPathValue("blockId", tt.blockID)

			w := httptest.NewRecorder()
			handler.DeleteAttachment(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}
