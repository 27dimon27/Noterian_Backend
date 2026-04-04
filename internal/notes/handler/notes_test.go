package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/dto"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes/handler/mocks"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"go.uber.org/mock/gomock"
)

func TestNoteHandler_GetNotes(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		expectedNotes := []models.Note{
			{ID: uuid.New(), UserID: userID, Title: "Note 1"},
			{ID: uuid.New(), UserID: userID, Title: "Note 2"},
		}

		mockUsecase.EXPECT().
			GetNotes(gomock.Any(), userID).
			Return(expectedNotes, nil)

		req := makeTestRequest("GET", "/notes", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.NotesResponse
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.Total != 2 {
			t.Errorf("expected 2 notes, got %d", len(response.Notes))
		}
		if response.Notes[0].Title != "Note 1" || response.Notes[1].Title != "Note 2" {
			t.Errorf("expected titles \"Note 1\" and \"Note 2\", got \"%s\" and \"%s\"", response.Notes[0].Title, response.Notes[1].Title)
		}
	})

	t.Run("unauthorized, no userID", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes", nil, nil, nil)
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetNotes(gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("GET", "/notes", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.GetNotes(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_CreateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	ctxWithAuth := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		noteReq := dto.NoteRequest{
			Title: "New Note",
		}

		newNoteID := uuid.New()
		expectedFromUsecase := &models.Note{
			ID:     newNoteID,
			UserID: userID,
			Title:  "New Note",
		}

		mockUsecase.EXPECT().
			CreateNote(gomock.Any(), gomock.Any()).
			Return(expectedFromUsecase, nil)

		req := makeTestRequest("POST", "/notes", noteReq, ctxWithAuth, nil)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected 201, got %d", w.Code)
		}

		var response dto.Note
		json.Unmarshal(w.Body.Bytes(), &response)

		if response.ID != newNoteID {
			t.Errorf("expected id %s, got %s", newNoteID, response.ID)
		}
		if response.Title != noteReq.Title {
			t.Errorf("expected title %s, got %s", noteReq.Title, response.Title)
		}
	})

	t.Run("empty body error", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes", nil, ctxWithAuth, nil)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for nil body, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID in context", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes", dto.NoteRequest{Title: "Note"}, nil, nil)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/notes", bytes.NewReader([]byte("{invalid json}")))
		req = req.WithContext(ctxWithAuth)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for bad json, got %d", w.Code)
		}
	})

	t.Run("usecase validation error", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateNote(gomock.Any(), gomock.Any()).
			Return(nil, notes.ErrInvalidNoteData)

		req := makeTestRequest("POST", "/notes", dto.NoteRequest{Title: ""}, ctxWithAuth, nil)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400 for validation error, got %d", w.Code)
		}
	})

	t.Run("internal server error", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateNote(gomock.Any(), gomock.Any()).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("POST", "/notes", dto.NoteRequest{Title: "Title"}, ctxWithAuth, nil)
		w := httptest.NewRecorder()

		handler.CreateNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_GetNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Test Note",
		}
		expectedBlocks := []models.Block{
			{
				ID:          uuid.New(),
				NoteID:      noteID,
				BlockTypeID: 1,
				Position:    0,
				Content:     "Hello world",
			},
			{
				ID:          uuid.New(),
				NoteID:      noteID,
				BlockTypeID: 2,
				Position:    1,
				Content:     "https://example.com",
			},
		}

		mockUsecase.EXPECT().
			GetNote(gomock.Any(), noteID, userID).
			Return(note, nil)

		mockUsecase.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(expectedBlocks, nil)

		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response dto.NoteResponse
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}
		if response.Note.ID != noteID {
			t.Errorf("expected note ID %v, got %v", noteID, response.Note.ID)
		}
		if len(response.Blocks) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(response.Blocks))
		}
		if response.Blocks[0].Content != expectedBlocks[0].Content {
			t.Errorf("block 0: expected content %s, got %s", expectedBlocks[0].Content, response.Blocks[0].Content)
		}
		if response.Blocks[1].BlockTypeID != expectedBlocks[1].BlockTypeID {
			t.Errorf("block 1: expected type %d, got %d", expectedBlocks[1].BlockTypeID, response.Blocks[1].BlockTypeID)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/invalid", nil, ctx, map[string]string{
			"noteId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized no userID", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, nil, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetNote(gomock.Any(), noteID, userID).
			Return(nil, notes.ErrForbidden)

		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetNote(gomock.Any(), noteID, userID).
			Return(nil, nil)

		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("internal error on GetNote", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetNote(gomock.Any(), noteID, userID).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})

	t.Run("internal error on GetBlocks", func(t *testing.T) {
		note := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  "Test Note",
		}

		mockUsecase.EXPECT().
			GetNote(gomock.Any(), noteID, userID).
			Return(note, nil)
		mockUsecase.EXPECT().
			GetBlocks(gomock.Any(), noteID).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("GET", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_UpdateNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)
	updateReq := dto.NoteRequest{
		Title: "Updated Title",
	}

	t.Run("success", func(t *testing.T) {
		updatedNote := &models.Note{
			ID:     noteID,
			UserID: userID,
			Title:  updateReq.Title,
		}

		mockUsecase.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
			Return(updatedNote, nil)

		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response dto.Note
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.Title != updateReq.Title {
			t.Errorf("expected title %s, got %s", updateReq.Title, response.Title)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/invalid", nil, ctx, map[string]string{
			"noteId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized no userID", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, nil, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
			Return(nil, notes.ErrNoteNotFound)

		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("invalid note data", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
			Return(nil, notes.ErrInvalidNoteData)

		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
			Return(nil, notes.ErrForbidden)

		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("PUT", "/notes/"+noteID.String(), updateReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_DeleteNote(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteNote(gomock.Any(), noteID, userID).
			Return(nil)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes/", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes/invalid", nil, ctx, map[string]string{
			"noteId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized, no userID", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes/"+noteID.String(), nil, nil, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteNote(gomock.Any(), noteID, userID).
			Return(notes.ErrNoteNotFound)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteNote(gomock.Any(), noteID, userID).
			Return(notes.ErrForbidden)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteNote(gomock.Any(), noteID, userID).
			Return(errors.New("db error"))

		req := makeTestRequest("DELETE", "/notes/"+noteID.String(), nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteNote(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func makeTestRequest(method, url string, body interface{}, ctx context.Context, pathVars map[string]string) *http.Request {
	var bodyReader io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		bodyReader = bytes.NewReader(b)
	}

	req := httptest.NewRequest(method, url, bodyReader)
	req.Header.Set("Content-Type", "application/json")

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	for key, value := range pathVars {
		req.SetPathValue(key, value)
	}

	return req
}
func TestNoteHandler_CreateBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	blockReq2 := dto.BlockRequest{
		BlockTypeID: 1,
		Position:    0,
	}

	t.Run("success", func(t *testing.T) {
		blockReq := dto.BlockRequest{
			BlockTypeID: 1,
			Position:    0,
			Content:     "haha",
		}

		createdBlock := &models.Block{
			ID:          blockID,
			NoteID:      noteID,
			BlockTypeID: blockReq.BlockTypeID,
			Position:    blockReq.Position,
			Content:     blockReq.Content,
		}

		mockUsecase.EXPECT().
			CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
			Return(createdBlock, nil)

		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response dto.Block
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.ID != blockID {
			t.Errorf("expected block ID %v, got %v", blockID, response.ID)
		}
		if response.BlockTypeID != blockReq.BlockTypeID {
			t.Errorf("expected type %d, got %d", blockReq.BlockTypeID, response.BlockTypeID)
		}
		if response.Content != blockReq.Content {
			t.Errorf("expected content %s, got %s", blockReq.Content, response.Content)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes//blocks", nil, ctx, nil)
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes/invalid/blocks", nil, ctx, map[string]string{
			"noteId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq2, nil, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
			Return(nil, notes.ErrNoteNotFound)

		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq2, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d. Body: %s", w.Code, w.Body.String())
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
			Return(nil, notes.ErrForbidden)

		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq2, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("invalid block type", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
			Return(nil, notes.ErrInvalidBlockType)

		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq2, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("POST", "/notes/"+noteID.String()+"/blocks", blockReq2, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.CreateBlock(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_GetBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)

	t.Run("success", func(t *testing.T) {
		block := &models.Block{
			ID:          blockID,
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "Content",
		}

		mockUsecase.EXPECT().
			GetBlock(gomock.Any(), blockID, noteID, userID).
			Return(block, nil)

		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}

		var response dto.Block
		if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if response.ID != blockID {
			t.Errorf("expected block ID %v, got %v", blockID, response.ID)
		}
		if response.Content != "Content" {
			t.Errorf("expected content 'Content', got '%s'", response.Content)
		}
		if response.BlockTypeID != 1 {
			t.Errorf("expected type ID 1, got %d", response.BlockTypeID)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes//blocks/"+blockID.String(), nil, ctx, map[string]string{
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("missing blockId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/", nil, ctx, map[string]string{
			"noteId": noteID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid noteId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/invalid/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  "invalid",
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/invalid", nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized, no userID", func(t *testing.T) {
		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, nil, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetBlock(gomock.Any(), blockID, noteID, userID).
			Return(nil, notes.ErrNoteNotFound)

		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetBlock(gomock.Any(), blockID, noteID, userID).
			Return(nil, notes.ErrBlockNotFound)

		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetBlock(gomock.Any(), blockID, noteID, userID).
			Return(nil, notes.ErrForbidden)

		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			GetBlock(gomock.Any(), blockID, noteID, userID).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("GET", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.GetBlock(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_UpdateBlockContent(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)
	
	updateReq := dto.UpdateBlockContentRequest{Content: "New Content"}
	pathVars := map[string]string{
		"noteId":  noteID.String(),
		"blockId": blockID.String(),
	}

	t.Run("success", func(t *testing.T) {
		updatedBlock := &models.Block{
			ID:      blockID,
			NoteID:  noteID,
			Content: "New Content",
		}

		mockUsecase.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, noteID, userID, updateReq.Content).
			Return(updatedBlock, nil)

		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var response dto.Block
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Content != updateReq.Content {
			t.Errorf("expected content %s, got %s", updateReq.Content, response.Content)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes//blocks/"+blockID.String(), updateReq, ctx, map[string]string{
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, nil, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, noteID, userID, updateReq.Content).
			Return(nil, notes.ErrNoteNotFound)

		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, noteID, userID, updateReq.Content).
			Return(nil, notes.ErrBlockNotFound)

		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, noteID, userID, updateReq.Content).
			Return(nil, notes.ErrForbidden)

		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			UpdateBlockContent(gomock.Any(), blockID, noteID, userID, updateReq.Content).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("PUT", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), updateReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.UpdateBlockContent(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_MoveBlock(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)
	
	moveURL := "/notes/" + noteID.String() + "/blocks/" + blockID.String() + "/move"
	pathVars := map[string]string{
		"noteId":  noteID.String(),
		"blockId": blockID.String(),
	}

	t.Run("success", func(t *testing.T) {
		moveReq := dto.MoveBlockRequest{NewPosition: 1}
		movedBlock := &models.Block{
			ID:       blockID,
			Position: 1,
		}

		mockUsecase.EXPECT().
			MoveBlock(gomock.Any(), blockID, noteID, userID, moveReq.NewPosition).
			Return(movedBlock, nil)

		req := makeTestRequest("PUT", moveURL, moveReq, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}

		var response dto.Block
		json.Unmarshal(w.Body.Bytes(), &response)
		if response.Position != moveReq.NewPosition {
			t.Errorf("expected position %d, got %d", moveReq.NewPosition, response.Position)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := makeTestRequest("PUT", moveURL, nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := makeTestRequest("PUT", moveURL, dto.MoveBlockRequest{NewPosition: 1}, nil, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			MoveBlock(gomock.Any(), blockID, noteID, userID, 1).
			Return(nil, notes.ErrNoteNotFound)

		req := makeTestRequest("PUT", moveURL, dto.MoveBlockRequest{NewPosition: 1}, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", w.Code)
		}
	})

	t.Run("invalid position", func(t *testing.T) {
		mockUsecase.EXPECT().
			MoveBlock(gomock.Any(), blockID, noteID, userID, 5).
			Return(nil, notes.ErrInvalidPosition)

		req := makeTestRequest("PUT", moveURL, dto.MoveBlockRequest{NewPosition: 5}, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			MoveBlock(gomock.Any(), blockID, noteID, userID, 1).
			Return(nil, errors.New("db error"))

		req := makeTestRequest("PUT", moveURL, dto.MoveBlockRequest{NewPosition: 1}, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.MoveBlock(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", w.Code)
		}
	})
}

func TestNoteHandler_DeleteBlock_Refactored(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()
	ctx := context.WithValue(context.Background(), types.UserIDKey, userID)
	
	pathVars := map[string]string{
		"noteId":  noteID.String(),
		"blockId": blockID.String(),
	}

	t.Run("success", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteBlock(gomock.Any(), blockID, noteID, userID).
			Return(nil)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status 204, got %d", w.Code)
		}
	})

	t.Run("missing noteId", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes//blocks/"+blockID.String(), nil, ctx, map[string]string{
			"blockId": blockID.String(),
		})
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("invalid blockId", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/invalid", nil, ctx, map[string]string{
			"noteId":  noteID.String(),
			"blockId": "invalid",
		})
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", w.Code)
		}
	})

	t.Run("unauthorized - no userID", func(t *testing.T) {
		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, nil, pathVars)
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status 401, got %d", w.Code)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteBlock(gomock.Any(), blockID, noteID, userID).
			Return(notes.ErrNoteNotFound)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", w.Code)
		}
	})

	t.Run("forbidden", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteBlock(gomock.Any(), blockID, noteID, userID).
			Return(notes.ErrForbidden)

		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status 403, got %d", w.Code)
		}
	})

	t.Run("internal error", func(t *testing.T) {
		mockUsecase.EXPECT().
			DeleteBlock(gomock.Any(), blockID, noteID, userID).
			Return(errors.New("db error"))

		req := makeTestRequest("DELETE", "/notes/"+noteID.String()+"/blocks/"+blockID.String(), nil, ctx, pathVars)
		w := httptest.NewRecorder()

		handler.DeleteBlock(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", w.Code)
		}
	})
}
