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

type MinIOService interface {
	UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, key string) error
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
}

type AttachmentRepository struct {
	db    *sql.DB
	minio MinIOService
}

func NewAttachmentRepository(db *sql.DB, minio MinIOService) *AttachmentRepository {
	return &AttachmentRepository{
		db:    db,
		minio: minio,
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
			return nil, nil
		}
		return nil, err
	}

	return attachment, nil
}

func (r *AttachmentRepository) UpdateAttachmentURL(ctx context.Context, attachmentID uuid.UUID, url string, expiresAt time.Time) error {
	var returnedURL string
	var returnedExpiresAt time.Time
	var returnedUpdatedAt time.Time

	err := r.db.QueryRowContext(ctx, UPDATE_ATTACHMENT_URL, url, expiresAt, time.Now(), attachmentID).Scan(
		&returnedURL,
		&returnedExpiresAt,
		&returnedUpdatedAt,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *AttachmentRepository) UploadAttachment(
	ctx context.Context,
	blockID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
	presignedURLExpiry time.Duration,
) (*models.Attachment, error) {
	attachmentID := uuid.New()
	minioKey := attachmentID.String()

	if err := r.minio.UploadFile(ctx, minioKey, fileReader, fileSize, mimeType); err != nil {
		return nil, err
	}

	presignedURL, err := r.minio.GeneratePresignedURL(ctx, minioKey, presignedURLExpiry)
	if err != nil {
		_ = r.minio.DeleteFile(ctx, minioKey)
		return nil, attachments.ErrFailedToGenerateURL
	}

	now := time.Now()
	attachment := &models.Attachment{
		ID:           attachmentID,
		BlockID:      blockID,
		MinioKey:     minioKey,
		AttachURL:    presignedURL,
		URLExpiresAt: now.Add(presignedURLExpiry),
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
		attachment.CreatedAt,
		attachment.UpdatedAt,
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
		_ = r.minio.DeleteFile(ctx, minioKey)
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

	if err := r.minio.DeleteFile(ctx, minioKey); err != nil {
		return err
	}

	return nil
}

func (r *AttachmentRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	var note models.Note
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, GET_NOTE_BY_ID, noteID).Scan(
		&note.ID, &note.UserID, &note.Title, &parentID, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return nil, err
		}
		note.ParentID = &pid
	}

	return &note, nil
}

func (r *AttachmentRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	var block models.Block

	err := r.db.QueryRowContext(ctx, GET_BLOCK_BY_ID, blockID).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	block.States = []models.BlockState{}

	return &block, nil
}
