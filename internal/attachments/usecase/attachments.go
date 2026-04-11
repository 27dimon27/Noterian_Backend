package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

//go:generate go run go.uber.org/mock/mockgen -source=attachments.go -destination=mocks/mock_attachments_repository.go -package=mocks

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
}

type NoteRepository interface {
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
}

type attachmentUsecase struct {
	attachmentRepo AttachmentRepository
	noteRepo       NoteRepository
}

func NewAttachmentUsecase(attachmentRepo AttachmentRepository, noteRepo NoteRepository) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		noteRepo:       noteRepo,
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

	attachment, err := u.attachmentRepo.UploadAttachment(ctx, blockID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		return nil, err
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
	note, err := u.noteRepo.GetNote(ctx, noteID)
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
	block, err := u.noteRepo.GetBlock(ctx, blockID)
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
