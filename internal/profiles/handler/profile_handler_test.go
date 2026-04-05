package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestProfileHandler_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockProfileUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "auth",
		Secure:     false,
	}

	handler := NewProfileHandler(mockUsecase, jwtConfig)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		expectedProfile := &models.Profile{
			ID:       userID,
			Username: "testuser",
		}

		mockUsecase.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(expectedProfile, nil)

		req := httptest.NewRequest("GET", "/profile", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.Profile
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ID != userID {
			t.Errorf("expected ID %v, got %v", userID, response.ID)
		}
		if response.Username != "testuser" {
			t.Errorf("expected Username 'testuser', got '%s'", response.Username)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/profile", nil)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		req := httptest.NewRequest("GET", "/profile", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.GetProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestProfileHandler_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockProfileUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "auth",
		Secure:     false,
	}

	handler := NewProfileHandler(mockUsecase, jwtConfig)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		updatedProfile := &models.Profile{
			ID:       userID,
			Username: "newusername",
		}

		mockUsecase.EXPECT().
			UpdateProfile(gomock.Any(), userID, gomock.Any()).
			Return(updatedProfile, nil)

		profileDTO := dto.Profile{
			ID:       userID,
			Username: "newusername",
		}
		bodyBytes, _ := json.Marshal(profileDTO)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.Profile
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Username != "newusername" {
			t.Errorf("expected Username 'newusername', got '%s'", response.Username)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/profile", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader([]byte("invalid")))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		profileDTO := dto.Profile{
			Username: "test",
		}
		bodyBytes, _ := json.Marshal(profileDTO)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateProfile(gomock.Any(), userID, gomock.Any()).
			Return(nil, errors.New("db error"))

		profileDTO := dto.Profile{
			Username: "newusername",
		}
		bodyBytes, _ := json.Marshal(profileDTO)

		req := httptest.NewRequest("PUT", "/profile", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.UpdateProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestProfileHandler_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockProfileUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "auth",
		Secure:     false,
	}

	handler := NewProfileHandler(mockUsecase, jwtConfig)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(nil)

		req := httptest.NewRequest("DELETE", "/profile", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		cookies := w.Result().Cookies()
		found := false
		for _, c := range cookies {
			if c.Name == "auth" && c.Value == "" && c.MaxAge == -1 {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected auth cookie to be deleted")
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := httptest.NewRequest("DELETE", "/profile", nil)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(errors.New("db error"))

		req := httptest.NewRequest("DELETE", "/profile", nil)
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		handler.DeleteProfile(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
