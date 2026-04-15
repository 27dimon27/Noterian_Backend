package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func setupTestHandler(t *testing.T) (*AuthHandler, *mocks.MockAuthUsecase, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUsecase := mocks.NewMockAuthUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "test_cookie",
		CookieTime: 3600,
		Secret:     "test_secret",
		Secure:     false,
	}

	handler := NewAuthHandler(mockUsecase, jwtConfig)
	return handler, mockUsecase, ctrl
}

func TestSignupUser_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	reqBody := dto.SignUpUser{
		Username: "testuser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	expectedUser := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "testuser", "Test1234").
		Return(expectedUser, nil)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var userResp dto.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if userResp.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", userResp.Username)
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("Expected cookie to be set")
	}
}

func TestSignupUser_EmptyBody(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/signup", nil)
	w := httptest.NewRecorder()

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %v", resp.Status)
	}
}

func TestSignupUser_InvalidJSON(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader([]byte("{invalid json")))
	w := httptest.NewRecorder()

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %v", resp.Status)
	}
}

func TestSignupUser_UserAlreadyExists(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	reqBody := dto.SignUpUser{
		Username: "existinguser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "existinguser", "Test1234").
		Return(nil, auth.ErrUserExist)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusConflict {
		t.Errorf("Expected status Conflict, got %v", resp.Status)
	}
}

func TestSignupUser_InvalidUsername(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	reqBody := dto.SignUpUser{
		Username: "invalid@username",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "invalid@username", "Test1234").
		Return(nil, auth.ErrInvalidUsername)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected status BadRequest, got %v", resp.Status)
	}
}

func TestSigninUser_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	reqBody := dto.SignInUser{
		Username: "testuser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	expectedUser := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "testuser", "Test1234").
		Return(expectedUser, nil)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status OK, got %v", resp.Status)
	}

	var userResp dto.UserResponse
	if err := json.NewDecoder(resp.Body).Decode(&userResp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if userResp.Username != "testuser" {
		t.Errorf("Expected username testuser, got %s", userResp.Username)
	}
}

func TestSigninUser_InvalidCredentials(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	reqBody := dto.SignInUser{
		Username: "testuser",
		Password: "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "testuser", "wrongpassword").
		Return(nil, auth.ErrBadCredentials)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", resp.Status)
	}
}

func TestSigninUser_UserNotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	reqBody := dto.SignInUser{
		Username: "nonexistent",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "nonexistent", "Test1234").
		Return(nil, auth.ErrUserNotExist)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected status Unauthorized, got %v", resp.Status)
	}
}

func TestLogOutUser(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	handler.LogOutUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("Expected status NoContent, got %v", resp.Status)
	}

	cookies := resp.Cookies()
	if len(cookies) == 0 {
		t.Error("Expected logout cookie to be set")
	}

	if cookies[0].Value != "" {
		t.Error("Expected empty cookie value")
	}

	if cookies[0].MaxAge != -1 {
		t.Errorf("Expected MaxAge -1, got %d", cookies[0].MaxAge)
	}
}
