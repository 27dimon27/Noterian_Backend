package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles/repository/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func setupTestRepository(t *testing.T) (*profileRepository, sqlmock.Sqlmock, *mocks.MockMinIOService, *gomock.Controller) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	mockMinio := mocks.NewMockMinIOService(ctrl)

	repo := NewProfileRepository(db, mockMinio, "test-bucket")

	return repo, mock, mockMinio, ctrl
}

func quoteSQL(sql string) string {
	return regexp.QuoteMeta(sql)
}

func TestProfileRepository_GetProfile(t *testing.T) {
	repo, mock, _, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	userID := uuid.New()
	expectedProfile := &models.Profile{
		ID:        userID,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(expectedProfile.ID, expectedProfile.Username, expectedProfile.CreatedAt, expectedProfile.UpdatedAt)

		mock.ExpectQuery(quoteSQL(GET_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		profile, err := repo.GetProfile(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile.ID, profile.ID)
		assert.Equal(t, expectedProfile.Username, profile.Username)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(GET_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		profile, err := repo.GetProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
		assert.Nil(t, profile)
	})
}

func TestProfileRepository_UpdateProfile(t *testing.T) {
	repo, mock, _, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	userID := uuid.New()
	profile := models.Profile{
		Username: "newusername",
	}
	expectedProfile := &models.Profile{
		ID:        userID,
		Username:  "newusername",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(expectedProfile.ID, expectedProfile.Username, expectedProfile.CreatedAt, expectedProfile.UpdatedAt)

		mock.ExpectQuery(quoteSQL(UPDATE_PROFILE_BY_USER_ID)).
			WithArgs(userID, profile.Username).
			WillReturnRows(rows)

		updatedProfile, err := repo.UpdateProfile(context.Background(), userID, profile)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile.Username, updatedProfile.Username)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(UPDATE_PROFILE_BY_USER_ID)).
			WithArgs(userID, profile.Username).
			WillReturnError(sql.ErrNoRows)

		updatedProfile, err := repo.UpdateProfile(context.Background(), userID, profile)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
		assert.Nil(t, updatedProfile)
	})
}

func TestProfileRepository_DeleteProfile(t *testing.T) {
	repo, mock, _, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)

		mock.ExpectQuery(quoteSQL(DELETE_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		err := repo.DeleteProfile(context.Background(), userID)

		assert.NoError(t, err)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(DELETE_PROFILE_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		err := repo.DeleteProfile(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
	})
}

func TestProfileRepository_GetAvatar(t *testing.T) {
	repo, mock, mockMinio, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	profileID := uuid.New()
	avatarID := uuid.New()
	minioKey := "test-key"

	t.Run("success with valid URL", func(t *testing.T) {
		expiresAt := time.Now().Add(profiles.PRESIGNED_URL_EXPIRY)
		rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(avatarID, profileID, minioKey, "https://example.com/avatar.jpg", expiresAt, time.Now(), time.Now())

		mock.ExpectQuery(quoteSQL(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		avatar, err := repo.GetAvatar(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Equal(t, avatarID, avatar.ID)
	})

	t.Run("expired URL - regenerates", func(t *testing.T) {
		expiredAt := time.Now().Add(-time.Hour)
		rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(avatarID, profileID, minioKey, "https://example.com/old.jpg", expiredAt, time.Now(), time.Now())

		mock.ExpectQuery(quoteSQL(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), repo.avatarBucket, minioKey, profiles.PRESIGNED_URL_EXPIRY).
			Return("https://example.com/new.jpg", nil)

		updateRows := sqlmock.NewRows([]string{"avatar_url", "url_expires_at", "updated_at"}).
			AddRow("https://example.com/new.jpg", time.Now().Add(profiles.PRESIGNED_URL_EXPIRY), time.Now())

		mock.ExpectQuery(quoteSQL(UPDATE_AVATAR_URL)).
			WithArgs("https://example.com/new.jpg", sqlmock.AnyArg(), sqlmock.AnyArg(), avatarID).
			WillReturnRows(updateRows)

		avatar, err := repo.GetAvatar(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Equal(t, "https://example.com/new.jpg", avatar.AvatarURL)
	})

	t.Run("avatar not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(GET_AVATAR_BY_PROFILE_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		avatar, err := repo.GetAvatar(context.Background(), profileID)

		assert.NoError(t, err)
		assert.Nil(t, avatar)
	})
}

func TestProfileRepository_UploadAvatar(t *testing.T) {
	repo, mock, mockMinio, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	profileID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte("fake image data"))

	t.Run("success - no existing avatar", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), repo.avatarBucket, gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(nil)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), repo.avatarBucket, gomock.Any(), profiles.PRESIGNED_URL_EXPIRY).
			Return("https://example.com/avatar.jpg", nil)

		rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(uuid.New(), profileID, "key", "https://example.com/avatar.jpg", time.Now().Add(profiles.PRESIGNED_URL_EXPIRY), time.Now(), time.Now())

		mock.ExpectQuery(quoteSQL(CREATE_AVATAR)).
			WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		avatar, err := repo.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

		assert.NoError(t, err)
		assert.NotNil(t, avatar)
	})

	t.Run("success - with existing avatar", func(t *testing.T) {
		deleteRows := sqlmock.NewRows([]string{"minio_key"}).AddRow("old-key")
		mock.ExpectQuery(quoteSQL(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnRows(deleteRows)

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), repo.avatarBucket, "old-key").
			Return(nil)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), repo.avatarBucket, gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(nil)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), repo.avatarBucket, gomock.Any(), profiles.PRESIGNED_URL_EXPIRY).
			Return("https://example.com/new-avatar.jpg", nil)

		rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(uuid.New(), profileID, "new-key", "https://example.com/new-avatar.jpg", time.Now().Add(profiles.PRESIGNED_URL_EXPIRY), time.Now(), time.Now())

		mock.ExpectQuery(quoteSQL(CREATE_AVATAR)).
			WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(rows)

		avatar, err := repo.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

		assert.NoError(t, err)
		assert.NotNil(t, avatar)
	})

	t.Run("upload fails - rolls back", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), repo.avatarBucket, gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(errors.New("upload failed"))

		avatar, err := repo.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

		assert.Error(t, err)
		assert.Nil(t, avatar)
	})
}

func TestProfileRepository_DeleteAvatar(t *testing.T) {
	repo, mock, mockMinio, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	profileID := uuid.New()
	minioKey := "test-key"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)

		mock.ExpectQuery(quoteSQL(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnRows(rows)

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), repo.avatarBucket, minioKey).
			Return(nil)

		err := repo.DeleteAvatar(context.Background(), profileID)

		assert.NoError(t, err)
	})

	t.Run("avatar not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(DELETE_AVATAR_BY_ID)).
			WithArgs(profileID).
			WillReturnError(sql.ErrNoRows)

		err := repo.DeleteAvatar(context.Background(), profileID)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrAvatarNotFound, err)
	})
}

func TestProfileRepository_ChangePassword(t *testing.T) {
	repo, mock, _, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	userID := uuid.New()
	newPassword := "newpassword123"

	t.Run("success", func(t *testing.T) {
		expectedProfile := &models.Profile{
			ID:        userID,
			Username:  "testuser",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
			AddRow(expectedProfile.ID, expectedProfile.Username, expectedProfile.CreatedAt, expectedProfile.UpdatedAt)

		mock.ExpectQuery(quoteSQL(CHANGE_PASSWORD_BY_USER_ID)).
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnRows(rows)

		profile, err := repo.ChangePassword(context.Background(), userID, newPassword)

		assert.NoError(t, err)
		assert.Equal(t, expectedProfile.ID, profile.ID)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(CHANGE_PASSWORD_BY_USER_ID)).
			WithArgs(userID, sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		profile, err := repo.ChangePassword(context.Background(), userID, newPassword)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
		assert.Nil(t, profile)
	})
}

func TestProfileRepository_GetPassword(t *testing.T) {
	repo, mock, _, ctrl := setupTestRepository(t)
	defer ctrl.Finish()
	defer repo.db.Close()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"password"}).AddRow(hashedPassword)

		mock.ExpectQuery(quoteSQL(GET_PASSWORD_BY_USER_ID)).
			WithArgs(userID).
			WillReturnRows(rows)

		password, err := repo.GetPassword(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, hashedPassword, password)
	})

	t.Run("user not found", func(t *testing.T) {
		mock.ExpectQuery(quoteSQL(GET_PASSWORD_BY_USER_ID)).
			WithArgs(userID).
			WillReturnError(sql.ErrNoRows)

		password, err := repo.GetPassword(context.Background(), userID)

		assert.Error(t, err)
		assert.Equal(t, profiles.ErrUserNotExist, err)
		assert.Nil(t, password)
	})
}
