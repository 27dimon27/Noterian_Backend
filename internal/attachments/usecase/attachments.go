package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_usecase_attachments.go -package=mocks

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, error)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, error)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) error
	GetHeader(ctx context.Context, noteID uuid.UUID) (*models.Header, error)
	UploadHeader(ctx context.Context, noteID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Header, error)
	DeleteHeader(ctx context.Context, noteID uuid.UUID) error
}

type attachmentUsecase struct {
	attachmentRepo AttachmentRepository
	notesClient    grpcclient.NotesServiceClient
}

func NewAttachmentUsecase(attachmentRepo AttachmentRepository, notesClient grpcclient.NotesServiceClient) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		notesClient:    notesClient,
	}
}

func (u *attachmentUsecase) GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, error) {
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
	userID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
	hasPosition bool,
	position int,
) (*models.Attachment, error) {
	blockTypeID, err := u.getBlockTypeByMimeType(mimeType)
	if err != nil {
		return nil, err
	}

	blocks, err := u.notesClient.GetBlocks(ctx, noteID, userID)
	if err != nil {
		return nil, u.mapGrpcError(err)
	}

	var blockPosition int
	if hasPosition {
		if position < 0 || position > len(blocks) {
			return nil, attachments.ErrInvalidPosition
		}
		blockPosition = position
	} else {
		blockPosition = len(blocks)
	}

	err = u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, 1)
	if err != nil {
		return nil, u.mapGrpcError(err)
	}

	createdBlock, err := u.notesClient.CreateBlock(ctx, userID, &notesgen.BlockResponse{
		NoteId:      noteID.String(),
		BlockTypeId: int32(blockTypeID),
		Position:    int32(blockPosition),
		Content:     "",
	})
	if err != nil {
		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			return nil, fmt.Errorf("create block failed: %w, and rollback failed: %w", err, shiftErr)
		}
		return nil, u.mapGrpcError(err)
	}

	blockID, err := uuid.Parse(createdBlock.Id)
	if err != nil {
		var errs []error
		errs = append(errs, fmt.Errorf("failed to parse block ID: %w", err))

		if _, deleteErr := u.notesClient.DeleteBlock(ctx, blockID, noteID, userID); deleteErr != nil {
			errs = append(errs, fmt.Errorf("failed to delete block during rollback: %w", deleteErr))
		}

		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			errs = append(errs, fmt.Errorf("failed to shift block positions during rollback: %w", shiftErr))
		}

		return nil, errors.Join(errs...)
	}

	attachment, err := u.attachmentRepo.UploadAttachment(ctx, blockID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		var errs []error
		errs = append(errs, fmt.Errorf("failed to parse block ID: %w", err))

		if _, deleteErr := u.notesClient.DeleteBlock(ctx, blockID, noteID, userID); deleteErr != nil {
			errs = append(errs, fmt.Errorf("failed to delete block during rollback: %w", deleteErr))
		}

		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			errs = append(errs, fmt.Errorf("failed to shift block positions during rollback: %w", shiftErr))
		}

		return nil, errors.Join(errs...)
	}

	return attachment, nil
}

func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) error {
	block, err := u.notesClient.GetBlock(ctx, blockID, noteID, userID)
	if err != nil {
		return u.mapGrpcError(err)
	}
	if block == nil {
		return attachments.ErrBlockNotFound
	}

	if err := u.attachmentRepo.DeleteAttachment(ctx, blockID); err != nil {
		return err
	}

	return nil
}

func (u *attachmentUsecase) GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, error) {
	header, err := u.attachmentRepo.GetHeader(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if header == nil {
		return nil, attachments.ErrHeaderNotFound
	}

	return header, nil
}

func (u *attachmentUsecase) UploadHeader(
	ctx context.Context,
	noteID uuid.UUID,
	userID uuid.UUID,
	fileName string,
	fileSize int64,
	mimeType string,
	fileReader io.Reader,
) (*models.Header, error) {
	header, err := u.attachmentRepo.UploadHeader(ctx, noteID, fileName, fileSize, mimeType, fileReader)
	if err != nil {
		return nil, err
	}

	return header, nil
}

func (u *attachmentUsecase) DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	err := u.attachmentRepo.DeleteHeader(ctx, noteID)
	if err != nil {
		return err
	}

	return nil
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

func (u *attachmentUsecase) mapGrpcError(err error) error {
	// можно добавить маппинг gRPC ошибок в доменные ошибки, например если пришел codes.NotFound - вернуть attachments.ErrNoteNotFound
	return err
}
