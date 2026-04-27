package repository

import (
	"context"
	"database/sql"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRepository(t *testing.T) (*noteRepository, sqlmock.Sqlmock, *sql.DB) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	repo := NewNoteRepository(db)
	return repo, mock, db
}

func TestGetNotes_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()
	noteID1 := uuid.New()
	noteID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID1, userID, "Note 1", nil, time.Now(), time.Now()).
		AddRow(noteID2, userID, "Note 2", nil, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
		WithArgs(userID).
		WillReturnRows(rows)

	notes, err := repo.GetNotes(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, notes, 2)
	assert.Equal(t, "Note 1", notes[0].Title)
	assert.Equal(t, "Note 2", notes[1].Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNotes_WithParentID(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()
	noteID := uuid.New()
	parentID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID, userID, "Child Note", parentID, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
		WithArgs(userID).
		WillReturnRows(rows)

	notes, err := repo.GetNotes(context.Background(), userID)

	assert.NoError(t, err)
	assert.Len(t, notes, 1)
	assert.NotNil(t, notes[0].ParentID)
	assert.Equal(t, parentID, *notes[0].ParentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNote_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID, userID, "Test Note", nil, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
		WithArgs(noteID).
		WillReturnRows(rows)

	note, err := repo.GetNote(context.Background(), noteID)

	assert.NoError(t, err)
	assert.NotNil(t, note)
	assert.Equal(t, noteID, note.ID)
	assert.Equal(t, "Test Note", note.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNote_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	note, err := repo.GetNote(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Nil(t, note)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocks_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID1 := uuid.New()
	blockID2 := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID1, noteID, 1, 0, "Block 1", time.Now(), time.Now()).
		AddRow(blockID2, noteID, 2, 1, "Block 2", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
		WithArgs(noteID).
		WillReturnRows(rows)

	blocks, err := repo.GetBlocks(context.Background(), noteID)

	assert.NoError(t, err)
	assert.Len(t, blocks, 2)
	assert.Equal(t, "Block 1", blocks[0].Content)
	assert.Equal(t, 0, blocks[0].Position)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNote_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()
	note := models.Note{
		UserID:    userID,
		Title:     "New Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	expectedID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(expectedID, userID, "New Note", nil, note.CreatedAt, note.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
		WithArgs(note.UserID, note.Title, sql.NullString{}, note.CreatedAt, note.UpdatedAt).
		WillReturnRows(rows)

	createdNote, err := repo.CreateNote(context.Background(), note)

	assert.NoError(t, err)
	assert.NotNil(t, createdNote)
	assert.Equal(t, expectedID, createdNote.ID)
	assert.Equal(t, "New Note", createdNote.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNote_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	userID := uuid.New()
	note := models.Note{
		Title:     "Updated Note",
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID, userID, "Updated Note", nil, time.Now(), note.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
		WithArgs(noteID, note.Title, sql.NullString{}, note.UpdatedAt).
		WillReturnRows(rows)

	updatedNote, err := repo.UpdateNote(context.Background(), noteID, note)

	assert.NoError(t, err)
	assert.NotNil(t, updatedNote)
	assert.Equal(t, "Updated Note", updatedNote.Title)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNote_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	rows := sqlmock.NewRows([]string{"id"}).AddRow(noteID)

	mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
		WithArgs(noteID).
		WillReturnRows(rows)

	err := repo.DeleteNote(context.Background(), noteID)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteNote_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
		WithArgs(noteID).
		WillReturnError(sql.ErrNoRows)

	err := repo.DeleteNote(context.Background(), noteID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrNoteNotFound, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBlock_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	block := models.Block{
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	expectedID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(expectedID, noteID, 1, 0, "", block.CreatedAt, block.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
		WithArgs(block.NoteID, block.BlockTypeID, block.Position, block.Content, block.CreatedAt, block.UpdatedAt).
		WillReturnRows(rows)

	createdBlock, err := repo.CreateBlock(context.Background(), block)

	assert.NoError(t, err)
	assert.NotNil(t, createdBlock)
	assert.Equal(t, expectedID, createdBlock.ID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlock_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	noteID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID, noteID, 1, 0, "Content", time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_BY_ID)).
		WithArgs(blockID).
		WillReturnRows(rows)

	block, err := repo.GetBlock(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, block)
	assert.Equal(t, blockID, block.ID)
	assert.Equal(t, "Content", block.Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockContent_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	noteID := uuid.New()
	newContent := "Updated content"
	updatedAt := time.Now()

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID, noteID, 1, 0, newContent, time.Now(), updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_CONTENT)).
		WithArgs(blockID, newContent, updatedAt).
		WillReturnRows(rows)

	updatedBlock, err := repo.UpdateBlockContent(context.Background(), blockID, newContent)

	assert.NoError(t, err)
	assert.NotNil(t, updatedBlock)
	assert.Equal(t, newContent, updatedBlock.Content)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 0, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 1))

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID, noteID, 1, 2, "Content", time.Now(), updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
		WithArgs(blockID, 2, updatedAt).
		WillReturnRows(rows)

	mock.ExpectCommit()

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2)

	assert.NoError(t, err)
	assert.NotNil(t, movedBlock)
	assert.Equal(t, 2, movedBlock.Position)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBlock_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	noteID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "note_id"}).AddRow(blockID, noteID)

	mock.ExpectQuery(regexp.QuoteMeta(DELETE_BLOCK)).
		WithArgs(blockID).
		WillReturnRows(rows)

	resultNoteID, err := repo.DeleteBlock(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, resultNoteID)
	assert.Equal(t, noteID, *resultNoteID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockFormatting_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	italic := false

	rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
		AddRow(0, 5, bold, italic, nil, nil).
		AddRow(6, 10, nil, nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(rows)

	formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, formatting)
	assert.Equal(t, blockID.String(), formatting.BlockID)
	assert.Len(t, formatting.Ranges, 2)
	assert.True(t, *formatting.Ranges[0].Bold)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocksFormatting_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID1 := uuid.New()
	blockID2 := uuid.New()
	bold := true

	rows := sqlmock.NewRows([]string{"block_id", "start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
		AddRow(blockID1.String(), 0, 5, bold, nil, nil, nil).
		AddRow(blockID1.String(), 6, 10, nil, nil, nil, nil).
		AddRow(blockID2.String(), 0, 3, nil, nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_FORMATTING)).
		WithArgs(pq.Array([]uuid.UUID{blockID1, blockID2})).
		WillReturnRows(rows)

	formattings, err := repo.GetBlocksFormatting(context.Background(), []uuid.UUID{blockID1, blockID2})

	assert.NoError(t, err)
	assert.Len(t, formattings, 2)
	assert.Len(t, formattings[blockID1.String()].Ranges, 2)
	assert.Len(t, formattings[blockID2.String()].Ranges, 1)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockFormatting_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	mock.ExpectBegin()

	rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(rows)

	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
		WithArgs(blockID, 0, 5, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	newRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
		AddRow(0, 5, bold, nil, nil, nil)

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(newRows)

	result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.Ranges, 1)
	assert.Equal(t, 0, result.Ranges[0].StartPos)
	assert.Equal(t, 5, result.Ranges[0].EndPos)
	assert.True(t, *result.Ranges[0].Bold)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestResetBlockFormatting_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(rows)

	result, err := repo.ResetBlockFormatting(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result.Ranges)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockType_Success(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockTypeID := 1

	rows := sqlmock.NewRows([]string{"id", "name"}).
		AddRow(1, "text")

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
		WithArgs(blockTypeID).
		WillReturnRows(rows)

	blockType, err := repo.GetBlockType(context.Background(), blockTypeID)

	assert.NoError(t, err)
	assert.NotNil(t, blockType)
	assert.Equal(t, "text", blockType.Name)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShiftBlockPositions_Up(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_UP)).
		WithArgs(noteID, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.ShiftBlockPositions(context.Background(), noteID, 2, 1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShiftBlockPositions_Down(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 3))

	err := repo.ShiftBlockPositions(context.Background(), noteID, 2, -1)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShiftBlockPositions_ZeroDirection(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	err := repo.ShiftBlockPositions(context.Background(), noteID, 2, 0)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockFormatting_Empty(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(rows)

	formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

	assert.NoError(t, err)
	assert.NotNil(t, formatting)
	assert.Empty(t, formatting.Ranges)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocksFormatting_EmptyList(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	formattings, err := repo.GetBlocksFormatting(context.Background(), []uuid.UUID{})

	assert.NoError(t, err)
	assert.NotNil(t, formattings)
	assert.Empty(t, formattings)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocks_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
		WithArgs(noteID).
		WillReturnError(sql.ErrConnDone)

	blocks, err := repo.GetBlocks(context.Background(), noteID)

	assert.Error(t, err)
	assert.Nil(t, blocks)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocks_RowsError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(uuid.New(), noteID, 1, 0, "Block 1", time.Now(), time.Now())
	rows.RowError(0, assert.AnError)

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
		WithArgs(noteID).
		WillReturnRows(rows)

	blocks, err := repo.GetBlocks(context.Background(), noteID)

	assert.Error(t, err)
	assert.Nil(t, blocks)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNote_WithParentID(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID, userID, "Child Note", parentID, time.Now(), time.Now())

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
		WithArgs(noteID).
		WillReturnRows(rows)

	note, err := repo.GetNote(context.Background(), noteID)

	assert.NoError(t, err)
	assert.NotNil(t, note)
	assert.NotNil(t, note.ParentID)
	assert.Equal(t, parentID, *note.ParentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNote_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
		WithArgs(noteID).
		WillReturnError(sql.ErrConnDone)

	note, err := repo.GetNote(context.Background(), noteID)

	assert.Error(t, err)
	assert.Nil(t, note)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNotes_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	notes, err := repo.GetNotes(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetNotes_RowsError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(uuid.New(), userID, "Note 1", nil, time.Now(), time.Now())
	rows.RowError(0, assert.AnError)

	mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
		WithArgs(userID).
		WillReturnRows(rows)

	notes, err := repo.GetNotes(context.Background(), userID)

	assert.Error(t, err)
	assert.Nil(t, notes)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNote_WithParentID(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()
	parentID := uuid.New()
	note := models.Note{
		UserID:    userID,
		Title:     "Child Note",
		ParentID:  &parentID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	expectedID := uuid.New()
	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(expectedID, userID, "Child Note", parentID, note.CreatedAt, note.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
		WithArgs(note.UserID, note.Title, sql.NullString{String: parentID.String(), Valid: true}, note.CreatedAt, note.UpdatedAt).
		WillReturnRows(rows)

	createdNote, err := repo.CreateNote(context.Background(), note)

	assert.NoError(t, err)
	assert.Equal(t, expectedID, createdNote.ID)
	assert.Equal(t, parentID, *createdNote.ParentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateNote_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	userID := uuid.New()
	note := models.Note{
		UserID:    userID,
		Title:     "New Note",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
		WithArgs(note.UserID, note.Title, sql.NullString{}, note.CreatedAt, note.UpdatedAt).
		WillReturnError(sql.ErrConnDone)

	createdNote, err := repo.CreateNote(context.Background(), note)

	assert.Error(t, err)
	assert.Nil(t, createdNote)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNote_WithParentID(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()
	note := models.Note{
		Title:     "Updated Note",
		ParentID:  &parentID,
		UpdatedAt: time.Now(),
	}

	rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
		AddRow(noteID, userID, "Updated Note", parentID, time.Now(), note.UpdatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
		WithArgs(noteID, note.Title, sql.NullString{String: parentID.String(), Valid: true}, note.UpdatedAt).
		WillReturnRows(rows)

	updatedNote, err := repo.UpdateNote(context.Background(), noteID, note)

	assert.NoError(t, err)
	assert.Equal(t, parentID, *updatedNote.ParentID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateNote_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	note := models.Note{
		Title:     "Updated Note",
		UpdatedAt: time.Now(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
		WithArgs(noteID, note.Title, sql.NullString{}, note.UpdatedAt).
		WillReturnError(sql.ErrConnDone)

	updatedNote, err := repo.UpdateNote(context.Background(), noteID, note)

	assert.Error(t, err)
	assert.Nil(t, updatedNote)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCreateBlock_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	block := models.Block{
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "",
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
		WithArgs(block.NoteID, block.BlockTypeID, block.Position, block.Content, block.CreatedAt, block.UpdatedAt).
		WillReturnError(sql.ErrConnDone)

	createdBlock, err := repo.CreateBlock(context.Background(), block)

	assert.Error(t, err)
	assert.Nil(t, createdBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlock_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_BY_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	block, err := repo.GetBlock(context.Background(), blockID)

	assert.NoError(t, err)
	assert.Nil(t, block)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlock_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_BY_ID)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)

	block, err := repo.GetBlock(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, block)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockContent_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_CONTENT)).
		WithArgs(blockID, "content", updatedAt).
		WillReturnError(sql.ErrNoRows)

	updatedBlock, err := repo.UpdateBlockContent(context.Background(), blockID, "content")

	assert.Error(t, err)
	assert.Equal(t, notes.ErrBlockNotFound, err)
	assert.Nil(t, updatedBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_MoveUp(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_UP)).
		WithArgs(noteID, 5, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 3))

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID, noteID, 1, 2, "Content", time.Now(), updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
		WithArgs(blockID, 2, updatedAt).
		WillReturnRows(rows)

	mock.ExpectCommit()

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 5, 2)

	assert.NoError(t, err)
	assert.Equal(t, 2, movedBlock.Position)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_BeginTxError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2)

	assert.Error(t, err)
	assert.Nil(t, movedBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_UpdatePositionsError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 0, 2, updatedAt).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2)

	assert.Error(t, err)
	assert.Nil(t, movedBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_UpdateBlockError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 0, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 3))
	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
		WithArgs(blockID, 2, updatedAt).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2)

	assert.Error(t, err)
	assert.Nil(t, movedBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMoveBlock_CommitError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	blockID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 0, 2, updatedAt).
		WillReturnResult(sqlmock.NewResult(0, 3))

	rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
		AddRow(blockID, noteID, 1, 2, "Content", time.Now(), updatedAt)

	mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
		WithArgs(blockID, 2, updatedAt).
		WillReturnRows(rows)
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	movedBlock, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2)

	assert.Error(t, err)
	assert.Nil(t, movedBlock)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBlock_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(DELETE_BLOCK)).
		WithArgs(blockID).
		WillReturnError(sql.ErrNoRows)

	resultNoteID, err := repo.DeleteBlock(context.Background(), blockID)

	assert.Error(t, err)
	assert.Equal(t, notes.ErrBlockNotFound, err)
	assert.Nil(t, resultNoteID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteBlock_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(DELETE_BLOCK)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)

	resultNoteID, err := repo.DeleteBlock(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, resultNoteID)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockFormatting_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)

	formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, formatting)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlocksFormatting_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID1 := uuid.New()
	blockID2 := uuid.New()

	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_FORMATTING)).
		WithArgs(pq.Array([]uuid.UUID{blockID1, blockID2})).
		WillReturnError(sql.ErrConnDone)

	formattings, err := repo.GetBlocksFormatting(context.Background(), []uuid.UUID{blockID1, blockID2})

	assert.Error(t, err)
	assert.Nil(t, formattings)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockFormatting_BeginTxError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockFormatting_GetRangesError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockFormatting_DeleteError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	mock.ExpectBegin()
	existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(existingRows)
	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUpdateBlockFormatting_InsertError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()
	bold := true
	formattingRange := models.FormattingRange{
		StartPos: 0,
		EndPos:   5,
		Bold:     &bold,
	}

	mock.ExpectBegin()
	existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
	mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnRows(existingRows)
	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
		WithArgs(blockID, 0, 5, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestResetBlockFormatting_BeginTxError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectBegin().WillReturnError(sql.ErrConnDone)

	result, err := repo.ResetBlockFormatting(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestResetBlockFormatting_DeleteError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnError(sql.ErrConnDone)
	mock.ExpectRollback()

	result, err := repo.ResetBlockFormatting(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestResetBlockFormatting_CommitError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockID := uuid.New()

	mock.ExpectBegin()
	mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
		WithArgs(blockID).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit().WillReturnError(sql.ErrConnDone)

	result, err := repo.ResetBlockFormatting(context.Background(), blockID)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockType_NotFound(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockTypeID := 999

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
		WithArgs(blockTypeID).
		WillReturnError(sql.ErrNoRows)

	blockType, err := repo.GetBlockType(context.Background(), blockTypeID)

	assert.NoError(t, err)
	assert.Nil(t, blockType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetBlockType_QueryError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	blockTypeID := 1

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, name FROM block_types WHERE id = $1")).
		WithArgs(blockTypeID).
		WillReturnError(sql.ErrConnDone)

	blockType, err := repo.GetBlockType(context.Background(), blockTypeID)

	assert.Error(t, err)
	assert.Nil(t, blockType)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShiftBlockPositions_UpError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_UP)).
		WithArgs(noteID, 2, updatedAt).
		WillReturnError(sql.ErrConnDone)

	err := repo.ShiftBlockPositions(context.Background(), noteID, 2, 1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestShiftBlockPositions_DownError(t *testing.T) {
	repo, mock, db := setupTestRepository(t)
	defer db.Close()

	noteID := uuid.New()
	updatedAt := time.Now()

	mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_DOWN)).
		WithArgs(noteID, 2, updatedAt).
		WillReturnError(sql.ErrConnDone)

	err := repo.ShiftBlockPositions(context.Background(), noteID, 2, -1)

	assert.Error(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
