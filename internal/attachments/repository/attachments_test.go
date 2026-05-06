package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	return "", nil
}

func TestAttachmentRepository_GetAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	attachmentID := uuid.New()
	blockID := uuid.New()
	minioKey := "test-key"
	expiredURL := "http://example.com/expired"
	newURL := "http://example.com/new"
	expiredTime := time.Now().Add(-2 * time.Hour)
	futureTime := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
		checkResult   func(t *testing.T, attachment *models.Attachment)
	}{
		{
			name: "success - valid URL",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{
					"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
				}).AddRow(
					attachmentID, blockID, minioKey, newURL, futureTime, time.Now(), time.Now(),
				)
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnRows(rows)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, attachment *models.Attachment) {
				require.NotNil(t, attachment)
				assert.Equal(t, attachmentID, attachment.ID)
				assert.Equal(t, blockID, attachment.BlockID)
				assert.Equal(t, minioKey, attachment.MinioKey)
				assert.Equal(t, newURL, attachment.AttachURL)
			},
		},
		{
			name: "URL expired - regenerate new URL",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{
					"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
				}).AddRow(
					attachmentID, blockID, minioKey, expiredURL, expiredTime, time.Now(), time.Now(),
				)
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnRows(rows)

				updateRows := sqlmock.NewRows([]string{
					"attach_url", "url_expires_at", "updated_at",
				}).AddRow(newURL, futureTime, time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_ATTACHMENT_URL)).
					WithArgs(attachmentID, newURL, sqlmock.AnyArg()).
					WillReturnRows(updateRows)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, attachment *models.Attachment) {
				require.NotNil(t, attachment)
				assert.Equal(t, newURL, attachment.AttachURL)
			},
		},
		{
			name: "attachment not found",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: nil,
			checkResult: func(t *testing.T, attachment *models.Attachment) {
				assert.Nil(t, attachment)
			},
		},
		{
			name: "database error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(errors.New("database connection error"))
			},
			expectedError: errors.New("database connection error"),
			checkResult: func(t *testing.T, attachment *models.Attachment) {
				assert.Nil(t, attachment)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMinio := &mockMinIOService{
				generatePresignedURLFunc: func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return newURL, nil
				},
			}

			repo := NewAttachmentRepository(db, mockMinio, "test-bucket")
			tt.setupMock()

			attachment, err := repo.GetAttachment(context.Background(), blockID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}
			tt.checkResult(t, attachment)

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAttachmentRepository_UpdateAttachmentURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	attachmentID := uuid.New()
	newURL := "http://example.com/new"
	expiresAt := time.Now().Add(time.Hour)

	tests := []struct {
		name          string
		setupMock     func()
		expectedError error
	}{
		{
			name: "success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{
					"attach_url", "url_expires_at", "updated_at",
				}).AddRow(newURL, expiresAt, time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_ATTACHMENT_URL)).
					WithArgs(attachmentID, newURL, expiresAt).
					WillReturnRows(rows)
			},
			expectedError: nil,
		},
		{
			name: "database error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_ATTACHMENT_URL)).
					WithArgs(attachmentID, newURL, expiresAt).
					WillReturnError(errors.New("update failed"))
			},
			expectedError: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := NewAttachmentRepository(db, nil, "test-bucket")
			tt.setupMock()

			err := repo.UpdateAttachmentURL(context.Background(), attachmentID, newURL, expiresAt)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAttachmentRepository_UploadAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	blockID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileContent := []byte("test data")
	presignedURL := "http://example.com/presigned"
	now := time.Now()

	tests := []struct {
		name          string
		setupMock     func(mockMinio *mockMinIOService)
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(mockMinio *mockMinIOService) {
				// Check existing attachment - none found
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return nil
				}
				mockMinio.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return presignedURL, nil
				}

				// Create attachment
				attachmentID := uuid.New()
				minioKey := attachmentID.String()
				urlExpiresAt := now.Add(attachments.PRESIGNED_URL_EXPIRY)

				rows := sqlmock.NewRows([]string{
					"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
				}).AddRow(
					attachmentID, blockID, minioKey, presignedURL, urlExpiresAt, now, now,
				)
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_ATTACHMENT)).
					WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presignedURL, sqlmock.AnyArg()).
					WillReturnRows(rows)
			},
			expectedError: nil,
		},
		{
			name: "error - attachment already exists",
			setupMock: func(mockMinio *mockMinIOService) {
				rows := sqlmock.NewRows([]string{
					"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
				}).AddRow(
					uuid.New(), blockID, "existing-key", "http://example.com/existing", now.Add(time.Hour), now, now,
				)
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnRows(rows)
			},
			expectedError: attachments.ErrBlockAlreadyHasAttach,
		},
		{
			name: "error - check existing attachment database error",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - minio upload fails",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					return errors.New("minio upload error")
				}
			},
			expectedError: attachments.ErrFailedToUpload,
		},
		{
			name: "error - generate presigned URL fails with cleanup",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				var uploadedKey string
				mockMinio.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					uploadedKey = key
					return nil
				}
				mockMinio.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return "", errors.New("presigned url error")
				}
				mockMinio.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					assert.Equal(t, uploadedKey, key)
					return nil
				}
			},
			expectedError: attachments.ErrFailedToGenerateURL,
		},
		{
			name: "error - create attachment database error with cleanup",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				var uploadedKey string
				mockMinio.uploadFileFunc = func(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error {
					uploadedKey = key
					return nil
				}
				mockMinio.generatePresignedURLFunc = func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
					return presignedURL, nil
				}
				mockMinio.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					assert.Equal(t, uploadedKey, key)
					return nil
				}

				mock.ExpectQuery(regexp.QuoteMeta(CREATE_ATTACHMENT)).
					WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presignedURL, sqlmock.AnyArg()).
					WillReturnError(errors.New("database insert error"))
			},
			expectedError: errors.New("database insert error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Создаем новый mock для каждого теста
			mockMinio := &mockMinIOService{}
			tt.setupMock(mockMinio)

			repo := NewAttachmentRepository(db, mockMinio, "test-bucket")
			attachment, err := repo.UploadAttachment(
				context.Background(),
				blockID,
				fileName,
				fileSize,
				mimeType,
				bytes.NewReader(fileContent),
			)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError, err)
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}

			// Проверяем ожидания
			if err := mock.ExpectationsWereMet(); err != nil && tt.expectedError == nil {
				t.Errorf("unexpected mock expectations: %v", err)
			}
		})
	}
}

func TestAttachmentRepository_DeleteAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	blockID := uuid.New()
	minioKey := "test-key"

	tests := []struct {
		name          string
		setupMock     func(mockMinio *mockMinIOService)
		expectedError error
	}{
		{
			name: "success",
			setupMock: func(mockMinio *mockMinIOService) {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_ATTACHMENT_BY_ID)).
					WithArgs(blockID).
					WillReturnRows(rows)

				mockMinio.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					assert.Equal(t, minioKey, key)
					return nil
				}
			},
			expectedError: nil,
		},
		{
			name: "error - attachment not found",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_ATTACHMENT_BY_ID)).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)
			},
			expectedError: attachments.ErrAttachmentNotFound,
		},
		{
			name: "error - database error",
			setupMock: func(mockMinio *mockMinIOService) {
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_ATTACHMENT_BY_ID)).
					WithArgs(blockID).
					WillReturnError(errors.New("database error"))
			},
			expectedError: errors.New("database error"),
		},
		{
			name: "error - minio delete fails",
			setupMock: func(mockMinio *mockMinIOService) {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_ATTACHMENT_BY_ID)).
					WithArgs(blockID).
					WillReturnRows(rows)

				mockMinio.deleteFileFunc = func(ctx context.Context, bucketName, key string) error {
					return errors.New("minio delete error")
				}
			},
			expectedError: errors.New("minio delete error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockMinio := &mockMinIOService{}
			tt.setupMock(mockMinio)

			repo := NewAttachmentRepository(db, mockMinio, "test-bucket")
			err := repo.DeleteAttachment(context.Background(), blockID)

			if tt.expectedError != nil {
				assert.Error(t, err)
				if tt.expectedError == attachments.ErrAttachmentNotFound {
					assert.Equal(t, tt.expectedError, err)
				}
			} else {
				assert.NoError(t, err)
			}

			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAttachmentRepository_EdgeCases(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	t.Run("GetAttachment with empty result set", func(t *testing.T) {
		blockID := uuid.New()

		mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
			WithArgs(blockID).
			WillReturnRows(sqlmock.NewRows([]string{
				"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
			}))

		repo := NewAttachmentRepository(db, nil, "test-bucket")
		attachment, err := repo.GetAttachment(context.Background(), blockID)

		assert.NoError(t, err)
		assert.Nil(t, attachment)
		assert.NoError(t, mock.ExpectationsWereMet())
	})
}

func TestSQLQueriesConstants(t *testing.T) {
	// Проверяем, что все SQL запросы определены корректно
	assert.NotEmpty(t, CREATE_ATTACHMENT, "CREATE_ATTACHMENT query should not be empty")
	assert.NotEmpty(t, GET_ATTACHMENT_BY_BLOCK_ID, "GET_ATTACHMENT_BY_BLOCK_ID query should not be empty")
	assert.NotEmpty(t, UPDATE_ATTACHMENT_URL, "UPDATE_ATTACHMENT_URL query should not be empty")
	assert.NotEmpty(t, DELETE_ATTACHMENT_BY_ID, "DELETE_ATTACHMENT_BY_ID query should not be empty")
	assert.NotEmpty(t, GET_NOTE_BY_ID, "GET_NOTE_BY_ID query should not be empty")
	assert.NotEmpty(t, GET_BLOCK_BY_ID, "GET_BLOCK_BY_ID query should not be empty")

	// Проверяем синтаксис ключевых слов
	assert.Contains(t, CREATE_ATTACHMENT, "INSERT INTO", "CREATE_ATTACHMENT should be INSERT statement")
	assert.Contains(t, GET_ATTACHMENT_BY_BLOCK_ID, "SELECT", "GET_ATTACHMENT_BY_BLOCK_ID should be SELECT statement")
	assert.Contains(t, UPDATE_ATTACHMENT_URL, "UPDATE", "UPDATE_ATTACHMENT_URL should be UPDATE statement")
	assert.Contains(t, DELETE_ATTACHMENT_BY_ID, "DELETE", "DELETE_ATTACHMENT_BY_ID should be DELETE statement")
	assert.Contains(t, GET_NOTE_BY_ID, "SELECT", "GET_NOTE_BY_ID should be SELECT statement")
	assert.Contains(t, GET_BLOCK_BY_ID, "SELECT", "GET_BLOCK_BY_ID should be SELECT statement")

	// Проверяем наличие таблиц
	assert.Contains(t, CREATE_ATTACHMENT, "attachments", "CREATE_ATTACHMENT should reference attachments table")
	assert.Contains(t, GET_ATTACHMENT_BY_BLOCK_ID, "attachments", "GET_ATTACHMENT_BY_BLOCK_ID should reference attachments table")
	assert.Contains(t, UPDATE_ATTACHMENT_URL, "attachments", "UPDATE_ATTACHMENT_URL should reference attachments table")
	assert.Contains(t, DELETE_ATTACHMENT_BY_ID, "attachments", "DELETE_ATTACHMENT_BY_ID should reference attachments table")
	assert.Contains(t, GET_NOTE_BY_ID, "notes", "GET_NOTE_BY_ID should reference notes table")
	assert.Contains(t, GET_BLOCK_BY_ID, "blocks", "GET_BLOCK_BY_ID should reference blocks table")
}

func TestAttachmentRepository_StructCreation(t *testing.T) {
	// Тестируем создание репозитория с nil зависимостями
	repo := NewAttachmentRepository(nil, nil, "test-bucket")
	assert.NotNil(t, repo, "Repository should be created even with nil dependencies")
	assert.Equal(t, "test-bucket", repo.attachmentBucket, "Bucket name should be set correctly")
	assert.Nil(t, repo.db, "DB should be nil")
	assert.Nil(t, repo.minio, "MinIO should be nil")
}

func TestAttachmentRepository_GetAttachmentWithExpiredURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	attachmentID := uuid.New()
	blockID := uuid.New()
	minioKey := "test-key"
	expiredURL := "http://example.com/expired"
	newURL := "http://example.com/new"
	expiredTime := time.Now().Add(-2 * time.Hour)
	newExpiryTime := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

	mockMinio := &mockMinIOService{
		generatePresignedURLFunc: func(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error) {
			assert.Equal(t, minioKey, key)
			return newURL, nil
		},
	}

	// Expect query for expired attachment
	rows := sqlmock.NewRows([]string{
		"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at",
	}).AddRow(
		attachmentID, blockID, minioKey, expiredURL, expiredTime, time.Now(), time.Now(),
	)
	mock.ExpectQuery(regexp.QuoteMeta(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(rows)

	// Expect update query
	updateRows := sqlmock.NewRows([]string{
		"attach_url", "url_expires_at", "updated_at",
	}).AddRow(newURL, newExpiryTime, time.Now())
	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_ATTACHMENT_URL)).
		WithArgs(attachmentID, newURL, sqlmock.AnyArg()).
		WillReturnRows(updateRows)

	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")
	attachment, err := repo.GetAttachment(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, attachment)
	assert.Equal(t, newURL, attachment.AttachURL)
	assert.NoError(t, mock.ExpectationsWereMet())
}
