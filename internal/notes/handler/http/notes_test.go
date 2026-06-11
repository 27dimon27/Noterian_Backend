package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/logger"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/http/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var log = logger.Init()

func TestNoteHandler_GetNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		expectedNotes := []models.Note{
			{ID: uuid.New(), UserID: userID, Title: "Note 1"},
			{ID: uuid.New(), UserID: userID, Title: "Note 2"},
		}

		mockUsecase.EXPECT().GetNotes(gomock.Any(), userID).Return(expectedNotes, nil)

		req := httptest.NewRequest(http.MethodGet, "/notes", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response dto.NotesResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Len(t, response.Notes, 2)
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notes", nil)
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error", func(t *testing.T) {
		mockUsecase.EXPECT().GetNotes(gomock.Any(), userID).Return(nil, errors.New("db error"))

		req := httptest.NewRequest(http.MethodGet, "/notes", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		assert.Equal(t, http.StatusInternalServerError, w.Code)
	})
}

func TestNoteHandler_GetNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, UserID: userID, Title: "Test Note"}
		blocks := []models.Block{{ID: uuid.New(), Content: "Block"}}
		formattings := map[string]models.BlockFormatting{}

		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, blocks, formattings, nil)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid note ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notes/invalid", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", "invalid")
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("missing note ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/notes/", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(nil, nil, nil, notes.ErrNoteNotFound)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(nil, nil, nil, notes.ErrForbidden)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestNoteHandler_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		noteRequest := dto.NoteRequest{
			Title:      "New Note",
			IsPublic:   false,
			IsFavorite: false,
			Icon:       "📝",
		}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		createdNote := &models.Note{ID: uuid.New(), UserID: userID, Title: "New Note"}

		mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)

		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader([]byte("{invalid json}")))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("unauthorized", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: "New Note"}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("usecase error - invalid data", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: ""}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, notes.ErrInvalidNoteData)

		req := httptest.NewRequest(http.MethodPost, "/notes", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		noteRequest := dto.NoteRequest{
			Title:      "Updated Note",
			IsPublic:   true,
			IsFavorite: true,
			Icon:       "⭐",
		}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		updatedNote := &models.Note{ID: noteID, UserID: userID, Title: "Updated Note"}

		mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).Return(updatedNote, nil)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("note not found", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: "Updated"}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, notes.ErrNoteNotFound)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String(), bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(notes.ErrForbidden)

		req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestNoteHandler_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		blockRequest := dto.BlockRequest{
			BlockTypeID: 1,
			Position:    0,
			Content:     "Hello",
		}
		body, err := json.Marshal(blockRequest)
		require.NoError(t, err)

		createdBlock := &models.Block{ID: uuid.New(), NoteID: noteID, BlockTypeID: 1, Position: 0}

		mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(createdBlock, nil)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		assert.Equal(t, http.StatusCreated, w.Code)
	})

	t.Run("invalid block type", func(t *testing.T) {
		blockRequest := dto.BlockRequest{BlockTypeID: 0, Position: 0}
		body, err := json.Marshal(blockRequest)
		require.NoError(t, err)

		mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidBlockType)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/blocks", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNoteHandler_UpdateBlockContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		updateRequest := dto.UpdateBlockContentRequest{Content: "New content"}
		body, err := json.Marshal(updateRequest)
		require.NoError(t, err)

		updatedBlock := &models.Block{ID: blockID, Content: "New content"}

		mockUsecase.EXPECT().UpdateBlockContent(gomock.Any(), blockID, noteID, userID, "New content").Return(updatedBlock, nil)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/content", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestNoteHandler_MoveBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		moveRequest := dto.MoveBlockRequest{NewPosition: 5}
		body, err := json.Marshal(moveRequest)
		require.NoError(t, err)

		movedBlock := &models.Block{ID: blockID, Position: 5}

		mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(movedBlock, nil)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/move", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid position", func(t *testing.T) {
		moveRequest := dto.MoveBlockRequest{NewPosition: -1}
		body, err := json.Marshal(moveRequest)
		require.NoError(t, err)

		mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, -1).Return(nil, notes.ErrInvalidPosition)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/move", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNoteHandler_DeleteBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestNoteHandler_UpdateBlockFormatting(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		formattingRequest := dto.FormattingRange{
			StartPos: 0,
			EndPos:   5,
			Bold:     boolPtr(true),
		}
		body, err := json.Marshal(formattingRequest)
		require.NoError(t, err)

		updatedFormatting := &models.BlockFormatting{BlockID: blockID.String()}

		mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(updatedFormatting, nil)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.UpdateBlockFormatting(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid formatting range", func(t *testing.T) {
		formattingRequest := dto.FormattingRange{StartPos: 10, EndPos: 5}
		body, err := json.Marshal(formattingRequest)
		require.NoError(t, err)

		mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidFormattingRange)

		req := httptest.NewRequest(http.MethodPut, "/notes/"+noteID.String()+"/blocks/"+blockID.String()+"/formatting", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("blockId", blockID.String())
		w := httptest.NewRecorder()

		handler.UpdateBlockFormatting(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNoteHandler_GetSubnotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		subnotes := []models.Note{
			{ID: uuid.New(), Title: "Subnote 1"},
			{ID: uuid.New(), Title: "Subnote 2"},
		}

		mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return(subnotes, nil)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/subnote", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetSubnotes(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestNoteHandler_CreateSubnote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success with position", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: "New Subnote"}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		createdNote := &models.Note{ID: uuid.New(), Title: "New Subnote"}
		blockID := uuid.New()

		mockUsecase.EXPECT().CreateSubnote(gomock.Any(), noteID, userID, gomock.Any(), true, 2).Return(createdNote, blockID, nil)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/subnote?position=2", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.CreateSubnote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("success without position", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: "New Subnote"}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		createdNote := &models.Note{ID: uuid.New(), Title: "New Subnote"}
		blockID := uuid.New()

		mockUsecase.EXPECT().CreateSubnote(gomock.Any(), noteID, userID, gomock.Any(), false, 0).Return(createdNote, blockID, nil)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/subnote", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.CreateSubnote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("invalid position query param", func(t *testing.T) {
		noteRequest := dto.NoteRequest{Title: "New Subnote"}
		body, err := json.Marshal(noteRequest)
		require.NoError(t, err)

		req := httptest.NewRequest(http.MethodPost, "/notes/"+noteID.String()+"/subnote?position=invalid", bytes.NewReader(body))
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.CreateSubnote(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestNoteHandler_DeleteSubnote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().DeleteSubnote(gomock.Any(), noteID, subnoteID, userID).Return(nil)

		req := httptest.NewRequest(http.MethodDelete, "/notes/"+noteID.String()+"/subnote/"+subnoteID.String(), nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		req.SetPathValue("subnoteId", subnoteID.String())
		w := httptest.NewRecorder()

		handler.DeleteSubnote(w, req)

		assert.Equal(t, http.StatusNoContent, w.Code)
	})
}

func TestNoteHandler_GetNotePDF(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	userID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		pdfBuffer := bytes.NewBufferString("PDF content")

		mockUsecase.EXPECT().GenerateNotePDF(gomock.Any(), noteID, userID).Return(pdfBuffer, nil)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/pdf", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetNotePDF(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		assert.Contains(t, w.Header().Get("Content-Disposition"), "attachment")
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().GenerateNotePDF(gomock.Any(), noteID, userID).Return(nil, notes.ErrForbidden)

		req := httptest.NewRequest(http.MethodGet, "/notes/"+noteID.String()+"/pdf", nil)
		req = req.WithContext(context.WithValue(req.Context(), types.UserIDKey, userID))
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetNotePDF(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}

func TestNoteHandler_GetPublicNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase, log)

	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := &models.Note{ID: noteID, Title: "Public Note", Icon: "🌐"}

		mockUsecase.EXPECT().GetPublicNote(gomock.Any(), noteID).Return(note, nil)

		req := httptest.NewRequest(http.MethodGet, "/public/notes/"+noteID.String(), nil)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetPublicNote(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().GetPublicNote(gomock.Any(), noteID).Return(nil, notes.ErrNoteNotFound)

		req := httptest.NewRequest(http.MethodGet, "/public/notes/"+noteID.String(), nil)
		req.SetPathValue("noteId", noteID.String())
		w := httptest.NewRecorder()

		handler.GetPublicNote(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}
