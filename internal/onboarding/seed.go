package onboarding

import (
	"context"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

//go:generate mockgen -source=seed.go -destination=mocks/mock_onboarding.go -package=mocks

// Repository is the subset of the notes repository required to seed an
// onboarding note. It intentionally only exposes the create operations so the
// seeder cannot mutate or read existing user data.
type Repository interface {
	CreateNote(ctx context.Context, note models.Note) (*models.Note, error)
	CreateBlock(ctx context.Context, block models.Block) (*models.Block, error)
}

type Seeder struct {
	repo Repository
}

func NewSeeder(repo Repository) *Seeder {
	return &Seeder{repo: repo}
}

// Block type IDs as seeded in migrations/001_init.sql. Kept local to the
// package to avoid coupling onboarding to other layers.
const (
	blockTypeText  = 1
	blockTypeCode  = 3
	blockTypeQuote = 4
)

const onboardingNoteTitle = "Добро пожаловать в Noterian! 👋"
const onboardingNoteIcon = ""

type seedBlock struct {
	typeID  int
	content string
}

var seedBlocks = []seedBlock{
	{
		typeID:  blockTypeText,
		content: "Привет! Это твоя первая заметка. Здесь можно собирать всё, что важно: идеи, планы, конспекты лекций или мысли о прошедшем дне.",
	},
	{
		typeID:  blockTypeText,
		content: "Текст можно форматировать: делать жирным, курсивом или подчёркивать. Выдели любой фрагмент текста, чтобы открыть меню форматирования.",
	},
	{
		typeID:  blockTypeText,
		content: "Чтобы добавить новый блок, нажми Enter в конце строки или воспользуйся кнопкой «+» слева от блока. Блоки бывают разные: текст, изображение, аудио и видео.",
	},
	{
		typeID:  blockTypeText,
		content: "Заметки можно вкладывать друг в друга, к ним можно добавлять обложку и иконку. Их можно отмечать избранными и общими или экспортировать в PDF.",
	},
	{
		typeID:  blockTypeText,
		content: "Удачи! Когда будешь готов начать с чистого листа, просто удали эту заметку. 🚀",
	},
}

// SeedOnboardingNote creates the initial welcome note and its demo blocks for
// the given user. Callers should treat failures as best-effort: the user has
// already been created upstream and onboarding content is non-critical.
func (s *Seeder) SeedOnboardingNote(ctx context.Context, userID uuid.UUID) error {
	note, err := s.repo.CreateNote(ctx, models.Note{
		UserID: userID,
		Title:  onboardingNoteTitle,
		Icon:   onboardingNoteIcon,
	})
	if err != nil {
		return err
	}

	for i, b := range seedBlocks {
		if _, err := s.repo.CreateBlock(ctx, models.Block{
			NoteID:      note.ID,
			BlockTypeID: b.typeID,
			Position:    i,
			Content:     b.content,
		}); err != nil {
			return err
		}
	}

	return nil
}
