package usecase

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/attachments/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	notesgen "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
)

//go:generate mockgen -source=attachments.go -destination=mocks/mock_usecase_attachments.go -package=mocks

type AttachmentRepository interface {
	GetAttachment(ctx context.Context, blockID uuid.UUID) (*models.Attachment, types.AppErrorInterface)
	UploadAttachment(ctx context.Context, blockID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Attachment, types.AppErrorInterface)
	DeleteAttachment(ctx context.Context, blockID uuid.UUID) types.AppErrorInterface
	GetHeader(ctx context.Context, noteID uuid.UUID) (*models.Header, types.AppErrorInterface)
	UploadHeader(ctx context.Context, noteID uuid.UUID, fileName string, fileSize int64, mimeType string, fileReader io.Reader) (*models.Header, types.AppErrorInterface)
	DeleteHeader(ctx context.Context, noteID uuid.UUID) types.AppErrorInterface
}

type attachmentUsecase struct {
	attachmentRepo AttachmentRepository
	notesClient    grpcclient.NotesServiceClient
	logger         *slog.Logger
}

func NewAttachmentUsecase(attachmentRepo AttachmentRepository, notesClient grpcclient.NotesServiceClient, logger *slog.Logger) *attachmentUsecase {
	return &attachmentUsecase{
		attachmentRepo: attachmentRepo,
		notesClient:    notesClient,
		logger:         logger,
	}
}

func (u *attachmentUsecase) GetAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) (*models.Attachment, types.AppErrorInterface) {
	attachment, customErr := u.attachmentRepo.GetAttachment(ctx, blockID)
	if customErr != nil {
		return nil, customErr
	}

	if attachment == nil {
		return nil, &types.AppError{
			Err:        attachments.ErrAttachmentNotFound,
			PublicMsg:  attachments.PublicMsgErrAttachmentNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "GetAttachment",
		}
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
) (*models.Attachment, types.AppErrorInterface) {
	blockTypeID, customErr := u.getBlockTypeByMimeType(mimeType)
	if customErr != nil {
		return nil, customErr
	}

	blocks, err := u.notesClient.GetBlocks(ctx, noteID, userID)
	if err != nil {
		return nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "UploadAttachment",
		}
	}

	var blockPosition int
	if hasPosition {
		if position < 0 || position > len(blocks) {
			return nil, &types.AppError{
				Err:        attachments.ErrInvalidPosition,
				PublicMsg:  attachments.PublicMsgErrInvalidPosition,
				StatusCode: 400,
				Layer:      "usecase",
				Op:         "UploadAttachment",
			}
		}
		blockPosition = position
	} else {
		blockPosition = len(blocks)
	}

	if err := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, 1); err != nil {
		return nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "UploadAttachment",
		}
	}

	createdBlock, grpcErr := u.notesClient.CreateBlock(ctx, userID, &notesgen.BlockResponse{
		NoteId:      noteID.String(),
		BlockTypeId: int32(blockTypeID),
		Position:    int32(blockPosition),
		Content:     "",
	})
	if grpcErr != nil {
		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			return nil, &types.AppError{
				Err:        fmt.Errorf("grpc error: %w; %w", grpcErr, shiftErr),
				PublicMsg:  attachments.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "usecase",
				Op:         "UploadAttachment",
			}
		}
		return nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", grpcErr),
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "UploadAttachment",
		}
	}

	blockID, parseErr := uuid.Parse(createdBlock.Id)
	if parseErr != nil {
		if _, deleteErr := u.notesClient.DeleteBlock(ctx, blockID, noteID, userID); deleteErr != nil {
			parseErr = errors.Join(parseErr, fmt.Errorf("grpc error: %w", deleteErr))
		}

		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			parseErr = errors.Join(parseErr, fmt.Errorf("grpc error: %w", shiftErr))
		}

		return nil, &types.AppError{
			Err:        parseErr,
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "UploadAttachment",
		}
	}

	attachment, customErr := u.attachmentRepo.UploadAttachment(ctx, blockID, fileName, fileSize, mimeType, fileReader)
	if customErr != nil {
		err := customErr.Unwrap()
		errPublicMsg := customErr.PublicMessage()
		errStatus := customErr.Code()

		if _, deleteErr := u.notesClient.DeleteBlock(ctx, blockID, noteID, userID); deleteErr != nil {
			err = errors.Join(err, fmt.Errorf("grpc error: %w", deleteErr))
			errPublicMsg = attachments.PublicMsgErrInternalServer
			errStatus = 500
		}

		if shiftErr := u.notesClient.ShiftBlockPositions(ctx, noteID, blockPosition, -1); shiftErr != nil {
			err = errors.Join(err, fmt.Errorf("grpc error: %w", shiftErr))
			errPublicMsg = attachments.PublicMsgErrInternalServer
			errStatus = 500
		}

		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  errPublicMsg,
			StatusCode: errStatus,
			Layer:      "usecase",
			Op:         "UploadAttachment",
		}
	}

	return attachment, nil
}

func (u *attachmentUsecase) DeleteAttachment(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, userID uuid.UUID) types.AppErrorInterface {
	block, err := u.notesClient.GetBlock(ctx, blockID, noteID, userID)
	if err != nil {
		return &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  attachments.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "DeleteAttachment",
		}
	}

	if block == nil {
		return &types.AppError{
			Err:        attachments.ErrBlockNotFound,
			PublicMsg:  attachments.PublicMsgErrBlockNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "DeleteAttachment",
		}
	}

	if customErr := u.attachmentRepo.DeleteAttachment(ctx, blockID); customErr != nil {
		return customErr
	}

	return nil
}

func (u *attachmentUsecase) GetHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Header, types.AppErrorInterface) {
	header, customErr := u.attachmentRepo.GetHeader(ctx, noteID)
	if customErr != nil {
		return nil, customErr
	}

	if header == nil {
		return nil, &types.AppError{
			Err:        attachments.ErrHeaderNotFound,
			PublicMsg:  attachments.PublicMsgErrHeaderNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "GetHeader",
		}
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
) (*models.Header, types.AppErrorInterface) {
	header, customErr := u.attachmentRepo.UploadHeader(ctx, noteID, fileName, fileSize, mimeType, fileReader)
	if customErr != nil {
		return nil, customErr
	}

	return header, nil
}

func (u *attachmentUsecase) DeleteHeader(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface {
	if customErr := u.attachmentRepo.DeleteHeader(ctx, noteID); customErr != nil {
		return customErr
	}

	return nil
}

func (u *attachmentUsecase) getBlockTypeByMimeType(mimeType string) (int, types.AppErrorInterface) {
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
	return 0, &types.AppError{
		Err:        attachments.ErrInvalidMimeType,
		PublicMsg:  attachments.PublicMsgErrInvalidMimeType,
		StatusCode: 400,
		Layer:      "usecase",
		Op:         "getBlockTypeByMimeType",
	}
}
