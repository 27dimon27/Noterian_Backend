package usecase

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/pdf"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
)

//go:generate mockgen -source=notes.go -destination=mocks/mock_usecase_notes.go -package=mocks

type NoteRepository interface {
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, types.AppErrorInterface)
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, types.AppErrorInterface)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, types.AppErrorInterface)
	GetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.BlockFormatting, types.AppErrorInterface)
	GetBlocksFormatting(ctx context.Context, blockIDs []uuid.UUID) (map[string]models.BlockFormatting, types.AppErrorInterface)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, types.AppErrorInterface)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, types.AppErrorInterface)
	DeleteNote(ctx context.Context, noteID uuid.UUID) types.AppErrorInterface
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, types.AppErrorInterface)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, types.AppErrorInterface)
	GetBlockType(ctx context.Context, blockTypeID int) (*models.BlockType, types.AppErrorInterface)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string) (*models.Block, types.AppErrorInterface)
	MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int) (*models.Block, types.AppErrorInterface)
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, types.AppErrorInterface)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) types.AppErrorInterface
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, types.AppErrorInterface)
	GetSubnotes(ctx context.Context, noteID uuid.UUID) ([]models.Note, types.AppErrorInterface)
}

type noteUsecase struct {
	noteRepository    NoteRepository
	attachmentsClient grpcclient.AttachmentsServiceClient
	logger            *slog.Logger
}

func NewNoteUsecase(noteRepository NoteRepository, attachmentsClient grpcclient.AttachmentsServiceClient, logger *slog.Logger) *noteUsecase {
	return &noteUsecase{
		noteRepository:    noteRepository,
		attachmentsClient: attachmentsClient,
		logger:            logger,
	}
}

func (u *noteUsecase) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, types.AppErrorInterface) {
	notes, err := u.noteRepository.GetNotes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (u *noteUsecase) GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, types.AppErrorInterface) {
	note, customErr := u.checkNoteAccess(ctx, noteID, userID)
	if customErr != nil {
		return nil, nil, nil, customErr
	}

	blocks, customErr := u.noteRepository.GetBlocks(ctx, note.ID)
	if customErr != nil {
		return nil, nil, nil, customErr
	}

	blockIDs := make([]uuid.UUID, len(blocks))
	for i, block := range blocks {
		blockIDs[i] = block.ID

		if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
			attachment, err := u.attachmentsClient.GetAttachment(ctx, block.ID, noteID, userID)
			if err != nil {
				continue
			}
			blocks[i].Content = attachment.AttachUrl
		}
	}

	header, err := u.attachmentsClient.GetHeader(ctx, noteID, userID)
	if err != nil {
		return nil, nil, nil, &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "GetNote",
		}
	}

	if header != nil {
		note.HeaderURL = header.HeaderUrl
	}

	formattings, customErr := u.noteRepository.GetBlocksFormatting(ctx, blockIDs)
	if customErr != nil {
		return nil, nil, nil, customErr
	}

	return note, blocks, formattings, nil
}

func (u *noteUsecase) GetPublicNote(ctx context.Context, noteID uuid.UUID) (*models.Note, types.AppErrorInterface) {
	note, err := u.noteRepository.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil || !note.IsPublic {
		return nil, &types.AppError{
			Err:        notes.ErrNoteNotFound,
			PublicMsg:  notes.PublicMsgErrNoteNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "GetPublicNote",
		}
	}

	return note, nil
}

func (u *noteUsecase) CreateNote(ctx context.Context, note models.Note) (*models.Note, types.AppErrorInterface) {
	if note.Title == "" {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidNoteData,
			PublicMsg:  notes.PublicMsgErrInvalidNoteData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "CreateNote",
		}
	}

	createdNote, err := u.noteRepository.CreateNote(ctx, note)
	if err != nil {
		return nil, err
	}

	return createdNote, nil
}

func (u *noteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	if note.Title == "" {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidNoteData,
			PublicMsg:  notes.PublicMsgErrInvalidNoteData,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "UpdateNote",
		}
	}

	updatedNote, err := u.noteRepository.UpdateNote(ctx, noteID, note)
	if err != nil {
		return nil, err
	}

	return updatedNote, nil
}

func (u *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface {
	_, customErr := u.checkNoteAccess(ctx, noteID, userID)
	if customErr != nil {
		return customErr
	}

	blocks, customErr := u.noteRepository.GetBlocks(ctx, noteID)
	if customErr != nil {
		return customErr
	}

	for _, block := range blocks {
		if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
			if err := u.attachmentsClient.DeleteAttachment(ctx, block.ID, noteID, userID); err != nil {
				continue
			}
		}
	}

	if err := u.attachmentsClient.DeleteHeader(ctx, noteID, userID); err != nil {
		return &types.AppError{
			Err:        fmt.Errorf("grpc error: %w", err),
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "DeleteNote",
		}
	}

	if customErr := u.noteRepository.DeleteNote(ctx, noteID); customErr != nil {
		return customErr
	}

	return nil
}

func (u *noteUsecase) CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	if block.BlockTypeID <= 0 {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidBlockType,
			PublicMsg:  notes.PublicMsgErrInvalidBlockType,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "CreateBlock",
		}
	}

	block.NoteID = noteID
	block.Content = ""

	blocks, err := u.noteRepository.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if block.Position < 0 || block.Position > len(blocks) {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidPosition,
			PublicMsg:  notes.PublicMsgErrInvalidPosition,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "CreateBlock",
		}
	} else {
		err := u.noteRepository.ShiftBlockPositions(ctx, noteID, block.Position, 1)
		if err != nil {
			return nil, err
		}
	}

	createdBlock, err := u.noteRepository.CreateBlock(ctx, block)
	if err != nil {
		return nil, err
	}

	return createdBlock, nil
}

func (u *noteUsecase) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	updatedBlock, err := u.noteRepository.UpdateBlockContent(ctx, blockID, content)
	if err != nil {
		return nil, err
	}

	return updatedBlock, nil
}

func (u *noteUsecase) MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	block, err := u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	if block.Position == newPosition {
		return block, nil
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if newPosition < 0 || newPosition > len(blocks) {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidPosition,
			PublicMsg:  notes.PublicMsgErrInvalidPosition,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "MoveBlock",
		}
	}

	updatedBlock, err := u.noteRepository.MoveBlock(ctx, noteID, blockID, block.Position, newPosition)
	if err != nil {
		return nil, err
	}

	return updatedBlock, nil
}

func (u *noteUsecase) DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface {
	_, customErr := u.checkNoteAccess(ctx, noteID, userID)
	if customErr != nil {
		return customErr
	}

	block, customErr := u.checkBlockAccess(ctx, noteID, blockID)
	if customErr != nil {
		return customErr
	}

	if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
		err := u.attachmentsClient.DeleteAttachment(ctx, blockID, noteID, userID)
		if err != nil {
			return &types.AppError{
				Err:        fmt.Errorf("grpc error: %w", err),
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "usecase",
				Op:         "DeleteBlock",
			}
		}
	}

	blockNoteID, customErr := u.noteRepository.DeleteBlock(ctx, blockID)
	if customErr != nil {
		return customErr
	}

	if blockNoteID == nil {
		return &types.AppError{
			Err:        notes.ErrBlockNotFound,
			PublicMsg:  notes.PublicMsgErrBlockNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "DeleteBlock",
		}
	}

	customErr = u.noteRepository.ShiftBlockPositions(ctx, noteID, block.Position, -1)
	if customErr != nil {
		return customErr
	}

	return nil
}

func (u *noteUsecase) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	block, err := u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	blockType, err := u.noteRepository.GetBlockType(ctx, block.BlockTypeID)
	if err != nil {
		return nil, err
	}

	if blockType == nil {
		return nil, &types.AppError{
			Err:        notes.ErrBlockTypeNotFound,
			PublicMsg:  notes.PublicMsgErrBlockTypeNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "UpdateBlockFormatting",
		}
	}

	if blockType.Name == "image" {
		if formattingRange.Bold != nil || formattingRange.Italic != nil || formattingRange.Underline != nil {
			return nil, &types.AppError{
				Err:        notes.ErrInvalidFormattingForImageBlock,
				PublicMsg:  notes.PublicMsgErrInvalidFormattingForImageBlock,
				StatusCode: 400,
				Layer:      "usecase",
				Op:         "UpdateBlockFormatting",
			}
		}
	} else if blockType.Name != "text" {
		return nil, &types.AppError{
			Err:        notes.ErrFormattingNotSupported,
			PublicMsg:  notes.PublicMsgErrFormattingNotSupported,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "UpdateBlockFormatting",
		}
	}

	if formattingRange.StartPos < 0 || formattingRange.EndPos > len(block.Content) || formattingRange.StartPos >= formattingRange.EndPos {
		return nil, &types.AppError{
			Err:        notes.ErrInvalidFormattingRange,
			PublicMsg:  notes.PublicMsgErrInvalidFormattingRange,
			StatusCode: 400,
			Layer:      "usecase",
			Op:         "UpdateBlockFormatting",
		}
	}

	blockFormatting, err := u.noteRepository.UpdateBlockFormatting(ctx, blockID, formattingRange)
	if err != nil {
		return nil, err
	}

	return blockFormatting, nil
}

func (u *noteUsecase) GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	subnotes, err := u.noteRepository.GetSubnotes(ctx, noteID)
	if err != nil {
		return nil, err
	}

	return subnotes, nil
}

func (u *noteUsecase) CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note, hasPosition bool, position int) (*models.Note, uuid.UUID, types.AppErrorInterface) {
	_, err := u.checkNoteAccess(ctx, parentNoteID, userID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, parentNoteID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	blockPosition := len(blocks)

	if hasPosition {
		if position < 0 || position > len(blocks) {
			return nil, uuid.Nil, &types.AppError{
				Err:        notes.ErrInvalidPosition,
				PublicMsg:  notes.PublicMsgErrInvalidPosition,
				StatusCode: 400,
				Layer:      "usecase",
				Op:         "CreateSubnote",
			}
		}

		blockPosition = position
	}

	block := models.Block{
		NoteID:      parentNoteID,
		BlockTypeID: 5, // пока константа подзаметки, потом вынести в перменные
		Position:    blockPosition,
		Content:     "",
	}

	if err := u.noteRepository.ShiftBlockPositions(ctx, parentNoteID, blockPosition, 1); err != nil {
		return nil, uuid.Nil, err
	}

	createdBlock, err := u.noteRepository.CreateBlock(ctx, block)
	if err != nil {
		if shiftErr := u.noteRepository.ShiftBlockPositions(ctx, parentNoteID, blockPosition, -1); shiftErr != nil {
			return nil, uuid.Nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", err.Unwrap(), shiftErr.Unwrap()),
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "usecase",
				Op:         "CreateSubnote",
			}
		}
		return nil, uuid.Nil, err
	}

	createdNote, err := u.noteRepository.CreateNote(ctx, note)
	if err != nil {
		if _, delErr := u.noteRepository.DeleteBlock(ctx, createdBlock.ID); delErr != nil {
			return nil, uuid.Nil, &types.AppError{
				Err:        fmt.Errorf("%w; %w", err.Unwrap(), delErr.Unwrap()),
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "usecase",
				Op:         "CreateSubnote",
			}
		}
		return nil, uuid.Nil, err
	}

	return createdNote, createdBlock.ID, nil
}

func (u *noteUsecase) DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) types.AppErrorInterface {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, noteID)
	if err != nil {
		return err
	}

	for _, block := range blocks {
		if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
			if err := u.attachmentsClient.DeleteAttachment(ctx, block.ID, noteID, userID); err != nil {
				continue
			}
		}
	}

	if err := u.noteRepository.DeleteNote(ctx, subnoteID); err != nil {
		return err
	}

	return nil
}

func (u *noteUsecase) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.Block, types.AppErrorInterface) {
	block, err := u.noteRepository.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	return block, nil
}

func (u *noteUsecase) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) types.AppErrorInterface {
	if err := u.noteRepository.ShiftBlockPositions(ctx, noteID, fromPosition, 1); err != nil {
		return err
	}

	return nil
}

func (u *noteUsecase) GenerateNotePDF(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*bytes.Buffer, types.AppErrorInterface) {
	note, customErr := u.checkNoteAccess(ctx, noteID, userID)
	if customErr != nil {
		return nil, customErr
	}

	blocks, customErr := u.noteRepository.GetBlocks(ctx, note.ID)
	if customErr != nil {
		return nil, customErr
	}

	blockIDs := make([]uuid.UUID, len(blocks))
	for i, block := range blocks {
		blockIDs[i] = block.ID

		if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
			attachment, err := u.attachmentsClient.GetAttachment(ctx, block.ID, noteID, userID)
			if err != nil {
				continue
			}

			blocks[i].Content = attachment.AttachUrl
		}
	}

	header, err := u.attachmentsClient.GetHeader(ctx, noteID, userID)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "GenerateNotePDF",
		}
	}

	headerURL := ""
	if header != nil {
		headerURL = header.HeaderUrl
	}

	formattings, customErr := u.noteRepository.GetBlocksFormatting(ctx, blockIDs)
	if customErr != nil {
		return nil, customErr
	}

	subnotes, customErr := u.noteRepository.GetSubnotes(ctx, noteID)
	if customErr != nil {
		return nil, customErr
	}

	subnotesMap := make(map[string]models.Note)

	for _, block := range blocks {
		if block.BlockTypeID == 5 {
			for _, subnote := range subnotes {
				if subnote.ID.String() == block.Content {
					subnotesMap[block.ID.String()] = subnote
					break
				}
			}
		}
	}

	noteContent := &pdf.NoteContent{
		Note:       note,
		Blocks:     blocks,
		Formatting: formattings,
		Subnotes:   subnotesMap,
		HeaderURL:  headerURL,
	}

	pdfBuffer, err := pdf.GeneratePDF(noteContent)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "usecase",
			Op:         "GenerateNotePDF",
		}
	}

	return pdfBuffer, nil
}

func (u *noteUsecase) checkNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, types.AppErrorInterface) {
	note, err := u.noteRepository.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, &types.AppError{
			Err:        notes.ErrNoteNotFound,
			PublicMsg:  notes.PublicMsgErrNoteNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "checkNoteAccess",
		}
	}

	if !note.IsPublic && note.UserID != userID {
		return nil, &types.AppError{
			Err:        notes.ErrForbidden,
			PublicMsg:  notes.PublicMsgErrForbidden,
			StatusCode: 403,
			Layer:      "usecase",
			Op:         "checkNoteAccess",
		}
	}

	return note, nil
}

func (u *noteUsecase) checkBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, types.AppErrorInterface) {
	block, err := u.noteRepository.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, &types.AppError{
			Err:        notes.ErrBlockNotFound,
			PublicMsg:  notes.PublicMsgErrBlockNotFound,
			StatusCode: 404,
			Layer:      "usecase",
			Op:         "checkBlockAccess",
		}
	}

	if block.NoteID != noteID {
		return nil, &types.AppError{
			Err:        notes.ErrForbidden,
			PublicMsg:  notes.PublicMsgErrForbidden,
			StatusCode: 403,
			Layer:      "usecase",
			Op:         "checkBlockAccess",
		}
	}

	return block, nil
}
