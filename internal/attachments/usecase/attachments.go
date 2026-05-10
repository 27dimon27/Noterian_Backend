package usecase

import (
	"context"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_usecase_attachments.go -package=mocks

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
}

type NoteRepository interface {
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error)
}

type attachmentUsecase struct {
	attachmentRepository AttachmentRepository
	noteRepository       NoteRepository
}

func NewAttachmentUsecase(attachmentRepository AttachmentRepository, noteRepository NoteRepository) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepository: attachmentRepository,
		noteRepository:       noteRepository,
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

	attachment, err := u.attachmentRepository.GetAttachment(ctx, blockID)
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
	userID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
	hasPosition bool,
	position int,
) (*models.Attachment, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	blockTypeID, err := u.getBlockTypeByMimeType(mimeType)
	if err != nil {
		return nil, err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	var blockPosition int
	if hasPosition {
		if position < 0 || position > len(blocks) {
			return nil, notes.ErrInvalidPosition
		}
		blockPosition = position
	} else {
		blockPosition = len(blocks)
	}

	block := models.Block{
		NoteID:      noteID,
		BlockTypeID: blockTypeID,
		Position:    blockPosition,
		Content:     "",
	}

	err = u.noteRepository.ShiftBlockPositions(ctx, noteID, blockPosition, 1)
	if err != nil {
		return nil, err
	}

	createdBlock, err := u.noteRepository.CreateBlock(ctx, block)
	if err != nil {
		_ = u.noteRepository.ShiftBlockPositions(ctx, noteID, blockPosition, -1)
		return nil, err
	}

	attachment, err := u.attachmentRepository.UploadAttachment(ctx, createdBlock.ID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		_, _ = u.noteRepository.DeleteBlock(ctx, createdBlock.ID)
		_ = u.noteRepository.ShiftBlockPositions(ctx, noteID, blockPosition, -1)
		return nil, err
	}

	return attachment, nil
}

func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	block, err := u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return err
	}

	if err := u.attachmentRepository.DeleteAttachment(ctx, blockID); err != nil {
		return err
	}

	_, err = u.noteRepository.DeleteBlock(ctx, blockID)
	if err != nil {
		return err
	}

	return u.noteRepository.ShiftBlockPositions(ctx, noteID, block.Position, -1)
}

func (u *attachmentUsecase) checkNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.noteRepository.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, attachments.ErrNoteNotFound
	}

	if !note.IsPublic && note.UserID != userID {
		return nil, attachments.ErrForbidden
	}

	return note, nil
}

func (u *attachmentUsecase) checkBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, error) {
	block, err := u.noteRepository.GetBlock(ctx, blockID)
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

func (u *attachmentUsecase) getBlockTypeByMimeType(mimeType string) (int, error) {
	if attachments.AllowedMimeTypesForImage[mimeType] {
		return 2, nil
	}
	if attachments.AllowedMimeTypesForGIF[mimeType] {
		return 2, nil
	}
	if attachments.AllowedMimeTypesForAudio[mimeType] {
		return 6, nil
	}
	if attachments.AllowedMimeTypesForVideo[mimeType] {
		return 7, nil
	}
	return 0, attachments.ErrInvalidMimeType
}
