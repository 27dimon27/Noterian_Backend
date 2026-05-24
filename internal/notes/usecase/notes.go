package usecase

import (
	"bytes"
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/grpcclient"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/pdf"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

//go:generate mockgen -source=notes.go -destination=mocks/mock_usecase_notes.go -package=mocks

type NoteRepository interface {
	GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	GetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.BlockFormatting, error)
	GetBlocksFormatting(ctx context.Context, blockIDs []uuid.UUID) (map[string]models.BlockFormatting, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, error)
	GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
	GetBlockType(ctx context.Context, blockTypeID int) (*models.BlockType, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string) (*models.Block, error)
	MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	GetSubnotes(ctx context.Context, noteID uuid.UUID) ([]models.Note, error)
}

type noteUsecase struct {
	noteRepository    NoteRepository
	attachmentsClient grpcclient.AttachmentsServiceClient
}

func NewNoteUsecase(noteRepository NoteRepository, attachmentsClient grpcclient.AttachmentsServiceClient) *noteUsecase {
	return &noteUsecase{
		noteRepository:    noteRepository,
		attachmentsClient: attachmentsClient,
	}
}

func (u *noteUsecase) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	notes, err := u.noteRepository.GetNotes(ctx, userID)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

func (u *noteUsecase) GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, []models.Block, map[string]models.BlockFormatting, error) {
	note, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, nil, nil, err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, note.ID)
	if err != nil {
		return nil, nil, nil, err
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
		if status.Code(err) != codes.NotFound {
			return nil, nil, nil, err
		}
	}

	if header != nil {
		note.HeaderURL = header.HeaderUrl
	}

	formattings, err := u.noteRepository.GetBlocksFormatting(ctx, blockIDs)
	if err != nil {
		return nil, nil, nil, err
	}

	return note, blocks, formattings, nil
}

func (u *noteUsecase) CreateNote(ctx context.Context, note models.Note) (*models.Note, error) {
	if note.Title == "" {
		return nil, notes.ErrInvalidNoteData
	}

	return u.noteRepository.CreateNote(ctx, note)
}

func (u *noteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, note models.Note) (*models.Note, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	if note.Title == "" {
		return nil, notes.ErrInvalidNoteData
	}

	return u.noteRepository.UpdateNote(ctx, noteID, note)
}

func (u *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
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
			err = u.attachmentsClient.DeleteAttachment(ctx, block.ID, noteID, userID)
			if err != nil {
				continue
			}
		}
	}

	err = u.attachmentsClient.DeleteHeader(ctx, noteID, userID)
	if err != nil && status.Code(err) != codes.NotFound {
		return err
	}

	return u.noteRepository.DeleteNote(ctx, noteID)
}

func (u *noteUsecase) CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	if block.BlockTypeID <= 0 {
		return nil, notes.ErrInvalidBlockType
	}

	block.NoteID = noteID
	block.Content = ""

	blocks, err := u.noteRepository.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if block.Position < 0 || block.Position > len(blocks) {
		return nil, notes.ErrInvalidPosition
	} else {
		err = u.noteRepository.ShiftBlockPositions(ctx, noteID, block.Position, 1)
		if err != nil {
			return nil, err
		}
	}

	return u.noteRepository.CreateBlock(ctx, block)
}

func (u *noteUsecase) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	return u.noteRepository.UpdateBlockContent(ctx, blockID, content)
}

func (u *noteUsecase) MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error) {
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
		return nil, notes.ErrInvalidPosition
	}

	return u.noteRepository.MoveBlock(ctx, noteID, blockID, block.Position, newPosition)
}

func (u *noteUsecase) DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	block, err := u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return err
	}

	if block.BlockTypeID != 1 && block.BlockTypeID != 5 {
		err = u.attachmentsClient.DeleteAttachment(ctx, blockID, noteID, userID)
		if err != nil {
			return err
		}
	}

	blockNoteID, err := u.noteRepository.DeleteBlock(ctx, blockID)
	if err != nil {
		return err
	}

	if blockNoteID == nil {
		return notes.ErrBlockNotFound
	}

	return u.noteRepository.ShiftBlockPositions(ctx, noteID, block.Position, -1)
}

func (u *noteUsecase) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error) {
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
		return nil, notes.ErrBlockTypeNotFound
	}

	if blockType.Name == "image" {
		if formattingRange.Bold != nil || formattingRange.Italic != nil || formattingRange.Underline != nil {
			return nil, notes.ErrInvalidFormattingForImageBlock
		}
	} else if blockType.Name != "text" {
		return nil, notes.ErrFormattingNotSupported
	}

	if formattingRange.StartPos < 0 || formattingRange.EndPos > len(block.Content) || formattingRange.StartPos >= formattingRange.EndPos {
		return nil, notes.ErrInvalidFormattingRange
	}

	return u.noteRepository.UpdateBlockFormatting(ctx, blockID, formattingRange)
}

func (u *noteUsecase) GetSubnotes(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) ([]models.Note, error) {
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

func (u *noteUsecase) CreateSubnote(ctx context.Context, parentNoteID uuid.UUID, userID uuid.UUID, note models.Note, hasPosition bool, position int) (*models.Note, uuid.UUID, error) {
	_, err := u.checkNoteAccess(ctx, parentNoteID, userID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, parentNoteID)
	if err != nil {
		return nil, uuid.Nil, err
	}

	var blockPosition int
	if hasPosition {
		if position < 0 || position > len(blocks) {
			return nil, uuid.Nil, notes.ErrInvalidPosition
		}
		blockPosition = position
	} else {
		blockPosition = len(blocks)
	}

	block := models.Block{
		NoteID:      parentNoteID,
		BlockTypeID: 5, // пока константа подзаметки, потом вынести в перменные
		Position:    blockPosition,
		Content:     "",
	}

	err = u.noteRepository.ShiftBlockPositions(ctx, parentNoteID, blockPosition, 1)
	if err != nil {
		return nil, uuid.Nil, err
	}

	createdBlock, err := u.noteRepository.CreateBlock(ctx, block)
	if err != nil {
		_ = u.noteRepository.ShiftBlockPositions(ctx, parentNoteID, blockPosition, -1)
		return nil, uuid.Nil, err
	}

	createdNote, err := u.noteRepository.CreateNote(ctx, note)
	if err != nil {
		return nil, uuid.Nil, err
	}

	return createdNote, createdBlock.ID, nil
}

func (u *noteUsecase) DeleteSubnote(ctx context.Context, noteID uuid.UUID, subnoteID uuid.UUID, userID uuid.UUID) error {
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
			err = u.attachmentsClient.DeleteAttachment(ctx, block.ID, noteID, userID)
			if err != nil {
				continue
			}
		}
	}

	err = u.noteRepository.DeleteNote(ctx, subnoteID)
	if err != nil {
		return err
	}

	return nil
}

func (u *noteUsecase) GetBlock(ctx context.Context, blockID, noteID, userID uuid.UUID) (*models.Block, error) {
	block, err := u.noteRepository.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}
	return block, nil
}

func (u *noteUsecase) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition, direction int) error {
	err := u.noteRepository.ShiftBlockPositions(ctx, noteID, fromPosition, 1)
	if err != nil {
		return err
	}
	return nil
}

func (u *noteUsecase) GenerateNotePDF(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*bytes.Buffer, error) {
	note, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	blocks, err := u.noteRepository.GetBlocks(ctx, note.ID)
	if err != nil {
		return nil, err
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
		if status.Code(err) != codes.NotFound {
			return nil, err
		}
	}

	headerURL := ""
	if header != nil {
		headerURL = header.HeaderUrl
	}

	formattings, err := u.noteRepository.GetBlocksFormatting(ctx, blockIDs)
	if err != nil {
		return nil, err
	}

	subnotes, err := u.noteRepository.GetSubnotes(ctx, noteID)
	if err != nil {
		return nil, err
	}

	noteContent := &pdf.NoteContent{
		Note:       note,
		Blocks:     blocks,
		Formatting: formattings,
		Subnotes:   subnotes,
		HeaderURL:  headerURL,
	}

	pdfBuffer, err := pdf.GeneratePDF(noteContent)
	if err != nil {
		return nil, err
	}

	return pdfBuffer, nil
}

func (u *noteUsecase) checkNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.noteRepository.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if !note.IsPublic && note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	return note, nil
}

func (u *noteUsecase) checkBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, error) {
	block, err := u.noteRepository.GetBlock(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, notes.ErrBlockNotFound
	}

	if block.NoteID != noteID {
		return nil, notes.ErrForbidden
	}

	return block, nil
}
