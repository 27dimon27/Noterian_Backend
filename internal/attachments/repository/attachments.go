package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_repository_minio.go -package=mocks

type MinIOService interface {
	UploadFile(ctx context.Context, bucketName, key string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, bucketName, key string) error
	GeneratePresignedURL(ctx context.Context, bucketName, key string, expiry time.Duration) (string, error)
}

type AttachmentRepository struct {
	db               *sql.DB
	minio            MinIOService
	attachmentBucket string
	headerBucket     string
	logger           *slog.Logger
}

func NewAttachmentRepository(db *sql.DB, minio MinIOService, attachmentBucket, headerBucket string, logger *slog.Logger) *AttachmentRepository {
	return &AttachmentRepository{
		db:               db,
		minio:            minio,
		attachmentBucket: attachmentBucket,
		headerBucket:     headerBucket,
		logger:           logger,
	}
}

func (r *AttachmentRepository) GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error) {
	attachment := &models.Attachment{}

	err := r.db.QueryRowContext(ctx, GET_ATTACHMENT_BY_BLOCK_ID, blockID).Scan(
		&attachment.ID,
		&attachment.BlockID,
		&attachment.MinioKey,
		&attachment.AttachURL,
		&attachment.URLExpiresAt,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Attachment not found")
			return nil, attachments.ErrAttachmentNotFound
		}
		r.logger.Error("Internal server error", "error", err)
		return nil, err
	}

	if time.Now().After(attachment.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, attachment.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			r.logger.Error("Internal server error", "error", err)
			return nil, err
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		err = r.updateAttachmentURL(ctx, attachment.ID, newURL, newExpiry)
		if err != nil {
			r.logger.Error("Internal server error", "error", err)
			return nil, err
		}

		attachment.AttachURL = newURL
		attachment.URLExpiresAt = newExpiry
		attachment.UpdatedAt = time.Now()
	}

	return attachment, nil
}

func (r *AttachmentRepository) UploadAttachment(
	ctx context.Context,
	blockID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Attachment, error) {
	existingAttach, err := r.GetAttachment(ctx, blockID)
	if err != nil && !errors.Is(err, attachments.ErrAttachmentNotFound) {
		r.logger.Error("Internal server error", "error", err)
		return nil, err
	}

	if existingAttach != nil {
		r.logger.Warn("Block already has aatch")
		return nil, attachments.ErrBlockAlreadyHasAttach
	}

	attachmentID := uuid.New()
	minioKey := attachmentID.String()

	if err := r.minio.UploadFile(ctx, r.attachmentBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		r.logger.Error("Internal server error", "error", err)
		return nil, attachments.ErrFailedToUpload
	}

	presignedURL, err := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, minioKey, attachments.PRESIGNED_URL_EXPIRY)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		if delErr := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); delErr != nil {
			r.logger.Error("Internal server error", "error", delErr)
			return nil, fmt.Errorf("generate presigned URL failed: %w, and cleanup failed: %w", err, delErr)
		}
		return nil, attachments.ErrFailedToGenerateURL
	}

	now := time.Now()
	attachment := &models.Attachment{
		ID:           attachmentID,
		BlockID:      blockID,
		MinioKey:     minioKey,
		AttachURL:    presignedURL,
		URLExpiresAt: now.Add(attachments.PRESIGNED_URL_EXPIRY),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = r.db.QueryRowContext(
		ctx,
		CREATE_ATTACHMENT,
		attachment.ID,
		attachment.BlockID,
		attachment.MinioKey,
		attachment.AttachURL,
		attachment.URLExpiresAt,
	).Scan(
		&attachment.ID,
		&attachment.BlockID,
		&attachment.MinioKey,
		&attachment.AttachURL,
		&attachment.URLExpiresAt,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		if delErr := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); delErr != nil {
			r.logger.Error("Internal server error", "error", delErr)
			return nil, fmt.Errorf("generate presigned URL failed: %w, and cleanup failed: %w", err, delErr)
		}
		return nil, err
	}

	return attachment, nil
}

func (r *AttachmentRepository) DeleteAttachment(ctx context.Context, blockID uuid.UUID) error {
	var minioKey string

	err := r.db.QueryRowContext(ctx, DELETE_ATTACHMENT_BY_ID, blockID).Scan(&minioKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Attachment not found")
			return attachments.ErrAttachmentNotFound
		}
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	if err := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); err != nil {
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	return nil
}

func (r *AttachmentRepository) GetHeader(ctx context.Context, noteID uuid.UUID) (*models.Header, error) {
	header := &models.Header{}

	err := r.db.QueryRowContext(ctx, GET_HEADER_BY_NOTE_ID, noteID).Scan(
		&header.ID,
		&header.NoteID,
		&header.MinioKey,
		&header.HeaderURL,
		&header.URLExpiresAt,
		&header.CreatedAt,
		&header.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Attachment not found")
			return nil, attachments.ErrHeaderNotFound
		}
		r.logger.Error("Internal server error", "error", err)
		return nil, err
	}

	if time.Now().After(header.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.headerBucket, header.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			r.logger.Error("Internal server error", "error", err)
			return nil, err
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		err = r.updateHeaderURL(ctx, header.ID, newURL, newExpiry)
		if err != nil {
			r.logger.Error("Internal server error", "error", err)
			return nil, err
		}

		header.HeaderURL = newURL
		header.URLExpiresAt = newExpiry
		header.UpdatedAt = time.Now()
	}

	return header, nil
}

func (r *AttachmentRepository) UploadHeader(
	ctx context.Context,
	noteID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Header, error) {
	err := r.DeleteHeader(ctx, noteID)
	if err != nil && !errors.Is(err, attachments.ErrHeaderNotFound) {
		r.logger.Error("Internal server error", "error", err)
		return nil, err
	}

	headerID := uuid.New()
	minioKey := headerID.String()

	if err := r.minio.UploadFile(ctx, r.headerBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		r.logger.Error("Internal server error", "error", err)
		return nil, attachments.ErrFailedToUpload
	}

	presignedURL, err := r.minio.GeneratePresignedURL(ctx, r.headerBucket, minioKey, attachments.PRESIGNED_URL_EXPIRY)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		if delErr := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); delErr != nil {
			r.logger.Error("Internal server error", "error", delErr)
			return nil, fmt.Errorf("generate presigned URL failed: %w, and cleanup failed: %w", err, delErr)
		}
		return nil, attachments.ErrFailedToGenerateURL
	}

	now := time.Now()
	header := &models.Header{
		ID:           headerID,
		NoteID:       noteID,
		MinioKey:     minioKey,
		HeaderURL:    presignedURL,
		URLExpiresAt: now.Add(attachments.PRESIGNED_URL_EXPIRY),
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	err = r.db.QueryRowContext(
		ctx,
		CREATE_HEADER,
		header.ID,
		header.NoteID,
		header.MinioKey,
		header.HeaderURL,
		header.URLExpiresAt,
	).Scan(
		&header.ID,
		&header.NoteID,
		&header.MinioKey,
		&header.HeaderURL,
		&header.URLExpiresAt,
		&header.CreatedAt,
		&header.UpdatedAt,
	)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		if delErr := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); delErr != nil {
			r.logger.Error("Internal server error", "error", delErr)
			return nil, fmt.Errorf("generate presigned URL failed: %w, and cleanup failed: %w", err, delErr)
		}
		return nil, err
	}

	return header, nil
}

func (r *AttachmentRepository) DeleteHeader(ctx context.Context, noteID uuid.UUID) error {
	var minioKey string

	err := r.db.QueryRowContext(ctx, DELETE_HEADER_BY_NOTE_ID, noteID).Scan(&minioKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Header not found")
			return attachments.ErrHeaderNotFound
		}
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	if err := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); err != nil {
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	return nil
}

func (r *AttachmentRepository) updateAttachmentURL(ctx context.Context, attachmentID uuid.UUID, url string, expiresAt time.Time) error {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	err := r.db.QueryRowContext(ctx, UPDATE_ATTACHMENT_URL, attachmentID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	return nil
}

func (r *AttachmentRepository) updateHeaderURL(ctx context.Context, headerID uuid.UUID, url string, expiresAt time.Time) error {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	err := r.db.QueryRowContext(ctx, UPDATE_HEADER_URL, headerID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	)
	if err != nil {
		r.logger.Error("Internal server error", "error", err)
		return err
	}

	return nil
}
