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
	"go.uber.org/mock/gomock"
)

func setupTestHandler(t *testing.T) (*NoteHandler, *mocks.MockNoteUsecase, *gomock.Controller) {
	ctrl := gomock.NewController(t)
	mockUsecase := mocks.NewMockNoteUsecase(ctrl)
	handler := NewNoteHandler(mockUsecase)
	return handler, mockUsecase, ctrl
}

func addUserIDToContext(r *http.Request, userID uuid.UUID) *http.Request {
	ctx := context.WithValue(r.Context(), types.UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestMoveBlock(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	moveRequest := dto.MoveBlockRequest{
		NewPosition: 5,
	}

	movedBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    5,
		Content:     "Test content",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		noteIDPath     string
		blockIDPath    string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(movedBlock, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			blockIDPath:    blockID.String(),
			body:           moveRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "",
			body:           moveRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			blockIDPath:    blockID.String(),
			body:           moveRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "invalid-uuid",
			body:           moveRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			body:           nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			body:           moveRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "note not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "block not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(nil, notes.ErrBlockNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "forbidden",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "invalid position",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(nil, notes.ErrInvalidPosition)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "internal error",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        moveRequest,
			setupMock: func() {
				mockUsecase.EXPECT().MoveBlock(gomock.Any(), blockID, noteID, userID, 5).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(jsonBody)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.noteIDPath+"/blocks/"+tt.blockIDPath+"/move", bodyReader)
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("blockId", tt.blockIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.MoveBlock(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDeleteBlock(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name           string
		noteIDPath     string
		blockIDPath    string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "note not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "block not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(notes.ErrBlockNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "forbidden",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "internal error",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteBlock(gomock.Any(), blockID, noteID, userID).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.noteIDPath+"/blocks/"+tt.blockIDPath, nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("blockId", tt.blockIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.DeleteBlock(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdateBlockFormatting(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	boldTrue := true
	textAlignCenter := 1

	formattingRequest := dto.FormattingRange{
		StartPos:  0,
		EndPos:    5,
		Bold:      &boldTrue,
		TextAlign: &textAlignCenter,
	}

	updatedFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{StartPos: 0, EndPos: 5, Bold: &boldTrue},
		},
	}

	tests := []struct {
		name           string
		noteIDPath     string
		blockIDPath    string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(updatedFormatting, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			blockIDPath:    blockID.String(),
			body:           formattingRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "",
			body:           formattingRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			blockIDPath:    blockID.String(),
			body:           formattingRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "invalid-uuid",
			body:           formattingRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			body:           nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			body:           formattingRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "note not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "block not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrBlockNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "forbidden",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "invalid block type",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidBlockType)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid formatting for image block",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidFormattingForImageBlock)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "formatting not supported",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrFormattingNotSupported)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "invalid formatting range",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidFormattingRange)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:        "internal error",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        formattingRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockFormatting(gomock.Any(), blockID, noteID, userID, gomock.Any()).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(jsonBody)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.noteIDPath+"/blocks/"+tt.blockIDPath+"/formatting", bodyReader)
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("blockId", tt.blockIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.UpdateBlockFormatting(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestResetBlockFormatting(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	resetFormatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	tests := []struct {
		name           string
		noteIDPath     string
		blockIDPath    string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).Return(resetFormatting, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:        "note not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "block not found",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).Return(nil, notes.ErrBlockNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:        "forbidden",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:        "internal error",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().ResetBlockFormatting(gomock.Any(), blockID, noteID, userID).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.noteIDPath+"/blocks/"+tt.blockIDPath+"/formatting", nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("blockId", tt.blockIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.ResetBlockFormatting(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateSubnote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	parentNoteID := uuid.New()
	subnoteID := uuid.New()

	noteRequest := dto.NoteRequest{
		Title:    "New Subnote",
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
		name           string
		noteIDPath     string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: parentNoteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateSubnote(gomock.Any(), parentNoteID, userID, gomock.Any()).Return(createdNote, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     parentNoteID.String(),
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "empty body",
			noteIDPath:     parentNoteID.String(),
			body:           nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "note not found",
			noteIDPath: parentNoteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateSubnote(gomock.Any(), parentNoteID, userID, gomock.Any()).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden",
			noteIDPath: parentNoteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateSubnote(gomock.Any(), parentNoteID, userID, gomock.Any()).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "internal error",
			noteIDPath: parentNoteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateSubnote(gomock.Any(), parentNoteID, userID, gomock.Any()).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(jsonBody)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req := httptest.NewRequest(http.MethodPost, "/notes/"+tt.noteIDPath+"/subnotes", bodyReader)
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.CreateSubnote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDeleteSubnote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID := uuid.New()

	tests := []struct {
		name           string
		noteIDPath     string
		subnoteIDPath  string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:          "success",
			noteIDPath:    noteID.String(),
			subnoteIDPath: subnoteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteSubnote(gomock.Any(), noteID, subnoteID, userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			subnoteIDPath:  subnoteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing subnote ID",
			noteIDPath:     noteID.String(),
			subnoteIDPath:  "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			subnoteIDPath:  subnoteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid subnote ID",
			noteIDPath:     noteID.String(),
			subnoteIDPath:  "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			subnoteIDPath:  subnoteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:          "note not found",
			noteIDPath:    noteID.String(),
			subnoteIDPath: subnoteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteSubnote(gomock.Any(), noteID, subnoteID, userID).Return(notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:          "forbidden",
			noteIDPath:    noteID.String(),
			subnoteIDPath: subnoteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteSubnote(gomock.Any(), noteID, subnoteID, userID).Return(notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:          "internal error",
			noteIDPath:    noteID.String(),
			subnoteIDPath: subnoteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteSubnote(gomock.Any(), noteID, subnoteID, userID).Return(errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.noteIDPath+"/subnotes/"+tt.subnoteIDPath, nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("subnoteId", tt.subnoteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.DeleteSubnote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetSubnotes(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID := uuid.New()
	parentID := noteID

	subnotes := []models.Note{
		{
			ID:        subnoteID,
			UserID:    userID,
			Title:     "Subnote 1",
			ParentID:  &parentID,
			IsPublic:  false,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	tests := []struct {
		name           string
		noteIDPath     string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return(subnotes, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "note not found",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "internal error",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return(nil, errors.New("internal error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:       "empty subnotes",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetSubnotes(gomock.Any(), noteID, userID).Return([]models.Note{}, nil)
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.noteIDPath+"/subnotes", nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.GetSubnotes(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetNotes(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	subnoteID := uuid.New()
	parentID := uuid.New()

	notes_list := []models.Note{
		{ID: noteID, UserID: userID, Title: "Note 1", IsPublic: false, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}
	subnotes := map[string][]models.Note{
		noteID.String(): {{ID: subnoteID, UserID: userID, Title: "Subnote 1", ParentID: &parentID, IsPublic: false, CreatedAt: time.Now(), UpdatedAt: time.Now()}},
	}

	tests := []struct {
		name           string
		setupMock      func()
		expectedStatus int
		expectedBody   interface{}
	}{
		{
			name: "success",
			setupMock: func() {
				mockUsecase.EXPECT().GetNotes(gomock.Any(), userID).Return(notes_list, subnotes, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "unauthorized - no userID",
			setupMock: func() {
				// no mock call expected
			},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "internal server error",
			setupMock: func() {
				mockUsecase.EXPECT().GetNotes(gomock.Any(), userID).Return(nil, nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/notes", nil)
			if tt.name != "unauthorized - no userID" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.GetNotes(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestGetNote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	note := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Test Note",
		IsPublic:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	blocks := []models.Block{
		{ID: blockID, NoteID: noteID, BlockTypeID: 1, Position: 0, Content: "Test content", CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	blockFormattings := map[string]models.BlockFormatting{
		blockID.String(): {BlockID: blockID.String(), Ranges: []models.FormattingRange{}},
	}

	tests := []struct {
		name           string
		noteIDPath     string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil)
				mockUsecase.EXPECT().GetBlocksWithFormatting(gomock.Any(), noteID).Return(blocks, blockFormattings, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid-uuid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "forbidden",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "note not found",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(nil, nil)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "blocks error",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().GetNote(gomock.Any(), noteID, userID).Return(note, nil)
				mockUsecase.EXPECT().GetBlocksWithFormatting(gomock.Any(), noteID).Return(nil, nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/notes/"+tt.noteIDPath, nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.GetNote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateNote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	noteRequest := dto.NoteRequest{
		Title:    "New Note",
		IsPublic: false,
	}

	createdNote := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "New Note",
		IsPublic:  false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name: "success",
			body: noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(createdNote, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "empty body",
			body:           nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid json",
			body:           "invalid json",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name: "invalid note data",
			body: noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, notes.ErrInvalidNoteData)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "internal error",
			body: noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateNote(gomock.Any(), gomock.Any()).Return(nil, errors.New("db error"))
			},
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(jsonBody)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req := httptest.NewRequest(http.MethodPost, "/notes", bodyReader)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.CreateNote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdateNote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	noteRequest := dto.NoteRequest{
		Title:    "Updated Note",
		IsPublic: true,
	}

	updatedNote := &models.Note{
		ID:        noteID,
		UserID:    userID,
		Title:     "Updated Note",
		IsPublic:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name           string
		noteIDPath     string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: noteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).Return(updatedNote, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid",
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			body:           noteRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "note not found",
			noteIDPath: noteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "invalid data",
			noteIDPath: noteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).Return(nil, notes.ErrInvalidNoteData)
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:       "forbidden",
			noteIDPath: noteID.String(),
			body:       noteRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateNote(gomock.Any(), noteID, gomock.Any(), userID).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.noteIDPath, bytes.NewReader(jsonBody))
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.UpdateNote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestDeleteNote(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()

	tests := []struct {
		name           string
		noteIDPath     string
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(nil)
			},
			expectedStatus: http.StatusNoContent,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid",
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "note not found",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden",
			noteIDPath: noteID.String(),
			setupMock: func() {
				mockUsecase.EXPECT().DeleteNote(gomock.Any(), noteID, userID).Return(notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodDelete, "/notes/"+tt.noteIDPath, nil)
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.DeleteNote(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestCreateBlock(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	blockRequest := dto.BlockRequest{
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
	}

	createdBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		noteIDPath     string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:       "success",
			noteIDPath: noteID.String(),
			body:       blockRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(createdBlock, nil)
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			body:           blockRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			noteIDPath:     noteID.String(),
			body:           nil,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			body:           blockRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "note not found",
			noteIDPath: noteID.String(),
			body:       blockRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, notes.ErrNoteNotFound)
			},
			expectedStatus: http.StatusNotFound,
		},
		{
			name:       "forbidden",
			noteIDPath: noteID.String(),
			body:       blockRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, notes.ErrForbidden)
			},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:       "invalid block type",
			noteIDPath: noteID.String(),
			body:       blockRequest,
			setupMock: func() {
				mockUsecase.EXPECT().CreateBlock(gomock.Any(), noteID, userID, gomock.Any()).Return(nil, notes.ErrInvalidBlockType)
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *bytes.Reader
			if tt.body != nil {
				jsonBody, _ := json.Marshal(tt.body)
				bodyReader = bytes.NewReader(jsonBody)
			} else {
				bodyReader = bytes.NewReader([]byte{})
			}
			req := httptest.NewRequest(http.MethodPost, "/notes/"+tt.noteIDPath+"/blocks", bodyReader)
			req.SetPathValue("noteId", tt.noteIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.CreateBlock(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUpdateBlockContent(t *testing.T) {
	handler, mockUsecase, ctrl := setupTestHandler(t)
	defer ctrl.Finish()

	userID := uuid.New()
	noteID := uuid.New()
	blockID := uuid.New()

	updateRequest := dto.UpdateBlockContentRequest{
		Content: "Updated content",
	}

	updatedBlock := &models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Updated content",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	tests := []struct {
		name           string
		noteIDPath     string
		blockIDPath    string
		body           interface{}
		setupMock      func()
		expectedStatus int
	}{
		{
			name:        "success",
			noteIDPath:  noteID.String(),
			blockIDPath: blockID.String(),
			body:        updateRequest,
			setupMock: func() {
				mockUsecase.EXPECT().UpdateBlockContent(gomock.Any(), blockID, noteID, userID, "Updated content").Return(updatedBlock, nil)
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing note ID",
			noteIDPath:     "",
			blockIDPath:    blockID.String(),
			body:           updateRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "",
			body:           updateRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid note ID",
			noteIDPath:     "invalid",
			blockIDPath:    blockID.String(),
			body:           updateRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid block ID",
			noteIDPath:     noteID.String(),
			blockIDPath:    "invalid",
			body:           updateRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthorized",
			noteIDPath:     noteID.String(),
			blockIDPath:    blockID.String(),
			body:           updateRequest,
			setupMock:      func() {},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonBody, _ := json.Marshal(tt.body)
			req := httptest.NewRequest(http.MethodPut, "/notes/"+tt.noteIDPath+"/blocks/"+tt.blockIDPath+"/content", bytes.NewReader(jsonBody))
			req.SetPathValue("noteId", tt.noteIDPath)
			req.SetPathValue("blockId", tt.blockIDPath)
			if tt.name != "unauthorized" {
				req = addUserIDToContext(req, userID)
			}
			w := httptest.NewRecorder()

			tt.setupMock()
			handler.UpdateBlockContent(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}
