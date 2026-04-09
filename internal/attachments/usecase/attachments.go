package usecase

import (
	"context"
	"io"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader, presignedURLExpiry time.Duration) (*models.Attachment, error)
	UpdateAttachmentURL(ctx context.Context, attachmentID uuid.UUID, url string, expiresAt time.Time) error
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
}

type MinIOService interface {
	GeneratePresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	UploadFile(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
	DeleteFile(ctx context.Context, key string) error
}

type attachmentUsecase struct {
	attachmentRepo AttachmentRepository
	minioService   MinIOService
}

func NewAttachmentUsecase(attachmentRepo AttachmentRepository, minioService MinIOService) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		minioService:   minioService,
	}
}

func (u *attachmentUsecase) GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	attachment, err := u.attachmentRepo.GetAttachment(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if attachment == nil {
		return nil, attachments.ErrAttachmentNotFound
	}

	if time.Now().After(attachment.URLExpiresAt) {
		newURL, err := u.minioService.GeneratePresignedURL(ctx, attachment.MinioKey, attachments.PRESIGNED_URL_EXPIRY)
		if err != nil {
			return nil, err
		}

		newExpiry := time.Now().Add(attachments.PRESIGNED_URL_EXPIRY)

		err = u.attachmentRepo.UpdateAttachmentURL(ctx, attachment.ID, newURL, newExpiry)
		if err != nil {
			return nil, err
		}

		attachment.AttachURL = newURL
		attachment.URLExpiresAt = newExpiry
		attachment.UpdatedAt = time.Now()
	}

	return attachment, nil
}

func (u *attachmentUsecase) UploadAttachment(
	ctx context.Context,
	noteID uuid.UUID,
	blockID uuid.UUID,
	userID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Attachment, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	if !attachments.AllowedMimeTypes[mimeType] {
		return nil, attachments.ErrInvalidMimeType
	}

	if fileSize > attachments.MAX_FILE_SIZE {
		return nil, attachments.ErrFileTooLarge
	}

	existingAttach, err := u.attachmentRepo.GetAttachment(ctx, blockID)
	if err != nil {
		return nil, err
	}
	if existingAttach != nil {
		return nil, attachments.ErrBlockAlreadyHasAttach
	}

	attachment, err := u.attachmentRepo.UploadAttachment(ctx, blockID, fileName, fileSize, mimeType, fileReader, attachments.PRESIGNED_URL_EXPIRY)
	if err != nil {
		return nil, attachments.ErrFailedToUpload
	}

	return attachment, nil
}

func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return err
	}

	if err := u.attachmentRepo.DeleteAttachment(ctx, blockID); err != nil {
		return err
	}

	return nil
}

func (u *attachmentUsecase) checkNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.attachmentRepo.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, attachments.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, attachments.ErrForbidden
	}

	return note, nil
}

func (u *attachmentUsecase) checkBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, error) {
	block, err := u.attachmentRepo.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, attachments.ErrBlockNotFound
	}

	if block.NoteID != noteID {
		return nil, attachments.ErrForbidden
	}

	return block, nil
}
