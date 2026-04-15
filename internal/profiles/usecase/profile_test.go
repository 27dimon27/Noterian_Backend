package usecase

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupTestUsecase(t *testing.T) (*profileUsecase, *mocks.MockProfileRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockProfileRepository(ctrl)
	usecase := NewProfileUsecase(mockRepo)
	return usecase, mockRepo, ctrl
}

func TestProfileUsecase_GetProfile(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedProfile := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(expectedProfile, nil)

		profile, err := usecase.GetProfile(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile, profile)
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetProfile(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		profile, err := usecase.GetProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, profile)
	})
}

func TestProfileUsecase_UpdateProfile(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	validProfile := models.Profile{
		ID:       userID,
		Username: "validuser",
	}
	updatedProfile := &models.Profile{
		ID:       userID,
		Username: "validuser",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			UpdateProfile(gomock.Any(), userID, validProfile).
			Return(updatedProfile, nil)

		profile, err := usecase.UpdateProfile(context.Background(), userID, validProfile)

		assert.NoError(t, err)
		assert.Equal(t, updatedProfile, profile)
	})

	t.Run("empty username", func(t *testing.T) {
		invalidProfile := models.Profile{
			ID:       userID,
			Username: "",
		}

		profile, err := usecase.UpdateProfile(context.Background(), userID, invalidProfile)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrInvalidProfileData, err)
		assert.Nil(t, profile)
	})
}

func TestProfileUsecase_DeleteProfile(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(nil)

		err := usecase.DeleteProfile(context.Background(), userID)

		assert.NoError(t, err)
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteProfile(gomock.Any(), userID).
			Return(errors.New("delete failed"))

		err := usecase.DeleteProfile(context.Background(), userID)

		assert.Error(t, err)
	})
}

func TestProfileUsecase_GetAvatar(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	profileID := uuid.New()
	expectedAvatar := &models.Avatar{
		ID:        uuid.New(),
		ProfileID: profileID,
		AvatarURL: "https://example.com/avatar.jpg",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAvatar(gomock.Any(), profileID).
			Return(expectedAvatar, nil)

		avatar, err := usecase.GetAvatar(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Equal(t, expectedAvatar, avatar)
	})

	t.Run("avatar not found - returns nil", func(t *testing.T) {
		mockRepo.EXPECT().
			GetAvatar(gomock.Any(), profileID).
			Return(nil, nil)

		avatar, err := usecase.GetAvatar(context.Background(), profileID)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrAvatarNotFound, err)
		assert.Nil(t, avatar)
	})
}

func TestProfileUsecase_UploadAvatar(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	profileID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte("fake image data"))

	expectedAvatar := &models.Avatar{
		ID:        uuid.New(),
		ProfileID: profileID,
		AvatarURL: "https://example.com/avatar.jpg",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			UploadAvatar(gomock.Any(), profileID, fileName, fileSize, mimeType, gomock.Any()).
			Return(expectedAvatar, nil)

		avatar, err := usecase.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

		assert.NoError(t, err)
		assert.Equal(t, expectedAvatar, avatar)
	})

	t.Run("repository returns error", func(t *testing.T) {
		mockRepo.EXPECT().
			UploadAvatar(gomock.Any(), profileID, fileName, fileSize, mimeType, gomock.Any()).
			Return(nil, errors.New("upload failed"))

		avatar, err := usecase.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

		assert.Error(t, err)
		assert.Nil(t, avatar)
	})
}

func TestProfileUsecase_DeleteAvatar(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	profileID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			DeleteAvatar(gomock.Any(), profileID).
			Return(nil)

		err := usecase.DeleteAvatar(context.Background(), profileID)

		assert.NoError(t, err)
	})
}

func TestProfileUsecase_ChangePassword(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correct123"), bcrypt.DefaultCost)
	updatedProfile := &models.Profile{
		ID:       userID,
		Username: "testuser",
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().
			GetPassword(gomock.Any(), userID).
			Return(hashedPassword, nil)

		mockRepo.EXPECT().
			ChangePassword(gomock.Any(), userID, "new123").
			Return(updatedProfile, nil)

		profile, err := usecase.ChangePassword(context.Background(), userID, "correct123", "new123")

		assert.NoError(t, err)
		assert.Equal(t, updatedProfile, profile)
	})

	t.Run("wrong old password", func(t *testing.T) {
		mockRepo.EXPECT().
			GetPassword(gomock.Any(), userID).
			Return(hashedPassword, nil)

		profile, err := usecase.ChangePassword(context.Background(), userID, "wrong123", "new123")

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrWrongPassword, err)
		assert.Nil(t, profile)
	})

	t.Run("user not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetPassword(gomock.Any(), userID).
			Return(nil, profiles.ErrUserNotExist)

		profile, err := usecase.ChangePassword(context.Background(), userID, "old123", "new123")

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
		assert.Nil(t, profile)
	})
}
