package usecase

import (
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/repository"
	"github.com/google/uuid"
)

type NoteUsecase interface {
	GetNotesByUserID(userID uuid.UUID) ([]models.Note, error)
	GetNoteByID(noteID uuid.UUID) (*models.Note, error)
	GetBlocksByNoteID(noteID uuid.UUID) ([]models.Block, error)
}

type noteUsecase struct {
	noteRepo repository.NoteRepository
}

func NewNoteUsecase(noteRepo repository.NoteRepository) NoteUsecase {
	return &noteUsecase{
		noteRepo: noteRepo,
	}
}

func (u *noteUsecase) GetNotesByUserID(userID uuid.UUID) ([]models.Note, error) {
	return u.noteRepo.GetNotesByUserID(userID)
}

func (u *noteUsecase) GetNoteByID(noteID uuid.UUID) (*models.Note, error) {
	return u.noteRepo.GetNoteByID(noteID)
}

func (u *noteUsecase) GetBlocksByNoteID(noteID uuid.UUID) ([]models.Block, error) {
	return u.noteRepo.GetBlocksByNoteID(noteID)
}
