package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_usecase_attachments.go -package=mocks

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
}

type NoteUsecase interface {
	CheckNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error)
	CheckBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, error)
}

type attachmentUsecase struct {
	attachmentRepo AttachmentRepository
	noteUsecase    NoteUsecase
}

func NewAttachmentUsecase(attachmentRepo AttachmentRepository, noteUsecase NoteUsecase) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		noteUsecase:    noteUsecase,
	}
}

func (u *attachmentUsecase) GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error) {
	_, err := u.noteUsecase.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.noteUsecase.CheckBlockAccess(ctx, noteID, blockID)
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
	_, err := u.noteUsecase.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.noteUsecase.CheckBlockAccess(ctx, noteID, blockID)
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
	_, err := u.noteUsecase.CheckNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	_, err = u.noteUsecase.CheckBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return err
	}

	if err := u.attachmentRepo.DeleteAttachment(ctx, blockID); err != nil {
		return err
	}

	return nil
}
