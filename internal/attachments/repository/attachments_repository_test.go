package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestNewAttachmentRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	if repo == nil {
		t.Errorf("expected non-nil repository")
	}
	if repo.db != db {
		t.Errorf("expected db to be set")
	}
}

func TestGetAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	blockID := uuid.New()
	attachmentID := uuid.New()
	now := time.Now()
	futureTime := now.Add(attachments.PRESIGNED_URL_EXPIRY)

	t.Run("success - valid URL", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(attachmentID, blockID, "test-key", "https://example.com/file", futureTime, now, now)

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnRows(rows)

		attachment, err := repo.GetAttachment(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if attachment.ID != attachmentID {
			t.Errorf("expected ID %v, got %v", attachmentID, attachment.ID)
		}
	})

	t.Run("attachment not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		attachment, err := repo.GetAttachment(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if attachment != nil {
			t.Errorf("expected nil attachment, got %v", attachment)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetAttachment(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUpdateAttachmentURL(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	attachmentID := uuid.New()
	url := "https://example.com/file"
	expiresAt := time.Now().Add(time.Hour)
	updatedAt := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"attach_url", "url_expires_at", "updated_at"}).
			AddRow(url, expiresAt, updatedAt)

		mock.ExpectQuery("UPDATE attachments SET attach_url = \\$1, url_expires_at = \\$2, updated_at = \\$3 WHERE id = \\$4 RETURNING attach_url, url_expires_at, updated_at").
			WithArgs(url, expiresAt, sqlmock.AnyArg(), attachmentID).
			WillReturnRows(rows)

		err := repo.UpdateAttachmentURL(context.Background(), attachmentID, url, expiresAt)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("UPDATE attachments SET attach_url = \\$1, url_expires_at = \\$2, updated_at = \\$3 WHERE id = \\$4 RETURNING attach_url, url_expires_at, updated_at").
			WithArgs(url, expiresAt, sqlmock.AnyArg(), attachmentID).
			WillReturnError(errors.New("db error"))

		err := repo.UpdateAttachmentURL(context.Background(), attachmentID, url, expiresAt)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUploadAttachment(t *testing.T) {
	blockID := uuid.New()
	attachmentID := uuid.New()
	fileName := "test.jpg"
	fileSize := int64(1024)
	mimeType := "image/jpeg"
	fileReader := bytes.NewReader([]byte{0xFF, 0xD8, 0xFF, 0xE0})
	now := time.Now()
	futureTime := now.Add(attachments.PRESIGNED_URL_EXPIRY)

	t.Run("success", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		mock.ExpectQuery("SELECT .* FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(nil)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
			Return("https://example.com/file", nil)

		mock.ExpectQuery("INSERT INTO attachments").
			WithArgs(
				sqlmock.AnyArg(),           // $1 (id) - ТУТ БЫЛА ОШИБКА
				blockID,                    // $2 (block_id)
				sqlmock.AnyArg(),           // $3 (minio_key)
				"https://example.com/file", // $4 (url)
				sqlmock.AnyArg(),           // $5 (expires)
				sqlmock.AnyArg(),           // $6 (created)
				sqlmock.AnyArg(),           // $7 (updated)
			).
			WillReturnRows(sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
				AddRow(attachmentID, blockID, "test-key", "https://example.com/file", time.Now(), time.Now(), time.Now()))

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if err != nil {
			t.Errorf("unexpected err: %s", err) 
		}
		
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Errorf("unfulfilled expectations: %s", err)
		}
	})


	t.Run("block already has attachment", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		existingRows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(attachmentID, blockID, "existing-key", "https://example.com/file", futureTime, now, now)

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnRows(existingRows)

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrBlockAlreadyHasAttach) {
			t.Errorf("expected ErrBlockAlreadyHasAttach, got %v", err)
		}
	})

	t.Run("check existing attachment error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("upload to minio error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(errors.New("upload error"))

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrFailedToUpload) {
			t.Errorf("expected ErrFailedToUpload, got %v", err)
		}
	})

	t.Run("generate presigned URL error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(nil)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
			Return("", errors.New("presign error"))

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), "test-bucket", gomock.Any()).
			Return(nil)

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if !errors.Is(err, attachments.ErrFailedToGenerateURL) {
			t.Errorf("expected ErrFailedToGenerateURL, got %v", err)
		}
	})

	t.Run("insert into database error", func(t *testing.T) {
		db, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("cant create mock: %s", err)
		}
		defer db.Close()

		ctrl := gomock.NewController(t)
		defer ctrl.Finish()

		mockMinio := mocks.NewMockMinIOService(ctrl)
		repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

		mock.ExpectQuery("SELECT id, block_id, minio_key, attach_url, url_expires_at, created_at, updated_at FROM attachments WHERE block_id = \\$1").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		mockMinio.EXPECT().
			UploadFile(gomock.Any(), "test-bucket", gomock.Any(), gomock.Any(), fileSize, mimeType).
			Return(nil)

		mockMinio.EXPECT().
			GeneratePresignedURL(gomock.Any(), "test-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
			Return("https://example.com/file", nil)

		mock.ExpectQuery("INSERT INTO attachments").
			WithArgs(attachmentID, blockID, "test-key", "https://example.com/file", futureTime, now, now).
			WillReturnError(errors.New("insert error"))

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), "test-bucket", gomock.Any()).
			Return(nil)

		_, err = repo.UploadAttachment(context.Background(), blockID, fileName, fileSize, mimeType, fileReader)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestDeleteAttachment(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockMinio := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, mockMinio, "test-bucket")

	blockID := uuid.New()
	minioKey := "test-key"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)

		mock.ExpectQuery("DELETE FROM attachments WHERE block_id = \\$1 RETURNING minio_key").
			WithArgs(blockID).
			WillReturnRows(rows)

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), "test-bucket", minioKey).
			Return(nil)

		err := repo.DeleteAttachment(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("attachment not found", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM attachments WHERE block_id = \\$1 RETURNING minio_key").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		err := repo.DeleteAttachment(context.Background(), blockID)

		if !errors.Is(err, attachments.ErrAttachmentNotFound) {
			t.Errorf("expected ErrAttachmentNotFound, got %v", err)
		}
	})

	t.Run("delete query error", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM attachments WHERE block_id = \\$1 RETURNING minio_key").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteAttachment(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("minio delete error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"minio_key"}).AddRow(minioKey)

		mock.ExpectQuery("DELETE FROM attachments WHERE block_id = \\$1 RETURNING minio_key").
			WithArgs(blockID).
			WillReturnRows(rows)

		mockMinio.EXPECT().
			DeleteFile(gomock.Any(), "test-bucket", minioKey).
			Return(errors.New("minio delete error"))

		err := repo.DeleteAttachment(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
