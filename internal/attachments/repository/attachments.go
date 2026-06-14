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
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
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

func (r *AttachmentRepository) GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, types.AppErrorInterface) {
	attachment := &models.Attachment{}

	if err := r.db.QueryRowContext(ctx, GET_ATTACHMENT_BY_BLOCK_ID, blockID).Scan(
		&attachment.ID,
		&attachment.BlockID,
		&attachment.MinioKey,
		&attachment.AttachURL,
		&attachment.URLExpiresAt,
		&attachment.CreatedAt,
		&attachment.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        attachments.ErrAttachmentNotFound,
				PublicMsg:  attachments.PublicMsgErrAttachmentNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetAttachment",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetAttachment",
		}
	}

	if time.Now().After(attachment.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, attachment.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetAttachment",
			}
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		if err := r.updateAttachmentURL(ctx, attachment.ID, newURL, newExpiry); err != nil {
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
) (*models.Attachment, types.AppErrorInterface) {
	existingAttach, err := r.GetAttachment(ctx, blockID)
	if err != nil && !err.Is(attachments.ErrAttachmentNotFound) {
		return nil, err
	}

	if existingAttach != nil {
		return nil, &types.AppError{
			Err:        attachments.ErrBlockAlreadyHasAttach,
			PublicMsg:  attachments.PublicMsgErrBlockAlreadyHasAttach,
			StatusCode: 409,
			Layer:      "repo",
			Op:         "UploadAttachment",
		}
	}

	attachmentID := uuid.New()
	minioKey := attachmentID.String()

	if err := r.minio.UploadFile(ctx, r.attachmentBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAttachment",
		}
	}

	presignedURL, generateErr := r.minio.GeneratePresignedURL(ctx, r.attachmentBucket, minioKey, attachments.PRESIGNED_URL_EXPIRY)
	if generateErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", generateErr, delErr),
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadAttachment",
			}
		}
		return nil, &types.AppError{
			Err:        generateErr,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAttachment",
		}
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

	if queryErr := r.db.QueryRowContext(
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
	); queryErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", queryErr, delErr),
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadAttachment",
			}
		}
		return nil, &types.AppError{
			Err:        queryErr,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadAttachment",
		}
	}

	return attachment, nil
}

func (r *AttachmentRepository) DeleteAttachment(ctx context.Context, blockID uuid.UUID) types.AppErrorInterface {
	var minioKey string

	if err := r.db.QueryRowContext(ctx, DELETE_ATTACHMENT_BY_ID, blockID).Scan(&minioKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.AppError{
				Err:        attachments.ErrAttachmentNotFound,
				PublicMsg:  attachments.PublicMsgErrAttachmentNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteAttachment",
			}
		}
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteAttachment",
		}
	}

	if err := r.minio.DeleteFile(ctx, r.attachmentBucket, minioKey); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteAttachment",
		}
	}

	return nil
}

func (r *AttachmentRepository) GetHeader(ctx context.Context, noteID uuid.UUID) (*models.Header, types.AppErrorInterface) {
	header := &models.Header{}

	if err := r.db.QueryRowContext(ctx, GET_HEADER_BY_NOTE_ID, noteID).Scan(
		&header.ID,
		&header.NoteID,
		&header.MinioKey,
		&header.HeaderURL,
		&header.URLExpiresAt,
		&header.CreatedAt,
		&header.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        attachments.ErrHeaderNotFound,
				PublicMsg:  attachments.PublicMsgErrHeaderNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetHeader",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetHeader",
		}
	}

	if time.Now().After(header.URLExpiresAt) {
		newURL, err := r.minio.GeneratePresignedURL(ctx, r.headerBucket, header.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetHeader",
			}
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		if err := r.updateHeaderURL(ctx, header.ID, newURL, newExpiry); err != nil {
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
) (*models.Header, types.AppErrorInterface) {
	if err := r.DeleteHeader(ctx, noteID); err != nil && !err.Is(attachments.ErrHeaderNotFound) {
		return nil, err
	}

	headerID := uuid.New()
	minioKey := headerID.String()

	if err := r.minio.UploadFile(ctx, r.headerBucket, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadHeader",
		}
	}

	presignedURL, generateErr := r.minio.GeneratePresignedURL(ctx, r.headerBucket, minioKey, attachments.PRESIGNED_URL_EXPIRY)
	if generateErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", generateErr, delErr),
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadHeader",
			}
		}
		return nil, &types.AppError{
			Err:        generateErr,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadHeader",
		}
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

	if queryErr := r.db.QueryRowContext(
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
	); queryErr != nil {
		if delErr := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); delErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", queryErr, delErr),
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "UploadHeader",
			}
		}
		return nil, &types.AppError{
			Err:        queryErr,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UploadHeader",
		}
	}

	return header, nil
}

func (r *AttachmentRepository) DeleteHeader(ctx context.Context, noteID uuid.UUID) types.AppErrorInterface {
	var minioKey string

	if err := r.db.QueryRowContext(ctx, DELETE_HEADER_BY_NOTE_ID, noteID).Scan(&minioKey); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.AppError{
				Err:        attachments.ErrHeaderNotFound,
				PublicMsg:  attachments.PublicMsgErrHeaderNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteHeader",
			}
		}
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteHeader",
		}
	}

	if err := r.minio.DeleteFile(ctx, r.headerBucket, minioKey); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteHeader",
		}
	}

	return nil
}

func (r *AttachmentRepository) updateAttachmentURL(ctx context.Context, attachmentID uuid.UUID, url string, expiresAt time.Time) types.AppErrorInterface {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	if err := r.db.QueryRowContext(ctx, UPDATE_ATTACHMENT_URL, attachmentID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "updateAttachmentURL",
		}
	}

	return nil
}

func (r *AttachmentRepository) updateHeaderURL(ctx context.Context, headerID uuid.UUID, url string, expiresAt time.Time) types.AppErrorInterface {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	if err := r.db.QueryRowContext(ctx, UPDATE_HEADER_URL, headerID, url, expiresAt).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	); err != nil {
		return &types.AppError{
			Err:        err,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "updateHeaderURL",
		}
	}

	return nil
}
