// usecase/notes_test.go
package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var log = logger.Init()

func TestNoteUsecase_GetNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	expectedNotes := []models.Note{
		{ID: uuid.New(), UserID: userID, Title: "Note 1"},
		{ID: uuid.New(), UserID: userID, Title: "Note 2"},
	}

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().GetNotes(gomock.Any(), userID).Return(expectedNotes, nil)

		notes, err := usecase.GetNotes(context.Background(), userID)

		assert.NoError(t, err)
		assert.Equal(t, expectedNotes, notes)
	})

	t.Run("repository error", func(t *testing.T) {
		expectedErr := errors.New("db error")
		mockRepo.EXPECT().GetNotes(gomock.Any(), userID).Return(nil, expectedErr)

		notes, err := usecase.GetNotes(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, notes)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_GetNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		note, blocks, formattings, err := usecase.GetNote(context.Background(), noteID, userID)

		assert.Error(t, err)
		assert.Nil(t, note)
		assert.Nil(t, blocks)
		assert.Nil(t, formattings)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})

	t.Run("forbidden - private note of another user", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: uuid.New(), Title: "Private Note", IsPublic: false}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)

		noteResult, blocks, formattings, err := usecase.GetNote(context.Background(), noteID, userID)

		assert.Error(t, err)
		assert.Nil(t, noteResult)
		assert.Nil(t, blocks)
		assert.Nil(t, formattings)
		assert.Equal(t, notes.ErrForbidden, err)
	})

	t.Run("get blocks error", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID, IsPublic: false}
		expectedErr := errors.New("blocks error")

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(nil, expectedErr)

		noteResult, blocks, formattings, err := usecase.GetNote(context.Background(), noteID, userID)

		assert.Error(t, err)
		assert.Nil(t, noteResult)
		assert.Nil(t, blocks)
		assert.Nil(t, formattings)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_GetPublicNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	noteID := uuid.New()

	t.Run("success with public note", func(t *testing.T) {
		note := &models.Note{ID: noteID, IsPublic: true, Title: "Public Note"}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)

		result, err := usecase.GetPublicNote(context.Background(), noteID)

		assert.NoError(t, err)
		assert.Equal(t, note, result)
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		result, err := usecase.GetPublicNote(context.Background(), noteID)

		assert.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("note is private", func(t *testing.T) {
		note := &models.Note{ID: noteID, IsPublic: false}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)

		result, err := usecase.GetPublicNote(context.Background(), noteID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})
}

func TestNoteUsecase_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := models.Note{UserID: userID, Title: "New Note"}
		createdNote := &models.Note{ID: uuid.New(), UserID: userID, Title: "New Note"}

		mockRepo.EXPECT().CreateNote(gomock.Any(), note).Return(createdNote, nil)

		result, err := usecase.CreateNote(context.Background(), note)

		assert.NoError(t, err)
		assert.Equal(t, createdNote, result)
	})

	t.Run("empty title", func(t *testing.T) {
		note := models.Note{UserID: userID, Title: ""}

		result, err := usecase.CreateNote(context.Background(), note)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidNoteData, err)
	})

	t.Run("repository error", func(t *testing.T) {
		note := models.Note{UserID: userID, Title: "New Note"}
		expectedErr := errors.New("db error")

		mockRepo.EXPECT().CreateNote(gomock.Any(), note).Return(nil, expectedErr)

		result, err := usecase.CreateNote(context.Background(), note)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		existingNote := &models.Note{ID: noteID, UserID: userID, IsPublic: false}
		updatedNoteData := models.Note{Title: "Updated Title", IsPublic: true}
		updatedNote := &models.Note{ID: noteID, UserID: userID, Title: "Updated Title", IsPublic: true}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(existingNote, nil)
		mockRepo.EXPECT().UpdateNote(gomock.Any(), noteID, updatedNoteData).Return(updatedNote, nil)

		result, err := usecase.UpdateNote(context.Background(), noteID, userID, updatedNoteData)

		assert.NoError(t, err)
		assert.Equal(t, updatedNote, result)
	})

	t.Run("empty title", func(t *testing.T) {
		existingNote := &models.Note{ID: noteID, UserID: userID}
		updatedNoteData := models.Note{Title: ""}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(existingNote, nil)

		result, err := usecase.UpdateNote(context.Background(), noteID, userID, updatedNoteData)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidNoteData, err)
	})

	t.Run("note not found", func(t *testing.T) {
		updatedNoteData := models.Note{Title: "Updated"}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		result, err := usecase.UpdateNote(context.Background(), noteID, userID, updatedNoteData)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})

	t.Run("forbidden", func(t *testing.T) {
		existingNote := &models.Note{ID: noteID, UserID: uuid.New(), IsPublic: false}
		updatedNoteData := models.Note{Title: "Updated"}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(existingNote, nil)

		result, err := usecase.UpdateNote(context.Background(), noteID, userID, updatedNoteData)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrForbidden, err)
	})
}

func TestNoteUsecase_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("header not found - continue", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		blocks := []models.Block{}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
		mockAttachments.EXPECT().DeleteHeader(gomock.Any(), noteID, userID).Return(status.Error(codes.NotFound, "not found"))
		mockRepo.EXPECT().DeleteNote(gomock.Any(), noteID).Return(nil)

		err := usecase.DeleteNote(context.Background(), noteID, userID)

		assert.NoError(t, err)
	})

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		err := usecase.DeleteNote(context.Background(), noteID, userID)

		assert.Error(t, err)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})
}

func TestNoteUsecase_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success at end", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := models.Block{BlockTypeID: 1, Position: 2}
		existingBlocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		createdBlock := &models.Block{ID: uuid.New(), NoteID: noteID, BlockTypeID: 1, Position: 2}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 2, 1).Return(nil)
		mockRepo.EXPECT().CreateBlock(gomock.Any(), gomock.Any()).Return(createdBlock, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		assert.NoError(t, err)
		assert.Equal(t, createdBlock, result)
	})

	t.Run("success at beginning", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := models.Block{BlockTypeID: 1, Position: 0}
		existingBlocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		createdBlock := &models.Block{ID: uuid.New(), NoteID: noteID, BlockTypeID: 1, Position: 0}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, 1).Return(nil)
		mockRepo.EXPECT().CreateBlock(gomock.Any(), gomock.Any()).Return(createdBlock, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		assert.NoError(t, err)
		assert.Equal(t, createdBlock, result)
	})

	t.Run("invalid block type", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := models.Block{BlockTypeID: 0, Position: 0}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidBlockType, err)
	})

	t.Run("invalid position", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := models.Block{BlockTypeID: 1, Position: 5}
		existingBlocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidPosition, err)
	})

	t.Run("shift positions error", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := models.Block{BlockTypeID: 1, Position: 1}
		existingBlocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		expectedErr := errors.New("shift error")

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 1, 1).Return(expectedErr)

		result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_UpdateBlockContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Content: "old content"}
		updatedBlock := &models.Block{ID: blockID, NoteID: noteID, Content: "new content"}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().UpdateBlockContent(gomock.Any(), blockID, "new content").Return(updatedBlock, nil)

		result, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, "new content")

		assert.NoError(t, err)
		assert.Equal(t, updatedBlock, result)
	})

	t.Run("block not found", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, notes.ErrBlockNotFound)

		result, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, "content")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrBlockNotFound, err)
	})

	t.Run("block doesn't belong to note", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: uuid.New()}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)

		result, err := usecase.UpdateBlockContent(context.Background(), blockID, noteID, userID, "content")

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrForbidden, err)
	})
}

func TestNoteUsecase_MoveBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success move forward", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 1}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: blockID, Position: 1},
			{ID: uuid.New(), Position: 2},
		}
		movedBlock := &models.Block{ID: blockID, NoteID: noteID, Position: 2}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
		mockRepo.EXPECT().MoveBlock(gomock.Any(), noteID, blockID, 1, 2).Return(movedBlock, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 2)

		assert.NoError(t, err)
		assert.Equal(t, movedBlock, result)
	})

	t.Run("success move backward", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 2}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
			{ID: blockID, Position: 2},
		}
		movedBlock := &models.Block{ID: blockID, NoteID: noteID, Position: 1}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
		mockRepo.EXPECT().MoveBlock(gomock.Any(), noteID, blockID, 2, 1).Return(movedBlock, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 1)

		assert.NoError(t, err)
		assert.Equal(t, movedBlock, result)
	})

	t.Run("same position - no move", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 1}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 1)

		assert.NoError(t, err)
		assert.Equal(t, block, result)
	})

	t.Run("invalid position", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 1}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: blockID, Position: 1},
		}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)

		result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 5)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidPosition, err)
	})
}

func TestNoteUsecase_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success delete text block", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 1, BlockTypeID: 1}
		returnedNoteID := noteID

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().DeleteBlock(gomock.Any(), blockID).Return(&returnedNoteID, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 1, -1).Return(nil)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		assert.NoError(t, err)
	})

	t.Run("delete attachment error", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, Position: 1, BlockTypeID: 2}
		expectedErr := errors.New("delete attachment failed")

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockAttachments.EXPECT().DeleteAttachment(gomock.Any(), blockID, noteID, userID).Return(expectedErr)

		err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

		assert.Error(t, err)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_UpdateBlockFormatting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success for text block", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 1, Content: "Hello World"}
		blockType := &models.BlockType{ID: 1, Name: "text"}
		formattingRange := models.FormattingRange{StartPos: 0, EndPos: 5, Bold: boolPtr(true)}
		updatedFormatting := &models.BlockFormatting{BlockID: blockID.String(), Ranges: []models.FormattingRange{formattingRange}}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 1).Return(blockType, nil)
		mockRepo.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, formattingRange).Return(updatedFormatting, nil)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

		assert.NoError(t, err)
		assert.Equal(t, updatedFormatting, result)
	})

	t.Run("success for image block with alignment only", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 2, Content: "image.jpg"}
		blockType := &models.BlockType{ID: 2, Name: "image"}
		formattingRange := models.FormattingRange{StartPos: 0, EndPos: 8, TextAlign: intPtr(1)}
		updatedFormatting := &models.BlockFormatting{BlockID: blockID.String(), Ranges: []models.FormattingRange{formattingRange}}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 2).Return(blockType, nil)
		mockRepo.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, formattingRange).Return(updatedFormatting, nil)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

		assert.NoError(t, err)
		assert.Equal(t, updatedFormatting, result)
	})

	t.Run("invalid formatting for image block - bold", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 2}
		blockType := &models.BlockType{ID: 2, Name: "image"}
		formattingRange := models.FormattingRange{StartPos: 0, EndPos: 8, Bold: boolPtr(true)}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 2).Return(blockType, nil)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidFormattingForImageBlock, err)
	})

	t.Run("unsupported block type", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 3}
		blockType := &models.BlockType{ID: 3, Name: "video"}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 3).Return(blockType, nil)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, models.FormattingRange{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrFormattingNotSupported, err)
	})

	t.Run("invalid formatting range", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 1, Content: "Hello"}
		blockType := &models.BlockType{ID: 1, Name: "text"}
		formattingRange := models.FormattingRange{StartPos: 10, EndPos: 15}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 1).Return(blockType, nil)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrInvalidFormattingRange, err)
	})

	t.Run("block type not found", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 999}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
		mockRepo.EXPECT().GetBlockType(gomock.Any(), 999).Return(nil, notes.ErrBlockTypeNotFound)

		result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, models.FormattingRange{})

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, notes.ErrBlockTypeNotFound, err)
	})
}

func TestNoteUsecase_GetSubnotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		subnotes := []models.Note{
			{ID: uuid.New(), UserID: userID, Title: "Subnote 1", ParentID: &noteID},
			{ID: uuid.New(), UserID: userID, Title: "Subnote 2", ParentID: &noteID},
		}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetSubnotes(gomock.Any(), noteID).Return(subnotes, nil)

		result, err := usecase.GetSubnotes(context.Background(), noteID, userID)

		assert.NoError(t, err)
		assert.Equal(t, subnotes, result)
	})
}

func TestNoteUsecase_CreateSubnote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	parentNoteID := uuid.New()

	t.Run("success with specified position", func(t *testing.T) {
		note := &models.Note{ID: parentNoteID, UserID: userID}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		createdBlock := &models.Block{ID: uuid.New(), NoteID: parentNoteID, BlockTypeID: 5, Position: 1}
		createdNote := &models.Note{ID: uuid.New(), UserID: userID, Title: "New Subnote", ParentID: &parentNoteID}

		mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), parentNoteID).Return(blocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), parentNoteID, 1, 1).Return(nil)
		mockRepo.EXPECT().CreateBlock(gomock.Any(), gomock.Any()).Return(createdBlock, nil)
		mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)

		resultNote, resultBlockID, err := usecase.CreateSubnote(context.Background(), parentNoteID, userID,
			models.Note{Title: "New Subnote"}, true, 1)

		assert.NoError(t, err)
		assert.Equal(t, createdNote, resultNote)
		assert.Equal(t, createdBlock.ID, resultBlockID)
	})

	t.Run("success at end", func(t *testing.T) {
		note := &models.Note{ID: parentNoteID, UserID: userID}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}
		createdBlock := &models.Block{ID: uuid.New(), NoteID: parentNoteID, BlockTypeID: 5, Position: 2}
		createdNote := &models.Note{ID: uuid.New(), UserID: userID, Title: "New Subnote", ParentID: &parentNoteID}

		mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), parentNoteID).Return(blocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), parentNoteID, 2, 1).Return(nil)
		mockRepo.EXPECT().CreateBlock(gomock.Any(), gomock.Any()).Return(createdBlock, nil)
		mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)

		resultNote, resultBlockID, err := usecase.CreateSubnote(context.Background(), parentNoteID, userID,
			models.Note{Title: "New Subnote"}, false, 0)

		assert.NoError(t, err)
		assert.Equal(t, createdNote, resultNote)
		assert.Equal(t, createdBlock.ID, resultBlockID)
	})

	t.Run("invalid position", func(t *testing.T) {
		note := &models.Note{ID: parentNoteID, UserID: userID}
		blocks := []models.Block{
			{ID: uuid.New(), Position: 0},
			{ID: uuid.New(), Position: 1},
		}

		mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), parentNoteID).Return(blocks, nil)

		resultNote, resultBlockID, err := usecase.CreateSubnote(context.Background(), parentNoteID, userID,
			models.Note{Title: "New Subnote"}, true, 5)

		assert.Error(t, err)
		assert.Nil(t, resultNote)
		assert.Equal(t, uuid.Nil, resultBlockID)
		assert.Equal(t, notes.ErrInvalidPosition, err)
	})

	t.Run("shift positions error - rollback", func(t *testing.T) {
		note := &models.Note{ID: parentNoteID, UserID: userID}
		blocks := []models.Block{{ID: uuid.New(), Position: 0}}
		expectedErr := errors.New("shift error")

		mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), parentNoteID).Return(blocks, nil)
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), parentNoteID, 0, 1).Return(expectedErr)

		resultNote, resultBlockID, err := usecase.CreateSubnote(context.Background(), parentNoteID, userID,
			models.Note{Title: "New Subnote"}, true, 0)

		assert.Error(t, err)
		assert.Nil(t, resultNote)
		assert.Equal(t, uuid.Nil, resultBlockID)
		assert.Equal(t, expectedErr, err)
	})
}

func TestNoteUsecase_DeleteSubnote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		blocks := []models.Block{}

		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
		mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
		mockRepo.EXPECT().DeleteNote(gomock.Any(), subnoteID).Return(nil)

		err := usecase.DeleteSubnote(context.Background(), noteID, subnoteID, userID)

		assert.NoError(t, err)
	})
}

// usecase/notes_test.go - исправленная версия TestNoteUsecase_GenerateNotePDF

func TestNoteUsecase_GenerateNotePDF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("note not found", func(t *testing.T) {
		mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		pdfBuffer, err := usecase.GenerateNotePDF(context.Background(), noteID, userID)

		assert.Error(t, err)
		assert.Nil(t, pdfBuffer)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})
}

func TestNoteUsecase_ShiftBlockPositions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mocks.NewMockNoteRepository(ctrl)
	mockAttachments := mocks.NewMockAttachmentsServiceClient(ctrl)
	usecase := NewNoteUsecase(mockRepo, mockAttachments, log)

	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 2, 1).Return(nil)

		err := usecase.ShiftBlockPositions(context.Background(), noteID, 2, 1)

		assert.NoError(t, err)
	})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
