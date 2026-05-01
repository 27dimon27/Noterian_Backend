package usecase

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/usecase/mocks"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func setupTestUsecase(t *testing.T) (*noteUsecase, *mocks.MockNoteRepository, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockRepo := mocks.NewMockNoteRepository(ctrl)
	usecase := NewNoteUsecase(mockRepo)
	return usecase, mockRepo, ctrl
}

func TestGetNotes_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedNotes := []models.Note{
		{ID: uuid.New(), UserID: userID, Title: "Note 1"},
		{ID: uuid.New(), UserID: userID, Title: "Note 2"},
	}

	mockRepo.EXPECT().
		GetNotes(gomock.Any(), userID).
		Return(expectedNotes, nil)

	notes, _, err := usecase.GetNotes(context.Background(), userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedNotes, notes)
}

func TestGetNote_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	expectedNote := &models.Note{
		ID:     noteID,
		UserID: userID,
		Title:  "Test Note",
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(expectedNote, nil)

	note, err := usecase.GetNote(context.Background(), noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedNote, note)
}

func TestGetNote_NotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(nil, nil)

	note, err := usecase.GetNote(context.Background(), noteID, userID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrNoteNotFound, err)
	assert.Nil(t, note)
}

func TestGetNote_Forbidden(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	otherUserID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: otherUserID,
		Title:  "Someone's Note",
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)

	note, err := usecase.GetNote(context.Background(), noteID, userID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrForbidden, err)
	assert.Nil(t, note)
}

func TestCreateNote_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	note := models.Note{
		UserID: userID,
		Title:  "New Note",
	}

	createdNote := &models.Note{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "New Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockRepo.EXPECT().
		CreateNote(gomock.Any(), gomock.Any()).
		Return(createdNote, nil)

	result, err := usecase.CreateNote(context.Background(), note)

	assert.NoError(t, err)
	assert.Equal(t, createdNote, result)
}

func TestCreateNote_EmptyTitle(t *testing.T) {
	usecase, _, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	note := models.Note{
		UserID: uuid.New(),
		Title:  "",
	}

	result, err := usecase.CreateNote(context.Background(), note)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidNoteData, err)
	assert.Nil(t, result)
}

func TestUpdateNote_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	existingNote := &models.Note{
		ID:     noteID,
		UserID: userID,
		Title:  "Old Title",
	}
	updatedNote := models.Note{
		Title: "New Title",
	}
	expectedNote := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "New Title",
		UpdatedAt: time.Now(),
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(existingNote, nil)
	mockRepo.EXPECT().
		UpdateNote(gomock.Any(), noteID, gomock.Any()).
		Return(expectedNote, nil)

	result, err := usecase.UpdateNote(context.Background(), noteID, updatedNote, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedNote.Title, result.Title)
}

func TestDeleteNote_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	existingNote := &models.Note{
		ID:     noteID,
		UserID: userID,
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(existingNote, nil)
	mockRepo.EXPECT().
		DeleteNote(gomock.Any(), noteID).
		Return(nil)

	err := usecase.DeleteNote(context.Background(), noteID, userID)

	assert.NoError(t, err)
}

func TestCreateBlock_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := models.Block{
		BlockTypeID: 1,
		Position:    0,
	}
	existingBlocks := []models.Block{}

	createdBlock := &models.Block{
		ID:          uuid.New(),
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(existingBlocks, nil)
	mockRepo.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(nil)
	mockRepo.EXPECT().
		CreateBlock(gomock.Any(), gomock.Any()).
		Return(createdBlock, nil)

	result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, createdBlock, result)
}

func TestCreateBlock_WithExistingBlocks(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := models.Block{
		BlockTypeID: 1,
		Position:    2,
	}
	existingBlocks := []models.Block{
		{ID: uuid.New(), Position: 0},
		{ID: uuid.New(), Position: 1},
		{ID: uuid.New(), Position: 2},
	}

	createdBlock := &models.Block{
		ID:          uuid.New(),
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(existingBlocks, nil)
	mockRepo.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 2, 1).
		Return(nil)
	mockRepo.EXPECT().
		CreateBlock(gomock.Any(), gomock.Any()).
		Return(createdBlock, nil)

	result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestCreateBlock_InvalidType(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
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

	result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidBlockType, err)
	assert.Nil(t, result)
}

func TestCreateBlock_InvalidPosition(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := models.Block{
		BlockTypeID: 1,
		Position:    10,
	}
	existingBlocks := []models.Block{
		{ID: uuid.New(), Position: 0},
		{ID: uuid.New(), Position: 1},
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(existingBlocks, nil)

	result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidPosition, err)
	assert.Nil(t, result)
}

func TestMoveBlock_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
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
		MoveBlock(gomock.Any(), noteID, blockID, 0, 2).
		Return(movedBlock, nil)

	result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 2)

	assert.NoError(t, err)
	assert.Equal(t, 2, result.Position)
}

func TestMoveBlock_SamePosition(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
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

	assert.NoError(t, err)
	assert.Equal(t, 1, result.Position)
}

func TestMoveBlock_InvalidPosition(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:       blockID,
		NoteID:   noteID,
		Position: 0,
	}
	blocks := []models.Block{
		{ID: blockID, Position: 0},
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

	result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 5)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidPosition, err)
	assert.Nil(t, result)
}

func TestDeleteBlock_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:       blockID,
		NoteID:   noteID,
		Position: 1,
	}
	blockNoteID := noteID

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		DeleteBlock(gomock.Any(), blockID).
		Return(&blockNoteID, nil)
	mockRepo.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 1, -1).
		Return(nil)

	err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

	assert.NoError(t, err)
}

func TestUpdateBlockFormatting_ForTextBlock(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Content:     "Hello World",
	}
	blockType := &models.BlockType{
		ID:   1,
		Name: "text",
	}
	expectedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{formattingRange},
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 1).
		Return(blockType, nil)
	mockRepo.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, formattingRange).
		Return(expectedFormatting, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFormatting, result)
}

func TestUpdateBlockFormatting_ForImageBlock_ValidFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	textAlign := 1
	formattingRange := models.FormattingRange{
		StartPos:  0,
		EndPos:    9,
		TextAlign: &textAlign,
	}

	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 2,
		Content:     "image.jpg",
	}
	blockType := &models.BlockType{
		ID:   2,
		Name: "image",
	}
	expectedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{formattingRange},
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 2).
		Return(blockType, nil)
	mockRepo.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, formattingRange).
		Return(expectedFormatting, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFormatting, result)
}

func TestUpdateBlockFormatting_ForImageBlock_ValidRange(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	textAlign := 1
	formattingRange := models.FormattingRange{
		StartPos:  0,
		EndPos:    9,
		TextAlign: &textAlign,
	}

	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 2,
		Content:     "image.jpg",
	}
	blockType := &models.BlockType{
		ID:   2,
		Name: "image",
	}
	expectedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{formattingRange},
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 2).
		Return(blockType, nil)
	mockRepo.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, formattingRange).
		Return(expectedFormatting, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.NoError(t, err)
	assert.Equal(t, expectedFormatting, result)
}

func TestUpdateBlockFormatting_ForImageBlock_InvalidFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 2,
		Content:     "image.jpg",
	}
	blockType := &models.BlockType{
		ID:   2,
		Name: "image",
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 2).
		Return(blockType, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidFormattingForImageBlock, err)
	assert.Nil(t, result)
}

func TestUpdateBlockFormatting_InvalidRange(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	formattingRange := models.FormattingRange{
		StartPos: 10,
		EndPos:   5,
	}

	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Content:     "Hello",
	}
	blockType := &models.BlockType{
		ID:   1,
		Name: "text",
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 1).
		Return(blockType, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidFormattingRange, err)
	assert.Nil(t, result)
}

func TestGetBlocksWithFormatting_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()
	blocks := []models.Block{
		{ID: uuid.New(), NoteID: noteID, Position: 0},
		{ID: uuid.New(), NoteID: noteID, Position: 1},
	}
	formattings := map[string]models.BlockFormatting{
		blocks[0].ID.String(): {BlockID: blocks[0].ID.String(), Ranges: []models.FormattingRange{}},
	}

	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(blocks, nil)
	mockRepo.EXPECT().
		GetBlocksFormatting(gomock.Any(), gomock.Any()).
		Return(formattings, nil)

	resultBlocks, resultFormattings, err := usecase.GetBlocksWithFormatting(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Equal(t, blocks, resultBlocks)
	assert.Equal(t, formattings, resultFormattings)
}

func TestGetBlocksWithFormatting_EmptyBlocks(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()
	blocks := []models.Block{}

	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(blocks, nil)

	resultBlocks, resultFormattings, err := usecase.GetBlocksWithFormatting(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Equal(t, blocks, resultBlocks)
	assert.NotNil(t, resultFormattings)
	assert.Empty(t, resultFormattings)
}

func TestGetBlockFormatting_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{
		ID:     noteID,
		UserID: userID,
	}
	block := &models.Block{
		ID:     blockID,
		NoteID: noteID,
	}
	expectedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockFormatting(gomock.Any(), blockID).
		Return(expectedFormatting, nil)

	result, err := usecase.GetBlockFormatting(context.Background(), blockID, noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedFormatting, result)
}

func TestGetBlocks_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()
	expectedBlocks := []models.Block{
		{ID: uuid.New(), NoteID: noteID, Position: 0},
		{ID: uuid.New(), NoteID: noteID, Position: 1},
	}

	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(expectedBlocks, nil)

	blocks, err := usecase.GetBlocks(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Equal(t, expectedBlocks, blocks)
}

func TestGetBlocks_Error(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()

	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(nil, errors.New("database error"))

	blocks, err := usecase.GetBlocks(context.Background(), noteID)

	assert.Error(t, err)
	assert.Nil(t, blocks)
}

func TestUpdateNote_EmptyTitle(t *testing.T) {
	usecase, _, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := models.Note{Title: ""}

	result, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidNoteData, err)
	assert.Nil(t, result)
}

func TestUpdateNote_GetNoteError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := models.Note{Title: "Updated Title"}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(nil, errors.New("database error"))

	result, err := usecase.UpdateNote(context.Background(), noteID, note, userID)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeleteNote_GetNoteError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(nil, errors.New("database error"))

	err := usecase.DeleteNote(context.Background(), noteID, userID)

	assert.Error(t, err)
}

func TestCreateBlock_ShiftPositionsError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := models.Block{BlockTypeID: 1, Position: 0}
	existingBlocks := []models.Block{}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(existingBlocks, nil)
	mockRepo.EXPECT().
		ShiftBlockPositions(gomock.Any(), noteID, 0, 1).
		Return(errors.New("shift error"))

	result, err := usecase.CreateBlock(context.Background(), noteID, userID, block)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestMoveBlock_GetBlocksError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := &models.Block{ID: blockID, NoteID: noteID, Position: 0}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlocks(gomock.Any(), noteID).
		Return(nil, errors.New("database error"))

	result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, 2)

	assert.Error(t, err)
	assert.Nil(t, result)
}

func TestDeleteBlock_BlockNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := &models.Block{ID: blockID, NoteID: noteID, Position: 1}

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

	assert.Error(t, err)
	assert.Equal(t, notes.ErrBlockNotFound, err)
}

func TestResetBlockFormatting_Success(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := &models.Block{ID: blockID, NoteID: noteID}
	expectedFormatting := &models.BlockFormatting{BlockID: blockID.String(), Ranges: []models.FormattingRange{}}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		ResetBlockFormatting(gomock.Any(), blockID).
		Return(expectedFormatting, nil)

	result, err := usecase.ResetBlockFormatting(context.Background(), blockID, noteID, userID)

	assert.NoError(t, err)
	assert.Equal(t, expectedFormatting, result)
}

func TestResetBlockFormatting_NoteNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(nil, nil)

	result, err := usecase.ResetBlockFormatting(context.Background(), blockID, noteID, userID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrNoteNotFound, err)
	assert.Nil(t, result)
}

func TestResetBlockFormatting_BlockNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(nil, nil)

	result, err := usecase.ResetBlockFormatting(context.Background(), blockID, noteID, userID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrBlockNotFound, err)
	assert.Nil(t, result)
}

func TestUpdateBlockFormatting_BlockTypeNotFound(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 1, Content: "Hello"}
	formattingRange := models.FormattingRange{StartPos: 0, EndPos: 5}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 1).
		Return(nil, nil)

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrInvalidBlockType, err)
	assert.Nil(t, result)
}

func TestUpdateBlockFormatting_GetBlockTypeError(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	note := &models.Note{ID: noteID, UserID: userID}
	block := &models.Block{ID: blockID, NoteID: noteID, BlockTypeID: 1, Content: "Hello"}
	formattingRange := models.FormattingRange{StartPos: 0, EndPos: 5}

	mockRepo.EXPECT().
		GetNote(gomock.Any(), noteID).
		Return(note, nil)
	mockRepo.EXPECT().
		GetBlock(gomock.Any(), blockID).
		Return(block, nil)
	mockRepo.EXPECT().
		GetBlockType(gomock.Any(), 1).
		Return(nil, errors.New("database error"))

	result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, formattingRange)

	assert.Error(t, err)
	assert.Nil(t, result)
}
