package repository

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/repository/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func newMockRepository(t *testing.T) (*AttachmentRepository, sqlmock.Sqlmock, *gomock.Controller, *mocks.MockMinIOService) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}

	ctrl := gomock.NewController(t)
	minioMock := mocks.NewMockMinIOService(ctrl)
	repo := NewAttachmentRepository(db, minioMock, "attachments-bucket", "headers-bucket")

	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Logf("failed to close db: %v", err)
		}
		ctrl.Finish()
	})

	return repo, mock, ctrl, minioMock
}

func queryRegexp(query string) string {
	return regexp.QuoteMeta(strings.TrimSpace(query))
}

func TestGetAttachment_NotExpired(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(attachmentID, blockID, "attachment-key", "https://example.com/old", now.Add(time.Hour), now, now)

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment == nil {
		t.Fatal("expected attachment, got nil")
	}
	if attachment.ID != attachmentID {
		t.Fatalf("expected id %v, got %v", attachmentID, attachment.ID)
	}
	if attachment.AttachURL != "https://example.com/old" {
		t.Fatalf("expected URL old, got %s", attachment.AttachURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_ExpiredRefreshURL(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(attachmentID, blockID, "attachment-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", "attachment-key", attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/new", nil)
	mock.ExpectQuery(queryRegexp(UPDATE_ATTACHMENT_URL)).WithArgs(attachmentID, "https://example.com/new", sqlmock.AnyArg()).WillReturnRows(
		sqlmock.NewRows([]string{"attach_url", "url_expires_at", "updated_at"}).AddRow("https://example.com/new", time.Now().Add(attachments.PRESIGNED_URL_EXPIRY), time.Now()),
	)

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment == nil {
		t.Fatal("expected attachment, got nil")
	}
	if attachment.AttachURL != "https://example.com/new" {
		t.Fatalf("expected refreshed URL, got %s", attachment.AttachURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_NotFound(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if !errors.Is(err, attachments.ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_ExistingAttachment(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(attachmentID, blockID, "attachment-key", "https://example.com/old", now.Add(time.Hour), now, now)

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if !errors.Is(err, attachments.ErrBlockAlreadyHasAttach) {
		t.Fatalf("expected ErrBlockAlreadyHasAttach, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_UploadFailure(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "attachments-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(errors.New("upload failed"))

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if !errors.Is(err, attachments.ErrFailedToUpload) {
		t.Fatalf("expected ErrFailedToUpload, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_GenerateURLFailureDeletesFile(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "attachments-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("url creation failed"))
	minioMock.EXPECT().DeleteFile(gomock.Any(), "attachments-bucket", gomock.Any()).Return(nil)

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if !errors.Is(err, attachments.ErrFailedToGenerateURL) {
		t.Fatalf("expected ErrFailedToGenerateURL, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_Success(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "attachments-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/file", nil)

	mock.ExpectQuery(queryRegexp(CREATE_ATTACHMENT)).WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), "https://example.com/file", sqlmock.AnyArg()).WillReturnRows(
		sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(attachmentID, blockID, "attachment-key", "https://example.com/file", createdAt.Add(attachments.PRESIGNED_URL_EXPIRY), createdAt, updatedAt),
	)

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attachment == nil {
		t.Fatal("expected attachment, got nil")
	}
	if attachment.ID != attachmentID {
		t.Fatalf("expected id %v, got %v", attachmentID, attachment.ID)
	}
	if attachment.AttachURL != "https://example.com/file" {
		t.Fatalf("expected URL https://example.com/file, got %s", attachment.AttachURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAttachment_NotFound(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_ATTACHMENT_BY_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)

	err := repo.DeleteAttachment(context.Background(), blockID)
	if !errors.Is(err, attachments.ErrAttachmentNotFound) {
		t.Fatalf("expected ErrAttachmentNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAttachment_Success(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_ATTACHMENT_BY_ID)).WithArgs(blockID).WillReturnRows(
		sqlmock.NewRows([]string{"minio_key"}).AddRow("attachment-key"),
	)
	minioMock.EXPECT().DeleteFile(gomock.Any(), "attachments-bucket", "attachment-key").Return(nil)

	err := repo.DeleteAttachment(context.Background(), blockID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_ExpiredRefreshURL(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(headerID, noteID, "header-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", "header-key", attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/new", nil)
	mock.ExpectQuery(queryRegexp(UPDATE_HEADER_URL)).WithArgs(headerID, "https://example.com/new", sqlmock.AnyArg()).WillReturnRows(
		sqlmock.NewRows([]string{"header_url", "url_expires_at", "updated_at"}).AddRow("https://example.com/new", time.Now().Add(attachments.PRESIGNED_URL_EXPIRY), time.Now()),
	)

	header, err := repo.GetHeader(context.Background(), noteID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("expected header, got nil")
	}
	if header.HeaderURL != "https://example.com/new" {
		t.Fatalf("expected refreshed header URL, got %s", header.HeaderURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_Success(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "headers-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/header", nil)
	mock.ExpectQuery(queryRegexp(CREATE_HEADER)).WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), "https://example.com/header", sqlmock.AnyArg()).WillReturnRows(
		sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(headerID, noteID, "header-key", "https://example.com/header", createdAt.Add(attachments.PRESIGNED_URL_EXPIRY), createdAt, updatedAt),
	)

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("expected header, got nil")
	}
	if header.HeaderURL != "https://example.com/header" {
		t.Fatalf("expected header URL https://example.com/header, got %s", header.HeaderURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHeader_NotFound(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)

	err := repo.DeleteHeader(context.Background(), noteID)
	if !errors.Is(err, attachments.ErrHeaderNotFound) {
		t.Fatalf("expected ErrHeaderNotFound, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHeader_Success(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(
		sqlmock.NewRows([]string{"minio_key"}).AddRow("header-key"),
	)
	minioMock.EXPECT().DeleteFile(gomock.Any(), "headers-bucket", "header-key").Return(nil)

	err := repo.DeleteHeader(context.Background(), noteID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_NotFound(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)

	header, err := repo.GetHeader(context.Background(), noteID)
	if !errors.Is(err, attachments.ErrHeaderNotFound) {
		t.Fatalf("expected ErrHeaderNotFound, got %v", err)
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_ReplaceExistingHeader(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Second)
	updatedAt := createdAt

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(
		sqlmock.NewRows([]string{"minio_key"}).AddRow("old-header-key"),
	)
	minioMock.EXPECT().DeleteFile(gomock.Any(), "headers-bucket", "old-header-key").Return(nil)
	minioMock.EXPECT().UploadFile(gomock.Any(), "headers-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/header", nil)
	mock.ExpectQuery(queryRegexp(CREATE_HEADER)).WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), "https://example.com/header", sqlmock.AnyArg()).WillReturnRows(
		sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
			AddRow(headerID, noteID, "header-key", "https://example.com/header", createdAt.Add(attachments.PRESIGNED_URL_EXPIRY), createdAt, updatedAt),
	)

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("expected header, got nil")
	}
	if header.HeaderURL != "https://example.com/header" {
		t.Fatalf("expected header URL, got %s", header.HeaderURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_GenerateURLFailureDeletesFile(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "headers-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("url failed"))
	minioMock.EXPECT().DeleteFile(gomock.Any(), "headers-bucket", gomock.Any()).Return(nil)

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if !errors.Is(err, attachments.ErrFailedToGenerateURL) {
		t.Fatalf("expected ErrFailedToGenerateURL, got %v", err)
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHeader_DeleteFileError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(
		sqlmock.NewRows([]string{"minio_key"}).AddRow("header-key"),
	)
	minioMock.EXPECT().DeleteFile(gomock.Any(), "headers-bucket", "header-key").Return(errors.New("delete failed"))

	err := repo.DeleteHeader(context.Background(), noteID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_SuccessNoRefresh(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	future := time.Now().Add(time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(headerID, noteID, "header-key", "https://example.com/header", future, now, now)

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(rows)

	header, err := repo.GetHeader(context.Background(), noteID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if header == nil {
		t.Fatal("expected header, got nil")
	}
	if header.HeaderURL != "https://example.com/header" {
		t.Fatalf("expected header URL, got %s", header.HeaderURL)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_UploadFailure(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "headers-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(errors.New("upload failed"))

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if !errors.Is(err, attachments.ErrFailedToUpload) {
		t.Fatalf("expected ErrFailedToUpload, got %v", err)
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAttachment_DeleteFileError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_ATTACHMENT_BY_ID)).WithArgs(blockID).WillReturnRows(
		sqlmock.NewRows([]string{"minio_key"}).AddRow("attachment-key"),
	)
	minioMock.EXPECT().DeleteFile(gomock.Any(), "attachments-bucket", "attachment-key").Return(errors.New("delete failed"))

	err := repo.DeleteAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_DBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(errors.New("conn refused"))

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, attachments.ErrAttachmentNotFound) {
		t.Fatalf("expected non-not-found error, got %v", err)
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_ExpiredGeneratePresignedURLError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(attachmentID, blockID, "attachment-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", "attachment-key", attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("presign fail"))

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_ExpiredUpdateURLError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()
	attachmentID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(attachmentID, blockID, "attachment-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", "attachment-key", attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/new", nil)
	mock.ExpectQuery(queryRegexp(UPDATE_ATTACHMENT_URL)).WithArgs(attachmentID, "https://example.com/new", sqlmock.AnyArg()).WillReturnError(errors.New("update failed"))

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_GetAttachmentDBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(errors.New("conn refused"))

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadAttachment_CreateDBErrorDeletesFile(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "attachments-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "attachments-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/file", nil)
	mock.ExpectQuery(queryRegexp(CREATE_ATTACHMENT)).WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), "https://example.com/file", sqlmock.AnyArg()).WillReturnError(errors.New("insert failed"))
	minioMock.EXPECT().DeleteFile(gomock.Any(), "attachments-bucket", gomock.Any()).Return(nil)

	attachment, err := repo.UploadAttachment(context.Background(), blockID, "file.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteAttachment_DBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_ATTACHMENT_BY_ID)).WithArgs(blockID).WillReturnError(errors.New("conn refused"))

	err := repo.DeleteAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, attachments.ErrAttachmentNotFound) {
		t.Fatalf("expected non-not-found error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_DBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(errors.New("conn refused"))

	header, err := repo.GetHeader(context.Background(), noteID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, attachments.ErrHeaderNotFound) {
		t.Fatalf("expected non-not-found error, got %v", err)
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_ExpiredGeneratePresignedURLError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(headerID, noteID, "header-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", "header-key", attachments.PRESIGNED_URL_EXPIRY).Return("", errors.New("presign fail"))

	header, err := repo.GetHeader(context.Background(), noteID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetHeader_ExpiredUpdateURLError(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()
	headerID := uuid.New()
	past := time.Now().Add(-time.Hour).UTC().Truncate(time.Second)
	now := time.Now().UTC().Truncate(time.Second)

	rows := sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(headerID, noteID, "header-key", "https://example.com/old", past, now, now)

	mock.ExpectQuery(queryRegexp(GET_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnRows(rows)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", "header-key", attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/new", nil)
	mock.ExpectQuery(queryRegexp(UPDATE_HEADER_URL)).WithArgs(headerID, "https://example.com/new", sqlmock.AnyArg()).WillReturnError(errors.New("update failed"))

	header, err := repo.GetHeader(context.Background(), noteID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_DeleteHeaderDBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(errors.New("conn refused"))

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestUploadHeader_CreateDBErrorDeletesFile(t *testing.T) {
	repo, mock, _, minioMock := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(sql.ErrNoRows)
	minioMock.EXPECT().UploadFile(gomock.Any(), "headers-bucket", gomock.Any(), gomock.Any(), int64(123), "image/png").Return(nil)
	minioMock.EXPECT().GeneratePresignedURL(gomock.Any(), "headers-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).Return("https://example.com/header", nil)
	mock.ExpectQuery(queryRegexp(CREATE_HEADER)).WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), "https://example.com/header", sqlmock.AnyArg()).WillReturnError(errors.New("insert failed"))
	minioMock.EXPECT().DeleteFile(gomock.Any(), "headers-bucket", gomock.Any()).Return(nil)

	header, err := repo.UploadHeader(context.Background(), noteID, "header.png", 123, "image/png", bytes.NewReader([]byte("data")))
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if header != nil {
		t.Fatalf("expected nil header, got %#v", header)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestDeleteHeader_DBError(t *testing.T) {
	repo, mock, _, _ := newMockRepository(t)
	noteID := uuid.New()

	mock.ExpectQuery(queryRegexp(DELETE_HEADER_BY_NOTE_ID)).WithArgs(noteID).WillReturnError(errors.New("conn refused"))

	err := repo.DeleteHeader(context.Background(), noteID)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if errors.Is(err, attachments.ErrHeaderNotFound) {
		t.Fatalf("expected non-not-found error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}

func TestGetAttachment_ExpiredRefreshAndScanError(t *testing.T) {
	// Cover the rows-Scan error branch on attachment fetch by returning a row with bad type.
	repo, mock, _, _ := newMockRepository(t)
	blockID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow("not-a-uuid", blockID, "k", "u", time.Now(), time.Now(), time.Now())

	mock.ExpectQuery(queryRegexp(GET_ATTACHMENT_BY_BLOCK_ID)).WithArgs(blockID).WillReturnRows(rows)

	attachment, err := repo.GetAttachment(context.Background(), blockID)
	if err == nil {
		t.Fatal("expected scan error, got nil")
	}
	if attachment != nil {
		t.Fatalf("expected nil attachment, got %#v", attachment)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatal(err)
	}
}
