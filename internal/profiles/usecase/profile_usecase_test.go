package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestProfileUsecase_GetProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	usecase := NewProfileUsecase(mockRepo)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedProfile := &models.Profile{
			ID:       userID,
			Username: "testuser",
		}

		mockRepo.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(expectedProfile, nil)

		profile, err := usecase.GetProfile(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if profile.ID != userID {
			t.Errorf("expected ID %v, got %v", userID, profile.ID)
		}
		if profile.Username != "testuser" {
			t.Errorf("expected Username 'testuser', got '%s'", profile.Username)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetProfile(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(nil, profiles.ErrUserNotExist)

		_, err := usecase.GetProfile(context.Background(), userID)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	usecase := NewProfileUsecase(mockRepo)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}
		updatedProfile := &models.Profile{
			ID:       userID,
			Username: "newusername",
		}

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), userID, profile).
			Return(updatedProfile, nil)

		result, err := usecase.UpdateProfile(context.Background(), userID, profile)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Username != "newusername" {
			t.Errorf("expected Username 'newusername', got '%s'", result.Username)
		}
	})

	t.Run("empty username", func(t *testing.T) {
		profile := models.Profile{
			Username: "",
		}

		_, err := usecase.UpdateProfile(context.Background(), userID, profile)

		if !errors.Is(err, profiles.ErrInvalidProfileData) {
			t.Errorf("expected ErrInvalidProfileData, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), userID, profile).
			Return(nil, errors.New("db error"))

		_, err := usecase.UpdateProfile(context.Background(), userID, profile)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		profile := models.Profile{
			Username: "newusername",
		}

		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), userID, profile).
			Return(nil, profiles.ErrUserNotExist)

		_, err := usecase.UpdateProfile(context.Background(), userID, profile)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})
}

func TestProfileUsecase_DeleteProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockProfileRepository(ctrl)
	usecase := NewProfileUsecase(mockRepo)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(nil)

		err := usecase.DeleteProfile(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(errors.New("db error"))

		err := usecase.DeleteProfile(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(profiles.ErrUserNotExist)

		err := usecase.DeleteProfile(context.Background(), userID)

		if !errors.Is(err, profiles.ErrUserNotExist) {
			t.Errorf("expected ErrUserNotExist, got %v", err)
		}
	})
}
