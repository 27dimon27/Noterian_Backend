package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

func setupTestHandler(t *testing.T) (*NoteHandler, *mocks.MockNoteUsecase, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)
	return handler, mockUsecase, ctrl
}

func createContextWithUserID(userID uuid.UUID) context.Context {
	ctx := context.Background()
	return context.WithValue(ctx, types.UserIDKey, userID)
}

func TestGetNotes_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	expectedNotes := []models.Note{
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Note 1",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:        uuid.New(),
			UserID:    userID,
			Title:     "Note 2",
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	mockUsecase.EXPECT().
		GetNotes(gomock.Any(), userID).
		Return(expectedNotes, nil)

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	req = req.WithContext(createContextWithUserID(userID))
	w := httptest.NewRecorder()

	handler.GetNotes(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.NotesResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, 2, response.Total)
	assert.Equal(t, expectedNotes[0].Title, response.Notes[0].Title)
}

func TestGetNotes_Unauthorized(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	w := httptest.NewRecorder()

	handler.GetNotes(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestGetNotes_InternalError(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()

	mockUsecase.EXPECT().
		GetNotes(gomock.Any(), userID).
		Return(nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/notes", nil)
	req = req.WithContext(createContextWithUserID(userID))
	w := httptest.NewRecorder()

	handler.GetNotes(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestGetNote_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Test Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	blocks := []models.Block{
		{
			ID:          uuid.New(),
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "Hello",
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}
	formattings := make(map[string]models.BlockFormatting)

	mockUsecase.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(note, nil)
	mockUsecase.EXPECT().
		GetBlocksWithFormatting(gomock.Any(), noteID).
		Return(blocks, formattings, nil)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.GetNote(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetNote_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockUsecase.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.GetNote(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetNote_Forbidden(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockUsecase.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(nil, notes.ErrForbidden)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.GetNote(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateNote_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteReq := dto.NoteRequest{
		Title: "New Note",
	}
	body, _ := json.Marshal(noteReq)

	createdNote := &models.Note{
		ID:        uuid.New(),
		UserID:    userID,
		Title:     "New Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUsecase.EXPECT().
		CreateNote(gomock.Any(), gomock.Any()).
		Return(createdNote, nil)

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	w := httptest.NewRecorder()

	handler.CreateNote(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestCreateNote_InvalidData(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	body := []byte(`{"title": ""}`)

	mockUsecase.EXPECT().
		CreateNote(gomock.Any(), gomock.Any()).
		Return(nil, notes.ErrInvalidNoteData)

	req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	w := httptest.NewRecorder()

	handler.CreateNote(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateNote_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	noteReq := dto.NoteRequest{
		Title: "Updated Note",
	}
	body, _ := json.Marshal(noteReq)

	updatedNote := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Updated Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUsecase.EXPECT().
		UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
		Return(updatedNote, nil)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.UpdateNote(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteNote_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockUsecase.EXPECT().
		DeleteNote(gomock.Any(), noteID, userID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.DeleteNote(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestCreateBlock_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockReq := dto.BlockRequest{
		BlockTypeID: 1,
		Position:    0,
	}
	body, _ := json.Marshal(blockReq)

	createdBlock := &models.Block{
		ID:          uuid.New(),
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUsecase.EXPECT().
		CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
		Return(createdBlock, nil)

	req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.CreateBlock(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestUpdateBlockContent_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	updateReq := dto.UpdateBlockContentRequest{
		Content: "New content",
	}
	body, _ := json.Marshal(updateReq)

	updatedBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "New content",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUsecase.EXPECT().
		UpdateBlockContent(gomock.Any(), blockID, noteID, userID, "New content").
		Return(updatedBlock, nil)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.UpdateBlockContent(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestMoveBlock_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	moveReq := dto.MoveBlockRequest{
		NewPosition: 2,
	}
	body, _ := json.Marshal(moveReq)

	movedBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "Content",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mockUsecase.EXPECT().
		MoveBlock(gomock.Any(), blockID, noteID, userID, 2).
		Return(movedBlock, nil)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/move", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.MoveBlock(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDeleteBlock_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockUsecase.EXPECT().
		DeleteBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.DeleteBlock(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestUpdateBlockFormatting_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	bold := true
	formattingReq := dto.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}
	body, _ := json.Marshal(formattingReq)

	updatedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{
				StartPos: 0,
				EndPos:   5,
				Bold:     &bold,
			},
		},
	}

	mockUsecase.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).
		Return(updatedFormatting, nil)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.UpdateBlockFormatting(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestResetBlockFormatting_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	resetFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	mockUsecase.EXPECT().
		ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).
		Return(resetFormatting, nil)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.ResetBlockFormatting(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestGetNote_InvalidNoteID(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
	}{
		{"empty note id", "", http.StatusBadRequest},
		{"invalid uuid", "invalid-uuid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.noteID, nil)
			req = req.WithContext(createContextWithUserID(userID))
			req.SetPathValue("noteId", tt.noteID)
			w := httptest.NewRecorder()

			handler.GetNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetNote_GetBlocksWithFormattingError(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	note := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Test Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mockUsecase.EXPECT().
		GetNote(gomock.Any(), noteID, userID).
		Return(note, nil)
	mockUsecase.EXPECT().
		GetBlocksWithFormatting(gomock.Any(), noteID).
		Return(nil, nil, errors.New("database error"))

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.GetNote(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestUpdateNote_EmptyBody(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.UpdateNote(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateNote_InvalidNoteID(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	body := []byte(`{"title":"Updated"}`)

	tests := []struct {
		name       string
		noteID     string
		wantStatus int
	}{
		{"empty note id", "", http.StatusBadRequest},
		{"invalid uuid", "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.noteID, bytes.NewBuffer(body))
			req = req.WithContext(createContextWithUserID(userID))
			req.SetPathValue("noteId", tt.noteID)
			w := httptest.NewRecorder()

			handler.UpdateNote(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestUpdateNote_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	noteReq := dto.NoteRequest{Title: "Updated Note"}
	body, _ := json.Marshal(noteReq)

	mockUsecase.EXPECT().
		UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
		Return(nil, notes.ErrNoteNotFound)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.UpdateNote(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateNote_Forbidden(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	noteReq := dto.NoteRequest{Title: "Updated Note"}
	body, _ := json.Marshal(noteReq)

	mockUsecase.EXPECT().
		UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
		Return(nil, notes.ErrForbidden)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.UpdateNote(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestDeleteNote_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockUsecase.EXPECT().
		DeleteNote(gomock.Any(), noteID, userID).
		Return(notes.ErrNoteNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.DeleteNote(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestDeleteNote_Forbidden(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	mockUsecase.EXPECT().
		DeleteNote(gomock.Any(), noteID, userID).
		Return(notes.ErrForbidden)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.DeleteNote(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestGetBlock_InvalidIDs(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()

	tests := []struct {
		name       string
		noteID     string
		blockID    string
		wantStatus int
	}{
		{"empty note id", "", "block-id", http.StatusBadRequest},
		{"invalid note id", "invalid", "block-id", http.StatusBadRequest},
		{"empty block id", uuid.New().String(), "", http.StatusBadRequest},
		{"invalid block id", uuid.New().String(), "invalid", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.noteID+"/blocks/"+tt.blockID, nil)
			req = req.WithContext(createContextWithUserID(userID))
			req.SetPathValue("noteId", tt.noteID)
			req.SetPathValue("blockId", tt.blockID)
			w := httptest.NewRecorder()

			handler.GetBlock(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)
		})
	}
}

func TestGetBlock_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockUsecase.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil, notes.ErrBlockNotFound)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.GetBlock(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBlock_Forbidden(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockUsecase.EXPECT().
		GetBlock(gomock.Any(), blockID, noteID, userID).
		Return(nil, notes.ErrForbidden)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.GetBlock(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestCreateBlock_EmptyBody(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks", nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.CreateBlock(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestCreateBlock_NoteNotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockReq := dto.BlockRequest{BlockTypeID: 1, Position: 0}
	body, _ := json.Marshal(blockReq)

	mockUsecase.EXPECT().
		CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
		Return(nil, notes.ErrNoteNotFound)

	req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	w := httptest.NewRecorder()

	handler.CreateBlock(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestUpdateBlockContent_EmptyBody(t *testing.T) {
	handler, _, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.UpdateBlockContent(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestMoveBlock_InvalidPosition(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	moveReq := dto.MoveBlockRequest{NewPosition: 10}
	body, _ := json.Marshal(moveReq)

	mockUsecase.EXPECT().
		MoveBlock(gomock.Any(), blockID, noteID, userID, 10).
		Return(nil, notes.ErrInvalidPosition)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/move", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.MoveBlock(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBlockFormatting_InvalidFormatting(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	formattingReq := dto.FormattingRange{StartPos: -1, EndPos: 5}
	body, _ := json.Marshal(formattingReq)

	mockUsecase.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).
		Return(nil, notes.ErrInvalidFormattingRange)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.UpdateBlockFormatting(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestUpdateBlockFormatting_FormattingNotSupported(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	bold := true
	formattingReq := dto.FormattingRange{StartPos: 0, EndPos: 5, Bold: &bold}
	body, _ := json.Marshal(formattingReq)

	mockUsecase.EXPECT().
		UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).
		Return(nil, notes.ErrFormattingNotSupported)

	req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", bytes.NewBuffer(body))
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.UpdateBlockFormatting(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestResetBlockFormatting_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockUsecase.EXPECT().
		ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).
		Return(nil, notes.ErrBlockNotFound)

	req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.ResetBlockFormatting(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestGetBlockFormatting_Success(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	bold := true
	expectedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{StartPos: 0, EndPos: 5, Bold: &bold},
		},
	}

	mockUsecase.EXPECT().
		GetBlockFormatting(gomock.Any(), blockID, noteID, userID).
		Return(expectedFormatting, nil)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.GetBlockFormatting(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response dto.BlockFormatting
	err := json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, blockID.String(), response.BlockID)
}

func TestGetBlockFormatting_NotFound(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	mockUsecase.EXPECT().
		GetBlockFormatting(gomock.Any(), blockID, noteID, userID).
		Return(nil, notes.ErrBlockNotFound)

	req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", nil)
	req = req.WithContext(createContextWithUserID(userID))
	req.SetPathValue("noteId", noteID.String())
	req.SetPathValue("blockId", blockID.String())
	w := httptest.NewRecorder()

	handler.GetBlockFormatting(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}
