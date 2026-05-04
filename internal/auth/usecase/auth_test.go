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
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupTestUsecase(t *testing.T) (*authUsecase, *mocks.MockUserRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName: "test_cookie",
		CookieTime: 3600,
		Secret:     "test_secret",
	}

	usecase, err := NewAuthUsecase(mockRepo, jwtConfig)
	if err != nil {
		t.Fatalf("Failed to create usecase: %v", err)
	}

	return usecase, mockRepo, ctrl
}

func TestSignupUser_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "valid_user"
	password := "Test1234"

	expectedUser := &models.Profile{
		ID:       uuid.New(),
		Username: username,
	}

	mockRepo.EXPECT().
		CreateUser(ctx, username, password).
		Return(expectedUser, nil)

	user, err := usecase.SignupUser(ctx, username, password)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Error("Expected user, got nil")
	}
	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}
}

func TestSignupUser_InvalidUsername(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"starts with underscore", "_user", "Test1234"},
		{"starts with dot", ".user", "Test1234"},
		{"ends with underscore", "user_", "Test1234"},
		{"ends with dot", "user.", "Test1234"},
		{"contains double underscore", "user__name", "Test1234"},
		{"contains invalid chars", "user@name", "Test1234"},
		{"empty username", "", "Test1234"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := usecase.SignupUser(ctx, tc.username, tc.password)

			if err != auth.ErrInvalidUsername {
				t.Errorf("Expected ErrInvalidUsername, got %v", err)
			}
			if user != nil {
				t.Error("Expected nil user")
			}
		})
	}

	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
}

func TestSignupUser_InvalidPassword(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"too short", "validuser", "Tes"},
		{"no uppercase", "validuser", "test1234"},
		{"no digit", "validuser", "Testtest"},
		{"empty password", "validuser", ""},
		{"only lowercase", "validuser", "test"},
		{"only uppercase", "validuser", "TEST"},
		{"only digits", "validuser", "123456"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			user, err := usecase.SignupUser(ctx, tc.username, tc.password)

			if err != auth.ErrInvalidPassword {
				t.Errorf("Expected ErrInvalidPassword, got %v", err)
			}
			if user != nil {
				t.Error("Expected nil user")
			}
		})
	}

	mockRepo.EXPECT().CreateUser(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)
}

func TestSignupUser_UserAlreadyExists(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "existinguser"
	password := "Test1234"

	mockRepo.EXPECT().
		CreateUser(ctx, username, password).
		Return(nil, auth.ErrUserExist)

	user, err := usecase.SignupUser(ctx, username, password)

	if err != auth.ErrUserExist {
		t.Errorf("Expected ErrUserExist, got %v", err)
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

func TestSignupUser_RepositoryError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"
	expectedErr := errors.New("database error")

	mockRepo.EXPECT().
		CreateUser(ctx, username, password).
		Return(nil, expectedErr)

	user, err := usecase.SignupUser(ctx, username, password)

	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

func TestSigninUser_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	expectedUser := &models.Profile{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
	}

	mockRepo.EXPECT().
		GetUserByUsername(ctx, username).
		Return(expectedUser, nil)

	user, err := usecase.SigninUser(ctx, username, password)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	if user == nil {
		t.Error("Expected user, got nil")
	}
	if user.Username != username {
		t.Errorf("Expected username %s, got %s", username, user.Username)
	}
}

func TestSigninUser_WrongPassword(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	correctPassword := "Test1234"
	wrongPassword := "Wrong4567"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(correctPassword), bcrypt.DefaultCost)

	expectedUser := &models.Profile{
		ID:       uuid.New(),
		Username: username,
		Password: hashedPassword,
	}

	mockRepo.EXPECT().
		GetUserByUsername(ctx, username).
		Return(expectedUser, nil)

	user, err := usecase.SigninUser(ctx, username, wrongPassword)

	if err != auth.ErrBadCredentials {
		t.Errorf("Expected ErrBadCredentials, got %v", err)
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

func TestSigninUser_UserNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "nonexistent"
	password := "Test1234"

	mockRepo.EXPECT().
		GetUserByUsername(ctx, username).
		Return(nil, auth.ErrUserNotExist)

	user, err := usecase.SigninUser(ctx, username, password)

	if err != auth.ErrUserNotExist {
		t.Errorf("Expected ErrUserNotExist, got %v", err)
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

func TestSigninUser_RepositoryError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()
	username := "testuser"
	password := "Test1234"

	expectedErr := errors.New("database connection error")

	mockRepo.EXPECT().
		GetUserByUsername(ctx, username).
		Return(nil, expectedErr)

	user, err := usecase.SigninUser(ctx, username, password)

	if err != expectedErr {
		t.Errorf("Expected %v, got %v", expectedErr, err)
	}
	if user != nil {
		t.Error("Expected nil user")
	}
}

func TestSignupUser_ValidPasswords(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	ctx := context.Background()

	testCases := []struct {
		name     string
		username string
		password string
	}{
		{"minimal valid", "user1", "Te1s"},
		{"with special chars", "user2", "Test123!@#"},
		{"cyrillic uppercase", "user3", "Тест1234"},
		{"long password", "user4", "MyVeryLongPassword123"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			expectedUser := &models.Profile{
				ID:       uuid.New(),
				Username: tc.username,
			}

			mockRepo.EXPECT().
				CreateUser(ctx, tc.username, tc.password).
				Return(expectedUser, nil)

			user, err := usecase.SignupUser(ctx, tc.username, tc.password)
			if err != nil {
				t.Errorf("Expected no error for password '%s', got %v", tc.password, err)
			}
			if user == nil {
				t.Error("Expected user, got nil")
			}
		})
	}
}
