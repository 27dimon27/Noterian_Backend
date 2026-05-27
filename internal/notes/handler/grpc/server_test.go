// handler/grpc/server_test.go
package grpc

import (
	"context"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/grpc/mocks"
	notesgrpc "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/proto/notes/grpc/gen"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func TestServer_GetNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	noteID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:        noteID,
			UserID:    userID,
			Title:     "Test Note",
			IsPublic:  true,
			Icon:      "📝",
			HeaderURL: "http://example.com/header.jpg",
		}

		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil)

		req := &notesgrpc.GetNoteRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		}

		resp, err := server.GetNote(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, noteID.String(), resp.Id)
		assert.Equal(t, userID.String(), resp.UserId)
		assert.Equal(t, "Test Note", resp.Title)
		assert.True(t, resp.IsPublic)
		assert.Equal(t, "📝", resp.Icon)
	})

	t.Run("with parent ID", func(t *testing.T) {
		parentID := uuid.New()
		note := &models.Note{
			ID:       noteID,
			UserID:   userID,
			Title:    "Subnote",
			ParentID: &parentID,
		}

		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil, nil, nil)

		req := &notesgrpc.GetNoteRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		}

		resp, err := server.GetNote(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp.ParentId)
		assert.Equal(t, parentID.String(), *resp.ParentId)
	})

	t.Run("invalid note ID", func(t *testing.T) {
		req := &notesgrpc.GetNoteRequest{
			NoteId: "invalid-uuid",
			UserId: userID.String(),
		}

		resp, err := server.GetNote(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	t.Run("invalid user ID", func(t *testing.T) {
		req := &notesgrpc.GetNoteRequest{
			NoteId: noteID.String(),
			UserId: "invalid-uuid",
		}

		resp, err := server.GetNote(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	blockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		block := &models.Block{
			ID:          blockID,
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    2,
			Content:     "Block content",
		}

		mockUsecase.EXPECT().GetBlock(gomock.Any(), blockID, noteID, userID).Return(block, nil)

		req := &notesgrpc.GetBlockRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		}

		resp, err := server.GetBlock(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, blockID.String(), resp.Id)
		assert.Equal(t, noteID.String(), resp.NoteId)
		assert.Equal(t, int32(1), resp.BlockTypeId)
		assert.Equal(t, int32(2), resp.Position)
		assert.Equal(t, "Block content", resp.Content)
	})

	t.Run("invalid block ID", func(t *testing.T) {
		req := &notesgrpc.GetBlockRequest{
			BlockId: "invalid",
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		}

		resp, err := server.GetBlock(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_GetBlocks(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	noteID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID}
		blocks := []models.Block{
			{ID: uuid.New(), NoteID: noteID, BlockTypeID: 1, Position: 0, Content: "Block 1"},
			{ID: uuid.New(), NoteID: noteID, BlockTypeID: 2, Position: 1, Content: "image.jpg"},
		}

		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, blocks, nil, nil)

		req := &notesgrpc.GetBlocksRequest{
			NoteId: noteID.String(),
			UserId: userID.String(),
		}

		resp, err := server.GetBlocks(context.Background(), req)

		require.NoError(t, err)
		assert.Len(t, resp.Blocks, 2)
	})

	t.Run("invalid note ID", func(t *testing.T) {
		req := &notesgrpc.GetBlocksRequest{
			NoteId: "invalid",
			UserId: userID.String(),
		}

		resp, err := server.GetBlocks(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	userID := uuid.New()

	t.Run("missing block", func(t *testing.T) {
		req := &notesgrpc.CreateBlockRequest{
			UserId: userID.String(),
			Block:  nil,
		}

		resp, err := server.CreateBlock(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
		assert.Equal(t, "block payload is required", err.Error())
	})
}

func TestServer_ShiftBlockPositions(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().ShiftBlockPositions(gomock.Any(), noteID, 2, 1).Return(nil)

		req := &notesgrpc.ShiftBlockPositionsRequest{
			NoteId:       noteID.String(),
			FromPosition: 2,
			Direction:    1,
		}

		resp, err := server.ShiftBlockPositions(context.Background(), req)

		require.NoError(t, err)
		assert.NotNil(t, resp)
	})

	t.Run("invalid note ID", func(t *testing.T) {
		req := &notesgrpc.ShiftBlockPositionsRequest{
			NoteId:       "invalid",
			FromPosition: 2,
			Direction:    1,
		}

		resp, err := server.ShiftBlockPositions(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func TestServer_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	server := NewServer(mockUsecase)

	blockID := uuid.New()
	noteID := uuid.New()
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(nil)

		req := &notesgrpc.DeleteBlockRequest{
			BlockId: blockID.String(),
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		}

		resp, err := server.DeleteBlock(context.Background(), req)

		require.NoError(t, err)
		assert.Equal(t, noteID.String(), resp.NoteId)
	})

	t.Run("invalid block ID", func(t *testing.T) {
		req := &notesgrpc.DeleteBlockRequest{
			BlockId: "invalid",
			NoteId:  noteID.String(),
			UserId:  userID.String(),
		}

		resp, err := server.DeleteBlock(context.Background(), req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}
