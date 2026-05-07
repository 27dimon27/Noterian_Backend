package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/auth/usecase/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/config"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupTestUsecase(t *testing.T) (*authUsecase, *mocks.MockUserRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUserRepo := mocks.NewMockUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		Secret:     "test-secret",
		CookieName: "test-cookie",
		CookieTime: 24,
	}

	usecase, err := NewAuthUsecase(mockUserRepo, jwtConfig)
	require.NoError(t, err)

	return usecase, mockUserRepo, ctrl
}

func TestAuthUsecase_SignupUser_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: "validuser",
	}

	mockRepo.EXPECT().
		SignupUser(gomock.Any(), "validuser", "ValidPass123").
		Return(expectedUser, nil)

	user, err := usecase.SignupUser(context.Background(), "validuser", "ValidPass123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestAuthUsecase_SignupUser_InvalidUsername(t *testing.T) {
	usecase, _, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"empty username", "", "ValidPass123"},
		{"username with space", "invalid user", "ValidPass123"},
		{"username with @", "user@name", "ValidPass123"},
		{"starts with underscore", "_username", "ValidPass123"},
		{"ends with underscore", "username_", "ValidPass123"},
		{"double underscore", "user__name", "ValidPass123"},
		{"starts with dot", ".username", "ValidPass123"},
		{"ends with dot", "username.", "ValidPass123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := usecase.SignupUser(context.Background(), tc.username, tc.password)
			assert.Nil(t, user)
			assert.ErrorIs(t, err, auth.ErrInvalidUsername)
		})
	}
}

func TestAuthUsecase_SignupUser_ValidUsernameFormats(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"only letters", "username", "ValidPass123"},
		{"letters and numbers", "user123", "ValidPass123"},
		{"with underscore middle", "user_name", "ValidPass123"},
		{"with dot middle", "user.name", "ValidPass123"},
		{"russian letters", "пользователь", "ValidPass123"},
		{"mixed case", "UserName", "ValidPass123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			expectedUser := &models.Profile{
				ID:       userID,
				Username: tc.username,
			}

			mockRepo.EXPECT().
				SignupUser(gomock.Any(), tc.username, tc.password).
				Return(expectedUser, nil)

			user, err := usecase.SignupUser(context.Background(), tc.username, tc.password)
			assert.NoError(t, err)
			assert.Equal(t, expectedUser, user)
		})
	}
}

func TestAuthUsecase_SignupUser_InvalidPassword(t *testing.T) {
	usecase, _, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"empty password", "validuser", ""},
		{"too short", "validuser", "Ab1"},
		{"no uppercase", "validuser", "pass123"},
		{"no digit", "validuser", "Password"},
		{"all lowercase", "validuser", "password"},
		{"all uppercase", "validuser", "PASSWORD"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := usecase.SignupUser(context.Background(), tc.username, tc.password)
			assert.Nil(t, user)
			assert.ErrorIs(t, err, auth.ErrInvalidPassword)
		})
	}
}

func TestAuthUsecase_SignupUser_ValidPasswords(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"standard password", "user1", "ValidPass123"},
		{"with special chars", "user2", "Valid@Pass123"},
		{"long password", "user3", "VeryLongValidPassword123"},
		{"russian letters", "user4", "ВалидныйПароль123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			userID := uuid.New()
			expectedUser := &models.Profile{
				ID:       userID,
				Username: tc.username,
			}

			mockRepo.EXPECT().
				SignupUser(gomock.Any(), tc.username, tc.password).
				Return(expectedUser, nil)

			user, err := usecase.SignupUser(context.Background(), tc.username, tc.password)
			assert.NoError(t, err)
			assert.Equal(t, expectedUser, user)
		})
	}
}

func TestAuthUsecase_SignupUser_RepositoryError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().
		SignupUser(gomock.Any(), "validuser", "ValidPass123").
		Return(nil, errors.New("database error"))

	user, err := usecase.SignupUser(context.Background(), "validuser", "ValidPass123")

	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestAuthUsecase_SigninUser_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("ValidPass123"), bcrypt.DefaultCost)
	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: "validuser",
		Password: hashedPassword,
	}

	mockRepo.EXPECT().
		SigninUser(gomock.Any(), "validuser").
		Return(expectedUser, nil)

	user, err := usecase.SigninUser(context.Background(), "validuser", "ValidPass123")

	assert.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func TestAuthUsecase_SigninUser_WrongPassword(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("ValidPass123"), bcrypt.DefaultCost)
	user := &models.Profile{
		ID:       uuid.New(),
		Username: "validuser",
		Password: hashedPassword,
	}

	mockRepo.EXPECT().
		SigninUser(gomock.Any(), "validuser").
		Return(user, nil)

	validatedUser, err := usecase.SigninUser(context.Background(), "validuser", "WrongPass123")

	assert.Nil(t, validatedUser)
	assert.ErrorIs(t, err, auth.ErrBadCredentials)
}

func TestAuthUsecase_SigninUser_UserNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().
		SigninUser(gomock.Any(), "nonexistent").
		Return(nil, auth.ErrUserNotExist)

	user, err := usecase.SigninUser(context.Background(), "nonexistent", "ValidPass123")

	assert.Nil(t, user)
	assert.ErrorIs(t, err, auth.ErrUserNotExist)
}

func TestAuthUsecase_SigninUser_RepositoryError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	mockRepo.EXPECT().
		SigninUser(gomock.Any(), "validuser").
		Return(nil, errors.New("database error"))

	user, err := usecase.SigninUser(context.Background(), "validuser", "ValidPass123")

	assert.Nil(t, user)
	assert.Error(t, err)
}

func TestAuthUsecase_SigninUser_EmptyCredentials(t *testing.T) {
	usecase, _, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	// Note: SigninUser doesn't validate credentials itself,
	// it relies on the repository to handle empty values
	mockRepo := mocks.NewMockUserRepository(ctrl)
	usecase.userRepo = mockRepo

	mockRepo.EXPECT().
		SigninUser(gomock.Any(), "").
		Return(nil, auth.ErrUserNotExist)

	user, err := usecase.SigninUser(context.Background(), "", "")

	assert.Nil(t, user)
	assert.Error(t, err)
}
