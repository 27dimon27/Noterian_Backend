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
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestAttachmentRepository_GetAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	blockID := uuid.New()
	attachmentID := uuid.New()
	minioKey := "test-key"
	presignedURL := "http://minio.example.com/test-key"
	now := time.Now()
	expiredAt := now.Add(-time.Hour)

	getAttachmentQuery := regexp.QuoteMeta(`SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = $1`)
	updateQuery := regexp.QuoteMeta(`UPDATE attachments SET attach_url = $1, url_expires_at = $2, updated_at = $3 WHERE id = $4 RETURNING attach_url, url_expires_at, updated_at`)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
		wantNil   bool
	}{
		{
			name: "success with valid URL",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(attachmentID, blockID, minioKey, presignedURL, now.Add(time.Hour), now, now)
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnRows(rows)
			},
			wantErr: nil,
			wantNil: false,
		},
		{
			name: "success with expired URL - regenerates",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(attachmentID, blockID, minioKey, presignedURL, expiredAt, now, now)
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnRows(rows)

				newURL := "http://minio.example.com/new-key"
				mockMinio.EXPECT().GeneratePresignedURL(gomock.Any(), "test-bucket", minioKey, attachments.PRESIGNED_URL_EXPIRY).Return(newURL, nil)

				updateRows := sqlmock.NewRows([]string{"attach_url", "url_expires_at", "updated_at"}).
					AddRow(newURL, now.Add(attachments.PRESIGNED_URL_EXPIRY), now)
				mock.ExpectQuery(updateQuery).
					WithArgs(newURL, sqlmock.AnyArg(), sqlmock.AnyArg(), attachmentID).
					WillReturnRows(updateRows)
			},
			wantErr: nil,
			wantNil: false,
		},
		{
			name: "attachment not found",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: nil,
			wantNil: true,
		},
		{
			name: "database error",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
			wantNil: true,
		},
		{
			name: "failed to generate presigned URL",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(attachmentID, blockID, minioKey, presignedURL, expiredAt, now, now)
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnRows(rows)

				mockMinio.EXPECT().GeneratePresignedURL(gomock.Any(), "test-bucket", minioKey, attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("minio error"))
			},
			wantErr: errors.New("minio error"),
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			attachment, err := repo.GetAttachment(context.Background(), blockID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, attachment)
				} else {
					assert.NotNil(t, attachment)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAttachmentRepository_UploadAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	blockID := uuid.New()
	fileName := "test.png"
	fileSize := int64(1024)
	mimeType := "image/png"
	fileContent := []byte("test image content")

	getAttachmentQuery := regexp.QuoteMeta(`SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = $1`)
	createAttachmentQuery := regexp.QuoteMeta(`INSERT INTO attachments (id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at`)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.EXPECT().UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).Return(nil)

				mockMinio.EXPECT().GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("http://example.com/test", nil)

				rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(uuid.New(), blockID, "test-key", "http://example.com/test", time.Now().Add(time.Hour), time.Now(), time.Now())
				mock.ExpectQuery(createAttachmentQuery).
					WillReturnRows(rows)
			},
			wantErr: nil,
		},
		{
			name: "block already has attachment",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
					AddRow(uuid.New(), blockID, "existing-key", "http://example.com/existing", time.Now().Add(time.Hour), time.Now(), time.Now())
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnRows(rows)
			},
			wantErr: attachments.ErrBlockAlreadyHasAttach,
		},
		{
			name: "failed to upload to MinIO",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.EXPECT().UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).Return(errors.New("upload failed"))
			},
			wantErr: attachments.ErrFailedToUpload,
		},
		{
			name: "failed to generate presigned URL - cleanup MinIO",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.EXPECT().UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).Return(nil)
				mockMinio.EXPECT().GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("generate failed"))
				mockMinio.EXPECT().DeleteFile(gomock.Any(), "test-bucket", gomock.Any()).Return(nil)
			},
			wantErr: attachments.ErrFailedToGenerateURL,
		},
		{
			name: "database insert fails - cleanup MinIO",
			setupMock: func() {
				mock.ExpectQuery(getAttachmentQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)

				mockMinio.EXPECT().UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).Return(nil)
				mockMinio.EXPECT().GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("http://example.com/test", nil)

				mock.ExpectQuery(createAttachmentQuery).
					WillReturnError(errors.New("database error"))
				mockMinio.EXPECT().DeleteFile(gomock.Any(), "test-bucket", gomock.Any()).Return(nil)
			},
			wantErr: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			reader := bytes.NewReader(fileContent)

			attachment, err := repo.UploadAttachment(
				context.Background(),
				blockID,
				fileName,
				fileSize,
				mimeType,
				reader,
			)

			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.wantErr.Error() != "" {
					assert.Equal(t, tt.wantErr.Error(), err.Error())
				}
				assert.Nil(t, attachment)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, attachment)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestAttachmentRepository_DeleteAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	blockID := uuid.New()
	minioKey := "test-key"

	deleteQuery := regexp.QuoteMeta(`DELETE FROM attachments WHERE block_id = $1 RETURNING minio_key`)

	tests := []struct {
		name      string
		setupMock func()
		wantErr   error
	}{
		{
			name: "success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery(deleteQuery).
					WithArgs(blockID).
					WillReturnRows(rows)
				mockMinio.EXPECT().DeleteFile(gomock.Any(), "test-bucket", minioKey).Return(nil)
			},
			wantErr: nil,
		},
		{
			name: "attachment not found",
			setupMock: func() {
				mock.ExpectQuery(deleteQuery).
					WithArgs(blockID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: attachments.ErrAttachmentNotFound,
		},
		{
			name: "database error",
			setupMock: func() {
				mock.ExpectQuery(deleteQuery).
					WithArgs(blockID).
					WillReturnError(errors.New("database error"))
			},
			wantErr: errors.New("database error"),
		},
		{
			name: "minio delete fails",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)
				mock.ExpectQuery(deleteQuery).
					WithArgs(blockID).
					WillReturnRows(rows)
				mockMinio.EXPECT().DeleteFile(gomock.Any(), "test-bucket", minioKey).Return(errors.New("minio delete failed"))
			},
			wantErr: errors.New("minio delete failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()

			err := repo.DeleteAttachment(context.Background(), blockID)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
