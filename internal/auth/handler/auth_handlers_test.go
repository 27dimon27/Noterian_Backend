package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestAuthHandler_SignupUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName:    "auth",
		CookieTimeJWT: 3600,
		Secret:        "secret-key",
		Secure:        false,
	}

	handler := NewAuthHandler(mockUsecase, jwtConfig)
	if handler == nil {
		t.Error("expected handler to be created, got nil")
	}

	validLogin := "testuser"
	validPassword := "validPassword123"
	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: validLogin,
	}

	t.Run("successful signup", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(gomock.Any(), validLogin, validPassword).
			Return(expectedUser, nil)

		signUpData := dto.SignUpUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signUpData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.UserResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ID != userID.String() {
			t.Errorf("expected id %s, got %s", userID.String(), response.ID)
		}
		if response.Username != validLogin {
			t.Errorf("expected login %s, got %s", validLogin, response.Username)
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
		if cookies[0].Name != jwtConfig.CookieName {
			t.Errorf("expected cookie name %s, got %s", jwtConfig.CookieName, cookies[0].Name)
		}
	})

	t.Run("user already exists", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, auth.ErrUserExist)

		signUpData := dto.SignUpUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signUpData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusConflict {
			t.Errorf("expected status StatusConflict, got %d", w.Code)
		}
	})

	t.Run("invalid login", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(gomock.Any(), "ab", validPassword).
			Return(nil, auth.ErrInvalidUsername)

		signUpData := dto.SignUpUser{
			Username: "ab",
			Password: validPassword,
		}
		
		body, err := json.Marshal(signUpData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid password", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(gomock.Any(), validLogin, "abc").
			Return(nil, auth.ErrInvalidPassword)

		signUpData := dto.SignUpUser{
			Username: validLogin,
			Password: "abc",
		}
		
		body, err := json.Marshal(signUpData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/signup", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, auth.ErrInternal)

		signUpData := dto.SignUpUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signUpData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signup", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
	t.Run("nil body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/signup", nil)
		req.Body = nil
		w := httptest.NewRecorder()

		handler.SignupUser(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}

func TestAuthHandler_SigninUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName:    "auth",
		CookieTimeJWT: 3600,
		Secret:        "secret-key",
		Secure:        false,
	}

	handler := NewAuthHandler(mockUsecase, jwtConfig)
	if handler == nil {
		t.Error("expected handler to be created, got nil")
	}

	validLogin := "testuser"
	validPassword := "validpassword123"
	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: validLogin,
	}

	t.Run("successful signin", func(t *testing.T) {
		mockUsecase.EXPECT().
			ValidateUser(gomock.Any(), validLogin, validPassword).
			Return(expectedUser, nil)

		signInData := dto.SignInUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signInData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signin", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.UserResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ID != userID.String() {
			t.Errorf("expected id %s, got %s", userID.String(), response.ID)
		}
		if response.Username != validLogin {
			t.Errorf("expected login %s, got %s", validLogin, response.Username)
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
		if cookies[0].Name != jwtConfig.CookieName {
			t.Errorf("expected cookie name %s, got %s", jwtConfig.CookieName, cookies[0].Name)
		}
	})

	t.Run("invalid credentials", func(t *testing.T) {
		mockUsecase.EXPECT().
			ValidateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, auth.ErrBadCredentials)

		signInData := dto.SignInUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signInData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signin", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			ValidateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, auth.ErrUserNotExist)

		signInData := dto.SignInUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signInData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signin", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("invalid request body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/signin", strings.NewReader("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().
			ValidateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, auth.ErrInternal)

		signInData := dto.SignInUser{
			Username: validLogin,
			Password: validPassword,
		}
		
		body, err := json.Marshal(signInData)
		if err != nil {
			t.Errorf("error in parsing json dto")
		}

		req := httptest.NewRequest("POST", "/signin", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("nil body", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/signin", nil)
		req.Body = nil
		w := httptest.NewRecorder()

		handler.SigninUser(w, req)

		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status 405, got %d", w.Code)
		}
	})
}

func TestAuthHandler_LogOutUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockAuthUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "auth",
		Secure:     false,
	}

	handler := &AuthHandler{
		authUsecase: mockUsecase,
		jwtConfig:   jwtConfig,
	}

	t.Run("successful logout", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/logout", nil)
		w := httptest.NewRecorder()

		handler.LogOutUser(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}

		cookies := w.Result().Cookies()
		if len(cookies) != 1 {
			t.Errorf("expected 1 cookie, got %d", len(cookies))
		}
		if cookies[0].Name != jwtConfig.CookieName {
			t.Errorf("expected cookie name %s, got %s", jwtConfig.CookieName, cookies[0].Name)
		}
		if cookies[0].Value != "" {
			t.Errorf("expected empty cookie value, got %s", cookies[0].Value)
		}
		if cookies[0].MaxAge != -1 {
			t.Errorf("expected max age -1, got %d", cookies[0].MaxAge)
		}
	})
}
