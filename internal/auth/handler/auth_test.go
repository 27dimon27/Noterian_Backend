package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestHandler(t *testing.T) (*AuthHandler, *mocks.MockAuthUsecase, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockAuthUsecase := mocks.NewMockAuthUsecase(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		CookieName: "test-cookie",
		CookieTime: 24 * time.Hour,
		Secure:     false,
	}

	handler := NewAuthHandler(mockAuthUsecase, jwtConfig)
	return handler, mockAuthUsecase, ctrl
}

func TestAuthHandler_SignupUser_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	signUpData := dto.SignUpUser{
		Username: "testuser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signUpData)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "testuser", "Test1234").
		Return(expectedUser, nil)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check cookie is set
	cookies := resp.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "test-cookie", cookies[0].Name)
	assert.NotEmpty(t, cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)
}

func TestAuthHandler_SignupUser_EmptyBody(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/signup", nil)
	w := httptest.NewRecorder()

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp map[string]string
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, auth.ErrInvalidInput.Error(), errResp["error"])
}

func TestAuthHandler_SignupUser_InvalidJSON(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader([]byte("invalid json")))
	w := httptest.NewRecorder()

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var errResp map[string]string
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, auth.ErrInvalidInput.Error(), errResp["error"])
}

func TestAuthHandler_SignupUser_UserAlreadyExists(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	signUpData := dto.SignUpUser{
		Username: "existinguser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signUpData)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "existinguser", "Test1234").
		Return(nil, auth.ErrUserExist)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	var errResp map[string]string
	err := json.NewDecoder(resp.Body).Decode(&errResp)
	require.NoError(t, err)
	assert.Equal(t, auth.ErrUserExist.Error(), errResp["error"])
}

func TestAuthHandler_SignupUser_InvalidUsername(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	signUpData := dto.SignUpUser{
		Username: "invalid username!",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signUpData)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "invalid username!", "Test1234").
		Return(nil, auth.ErrInvalidUsername)

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAuthHandler_SignupUser_InternalError(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	signUpData := dto.SignUpUser{
		Username: "testuser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signUpData)

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		CreateUser(gomock.Any(), "testuser", "Test1234").
		Return(nil, errors.New("database error"))

	handler.SignupUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

func TestAuthHandler_SigninUser_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	signInData := dto.SignInUser{
		Username: "testuser",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signInData)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "testuser", "Test1234").
		Return(expectedUser, nil)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check response body
	var userResp dto.UserResponse
	err := json.NewDecoder(resp.Body).Decode(&userResp)
	require.NoError(t, err)
	assert.Equal(t, userID.String(), userResp.ID)
	assert.Equal(t, "testuser", userResp.Username)
}

func TestAuthHandler_SigninUser_BadCredentials(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	signInData := dto.SignInUser{
		Username: "testuser",
		Password: "wrongpassword",
	}
	bodyBytes, _ := json.Marshal(signInData)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "testuser", "wrongpassword").
		Return(nil, auth.ErrBadCredentials)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_SigninUser_UserNotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	signInData := dto.SignInUser{
		Username: "nonexistent",
		Password: "Test1234",
	}
	bodyBytes, _ := json.Marshal(signInData)

	req := httptest.NewRequest(http.MethodPost, "/signin", bytes.NewReader(bodyBytes))
	w := httptest.NewRecorder()

	mockUsecase.EXPECT().
		ValidateUser(gomock.Any(), "nonexistent", "Test1234").
		Return(nil, auth.ErrUserNotExist)

	handler.SigninUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthHandler_LogOutUser_Success(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	http.SetCookie(w, &http.Cookie{
		Name:  "test-cookie",
		Value: "some-token",
	})

	handler.LogOutUser(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	cookies := resp.Cookies()
	assert.Len(t, cookies, 2)
	assert.Equal(t, "test-cookie", cookies[0].Name)
	assert.Equal(t, 0, cookies[0].MaxAge)
}
