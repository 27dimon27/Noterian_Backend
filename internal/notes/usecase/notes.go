package usecase

import (
	"context"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

type NoteRepository interface {
	GetNotesByUserID(ctx context.Context, userID uuid.UUID) ([]models.Note, error)
	GetNoteByID(ctx context.Context, noteID uuid.UUID) (*models.Note, error)
	GetBlocksByNoteID(ctx context.Context, noteID uuid.UUID) ([]models.Block, error)
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, error)
	DeleteNote(ctx context.Context, noteID uuid.UUID) error
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, error)
	GetBlockByID(ctx context.Context, blockID uuid.UUID) (*models.Block, error)
	UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string, updatedAt time.Time) (*models.Block, error)
	MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int, updatedAt time.Time) (*models.Block, error)
	DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error)
	ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int, updatedAt time.Time) error
}

type noteUsecase struct {
	noteRepo NoteRepository
}

func NewNoteUsecase(noteRepo NoteRepository) *noteUsecase {
	return &noteUsecase{
		noteRepo: noteRepo,
	}
}

func (u *noteUsecase) GetNotesByUserID(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	return u.noteRepo.GetNotesByUserID(ctx, userID)
}

func (u *noteUsecase) GetNoteByID(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) (*models.Note, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
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

func (u *noteUsecase) GetBlocksByNoteID(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	return u.noteRepo.GetBlocksByNoteID(ctx, noteID)
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

	existingNote, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if existingNote.UserID != userID {
		return nil, notes.ErrForbidden
	}

	note.ID = noteID
	note.UpdatedAt = time.Now()

	return u.noteRepo.UpdateNote(ctx, noteID, note)
}

func (u *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID, userID uuid.UUID) error {
	existingNote, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return err
	}

	if existingNote.UserID != userID {
		return notes.ErrForbidden
	}

	return u.noteRepo.DeleteNote(ctx, noteID)
}

func (u *noteUsecase) CreateBlock(ctx context.Context, noteID uuid.UUID, userID uuid.UUID, block models.Block) (*models.Block, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	if block.BlockTypeID <= 0 {
		return nil, notes.ErrInvalidBlockType
	}

	block.NoteID = noteID
	block.Content = ""
	block.CreatedAt = time.Now()
	block.UpdatedAt = time.Now()

	if block.Position <= 0 {
		blocks, err := u.noteRepo.GetBlocksByNoteID(ctx, noteID)
		if err != nil {
			return nil, err
		}

		block.Position = len(blocks)
	}

	return u.noteRepo.CreateBlock(ctx, block)
}

func (u *noteUsecase) GetBlockByID(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) (*models.Block, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	block, err := u.noteRepo.GetBlockByID(ctx, blockID)
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

func (u *noteUsecase) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, content string) (*models.Block, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	block, err := u.noteRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, notes.ErrBlockNotFound
	}

	if block.NoteID != noteID {
		return nil, notes.ErrForbidden
	}

	return u.noteRepo.UpdateBlockContent(ctx, blockID, content, time.Now())
}

func (u *noteUsecase) MoveBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID, newPosition int) (*models.Block, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if note == nil {
		return nil, notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return nil, notes.ErrForbidden
	}

	block, err := u.noteRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		return nil, err
	}

	if block == nil {
		return nil, notes.ErrBlockNotFound
	}

	if block.NoteID != noteID {
		return nil, notes.ErrForbidden
	}

	if block.Position == newPosition {
		return block, nil
	}

	blocks, err := u.noteRepo.GetBlocksByNoteID(ctx, noteID)
	if err != nil {
		return nil, err
	}

	if newPosition < 0 || newPosition >= len(blocks) {
		return nil, notes.ErrInvalidPosition
	}

	return u.noteRepo.MoveBlock(ctx, noteID, blockID, block.Position, newPosition, time.Now())
}

func (u *noteUsecase) DeleteBlock(ctx context.Context, blockID uuid.UUID, noteID uuid.UUID, userID uuid.UUID) error {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return err
	}

	if note == nil {
		return notes.ErrNoteNotFound
	}

	if note.UserID != userID {
		return notes.ErrForbidden
	}

	block, err := u.noteRepo.GetBlockByID(ctx, blockID)
	if err != nil {
		return err
	}

	if block == nil {
		return notes.ErrBlockNotFound
	}

	if block.NoteID != noteID {
		return notes.ErrForbidden
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
