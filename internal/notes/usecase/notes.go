package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

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
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string, updatedAt time.Time) (*models.Block, error)
	MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int, updatedAt time.Time) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int, updatedAt time.Time) error
	UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error)
	ResetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.BlockFormatting, error)
}

type noteUsecase struct {
	noteRepo NoteRepository
}

func NewNoteUsecase(noteRepo NoteRepository) *noteUsecase {
	return &noteUsecase{
		noteRepo: noteRepo,
	}
}

func (u *noteUsecase) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	return u.noteRepo.GetNotes(ctx, userID)
}

func (u *noteUsecase) GetNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	return note, nil
}

func (u *noteUsecase) GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	return u.noteRepo.GetBlocks(ctx, noteID)
}

func (u *noteUsecase) CreateNote(ctx context.Context, note models.Note) (*models.Note, error) {
	if note.Title == "" {
		return nil, notes.ErrInvalidNoteData
	}

	return u.noteRepo.CreateNote(ctx, note)
}

func (u *noteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note, userID uuid.UUID) (*models.Note, error) {
	if note.Title == "" {
		return nil, notes.ErrInvalidNoteData
	}

	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	note.ID = noteID
	note.UpdatedAt = time.Now()

	return u.noteRepo.UpdateNote(ctx, noteID, note)
}

func (u *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return err
	}

	return u.noteRepo.DeleteNote(ctx, noteID)
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
	block.CreatedAt = time.Now()
	block.UpdatedAt = time.Now()

	blocks, err := u.noteRepo.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if block.Position < 0 || block.Position > len(blocks) {
		return nil, notes.ErrInvalidPosition
	} else {
		err = u.noteRepo.ShiftBlockPositions(ctx, noteID, block.Position, 1, time.Now())
		if err != nil {
			return nil, err
		}
	}

	return u.noteRepo.CreateBlock(ctx, block)
}

func (u *noteUsecase) GetBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	block, err := u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	return block, nil
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

	return u.noteRepo.UpdateBlockContent(ctx, blockID, content, time.Now())
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

	blocks, err := u.noteRepo.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if newPosition < 0 || newPosition >= len(blocks) {
		return nil, notes.ErrInvalidPosition
	}

	return u.noteRepo.MoveBlock(ctx, noteID, blockID, block.Position, newPosition, time.Now())
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

	blockNoteID, err := u.noteRepo.DeleteBlock(ctx, blockID)
	if err != nil {
		return err
	}

	if blockNoteID == nil {
		return notes.ErrBlockNotFound
	}

	return u.noteRepo.ShiftBlockPositions(ctx, noteID, block.Position, -1, time.Now())
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

	blockType, err := u.noteRepo.GetBlockType(ctx, block.BlockTypeID)
	if err != nil {
		return nil, err
	}

	if blockType == nil {
		return nil, notes.ErrInvalidBlockType
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

	return u.noteRepo.UpdateBlockFormatting(ctx, blockID, formattingRange)
}

func (u *noteUsecase) ResetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	return u.noteRepo.ResetBlockFormatting(ctx, blockID)
}

func (u *noteUsecase) GetBlocksWithFormatting(ctx context.Context, noteID uuid.UUID) ([]models.Block, map[string]models.BlockFormatting, error) {
	blocks, err := u.noteRepo.GetBlocks(ctx, noteID)
	if err != nil {
		return nil, nil, err
	}

	if len(blocks) == 0 {
		return blocks, make(map[string]models.BlockFormatting), nil
	}

	blockIDs := make([]uuid.UUID, len(blocks))
	for i, block := range blocks {
		blockIDs[i] = block.ID
	}

	formattings, err := u.noteRepo.GetBlocksFormatting(ctx, blockIDs)
	if err != nil {
		return nil, nil, err
	}

	return blocks, formattings, nil
}

func (u *noteUsecase) GetBlockFormatting(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.BlockFormatting, error) {
	_, err := u.checkNoteAccess(ctx, noteID, userID)
	if err != nil {
		return nil, err
	}

	_, err = u.checkBlockAccess(ctx, noteID, blockID)
	if err != nil {
		return nil, err
	}

	return u.noteRepo.GetBlockFormatting(ctx, blockID)
}

func (u *noteUsecase) checkNoteAccess(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.noteRepo.GetNote(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	return note, nil
}

func (u *noteUsecase) checkBlockAccess(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID) (*models.Block, error) {
	block, err := u.noteRepo.GetBlock(ctx, blockID)
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
