package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/profiles"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

type mockMinIOService struct {
	uploadFileFunc           func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error
	deleteFileFunc           func(ctx context.Context, bucketName, key string) error
	generatePresignedURLFunc func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}

func (m *mockMinIOService) UploadFile(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
	if m.uploadFileFunc != nil {
		return m.uploadFileFunc(ctx, bucketName, key, reader, size, contentType)
	}
	return nil
}

func (m *mockMinIOService) DeleteFile(ctx context.Context, bucketName, key string) error {
	if m.deleteFileFunc != nil {
		return m.deleteFileFunc(ctx, bucketName, key)
	}
	return nil
}

func (m *mockMinIOService) GeneratePresignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
	if m.generatePresignedURLFunc != nil {
		return m.generatePresignedURLFunc(ctx, bucketName, key, expiry)
	}
	return "http://example.com/presigned", nil
}

func TestGetProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()
	expectedProfile := &models.Profile{
		ID:        userID,
		Username:  "testuser",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
					AddRow(userID, expectedProfile.Username, expectedProfile.CreatedAt, expectedProfile.UpdatedAt)
				mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id = \\$1").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id = \\$1").
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
		{
			name: "Database Error",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE id = \\$1").
					WithArgs(userID).
					WillReturnError(errors.New("connection failed"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			profile, err := repo.GetProfile(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedProfile.ID, profile.ID)
				assert.Equal(t, expectedProfile.Username, profile.Username)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetProfileByUsername(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()
	username := "testuser"

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
					AddRow(userID, username, time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE username = \\$1").
					WithArgs(username).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, username, created_at, updated_at FROM profiles WHERE username = \\$1").
					WithArgs(username).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			profile, err := repo.GetProfileByUsername(context.Background(), username)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()
	newUsername := "newusername"

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
					AddRow(userID, newUsername, time.Now(), time.Now())
				mock.ExpectQuery("UPDATE profiles SET username = \\$2, updated_at = now\\(\\) WHERE id = \\$1 RETURNING id, username, created_at, updated_at").
					WithArgs(userID, newUsername).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("UPDATE profiles SET username = \\$2, updated_at = now\\(\\) WHERE id = \\$1 RETURNING id, username, created_at, updated_at").
					WithArgs(userID, newUsername).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			profile, err := repo.UpdateProfile(context.Background(), userID, models.Profile{Username: newUsername})

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, newUsername, profile.Username)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteProfile(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id"}).AddRow(userID)
				mock.ExpectQuery("DELETE FROM profiles WHERE id = \\$1 RETURNING id").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("DELETE FROM profiles WHERE id = \\$1 RETURNING id").
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			err := repo.DeleteProfile(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileID := uuid.New()
	avatarID := uuid.New()
	minioKey := "test-key"
	avatarURL := "http://example.com/avatar.jpg"
	urlExpiresAt := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		checkNil  bool
	}{
		{
			name: "Success - Valid URL",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(avatarID, profileID, minioKey, avatarURL, urlExpiresAt, time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at FROM avatars WHERE profile_id = \\$1").
					WithArgs(profileID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "Avatar Not Found",
			setupMock: func() {
				mock.ExpectQuery("SELECT id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at FROM avatars WHERE profile_id = \\$1").
					WithArgs(profileID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr:  false,
			checkNil: true,
		},
		{
			name: "Expired URL - Regenerate",
			setupMock: func() {
				expiredTime := time.Now().Add(-1 * time.Hour)
				rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(avatarID, profileID, minioKey, avatarURL, expiredTime, time.Now(), time.Now())
				mock.ExpectQuery("SELECT id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at FROM avatars WHERE profile_id = \\$1").
					WithArgs(profileID).
					WillReturnRows(rows)

				mock.ExpectQuery("UPDATE avatars SET avatar_url = \\$2, url_expires_at = \\$3, updated_at = now\\(\\) WHERE id = \\$1 RETURNING avatar_url, url_expires_at, updated_at").
					WithArgs(avatarID, sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"avatar_url", "url_expires_at", "updated_at"}).
						AddRow("http://example.com/new-url.jpg", time.Now().Add(30*time.Minute), time.Now()))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			minioMock := &mockMinIOService{
				generatePresignedURLFunc: func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return "http://example.com/new-url.jpg", nil
				},
			}
			repo := NewProfileRepository(db, minioMock, "avatars")

			avatar, err := repo.GetAvatar(context.Background(), profileID)

			if tt.wantErr {
				assert.Error(t, err)
			} else if tt.checkNil {
				assert.NoError(t, err)
				assert.Nil(t, avatar)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, avatar)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUploadAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte("test image data"))

	tests := []struct {
		name       string
		setupMock  func()
		setupMinio func(*mockMinIOService)
		wantErr    bool
		errType    error
	}{
		{
			name: "Success",
			setupMock: func() {
				// Delete existing avatar (if any) - no rows found
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnError(sql.ErrNoRows)

				// Insert new avatar
				newAvatarID := uuid.New()
				newMinioKey := newAvatarID.String()
				newURL := "http://example.com/avatar.jpg"
				newExpiry := time.Now().Add(30 * time.Minute)

				rows := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(newAvatarID, profileID, newMinioKey, newURL, newExpiry, time.Now(), time.Now())
				mock.ExpectQuery("INSERT INTO avatars \\(id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, now\\(\\), now\\(\\)\\) RETURNING id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at").
					WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			setupMinio: func(m *mockMinIOService) {
				m.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return nil
				}
				m.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return "http://example.com/avatar.jpg", nil
				}
			},
			wantErr: false,
		},
		{
			name: "Delete Existing Avatar Success",
			setupMock: func() {
				// Delete existing avatar - found
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow("old-key")
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnRows(rows)

				// Insert new avatar
				newAvatarID := uuid.New()
				newMinioKey := newAvatarID.String()
				newURL := "http://example.com/avatar.jpg"
				newExpiry := time.Now().Add(30 * time.Minute)

				rows2 := sqlmock.NewRows([]string{"id", "profile_id", "minio_key", "avatar_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(newAvatarID, profileID, newMinioKey, newURL, newExpiry, time.Now(), time.Now())
				mock.ExpectQuery("INSERT INTO avatars \\(id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, now\\(\\), now\\(\\)\\) RETURNING id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at").
					WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnRows(rows2)
			},
			setupMinio: func(m *mockMinIOService) {
				m.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					return nil
				}
				m.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return nil
				}
				m.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return "http://example.com/avatar.jpg", nil
				}
			},
			wantErr: false,
		},
		{
			name: "MinIO Upload Failed",
			setupMock: func() {
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnError(sql.ErrNoRows)
			},
			setupMinio: func(m *mockMinIOService) {
				m.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return errors.New("minio upload failed")
				}
			},
			wantErr: true,
			errType: profiles.ErrFailedToUpload,
		},
		{
			name: "Database Insert Failed",
			setupMock: func() {
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnError(sql.ErrNoRows)

				mock.ExpectQuery("INSERT INTO avatars \\(id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at\\) VALUES \\(\\$1, \\$2, \\$3, \\$4, \\$5, now\\(\\), now\\(\\)\\) RETURNING id, profile_id, minio_key, avatar_url, url_expires_at, created_at, updated_at").
					WithArgs(sqlmock.AnyArg(), profileID, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
					WillReturnError(errors.New("insert failed"))
			},
			setupMinio: func(m *mockMinIOService) {
				m.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return nil
				}
				m.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return "http://example.com/avatar.jpg", nil
				}
				m.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					return nil
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			minioMock := &mockMinIOService{}
			if tt.setupMinio != nil {
				tt.setupMinio(minioMock)
			}
			repo := NewProfileRepository(db, minioMock, "avatars")

			avatar, err := repo.UploadAvatar(context.Background(), profileID, fileName, fileSize, mimeType, fileReader)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, avatar)
				assert.Equal(t, profileID, avatar.ProfileID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteAvatar(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	profileID := uuid.New()
	minioKey := "test-key"

	tests := []struct {
		name       string
		setupMock  func()
		setupMinio func(*mockMinIOService)
		wantErr    bool
		errType    error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnRows(rows)
			},
			setupMinio: func(m *mockMinIOService) {
				m.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					return nil
				}
			},
			wantErr: false,
		},
		{
			name: "Avatar Not Found",
			setupMock: func() {
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrAvatarNotFound,
		},
		{
			name: "MinIO Delete Failed",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery("DELETE FROM avatars WHERE profile_id = \\$1 RETURNING minio_key").
					WithArgs(profileID).
					WillReturnRows(rows)
			},
			setupMinio: func(m *mockMinIOService) {
				m.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					return errors.New("minio delete failed")
				}
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			minioMock := &mockMinIOService{}
			if tt.setupMinio != nil {
				tt.setupMinio(minioMock)
			}
			repo := NewProfileRepository(db, minioMock, "avatars")

			err := repo.DeleteAvatar(context.Background(), profileID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestChangePassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()
	newPassword := "newpass123"

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "username", "created_at", "updated_at"}).
					AddRow(userID, "testuser", time.Now(), time.Now())
				mock.ExpectQuery("UPDATE profiles SET password = \\$2, updated_at = now\\(\\) WHERE id = \\$1 RETURNING id, username, created_at, updated_at").
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("UPDATE profiles SET password = \\$2, updated_at = now\\(\\) WHERE id = \\$1 RETURNING id, username, created_at, updated_at").
					WithArgs(userID, sqlmock.AnyArg()).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			profile, err := repo.ChangePassword(context.Background(), userID, newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, profile)
				assert.Equal(t, userID, profile.ID)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetPassword(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	userID := uuid.New()
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "Success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"password"}).AddRow(hashedPassword)
				mock.ExpectQuery("SELECT password FROM profiles WHERE id = \\$1").
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "User Not Found",
			setupMock: func() {
				mock.ExpectQuery("SELECT password FROM profiles WHERE id = \\$1").
					WithArgs(userID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
			errType: profiles.ErrUserNotExist,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			repo := NewProfileRepository(db, &mockMinIOService{}, "avatars")

			password, err := repo.GetPassword(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, password)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
