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

func TestResetBlockFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Test content",
	}

	resetFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().ResetBlockFormatting(gomock.Any(), blockID).Return(resetFormatting, nil)
			},
			wantErr: false,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "forbidden - private note not owned",
			setupMock: func() {
				privateNote := &models.Note{
					ID:       noteID,
					UserID:   uuid.New(),
					Title:    "Private Note",
					IsPublic: false,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(privateNote, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "block not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrBlockNotFound,
		},
		{
			name: "block belongs to different note",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				otherNoteID := uuid.New()
				blockOther := &models.Block{
					ID:          blockID,
					NoteID:      otherNoteID,
					BlockTypeID: 1,
					Position:    0,
				}
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(blockOther, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "reset formatting error",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().ResetBlockFormatting(gomock.Any(), blockID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.ResetBlockFormatting(context.Background(), blockID, noteID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestGetSubnotes(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID1 := uuid.New()
	subnoteID2 := uuid.New()
	parentID := noteID

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Parent Note",
		IsPublic: false,
	}

	subnotes := []models.Note{
		{ID: subnoteID1, UserID: userID, Title: "Subnote 1", ParentID: &parentID, IsPublic: false},
		{ID: subnoteID2, UserID: userID, Title: "Subnote 2", ParentID: &parentID, IsPublic: false},
	}

	tests := []struct {
		name      string
		setupMock func()
		wantLen   int
		wantErr   bool
		errType   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetSubnotes(gomock.Any(), noteID).Return(subnotes, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success - no subnotes",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetSubnotes(gomock.Any(), noteID).Return([]models.Note{}, nil)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "forbidden - private note not owned",
			setupMock: func() {
				privateNote := &models.Note{
					ID:       noteID,
					UserID:   uuid.New(),
					Title:    "Private Note",
					IsPublic: false,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(privateNote, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "public note - accessible by other user",
			setupMock: func() {
				publicNote := &models.Note{
					ID:       noteID,
					UserID:   uuid.New(),
					Title:    "Public Note",
					IsPublic: true,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(publicNote, nil)
				mockRepo.EXPECT().GetSubnotes(gomock.Any(), noteID).Return(subnotes, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "get subnotes error",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetSubnotes(gomock.Any(), noteID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.GetSubnotes(context.Background(), noteID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.Len(t, result, tt.wantLen)
			}
		})
	}
}

func TestCreateSubnote(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	parentNoteID := uuid.New()
	subnoteID := uuid.New()

	parentNote := &models.Note{
		ID:       parentNoteID,
		UserID:   userID,
		Title:    "Parent Note",
		IsPublic: false,
	}

	newNote := models.Note{
		UserID:   userID,
		Title:    "New Subnote",
		ParentID: &parentNoteID,
		IsPublic: false,
	}

	createdNote := &models.Note{
		ID:        subnoteID,
		UserID:    userID,
		Title:     "New Subnote",
		ParentID:  &parentNoteID,
		IsPublic:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)
			},
			wantErr: false,
		},
		{
			name: "parent note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "forbidden - parent note not owned and not public",
			setupMock: func() {
				privateNote := &models.Note{
					ID:       parentNoteID,
					UserID:   uuid.New(),
					Title:    "Private Note",
					IsPublic: false,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(privateNote, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "public parent note - allowed",
			setupMock: func() {
				publicNote := &models.Note{
					ID:       parentNoteID,
					UserID:   uuid.New(),
					Title:    "Public Note",
					IsPublic: true,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(publicNote, nil)
				mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)
			},
			wantErr: false,
		},
		{
			name: "create note error",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "empty title validation",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				// Note creation will fail in CreateNote method due to empty title
				mockRepo.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, notes.ErrInvalidNoteData)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.CreateSubnote(context.Background(), parentNoteID, userID, newNote)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, createdNote.ID, result.ID)
				assert.Equal(t, createdNote.Title, result.Title)
			}
		})
	}
}

func TestDeleteSubnote(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	parentNoteID := uuid.New()
	subnoteID := uuid.New()

	parentNote := &models.Note{
		ID:       parentNoteID,
		UserID:   userID,
		Title:    "Parent Note",
		IsPublic: false,
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				mockRepo.EXPECT().DeleteNote(gomock.Any(), subnoteID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "parent note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "forbidden - parent note not owned and not public",
			setupMock: func() {
				privateNote := &models.Note{
					ID:       parentNoteID,
					UserID:   uuid.New(),
					Title:    "Private Note",
					IsPublic: false,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(privateNote, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "public parent note - allowed",
			setupMock: func() {
				publicNote := &models.Note{
					ID:       parentNoteID,
					UserID:   uuid.New(),
					Title:    "Public Note",
					IsPublic: true,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(publicNote, nil)
				mockRepo.EXPECT().DeleteNote(gomock.Any(), subnoteID).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "subnote not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				mockRepo.EXPECT().DeleteNote(gomock.Any(), subnoteID).Return(notes.ErrNoteNotFound)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "delete note error",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), parentNoteID).Return(parentNote, nil)
				mockRepo.EXPECT().DeleteNote(gomock.Any(), subnoteID).Return(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := usecase.DeleteSubnote(context.Background(), parentNoteID, subnoteID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetBlocksWithFormattingEmptyBlocks(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()

	mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return([]models.Block{}, nil)
	// GetBlocksFormatting should not be called

	blocks, formattings, err := usecase.GetBlocksWithFormatting(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Len(t, blocks, 0)
	assert.NotNil(t, formattings)
	assert.Len(t, formattings, 0)
}

func TestGetBlocksWithFormattingPartialFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()
	blockID1 := uuid.New()
	blockID2 := uuid.New()

	blocks := []models.Block{
		{ID: blockID1, NoteID: noteID, BlockTypeID: 1, Position: 0},
		{ID: blockID2, NoteID: noteID, BlockTypeID: 1, Position: 1},
	}

	blockIDs := []uuid.UUID{blockID1, blockID2}

	// Only blockID1 has formatting
	formattings := map[string]models.BlockFormatting{
		blockID1.String(): {
			BlockID: blockID1.String(),
			Ranges: []models.FormattingRange{
				{StartPos: 0, EndPos: 3, Bold: boolPtr(true)},
			},
		},
	}

	mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
	mockRepo.EXPECT().GetBlocksFormatting(gomock.Any(), blockIDs).Return(formattings, nil)

	resultBlocks, resultFormattings, err := usecase.GetBlocksWithFormatting(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Len(t, resultBlocks, 2)
	assert.Len(t, resultFormattings, 1)
	assert.Contains(t, resultFormattings, blockID1.String())
	assert.NotContains(t, resultFormattings, blockID2.String())
}

// Helper function
func boolPtr(b bool) *bool {
	return &b
}

func TestGetBlocksWithFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	noteID := uuid.New()
	blockID1 := uuid.New()
	blockID2 := uuid.New()

	blocks := []models.Block{
		{ID: blockID1, NoteID: noteID, BlockTypeID: 1, Position: 0, Content: "Block 1"},
		{ID: blockID2, NoteID: noteID, BlockTypeID: 1, Position: 1, Content: "Block 2"},
	}

	blockIDs := []uuid.UUID{blockID1, blockID2}

	formattings := map[string]models.BlockFormatting{
		blockID1.String(): {
			BlockID: blockID1.String(),
			Ranges: []models.FormattingRange{
				{StartPos: 0, EndPos: 3, Bold: boolPtr(true)},
			},
		},
		blockID2.String(): {
			BlockID: blockID2.String(),
			Ranges: []models.FormattingRange{
				{StartPos: 0, EndPos: 5, Italic: boolPtr(true)},
			},
		},
	}

	tests := []struct {
		name      string
		setupMock func()
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success with blocks and formatting",
			setupMock: func() {
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
				mockRepo.EXPECT().GetBlocksFormatting(gomock.Any(), blockIDs).Return(formattings, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "success with empty blocks",
			setupMock: func() {
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return([]models.Block{}, nil)
				// GetBlocksFormatting should not be called when blocks are empty
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "success with blocks but no formatting",
			setupMock: func() {
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
				mockRepo.EXPECT().GetBlocksFormatting(gomock.Any(), blockIDs).Return(map[string]models.BlockFormatting{}, nil)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "get blocks error",
			setupMock: func() {
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "get formatting error",
			setupMock: func() {
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
				mockRepo.EXPECT().GetBlocksFormatting(gomock.Any(), blockIDs).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			resultBlocks, resultFormattings, err := usecase.GetBlocksWithFormatting(context.Background(), noteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resultBlocks, tt.wantLen)
				assert.NotNil(t, resultFormattings)
			}
		})
	}
}

func TestGetBlockFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	block := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Test content",
	}

	formatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{StartPos: 0, EndPos: 3, Bold: boolPtr(true)},
			{StartPos: 4, EndPos: 10, Italic: boolPtr(true)},
		},
	}

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
		errType   error
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().GetBlockFormatting(gomock.Any(), blockID).Return(formatting, nil)
			},
			wantErr: false,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrNoteNotFound,
		},
		{
			name: "forbidden - private note not owned",
			setupMock: func() {
				privateNote := &models.Note{
					ID:       noteID,
					UserID:   uuid.New(),
					Title:    "Private Note",
					IsPublic: false,
				}
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(privateNote, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "block not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: true,
			errType: notes.ErrBlockNotFound,
		},
		{
			name: "block belongs to different note",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				otherNoteID := uuid.New()
				blockOther := &models.Block{
					ID:          blockID,
					NoteID:      otherNoteID,
					BlockTypeID: 1,
					Position:    0,
				}
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(blockOther, nil)
			},
			wantErr: true,
			errType: notes.ErrForbidden,
		},
		{
			name: "get formatting error",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().GetBlockFormatting(gomock.Any(), blockID).Return(nil, errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.GetBlockFormatting(context.Background(), blockID, noteID, userID)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.errType != nil {
					assert.ErrorIs(t, err, tt.errType)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, formatting.BlockID, result.BlockID)
				assert.Len(t, result.Ranges, 2)
			}
		})
	}
}

func TestCreateBlock(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	tests := []struct {
		name      string
		block     models.Block
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success - create at beginning with shift",
			block: models.Block{
				BlockTypeID: 1,
				Position:    0,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				existingBlocks := []models.Block{
					{ID: uuid.New(), NoteID: noteID, Position: 0},
					{ID: uuid.New(), NoteID: noteID, Position: 1},
				}
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)
				mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 0, 1).Return(nil)
				mockRepo.EXPECT().CreateBlock(gomock.Any(), gomock.Any()).Return(&models.Block{ID: blockID, Position: 0}, nil)
			},
			wantErr: false,
		},
		{
			name: "invalid block type",
			block: models.Block{
				BlockTypeID: 0,
				Position:    0,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
			},
			wantErr: true,
		},
		{
			name: "invalid position",
			block: models.Block{
				BlockTypeID: 1,
				Position:    10,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				existingBlocks := []models.Block{
					{ID: uuid.New(), NoteID: noteID, Position: 0},
					{ID: uuid.New(), NoteID: noteID, Position: 1},
				}
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(existingBlocks, nil)
			},
			wantErr: true,
		},
		{
			name: "note not found",
			block: models.Block{
				BlockTypeID: 1,
				Position:    0,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.CreateBlock(context.Background(), noteID, userID, tt.block)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestUpdateBlockFormatting(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	textBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Content:     "Hello World",
	}

	imageBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 2,
		Content:     "image.jpg",
	}

	textBlockType := &models.BlockType{ID: 1, Name: "text"}
	imageBlockType := &models.BlockType{ID: 2, Name: "image"}

	boldTrue := true
	// textAlignCenter := 1

	tests := []struct {
		name            string
		formattingRange models.FormattingRange
		setupMock       func()
		wantErr         bool
	}{
		{
			name: "success - text block formatting",
			formattingRange: models.FormattingRange{
				StartPos: 0,
				EndPos:   5,
				Bold:     &boldTrue,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(textBlock, nil)
				mockRepo.EXPECT().GetBlockType(gomock.Any(), textBlock.BlockTypeID).Return(textBlockType, nil)
				mockRepo.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, gomock.Any()).Return(&models.BlockFormatting{BlockID: blockID.String()}, nil)
			},
			wantErr: false,
		},
		{
			name: "error - image block with text formatting",
			formattingRange: models.FormattingRange{
				StartPos: 0,
				EndPos:   5,
				Bold:     &boldTrue,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(imageBlock, nil)
				mockRepo.EXPECT().GetBlockType(gomock.Any(), imageBlock.BlockTypeID).Return(imageBlockType, nil)
				// No UpdateBlockFormatting call expected
			},
			wantErr: true,
		},
		{
			name: "error - invalid range",
			formattingRange: models.FormattingRange{
				StartPos: 100,
				EndPos:   50,
				Bold:     &boldTrue,
			},
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(textBlock, nil)
				mockRepo.EXPECT().GetBlockType(gomock.Any(), textBlock.BlockTypeID).Return(textBlockType, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.UpdateBlockFormatting(context.Background(), blockID, noteID, userID, tt.formattingRange)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
		})
	}
}

func TestMoveBlock(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	block := &models.Block{
		ID:       blockID,
		NoteID:   noteID,
		Position: 2,
	}

	blocks := []models.Block{
		{Position: 0}, {Position: 1}, {Position: 2}, {Position: 3},
	}

	tests := []struct {
		name        string
		newPosition int
		setupMock   func()
		wantErr     bool
	}{
		{
			name:        "success - move up",
			newPosition: 0,
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
				mockRepo.EXPECT().MoveBlock(gomock.Any(), noteID, blockID, 2, 0).Return(block, nil)
			},
			wantErr: false,
		},
		{
			name:        "success - move down",
			newPosition: 3,
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
				mockRepo.EXPECT().MoveBlock(gomock.Any(), noteID, blockID, 2, 3).Return(block, nil)
			},
			wantErr: false,
		},
		{
			name:        "same position - no move",
			newPosition: 2,
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				// No GetBlocks or MoveBlock calls expected
			},
			wantErr: false,
		},
		{
			name:        "invalid position",
			newPosition: 10,
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().GetBlocks(gomock.Any(), noteID).Return(blocks, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			result, err := usecase.MoveBlock(context.Background(), blockID, noteID, userID, tt.newPosition)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.newPosition != 2 {
					assert.Nil(t, result)
				}
			} else {
				assert.NoError(t, err)
				if tt.newPosition != 2 {
					assert.NotNil(t, result)
				}
			}
		})
	}
}

func TestDeleteBlock(t *testing.T) {
	usecase, mockRepo, ctrl := setupTestUsecase(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:       noteID,
		UserID:   userID,
		Title:    "Test Note",
		IsPublic: false,
	}

	block := &models.Block{
		ID:       blockID,
		NoteID:   noteID,
		Position: 2,
	}

	blockNoteID := noteID

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(block, nil)
				mockRepo.EXPECT().DeleteBlock(gomock.Any(), blockID).Return(&blockNoteID, nil)
				mockRepo.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, block.Position, -1).Return(nil)
			},
			wantErr: false,
		},
		{
			name: "note not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(nil, nil)
			},
			wantErr: true,
		},
		{
			name: "block not found",
			setupMock: func() {
				mockRepo.EXPECT().GetNote(gomock.Any(), noteID).Return(note, nil)
				mockRepo.EXPECT().GetBlock(gomock.Any(), blockID).Return(nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := usecase.DeleteBlock(context.Background(), blockID, noteID, userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
