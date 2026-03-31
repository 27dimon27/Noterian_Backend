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

func TestCreateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName:    "auth",
		CookieTimeJWT: 3600,
		Secret:        "secret-key",
		Secure:        false,
	}

	usecase, err := NewAuthUsecase(mockRepo, jwtConfig)
	if err != nil {
		t.Fatalf("cant create usecase: %s", err)
	}

	validLogin := "testuser"
	validPassword := "validPassword123"
	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: validLogin,
	}

	t.Run("success creation", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateUser(gomock.Any(), validLogin, validPassword).
			Return(expectedUser, nil)

		user, err := usecase.CreateUser(context.Background(), validLogin, validPassword)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.ID != userID {
			t.Errorf("expected id %v, got %v", userID, user.ID)
		}
		if user.Username != validLogin {
			t.Errorf("expected username %s, got %s", validLogin, user.Username)
		}
	})

	t.Run("invalid login - empty", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), "", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid login - invalid characters", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), "test@user", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid login - starts with underscore", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), "_testuser", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid login - starts with dot", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), ".testuser", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid login - ends with underscore", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), "testuser_", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid login - contains double underscore", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), "test__user", validPassword)
		if !errors.Is(err, auth.ErrInvalidLogin) {
			t.Errorf("expected ErrInvalidLogin, got %v", err)
		}
	})

	t.Run("invalid password - empty", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), validLogin, "")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("invalid password - too short", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), validLogin, "abc")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("invalid password - no uppercase", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), validLogin, "validpassword123")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("invalid password - no digit", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), validLogin, "validPassword")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("invalid password - no uppercase and no digit", func(t *testing.T) {
		_, err := usecase.CreateUser(context.Background(), validLogin, "valide")
		if !errors.Is(err, auth.ErrInvalidPassword) {
			t.Errorf("expected ErrInvalidPassword, got %v", err)
		}
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			CreateUser(gomock.Any(), validLogin, validPassword).
			Return(nil, errors.New("db error"))

		_, err := usecase.CreateUser(context.Background(), validLogin, validPassword)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestValidateUser(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockUserRepository(ctrl)

	jwtConfig := config.JWTConfig{
		CookieName:    "auth",
		CookieTimeJWT: 3600,
		Secret:        "secret-key",
		Secure:        false,
	}

	usecase, err := NewAuthUsecase(mockRepo, jwtConfig)
	if err != nil {
		t.Fatalf("cant create usecase: %s", err)
	}

	validLogin := "testuser"
	validPassword := "validPassword123"

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(validPassword), bcrypt.DefaultCost)

	userID := uuid.New()
	expectedUser := &models.Profile{
		ID:       userID,
		Username: validLogin,
		Password: hashedPassword,
	}

	t.Run("success validation", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), validLogin).
			Return(expectedUser, nil)

		user, err := usecase.ValidateUser(context.Background(), validLogin, validPassword)
		if err != nil {
			t.Errorf("unexpected err: %s", err)
			return
		}

		if user.ID != userID {
			t.Errorf("expected id %v, got %v", userID, user.ID)
		}
		if user.Username != validLogin {
			t.Errorf("expected username %s, got %s", validLogin, user.Username)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), validLogin).
			Return(nil, auth.ErrUserNotExist)

		_, err := usecase.ValidateUser(context.Background(), validLogin, validPassword)
		if !errors.Is(err, auth.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})

	t.Run("wrong password", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), validLogin).
			Return(expectedUser, nil)

		_, err := usecase.ValidateUser(context.Background(), validLogin, "wrongPassword123")
		if !errors.Is(err, auth.ErrBadCredentials) {
			t.Errorf("expected ErrBadCredentials, got %v", err)
		}
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetUserByLogin(gomock.Any(), validLogin).
			Return(nil, errors.New("db error"))

		_, err := usecase.ValidateUser(context.Background(), validLogin, validPassword)
		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
