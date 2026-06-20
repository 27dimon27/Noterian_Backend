package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"log/slog"
	"os"
	"regexp"
	"strings"
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

func newTestRepo(t *testing.T) (*AttachmentRepository, sqlmock.Sqlmock, *mocks.MockMinIOService, func()) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	minioMock := mocks.NewMockMinIOService(ctrl)

	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	repo := NewAttachmentRepository(db, minioMock, "attach-bucket", "header-bucket", logger)

	cleanup := func() {
		_ = db.Close()
		ctrl.Finish()
	}

	return repo, mock, minioMock, cleanup
}

func attachmentRows(id, blockID uuid.UUID, minioKey, url string, expiresAt, createdAt, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "block_id", "minio_key", "attach_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(id, blockID, minioKey, url, expiresAt, createdAt, updatedAt)
}

func headerRows(id, noteID uuid.UUID, minioKey, url string, expiresAt, createdAt, updatedAt time.Time) *sqlmock.Rows {
	return sqlmock.NewRows([]string{"id", "note_id", "minio_key", "header_url", "url_expires_at", "created_at", "updated_at"}).
		AddRow(id, noteID, minioKey, url, expiresAt, createdAt, updatedAt)
}

func reQuote(s string) string {
	return regexp.QuoteMeta(strings.TrimSpace(s))
}

func TestGetAttachment_Success_NotExpired(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	attachID := uuid.New()
	now := time.Now()
	future := now.Add(time.Hour)

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(attachmentRows(attachID, blockID, "key-1", "http://example.com/1", future, now, now))

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, attachID, got.ID)
	assert.Equal(t, blockID, got.BlockID)
	assert.Equal(t, "key-1", got.MinioKey)
	assert.Equal(t, "http://example.com/1", got.AttachURL)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttachment_NotFound(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrAttachmentNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttachment_DBError(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	dbErr := errors.New("connection lost")

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(dbErr)

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.False(t, customErr.Is(attachments.ErrAttachmentNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttachment_ExpiredURL_RefreshedSuccessfully(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	attachID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(attachmentRows(attachID, blockID, "key-1", "http://old-url", past, now, now))

	newURL := "http://new-url"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", "key-1", attachments.PRESIGNED_URL_EXPIRY).
		Return(newURL, nil)

	mock.ExpectQuery(reQuote(UPDATE_ATTACHMENT_URL)).
		WithArgs(attachID, newURL, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"attach_url", "url_expires_at", "updated_at"}).
			AddRow(newURL, now.Add(time.Hour), now))

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, newURL, got.AttachURL)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttachment_ExpiredURL_PresignFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	attachID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(attachmentRows(attachID, blockID, "key-1", "http://old-url", past, now, now))

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", "key-1", attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("minio down"))

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetAttachment_ExpiredURL_UpdateFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	attachID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(attachmentRows(attachID, blockID, "key-1", "http://old-url", past, now, now))

	newURL := "http://new-url"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", "key-1", attachments.PRESIGNED_URL_EXPIRY).
		Return(newURL, nil)

	mock.ExpectQuery(reQuote(UPDATE_ATTACHMENT_URL)).
		WithArgs(attachID, newURL, sqlmock.AnyArg()).
		WillReturnError(errors.New("update failed"))

	got, customErr := repo.GetAttachment(context.Background(), blockID)
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_Success(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	now := time.Now()
	reader := strings.NewReader("file-bytes")

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(nil)

	presigned := "http://presigned-url"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_ATTACHMENT)).
		WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnRows(attachmentRows(uuid.New(), blockID, "generated-key", presigned, now.Add(time.Hour), now, now))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", reader)
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, blockID, got.BlockID)
	assert.Equal(t, presigned, got.AttachURL)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_AlreadyExists(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	attachID := uuid.New()
	now := time.Now()
	future := now.Add(time.Hour)

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnRows(attachmentRows(attachID, blockID, "key", "http://x", future, now, now))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrBlockAlreadyHasAttach))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_GetAttachmentDBErrorPropagates(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(errors.New("db is down"))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.False(t, customErr.Is(attachments.ErrBlockAlreadyHasAttach))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_UploadFileFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(errors.New("upload failed"))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_PresignFails_DeleteSucceeds(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("presign failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", gomock.Any()).
		Return(nil)

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_PresignFails_DeleteAlsoFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("presign failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", gomock.Any()).
		Return(errors.New("delete also failed"))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_DBInsertFails_DeleteSucceeds(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	presigned := "http://presigned-url"

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_ATTACHMENT)).
		WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", gomock.Any()).
		Return(nil)

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_DBInsertFails_DeleteAlsoFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	presigned := "http://presigned-url"

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(10), "text/plain").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_ATTACHMENT)).
		WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", gomock.Any()).
		Return(errors.New("cleanup failed too"))

	got, customErr := repo.UploadAttachment(context.Background(), blockID, "file.txt", 10, "text/plain", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAttachment_Success(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_ATTACHMENT_BY_ID)).
		WithArgs(blockID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("key-1"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", "key-1").
		Return(nil)

	customErr := repo.DeleteAttachment(context.Background(), blockID)
	require.Nil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAttachment_NotFound(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_ATTACHMENT_BY_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	customErr := repo.DeleteAttachment(context.Background(), blockID)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrAttachmentNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAttachment_DBError(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_ATTACHMENT_BY_ID)).
		WithArgs(blockID).
		WillReturnError(errors.New("db error"))

	customErr := repo.DeleteAttachment(context.Background(), blockID)
	require.NotNil(t, customErr)
	assert.False(t, customErr.Is(attachments.ErrAttachmentNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteAttachment_MinioDeleteFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_ATTACHMENT_BY_ID)).
		WithArgs(blockID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("key-1"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "attach-bucket", "key-1").
		Return(errors.New("minio delete failed"))

	customErr := repo.DeleteAttachment(context.Background(), blockID)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_Success_NotExpired(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	headerID := uuid.New()
	now := time.Now()
	future := now.Add(time.Hour)

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(headerRows(headerID, noteID, "h-key", "http://header-url", future, now, now))

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, headerID, got.ID)
	assert.Equal(t, noteID, got.NoteID)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_NotFound(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, got)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrHeaderNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_DBError(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(errors.New("db error"))

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_ExpiredURL_RefreshedSuccessfully(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	headerID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(headerRows(headerID, noteID, "h-key", "http://old", past, now, now))

	newURL := "http://new-header-url"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", "h-key", attachments.PRESIGNED_URL_EXPIRY).
		Return(newURL, nil)

	mock.ExpectQuery(reQuote(UPDATE_HEADER_URL)).
		WithArgs(headerID, newURL, sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"header_url", "url_expires_at", "updated_at"}).
			AddRow(newURL, now.Add(time.Hour), now))

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, newURL, got.HeaderURL)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_ExpiredURL_PresignFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	headerID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(headerRows(headerID, noteID, "h-key", "http://old", past, now, now))

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", "h-key", attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("presign failed"))

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetHeader_ExpiredURL_UpdateFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	headerID := uuid.New()
	now := time.Now()
	past := now.Add(-time.Hour)

	mock.ExpectQuery(reQuote(GET_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(headerRows(headerID, noteID, "h-key", "http://old", past, now, now))

	newURL := "http://new-header-url"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", "h-key", attachments.PRESIGNED_URL_EXPIRY).
		Return(newURL, nil)

	mock.ExpectQuery(reQuote(UPDATE_HEADER_URL)).
		WithArgs(headerID, newURL, sqlmock.AnyArg()).
		WillReturnError(errors.New("update failed"))

	got, customErr := repo.GetHeader(context.Background(), noteID)
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_Success_NoExistingHeader(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	presigned := "http://header-presigned"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_HEADER)).
		WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnRows(headerRows(uuid.New(), noteID, "new-key", presigned, now.Add(time.Hour), now, now))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("imagebytes12345678901"))
	require.Nil(t, customErr)
	require.NotNil(t, got)
	assert.Equal(t, noteID, got.NoteID)
	assert.Equal(t, presigned, got.HeaderURL)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_Success_ReplacesExistingHeader(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("old-key"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", "old-key").
		Return(nil)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	presigned := "http://header-presigned"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_HEADER)).
		WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnRows(headerRows(uuid.New(), noteID, "new-key", presigned, now.Add(time.Hour), now, now))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("imagebytes12345678901"))
	require.Nil(t, customErr)
	require.NotNil(t, got)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_DeleteHeaderDBErrorPropagates(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(errors.New("db is down"))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_UploadFileFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(errors.New("upload failed"))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_PresignFails_DeleteSucceeds(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("presign failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", gomock.Any()).
		Return(nil)

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_PresignFails_DeleteAlsoFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return("", errors.New("presign failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", gomock.Any()).
		Return(errors.New("cleanup failed too"))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_DBInsertFails_DeleteSucceeds(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	presigned := "http://header-presigned"

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_HEADER)).
		WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", gomock.Any()).
		Return(nil)

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_DBInsertFails_DeleteAlsoFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()
	presigned := "http://header-presigned"

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	minioMock.EXPECT().
		UploadFile(gomock.Any(), "header-bucket", gomock.Any(), gomock.Any(), int64(20), "image/png").
		Return(nil)

	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "header-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_HEADER)).
		WithArgs(sqlmock.AnyArg(), noteID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnError(errors.New("insert failed"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", gomock.Any()).
		Return(errors.New("cleanup failed too"))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadHeader_ExistingHeaderMinioDeleteFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("old-key"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", "old-key").
		Return(errors.New("delete failed"))

	got, customErr := repo.UploadHeader(context.Background(), noteID, "logo.png", 20, "image/png", strings.NewReader("data"))
	require.Nil(t, got)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHeader_Success(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("h-key"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", "h-key").
		Return(nil)

	customErr := repo.DeleteHeader(context.Background(), noteID)
	require.Nil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHeader_NotFound(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	customErr := repo.DeleteHeader(context.Background(), noteID)
	require.NotNil(t, customErr)
	assert.True(t, customErr.Is(attachments.ErrHeaderNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHeader_DBError(t *testing.T) {
	repo, mock, _, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnError(errors.New("db error"))

	customErr := repo.DeleteHeader(context.Background(), noteID)
	require.NotNil(t, customErr)
	assert.False(t, customErr.Is(attachments.ErrHeaderNotFound))

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteHeader_MinioDeleteFails(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	noteID := uuid.New()

	mock.ExpectQuery(reQuote(DELETE_HEADER_BY_NOTE_ID)).
		WithArgs(noteID).
		WillReturnRows(sqlmock.NewRows([]string{"minio_key"}).AddRow("h-key"))

	minioMock.EXPECT().
		DeleteFile(gomock.Any(), "header-bucket", "h-key").
		Return(errors.New("minio delete failed"))

	customErr := repo.DeleteHeader(context.Background(), noteID)
	require.NotNil(t, customErr)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUploadAttachment_ReaderIsConsumedRegardlessOfSize(t *testing.T) {
	repo, mock, minioMock, cleanup := newTestRepo(t)
	defer cleanup()

	blockID := uuid.New()
	now := time.Now()

	mock.ExpectQuery(reQuote(GET_ATTACHMENT_BY_BLOCK_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	var capturedReader io.Reader
	minioMock.EXPECT().
		UploadFile(gomock.Any(), "attach-bucket", gomock.Any(), gomock.Any(), int64(4), "text/plain").
		DoAndReturn(func(_ context.Context, _ string, _ string, r io.Reader, _ int64, _ string) error {
			capturedReader = r
			return nil
		})

	presigned := "http://presigned"
	minioMock.EXPECT().
		GeneratePresignedURL(gomock.Any(), "attach-bucket", gomock.Any(), attachments.PRESIGNED_URL_EXPIRY).
		Return(presigned, nil)

	mock.ExpectQuery(reQuote(CREATE_ATTACHMENT)).
		WithArgs(sqlmock.AnyArg(), blockID, sqlmock.AnyArg(), presigned, sqlmock.AnyArg()).
		WillReturnRows(attachmentRows(uuid.New(), blockID, "k", presigned, now.Add(time.Hour), now, now))

	_, customErr := repo.UploadAttachment(context.Background(), blockID, "f.txt", 4, "text/plain", strings.NewReader("data"))
	require.Nil(t, customErr)
	require.NotNil(t, capturedReader)

	require.NoError(t, mock.ExpectationsWereMet())
}
