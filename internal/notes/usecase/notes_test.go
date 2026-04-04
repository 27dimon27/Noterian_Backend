package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase/mocks"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestNoteUsecase_GetNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedNotes := []models.Note{
			{ID: uuid.New(), UserID: userID, Title: "Note 1"},
			{ID: uuid.New(), UserID: userID, Title: "Note 2"},
		}

		mockRepo.EXPECT().
			GetNotes(gomock.Any(), userID).
			Return(expectedNotes, nil)

		notes, err := usecase.GetNotes(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(notes) != 2 {
			t.Errorf("expected 2 notes, got %d", len(notes))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNotes(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetNotes(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestNoteUsecase_GetNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedNote := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Test Note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(expectedNote, nil)

		note, err := usecase.GetNote(context.Background(), noteID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if note.ID != noteID {
			t.Errorf("expected ID %v, got %v", noteID, note.ID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.GetNote(context.Background(), noteID, userID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("forbidden, wrong user", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
			Title:  "Test Note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.GetNote(context.Background(), noteID, userID)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetNote(context.Background(), noteID, userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestNoteUsecase_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedBlocks := []models.Block{
			{ID: uuid.New(), NoteID: noteID, Position: 0},
			{ID: uuid.New(), NoteID: noteID, Position: 1},
		}

		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(expectedBlocks, nil)

		blocks, err := usecase.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(blocks))
		}
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		_, err := usecase.GetBlocks(context.Background(), noteID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := models.Note{
			UserID: userID,
			Title:  "Valid Title",
		}
		createdNote := &models.Note{
			ID:     uuid.New(),
			UserID: userID,
			Title:  "Valid Title",
		}

		mockRepo.EXPECT().
			CreateNote(gomock.Any(), note).
			Return(createdNote, nil)

		result, err := usecase.CreateNote(context.Background(), note)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Title != "Valid Title" {
			t.Errorf("expected title 'Valid Title', got '%s'", result.Title)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		note := models.Note{
			UserID: userID,
			Title:  "",
		}

		_, err := usecase.CreateNote(context.Background(), note)

		if !errors.Is(err, notes.ErrInvalidNoteData) {
			t.Errorf("expected ErrInvalidNoteData, got %v", err)
		}
	})

	t.Run("repository error", func(t *testing.T) {
		note := models.Note{
			UserID: userID,
			Title:  "Valid Title",
		}

		mockRepo.EXPECT().
			CreateNote(gomock.Any(), note).
			Return(nil, errors.New("db error"))

		_, err := usecase.CreateNote(context.Background(), note)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestNoteUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existingNote := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Old Title",
		}
		updatedNote := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "New Title",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(existingNote, nil)
		mockRepo.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any()).
			Return(updatedNote, nil)

		note := models.Note{
			Title: "New Title",
		}
		result, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Title != "New Title" {
			t.Errorf("expected title 'New Title', got '%s'", result.Title)
		}
	})

	t.Run("empty title", func(t *testing.T) {
		note := models.Note{
			Title: "",
		}

		_, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

		if !errors.Is(err, notes.ErrInvalidNoteData) {
			t.Errorf("expected ErrInvalidNoteData, got %v", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		note := models.Note{
			Title: "New Title",
		}
		_, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		existingNote := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
			Title:  "Old Title",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(existingNote, nil)

		note := models.Note{
			Title: "New Title",
		}
		_, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNoteUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existingNote := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Test Note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(existingNote, nil)
		mockRepo.EXPECT().
			DeleteNote(gomock.Any(), noteID).
			Return(nil)

		err := usecase.DeleteNote(context.Background(), noteID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		err := usecase.DeleteNote(context.Background(), noteID, userID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		existingNote := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
			Title:  "Test Note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(existingNote, nil)

		err := usecase.DeleteNote(context.Background(), noteID, userID)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNoteUsecase_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Hello, world!",
		}
		blocks := []models.Block{}
		block := models.Block{
			BlockTypeID: 1,
			Position:    0,
		}
		createdBlock := &models.Block{
			ID:          blockID,
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)
		mockRepo.EXPECT().
			ShiftBlockPositions(gomock.Any(), noteID, 0, 1, gomock.Any()).
			Return(nil)
		mockRepo.EXPECT().
			CreateBlock(gomock.Any(), gomock.Any()).
			Return(createdBlock, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != blockID {
			t.Errorf("expected ID %v, got %v", blockID, result.ID)
		}
	})

	t.Run("success with shift", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Hello, world!",
		}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		block := models.Block{
			BlockTypeID: 1,
			Position:    1,
		}
		createdBlock := &models.Block{
			ID:          blockID,
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    1,
			Content:     "",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)
		mockRepo.EXPECT().
			ShiftBlockPositions(gomock.Any(), noteID, 1, 1, gomock.Any()).
			Return(nil)
		mockRepo.EXPECT().
			CreateBlock(gomock.Any(), gomock.Any()).
			Return(createdBlock, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Position != 1 {
			t.Errorf("expected position 1, got %d", result.Position)
		}
	})

	t.Run("invalid block type", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		block := models.Block{
			BlockTypeID: 0,
			Position:    0,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if !errors.Is(err, notes.ErrInvalidBlockType) {
			t.Errorf("expected ErrInvalidBlockType, got %v", err)
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
		}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		block := models.Block{
			BlockTypeID: 1,
			Position:    5,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)

		_, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if !errors.Is(err, notes.ErrInvalidPosition) {
			t.Errorf("expected ErrInvalidPosition, got %v", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		block := models.Block{
			BlockTypeID: 1,
			Position:    0,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		otherUserID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: otherUserID,
		}
		block := models.Block{
			BlockTypeID: 1,
			Position:    0,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)

		_, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNoteUsecase_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Note",
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		result, err := usecase.GetBlock(context.Background(), blockID, noteID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.ID != blockID {
			t.Errorf("expected ID %v, got %v", blockID, result.ID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.GetBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.GetBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("block belongs to different note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Note",
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		_, err := usecase.GetBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNoteUsecase_UpdateBlockContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	content := "New Content"

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: noteID,
		}
		updatedBlock := &models.Block{
			ID:      blockID,
			NoteID:  noteID,
			Content: content,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, content, gomock.Any()).
			Return(updatedBlock, nil)

		result, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, content)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Content != content {
			t.Errorf("expected content '%s', got '%s'", content, result.Content)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, content)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, content)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("forbidden - wrong note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		_, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, content)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})
}

func TestNoteUsecase_MoveBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success move to higher position", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 0,
		}
		blocks := []models.Block{
			{ID: blockID, Position: 0},
			{ID: uuid.New(), Position: 1},
			{ID: uuid.New(), Position: 2},
		}
		movedBlock := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 2,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)
		mockRepo.EXPECT().
			MoveBlock(gomock.Any(), noteID, blockID, 0, 2, gomock.Any()).
			Return(movedBlock, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 2)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Position != 2 {
			t.Errorf("expected position 2, got %d", result.Position)
		}
	})

	t.Run("success move to lower position", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 2,
		}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
			{ID: blockID, Position: 2},
		}
		movedBlock := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 0,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)
		mockRepo.EXPECT().
			MoveBlock(gomock.Any(), noteID, blockID, 2, 0, gomock.Any()).
			Return(movedBlock, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 0)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Position != 0 {
			t.Errorf("expected position 0, got %d", result.Position)
		}
	})

	t.Run("same position - no move", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 1,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 1)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if result.Position != 1 {
			t.Errorf("expected position 1, got %d", result.Position)
		}
	})

	t.Run("invalid position - negative", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 0,
		}
		blocks := []models.Block{
			{ID: blockID, Position: 0},
			{ID: uuid.New(), Position: 1},
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)

		_, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, -1)

		if !errors.Is(err, notes.ErrInvalidPosition) {
			t.Errorf("expected ErrInvalidPosition, got %v", err)
		}
	})

	t.Run("invalid position - out of range", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 0,
		}
		blocks := []models.Block{
			{ID: blockID, Position: 0},
			{ID: uuid.New(), Position: 1},
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(blocks, nil)

		_, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 5)

		if !errors.Is(err, notes.ErrInvalidPosition) {
			t.Errorf("expected ErrInvalidPosition, got %v", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		_, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 1)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		_, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 1)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})
}

func TestNoteUsecase_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 1,
		}
		returnedNoteID := noteID

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			DeleteBlock(gomock.Any(), blockID).
			Return(&returnedNoteID, nil)
		mockRepo.EXPECT().
			ShiftBlockPositions(gomock.Any(), noteID, 1, -1, gomock.Any()).
			Return(nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(nil, nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(nil, nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("forbidden - block belongs to different note", func(t *testing.T) {
		otherNoteID := uuid.New()
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:     blockID,
			NoteID: otherNoteID,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrForbidden) {
			t.Errorf("expected ErrForbidden, got %v", err)
		}
	})

	t.Run("delete returns nil noteID", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Title note",
		}
		block := &models.Block{
			ID:       blockID,
			NoteID:   noteID,
			Position: 1,
		}

		mockRepo.EXPECT().
			GetNote(gomock.Any(), noteID).
			Return(note, nil)
		mockRepo.EXPECT().
			GetBlock(gomock.Any(), blockID).
			Return(block, nil)
		mockRepo.EXPECT().
			DeleteBlock(gomock.Any(), blockID).
			Return(nil, nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})
}
