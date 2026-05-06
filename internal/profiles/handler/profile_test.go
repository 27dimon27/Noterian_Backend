package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

type mockProfileUsecase struct {
	getProfileFunc     func(ctx context.Context, userID uuid.UUID) (*models.Profile, error)
	updateProfileFunc  func(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error)
	deleteProfileFunc  func(ctx context.Context, userID uuid.UUID) error
	getAvatarFunc      func(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error)
	uploadAvatarFunc   func(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error)
	deleteAvatarFunc   func(ctx context.Context, profileID uuid.UUID) error
	changePasswordFunc func(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error)
}

func (m *mockProfileUsecase) GetProfile(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
	if m.getProfileFunc != nil {
		return m.getProfileFunc(ctx, userID)
	}
	return nil, nil
}

func (m *mockProfileUsecase) UpdateProfile(ctx context.Context, userID uuid.UUID, profile models.Profile) (*models.Profile, error) {
	if m.updateProfileFunc != nil {
		return m.updateProfileFunc(ctx, userID, profile)
	}
	return nil, nil
}

func (m *mockProfileUsecase) DeleteProfile(ctx context.Context, userID uuid.UUID) error {
	if m.deleteProfileFunc != nil {
		return m.deleteProfileFunc(ctx, userID)
	}
	return nil
}

func (m *mockProfileUsecase) GetAvatar(ctx context.Context, profileID uuid.UUID) (*models.Avatar, error) {
	if m.getAvatarFunc != nil {
		return m.getAvatarFunc(ctx, profileID)
	}
	return nil, nil
}

func (m *mockProfileUsecase) UploadAvatar(ctx context.Context, profileID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error) {
	if m.uploadAvatarFunc != nil {
		return m.uploadAvatarFunc(ctx, profileID, fileName, fileSize, mimeType, fileReader)
	}
	return nil, nil
}

func (m *mockProfileUsecase) DeleteAvatar(ctx context.Context, profileID uuid.UUID) error {
	if m.deleteAvatarFunc != nil {
		return m.deleteAvatarFunc(ctx, profileID)
	}
	return nil
}

func (m *mockProfileUsecase) ChangePassword(ctx context.Context, userID uuid.UUID, oldPassword, newPassword string) (*models.Profile, error) {
	if m.changePasswordFunc != nil {
		return m.changePasswordFunc(ctx, userID, oldPassword, newPassword)
	}
	return nil, nil
}

func TestGetProfile(t *testing.T) {
	tests := []struct {
		name           string
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "Success",
			contextUserID: uuid.New(),
			setupMock: func(mock *mockProfileUsecase) {
				mock.getProfileFunc = func(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
					return &models.Profile{
						ID:       userID,
						Username: "testuser",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   `"username":"testuser"`,
		},
		{
			name:           "Unauthorized - Invalid UserID",
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name:          "Internal Server Error",
			contextUserID: uuid.New(),
			setupMock: func(mock *mockProfileUsecase) {
				mock.getProfileFunc = func(ctx context.Context, userID uuid.UUID) (*models.Profile, error) {
					return nil, errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{CookieName: "token", Secure: false})

			req := httptest.NewRequest(http.MethodGet, "/profile", nil)
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.GetProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	userID := uuid.New()
	validProfile := ProfileRequest{Username: "newusername"}

	tests := []struct {
		name           string
		body           io.Reader
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "Success",
			body:          createJSONBody(validProfile),
			contextUserID: userID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.updateProfileFunc = func(ctx context.Context, id uuid.UUID, profile models.Profile) (*models.Profile, error) {
					return &models.Profile{
						ID:       id,
						Username: "newusername",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
			expectedBody:   "newusername",
		},
		{
			// В реальном коде при nil body не проверяется отдельно,
			// а ошибка возникает при попытке распарсить JSON
			name:           "Bad Request - Missing Body",
			body:           nil,
			contextUserID:  userID,
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидные данные профиля", // реальная ошибка из кода
		},
		{
			name:           "Unauthorized - Invalid UserID",
			body:           createJSONBody(validProfile),
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name:           "Bad Request - Invalid JSON",
			body:           strings.NewReader(`{invalid json`),
			contextUserID:  userID,
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидные данные профиля",
		},
		{
			name:          "Bad Request - Invalid Profile Data",
			body:          createJSONBody(validProfile),
			contextUserID: userID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.updateProfileFunc = func(ctx context.Context, id uuid.UUID, profile models.Profile) (*models.Profile, error) {
					return nil, profiles.ErrInvalidProfileData
				}
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидные данные профиля",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{})

			req := httptest.NewRequest(http.MethodPut, "/profile", tt.body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.UpdateProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

func TestDeleteProfile(t *testing.T) {
	tests := []struct {
		name           string
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
	}{
		{
			name:          "Success",
			contextUserID: uuid.New(),
			setupMock: func(mock *mockProfileUsecase) {
				mock.deleteProfileFunc = func(ctx context.Context, userID uuid.UUID) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "Unauthorized",
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "Internal Server Error",
			contextUserID: uuid.New(),
			setupMock: func(mock *mockProfileUsecase) {
				mock.deleteProfileFunc = func(ctx context.Context, userID uuid.UUID) error {
					return errors.New("database error")
				}
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{CookieName: "token", Secure: false})

			req := httptest.NewRequest(http.MethodDelete, "/profile", nil)
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.DeleteProfile(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetAvatar(t *testing.T) {
	avatarID := uuid.New()
	profileID := uuid.New()

	tests := []struct {
		name           string
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
	}{
		{
			name:          "Success",
			contextUserID: profileID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.getAvatarFunc = func(ctx context.Context, id uuid.UUID) (*models.Avatar, error) {
					return &models.Avatar{
						ID:        avatarID,
						ProfileID: id,
						AvatarURL: "http://example.com/avatar.jpg",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:          "Avatar Not Found",
			contextUserID: profileID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.getAvatarFunc = func(ctx context.Context, id uuid.UUID) (*models.Avatar, error) {
					return nil, profiles.ErrAvatarNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unauthorized",
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{})

			req := httptest.NewRequest(http.MethodGet, "/profile/avatar", nil)
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.GetAvatar(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUploadAvatar(t *testing.T) {
	profileID := uuid.New()

	tests := []struct {
		name           string
		contextUserID  interface{}
		fileContent    []byte
		fileName       string
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
	}{
		{
			name:          "Success",
			contextUserID: profileID,
			fileContent:   []byte{0xFF, 0xD8, 0xFF, 0xE0}, // JPEG magic bytes
			fileName:      "test.jpg",
			setupMock: func(mock *mockProfileUsecase) {
				mock.uploadAvatarFunc = func(ctx context.Context, id uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Avatar, error) {
					return &models.Avatar{
						ID:        uuid.New(),
						ProfileID: id,
						AvatarURL: "http://example.com/avatar.jpg",
					}, nil
				}
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Unauthorized",
			contextUserID:  "invalid",
			fileContent:    []byte{0xFF, 0xD8, 0xFF, 0xE0},
			fileName:       "test.jpg",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid MIME Type",
			contextUserID:  profileID,
			fileContent:    []byte{0x25, 0x50, 0x44, 0x46}, // PDF magic bytes
			fileName:       "test.pdf",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{})

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", tt.fileName)
			part.Write(tt.fileContent)
			writer.Close()

			req := httptest.NewRequest(http.MethodPost, "/profile/avatar", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.UploadAvatar(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	profileID := uuid.New()

	tests := []struct {
		name           string
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
	}{
		{
			name:          "Success",
			contextUserID: profileID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.deleteAvatarFunc = func(ctx context.Context, id uuid.UUID) error {
					return nil
				}
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:          "Avatar Not Found",
			contextUserID: profileID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.deleteAvatarFunc = func(ctx context.Context, id uuid.UUID) error {
					return profiles.ErrAvatarNotFound
				}
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "Unauthorized",
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{})

			req := httptest.NewRequest(http.MethodDelete, "/profile/avatar", nil)
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.DeleteAvatar(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestChangePassword(t *testing.T) {
	userID := uuid.New()
	validPassword := UpdatePasswordRequest{OldPassword: "oldpass123", NewPassword: "newpass456"}

	tests := []struct {
		name           string
		body           io.Reader
		contextUserID  interface{}
		setupMock      func(mock *mockProfileUsecase)
		expectedStatus int
		expectedBody   string
	}{
		{
			name:          "Success",
			body:          createJSONBody(validPassword),
			contextUserID: userID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.changePasswordFunc = func(ctx context.Context, id uuid.UUID, old, new string) (*models.Profile, error) {
					return &models.Profile{
						ID:       id,
						Username: "testuser",
					}, nil
				}
			},
			expectedStatus: http.StatusOK,
		},
		{
			// В реальном коде при nil body не проверяется отдельно,
			// а ошибка возникает при попытке распарсить JSON
			name:           "Bad Request - Missing Body",
			body:           nil,
			contextUserID:  userID,
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody:   "Невалидные данные пароля", // реальная ошибка из кода
		},
		{
			name:           "Unauthorized",
			body:           createJSONBody(validPassword),
			contextUserID:  "invalid",
			setupMock:      func(mock *mockProfileUsecase) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody:   "Невалидный UserID",
		},
		{
			name:          "Wrong Password",
			body:          createJSONBody(validPassword),
			contextUserID: userID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.changePasswordFunc = func(ctx context.Context, id uuid.UUID, old, new string) (*models.Profile, error) {
					return nil, profiles.ErrWrongPassword
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Неверный пароль",
		},
		{
			name:          "User Not Found",
			body:          createJSONBody(validPassword),
			contextUserID: userID,
			setupMock: func(mock *mockProfileUsecase) {
				mock.changePasswordFunc = func(ctx context.Context, id uuid.UUID, old, new string) (*models.Profile, error) {
					return nil, profiles.ErrUserNotExist
				}
			},
			expectedStatus: http.StatusNotFound,
			expectedBody:   "Пользователь не найден",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUsecase := &mockProfileUsecase{}
			if tt.setupMock != nil {
				tt.setupMock(mockUsecase)
			}

			handler := NewProfileHandler(mockUsecase, config.JWTConfig{})

			req := httptest.NewRequest(http.MethodPut, "/profile/password", tt.body)
			if tt.body != nil {
				req.Header.Set("Content-Type", "application/json")
			}
			ctx := context.WithValue(req.Context(), types.UserIDKey, tt.contextUserID)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()
			handler.ChangePassword(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedBody != "" {
				assert.Contains(t, w.Body.String(), tt.expectedBody)
			}
		})
	}
}

// Helper functions
type ProfileRequest struct {
	Username string `json:"username"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func createJSONBody(data interface{}) io.Reader {
	jsonData, _ := json.Marshal(data)
	return bytes.NewReader(jsonData)
}
