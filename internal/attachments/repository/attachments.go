package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
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
}

func NewAttachmentRepository(db *sql.DB, minio MinIOService, attachmentBucket string) *AttachmentRepository {
	return &AttachmentRepository{
		db:               db,
		minio:            minio,
		attachmentBucket: attachmentBucket,
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
			return nil, attachments.ErrAttachmentNotFound
		}
		return nil, err
	}

	if time.Now().After(attachment.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, attachment.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, err
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		err = r.updateAttachmentURL(ctx, attachment.ID, newURL, newExpiry)
		if err != nil {
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
	if err != nil {
		return nil, err
	}

	if existingAttach != nil {
		return nil, attachments.ErrBlockAlreadyHasAttach
	}

	attachmentID := uuid.New()
	minioKey := attachmentID.String()

	if err := r.minio.UploadFile(ctx, r.attachmentBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, attachments.ErrFailedToUpload
	}

	presignedURL, err := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, minioKey, attachments.PRESIGNED_URL_EXPIRY)
	if err != nil {
		_ = r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey)
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
		_ = r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey)
		return nil, err
	}

	return attachment, nil
}

func (r *AttachmentRepository) DeleteAttachment(ctx context.Context, blockID uuid.UUID) error {
	var minioKey string

	err := r.db.QueryRowContext(ctx, DELETE_ATTACHMENT_BY_ID, blockID).Scan(&minioKey)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return attachments.ErrAttachmentNotFound
		}
		return err
	}

	if err := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); err != nil {
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
		return err
	}

	return nil
}
