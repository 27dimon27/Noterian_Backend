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

func (u *noteUsecase) GetNoteByID(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	note, err := u.noteRepo.GetNoteByID(ctx, noteID)
	if err != nil {
		return nil, err
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

func (u *noteUsecase) UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, error) {
	if note.Title == "" {
		return nil, notes.ErrInvalidNoteData
	}

	note.ID = noteID
	note.UpdatedAt = time.Now()

	return u.noteRepo.UpdateNote(ctx, noteID, note)
}

func (u *noteUsecase) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	return u.noteRepo.DeleteNote(ctx, noteID)
}
