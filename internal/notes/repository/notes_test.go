// repository/notes_test.go (исправленная версия)
package repository

import (
	"context"
	"database/sql"
	"database/sql/driver"
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

// anyArgument представляет любой аргумент для sqlmock
type anyArgument struct{}

func (a anyArgument) Match(v driver.Value) bool {
	return true
}

func TestNoteRepository_GetNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	userID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "Note 1", nil, false, false, "📝", time.Now(), time.Now()).
			AddRow(uuid.New(), userID, "Note 2", nil, true, true, "⭐", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
			WithArgs(userID).
			WillReturnRows(rows)

		notes, err := repo.GetNotes(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, notes, 2)
	})

	t.Run("with parent_id", func(t *testing.T) {
		parentID := uuid.New()
		parentIDStr := parentID.String()
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "Subnote", parentIDStr, false, false, "📝", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
			WithArgs(userID).
			WillReturnRows(rows)

		notes, err := repo.GetNotes(context.Background(), userID)

		assert.NoError(t, err)
		assert.Len(t, notes, 1)
		assert.NotNil(t, notes[0].ParentID)
		assert.Equal(t, parentID, *notes[0].ParentID)
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(GET_NOTES_BY_USER)).
			WithArgs(userID).
			WillReturnError(sql.ErrConnDone)

		notes, err := repo.GetNotes(context.Background(), userID)

		assert.Error(t, err)
		assert.Nil(t, notes)
	})
}

func TestNoteRepository_GetNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(noteID, uuid.New(), "Test Note", nil, false, false, "📝", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
			WithArgs(noteID).
			WillReturnRows(rows)

		note, err := repo.GetNote(context.Background(), noteID)

		assert.NoError(t, err)
		assert.NotNil(t, note)
		assert.Equal(t, noteID, note.ID)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery(regexp.QuoteMeta(GET_NOTE_BY_ID)).
			WithArgs(noteID).
			WillReturnError(sql.ErrNoRows)

		note, err := repo.GetNote(context.Background(), noteID)

		assert.Error(t, err)
		assert.Nil(t, note)
		assert.Equal(t, notes.ErrNoteNotFound, err)
	})
}

func TestNoteRepository_GetBlocks(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(uuid.New(), noteID, 1, 0, "Text content", time.Now(), time.Now()).
			AddRow(uuid.New(), noteID, 2, 1, "image.jpg", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		assert.NoError(t, err)
		assert.Len(t, blocks, 2)
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_BY_NOTE)).
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		assert.NoError(t, err)
		assert.Len(t, blocks, 0)
	})
}

func TestNoteRepository_GetBlockType(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "text")

		mock.ExpectQuery("SELECT id, name FROM block_types WHERE id = \\$1").
			WithArgs(1).
			WillReturnRows(rows)

		blockType, err := repo.GetBlockType(context.Background(), 1)

		assert.NoError(t, err)
		assert.NotNil(t, blockType)
		assert.Equal(t, 1, blockType.ID)
		assert.Equal(t, "text", blockType.Name)
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name FROM block_types WHERE id = \\$1").
			WithArgs(999).
			WillReturnError(sql.ErrNoRows)

		blockType, err := repo.GetBlockType(context.Background(), 999)

		assert.Error(t, err)
		assert.Nil(t, blockType)
		assert.Equal(t, notes.ErrBlockTypeNotFound, err)
	})
}

func TestNoteRepository_CreateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	userID := uuid.New()

	t.Run("success without parent", func(t *testing.T) {
		note := models.Note{
			UserID:     userID,
			Title:      "New Note",
			IsPublic:   false,
			IsFavorite: false,
			Icon:       "📝",
		}
		createdAt := time.Now()
		updatedAt := time.Now()

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(uuid.New(), userID, "New Note", nil, false, false, "📝", createdAt, updatedAt)

		mock.ExpectQuery(regexp.QuoteMeta(CREATE_NOTE)).
			WithArgs(userID, "New Note", sql.NullString{}, false, false, "📝").
			WillReturnRows(rows)

		createdNote, err := repo.CreateNote(context.Background(), note)

		assert.NoError(t, err)
		assert.NotNil(t, createdNote)
	})
}

func TestNoteRepository_UpdateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		note := models.Note{
			Title:      "Updated Title",
			IsPublic:   true,
			IsFavorite: true,
			Icon:       "⭐",
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(noteID, uuid.New(), "Updated Title", nil, true, true, "⭐", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(UPDATE_NOTE)).
			WithArgs(noteID, "Updated Title", sql.NullString{}, true, true, "⭐").
			WillReturnRows(rows)

		updatedNote, err := repo.UpdateNote(context.Background(), noteID, note)

		assert.NoError(t, err)
		assert.NotNil(t, updatedNote)
		assert.Equal(t, "Updated Title", updatedNote.Title)
	})
}

func TestNoteRepository_DeleteNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(noteID)

		mock.ExpectQuery(regexp.QuoteMeta(DELETE_NOTE)).
			WithArgs(noteID).
			WillReturnRows(rows)

		err := repo.DeleteNote(context.Background(), noteID)

		assert.NoError(t, err)
	})
}

func TestNoteRepository_CreateBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		block := models.Block{
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "Hello",
		}

		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(uuid.New(), noteID, 1, 0, "Hello", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(CREATE_BLOCK)).
			WithArgs(noteID, 1, 0, "Hello").
			WillReturnRows(rows)

		createdBlock, err := repo.CreateBlock(context.Background(), block)

		assert.NoError(t, err)
		assert.NotNil(t, createdBlock)
	})
}

func TestNoteRepository_GetBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, uuid.New(), 1, 0, "Content", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_BY_ID)).
			WithArgs(blockID).
			WillReturnRows(rows)

		block, err := repo.GetBlock(context.Background(), blockID)

		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, blockID, block.ID)
	})
}

func TestNoteRepository_UpdateBlockContent(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, uuid.New(), 1, 0, "New Content", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_CONTENT)).
			WithArgs(blockID, "New Content").
			WillReturnRows(rows)

		block, err := repo.UpdateBlockContent(context.Background(), blockID, "New Content")

		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, "New Content", block.Content)
	})
}

func TestNoteRepository_MoveBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	blockID := uuid.New()

	t.Run("success move down", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_DOWN)).
			WithArgs(noteID, 1, 3).
			WillReturnResult(sqlmock.NewResult(0, 2))

		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 3, "Content", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
			WithArgs(blockID, 3).
			WillReturnRows(rows)

		mock.ExpectCommit()

		block, err := repo.MoveBlock(context.Background(), noteID, blockID, 1, 3)

		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, 3, block.Position)
	})

	t.Run("success move up", func(t *testing.T) {
		mock.ExpectBegin()

		mock.ExpectExec(regexp.QuoteMeta(UPDATE_BLOCKS_POSITION_UP)).
			WithArgs(noteID, 3, 1).
			WillReturnResult(sqlmock.NewResult(0, 2))

		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 1, "Content", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(UPDATE_BLOCK_POSITION)).
			WithArgs(blockID, 1).
			WillReturnRows(rows)

		mock.ExpectCommit()

		block, err := repo.MoveBlock(context.Background(), noteID, blockID, 3, 1)

		assert.NoError(t, err)
		assert.NotNil(t, block)
		assert.Equal(t, 1, block.Position)
	})
}

func TestNoteRepository_DeleteBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id"}).
			AddRow(blockID, noteID)

		mock.ExpectQuery(regexp.QuoteMeta(DELETE_BLOCK)).
			WithArgs(blockID).
			WillReturnRows(rows)

		returnedNoteID, err := repo.DeleteBlock(context.Background(), blockID)

		assert.NoError(t, err)
		assert.NotNil(t, returnedNoteID)
		assert.Equal(t, noteID, *returnedNoteID)
	})
}

func TestNoteRepository_ShiftBlockPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("shift up (direction > 0)", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_UP)).
			WithArgs(noteID, 2).
			WillReturnResult(sqlmock.NewResult(0, 3))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 2, 1)

		assert.NoError(t, err)
	})

	t.Run("shift down (direction < 0)", func(t *testing.T) {
		mock.ExpectExec(regexp.QuoteMeta(UPDATE_ALL_BLOCKS_POSITION_DOWN)).
			WithArgs(noteID, 2).
			WillReturnResult(sqlmock.NewResult(0, 3))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 2, -1)

		assert.NoError(t, err)
	})

	t.Run("direction zero - no operation", func(t *testing.T) {
		err := repo.ShiftBlockPositions(context.Background(), noteID, 2, 0)

		assert.NoError(t, err)
	})
}

func TestNoteRepository_GetBlockFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()

	t.Run("success with formatting", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 5, true, false, false, nil).
			AddRow(6, 10, false, true, true, 1)

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(rows)

		formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

		assert.NoError(t, err)
		assert.NotNil(t, formatting)
		assert.Equal(t, blockID.String(), formatting.BlockID)
		assert.Len(t, formatting.Ranges, 2)
	})

	t.Run("empty formatting", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(rows)

		formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

		assert.NoError(t, err)
		assert.NotNil(t, formatting)
		assert.Len(t, formatting.Ranges, 0)
	})
}

func TestNoteRepository_GetBlocksFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)

	t.Run("success", func(t *testing.T) {
		blockID1 := uuid.New()
		blockID2 := uuid.New()
		blockIDs := []uuid.UUID{blockID1, blockID2}

		rows := sqlmock.NewRows([]string{"block_id", "start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(blockID1.String(), 0, 5, true, false, false, nil).
			AddRow(blockID1.String(), 6, 10, false, true, false, nil).
			AddRow(blockID2.String(), 0, 3, false, false, true, 1)

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCKS_FORMATTING)).
			WithArgs(pq.Array(blockIDs)).
			WillReturnRows(rows)

		result, err := repo.GetBlocksFormatting(context.Background(), blockIDs)

		assert.NoError(t, err)
		assert.Len(t, result, 2)
		assert.Len(t, result[blockID1.String()].Ranges, 2)
		assert.Len(t, result[blockID2.String()].Ranges, 1)
	})

	t.Run("empty block IDs", func(t *testing.T) {
		result, err := repo.GetBlocksFormatting(context.Background(), []uuid.UUID{})

		assert.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestNoteRepository_UpdateBlockFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()

	t.Run("success create new formatting", func(t *testing.T) {
		mock.ExpectBegin()

		// Ожидаем запрос существующих форматирований
		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})
		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(existingRows)

		// Ожидаем удаление старых форматирований
		mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Ожидаем вставку нового форматирования - используем anyArgument для text_align
		mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
			WithArgs(blockID, 0, 5, true, false, false, anyArgument{}).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// После коммита - получаем обновленное форматирование
		formattingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 5, true, false, false, nil)

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(formattingRows)

		formattingRange := models.FormattingRange{StartPos: 0, EndPos: 5, Bold: boolPtr(true)}
		result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})

	t.Run("update existing formatting", func(t *testing.T) {
		mock.ExpectBegin()

		// Существующее форматирование
		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 10, true, false, false, nil)

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(existingRows)

		// Удаление старых
		mock.ExpectExec(regexp.QuoteMeta(DELETE_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		// Вставка новых диапазонов
		mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
			WithArgs(blockID, 0, 5, false, true, false, anyArgument{}).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectExec(regexp.QuoteMeta(INSERT_BLOCK_FORMATTING)).
			WithArgs(blockID, 5, 10, true, false, false, anyArgument{}).
			WillReturnResult(sqlmock.NewResult(1, 1))

		mock.ExpectCommit()

		// Возвращаем обновленное форматирование
		formattingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 5, false, true, false, nil).
			AddRow(5, 10, true, false, false, nil)

		mock.ExpectQuery(regexp.QuoteMeta(GET_BLOCK_FORMATTING)).
			WithArgs(blockID).
			WillReturnRows(formattingRows)

		formattingRange := models.FormattingRange{StartPos: 0, EndPos: 5, Bold: boolPtr(false), Italic: boolPtr(true)}
		result, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		assert.NoError(t, err)
		assert.NotNil(t, result)
	})
}

func TestNoteRepository_GetSubnotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	repo := NewNoteRepository(db)
	parentID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"}).
			AddRow(uuid.New(), uuid.New(), "Subnote 1", parentID, false, false, "📝", time.Now(), time.Now()).
			AddRow(uuid.New(), uuid.New(), "Subnote 2", parentID, true, false, "⭐", time.Now(), time.Now())

		mock.ExpectQuery(regexp.QuoteMeta(GET_SUBNOTES_BY_NOTE)).
			WithArgs(parentID).
			WillReturnRows(rows)

		subnotes, err := repo.GetSubnotes(context.Background(), parentID)

		assert.NoError(t, err)
		assert.Len(t, subnotes, 2)
	})

	t.Run("empty result", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "is_public", "is_favorite", "icon", "created_at", "updated_at"})

		mock.ExpectQuery(regexp.QuoteMeta(GET_SUBNOTES_BY_NOTE)).
			WithArgs(parentID).
			WillReturnRows(rows)

		subnotes, err := repo.GetSubnotes(context.Background(), parentID)

		assert.NoError(t, err)
		assert.Len(t, subnotes, 0)
	})
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}
