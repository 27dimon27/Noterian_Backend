package repository

import (
	"context"
	"database/sql"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)
	return db, mock
}

func TestGetNote(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	noteID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantNil   bool
		wantErr   bool
	}{
		{
			name: "success without parent",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Test Note", nil, false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "success with parent",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Test Note", parentID, false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name: "not found",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
					WithArgs(noteID).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name: "scan error - invalid parent_id",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Test Note", "invalid-uuid", false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantNil: true,
			wantErr: true,
		},
		{
			name: "query error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
					WithArgs(noteID).
					WillReturnError(errors.New("db error"))
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			note, err := repo.GetNote(context.Background(), noteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, note)
				} else {
					assert.NotNil(t, note)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetBlocks(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	noteID := uuid.New()
	blockID1 := uuid.New()
	blockID2 := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantLen   int
		wantErr   bool
	}{
		{
			name: "success with blocks",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
					AddRow(blockID1, noteID, 1, 0, "Content 1", time.Now(), time.Now()).
					AddRow(blockID2, noteID, 2, 1, "Content 2", time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantLen: 2,
			wantErr: false,
		},
		{
			name: "no blocks",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"})
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantLen: 0,
			wantErr: false,
		},
		{
			name: "query error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
					WithArgs(noteID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
		{
			name: "rows iteration error",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
					AddRow(blockID1, noteID, 1, 0, "Content 1", time.Now(), time.Now()).
					RowError(1, errors.New("row error"))
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
					WithArgs(noteID).
					WillReturnRows(rows)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			blocks, err := repo.GetBlocks(context.Background(), noteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, blocks, tt.wantLen)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetBlockType(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	tests := []struct {
		name        string
		blockTypeID int
		setupMock   func()
		wantNil     bool
		wantErr     bool
	}{
		{
			name:        "success",
			blockTypeID: 1,
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "name"}).
					AddRow(1, "text")
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
					WithArgs(1).
					WillReturnRows(rows)
			},
			wantNil: false,
			wantErr: false,
		},
		{
			name:        "not found",
			blockTypeID: 999,
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
					WithArgs(999).
					WillReturnError(sql.ErrNoRows)
			},
			wantNil: true,
			wantErr: false,
		},
		{
			name:        "query error",
			blockTypeID: 1,
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
					WithArgs(1).
					WillReturnError(errors.New("db error"))
			},
			wantNil: true,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			blockType, err := repo.GetBlockType(context.Background(), tt.blockTypeID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				if tt.wantNil {
					assert.Nil(t, blockType)
				} else {
					assert.NotNil(t, blockType)
				}
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateNote(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	userID := uuid.New()
	noteID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name      string
		note      models.Note
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success without parent",
			note: models.Note{
				UserID:   userID,
				Title:    "New Note",
				IsPublic: false,
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "New Note", nil, false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
					WithArgs(userID, "New Note", sql.NullString{Valid: false}, false).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "success with parent",
			note: models.Note{
				UserID:   userID,
				Title:    "New Note",
				ParentID: &parentID,
				IsPublic: true,
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "New Note", parentID, true, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
					WithArgs(userID, "New Note", sql.NullString{String: parentID.String(), Valid: true}, true).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query error",
			note: models.Note{
				UserID:   userID,
				Title:    "New Note",
				IsPublic: false,
			},
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
					WithArgs(userID, "New Note", sql.NullString{Valid: false}, false).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			note, err := repo.CreateNote(context.Background(), tt.note)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, note)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, note)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateNote(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	noteID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name      string
		note      models.Note
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success without parent",
			note: models.Note{
				Title:    "Updated Note",
				IsPublic: true,
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Updated Note", nil, true, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
					WithArgs(noteID, "Updated Note", sql.NullString{Valid: false}, true).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "success with parent",
			note: models.Note{
				Title:    "Updated Note",
				ParentID: &parentID,
				IsPublic: false,
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Updated Note", parentID, false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
					WithArgs(noteID, "Updated Note", sql.NullString{String: parentID.String(), Valid: true}, false).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "note not found",
			note: models.Note{
				Title:    "Updated Note",
				IsPublic: true,
			},
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
					WithArgs(noteID, "Updated Note", sql.NullString{Valid: false}, true).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "query error",
			note: models.Note{
				Title:    "Updated Note",
				IsPublic: true,
			},
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
					WithArgs(noteID, "Updated Note", sql.NullString{Valid: false}, true).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			note, err := repo.UpdateNote(context.Background(), noteID, tt.note)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, note)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, note)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestDeleteNote(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	noteID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
					WithArgs(noteID).
					WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(noteID))
			},
			wantErr: false,
		},
		{
			name: "note not found",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
					WithArgs(noteID).
					WillReturnError(sql.ErrNoRows)
			},
			wantErr: true,
		},
		{
			name: "query error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
					WithArgs(noteID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			err := repo.DeleteNote(context.Background(), noteID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestCreateBlock(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	noteID := uuid.New()
	blockID := uuid.New()

	tests := []struct {
		name      string
		block     models.Block
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			block: models.Block{
				NoteID:      noteID,
				BlockTypeID: 1,
				Position:    0,
				Content:     "Test content",
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
					AddRow(blockID, noteID, 1, 0, "Test content", time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
					WithArgs(noteID, 1, 0, "Test content").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "empty content",
			block: models.Block{
				NoteID:      noteID,
				BlockTypeID: 2,
				Position:    1,
				Content:     "",
			},
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
					AddRow(blockID, noteID, 2, 1, "", time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
					WithArgs(noteID, 2, 1, "").
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query error",
			block: models.Block{
				NoteID:      noteID,
				BlockTypeID: 1,
				Position:    0,
				Content:     "Test content",
			},
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
					WithArgs(noteID, 1, 0, "Test content").
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			block, err := repo.CreateBlock(context.Background(), tt.block)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, block)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, block)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestGetNotes(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	userID := uuid.New()
	noteID := uuid.New()
	parentID := uuid.New()

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "created_at", "updated_at"}).
					AddRow(noteID, userID, "Test Note", parentID, false, time.Now(), time.Now())
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
					WithArgs(userID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "query error",
			setupMock: func() {
				mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
					WithArgs(userID).
					WillReturnError(errors.New("db error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			notes, err := repo.GetNotes(context.Background(), userID)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// notes can be empty slice, but not nil
				assert.NotNil(t, notes)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}

func TestUpdateBlockFormatting(t *testing.T) {
	db, mock := setupTestDB(t)
	defer db.Close()

	repo := NewNoteRepository(db)

	blockID := uuid.New()
	boldTrue := true

	tests := []struct {
		name      string
		setupMock func()
		wantErr   bool
	}{
		{
			name: "success",
			setupMock: func() {
				mock.ExpectBegin()

				// Get existing formatting
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
					WithArgs(blockID).
					WillReturnRows(sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}))

				// Delete existing formatting
				mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
					WithArgs(blockID).
					WillReturnResult(sqlmock.NewResult(0, 1))

				// Insert new formatting - use sqlmock.AnyArg() for nil values or use proper type
				mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
					WithArgs(blockID, 0, 5, true, false, false, sqlmock.AnyArg()).
					WillReturnResult(sqlmock.NewResult(0, 1))

				mock.ExpectCommit()

				// After commit - get formatting
				rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
					AddRow(0, 5, true, false, false, nil)
				mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
					WithArgs(blockID).
					WillReturnRows(rows)
			},
			wantErr: false,
		},
		{
			name: "begin transaction error",
			setupMock: func() {
				mock.ExpectBegin().WillReturnError(errors.New("tx error"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setupMock()
			formattingRange := models.FormattingRange{
				StartPos: 0,
				EndPos:   5,
				Bold:     &boldTrue,
			}
			result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
			}
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
