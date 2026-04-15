package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestHandler(t *testing.T) (*ProfileHandler, *mocks.MockProfileUsecase, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUsecase := mocks.NewMockProfileUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "auth_token",
		Secure:     false,
	}

	handler := NewProfileHandler(mockUsecase, jwtConfig)
	return handler, mockUsecase, ctrl
}

func createContextWithUserID(userID uuid.UUID) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, types.UserIDKey, userID)
}

func createMultipartRequest(t *testing.T, fileName string, content []byte) (*http.Request, string) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", fileName)
	require.NoError(t, err)

	_, err = part.Write(content)
	require.NoError(t, err)

	err = writer.Close()
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/avatar", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	return req, writer.FormDataContentType()
}

func TestProfileHandler_GetProfile(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedProfile := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(expectedProfile, nil)

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Profile
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, userID, response.ID)
		assert.Equal(t, "testuser", response.Username)
	})

	t.Run("invalid user ID in context", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = req.WithContext(context.Background())
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase returns error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(nil, errors.New("database error"))

		req := httptest.NewRequest(http.MethodGet, "/profile", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	updatedProfile := &models.Profile{
		ID:       userID,
		Username: "newusername",
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateProfile(gomock.Any(), userID, gomock.Any()).
			Return(updatedProfile, nil)

		reqBody := `{"username": "newusername"}`
		req := httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Profile
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "newusername", response.Username)
	})

	t.Run("missing body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/profile", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		reqBody := `{"username": "newusername"}`
		req := httptest.NewRequest(http.MethodPut, "/profile", strings.NewReader(reqBody))
		req = req.WithContext(context.Background())
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestProfileHandler_DeleteProfile(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/profile", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("usecase returns error", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(errors.New("delete failed"))

		req := httptest.NewRequest(http.MethodDelete, "/profile", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestProfileHandler_GetAvatar(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedAvatar := &models.Avatar{
		ID:        uuid.New(),
		ProfileID: userID,
		AvatarURL: "https://example.com/avatar.jpg",
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetAvatar(gomock.Any(), userID).
			Return(expectedAvatar, nil)

		req := httptest.NewRequest(http.MethodGet, "/avatar", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.GetAvatar(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.Avatar
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, expectedAvatar.AvatarURL, response.AvatarURL)
	})

	t.Run("avatar not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetAvatar(gomock.Any(), userID).
			Return(nil, profiles.ErrAvatarNotFound)

		req := httptest.NewRequest(http.MethodGet, "/avatar", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.GetAvatar(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProfileHandler_UploadAvatar(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	avatar := &models.Avatar{
		ID:        uuid.New(),
		ProfileID: userID,
		AvatarURL: "https://example.com/new-avatar.jpg",
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			UploadAvatar(gomock.Any(), userID, gomock.Any(), gomock.Any(), "image/png", gomock.Any()).
			Return(avatar, nil)

		pngData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
		req, _ := createMultipartRequest(t, "test.png", pngData)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UploadAvatar(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)

		var response dto.Avatar
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, avatar.AvatarURL, response.AvatarURL)
	})

	t.Run("file too large", func(t *testing.T) {
		largeData := make([]byte, profiles.MAX_FILE_SIZE+1)
		req, _ := createMultipartRequest(t, "large.jpg", largeData)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UploadAvatar(w, req)

		assert.Equal(t, http.StatusRequestEntityTooLarge, w.Code)
	})

	t.Run("no file in request", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		writer.Close()

		req := httptest.NewRequest(http.MethodPost, "/avatar", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UploadAvatar(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})

	t.Run("invalid mime type", func(t *testing.T) {
		req, _ := createMultipartRequest(t, "test.txt", []byte("text content"))
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.UploadAvatar(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		req, _ := createMultipartRequest(t, "test.png", []byte("fake"))
		req = req.WithContext(context.Background())
		w := httptest.NewRecorder()

		handler.UploadAvatar(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}

func TestProfileHandler_DeleteAvatar(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAvatar(gomock.Any(), userID).
			Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/avatar", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.DeleteAvatar(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("avatar not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteAvatar(gomock.Any(), userID).
			Return(profiles.ErrAvatarNotFound)

		req := httptest.NewRequest(http.MethodDelete, "/avatar", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.DeleteAvatar(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestProfileHandler_ChangePassword(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	updatedProfile := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			ChangePassword(gomock.Any(), userID, "old123", "new123").
			Return(updatedProfile, nil)

		reqBody := `{"old_password": "old123", "new_password": "new123"}`
		req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.ChangePassword(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("missing body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/profile/password", nil)
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.ChangePassword(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("wrong password", func(t *testing.T) {
		mockUsecase.EXPECT().
			ChangePassword(gomock.Any(), userID, "wrong", "new123").
			Return(nil, profiles.ErrWrongPassword)

		reqBody := `{"old_password": "wrong", "new_password": "new123"}`
		req := httptest.NewRequest(http.MethodPut, "/profile/password", strings.NewReader(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(createContextWithUserID(userID))
		w := httptest.NewRecorder()

		handler.ChangePassword(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
