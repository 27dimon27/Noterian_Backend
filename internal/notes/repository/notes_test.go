package repository

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

func TestNewNoteRepository(t *testing.T) {
	db, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	if repo == nil {
		t.Errorf("expected non-nil repository")
	}
	if repo.db != db {
		t.Errorf("expected db to be set")
	}
}

func TestGetNotes(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	userID := uuid.New()
	noteID := uuid.New()
	now := time.Now()

	t.Run("success without parent_id", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Test Note", nil, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes").
			WithArgs(userID).
			WillReturnRows(rows)

		notes, err := repo.GetNotes(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(notes) != 1 {
			t.Errorf("expected 1 note, got %d", len(notes))
		}
	})

	t.Run("success with parent_id", func(t *testing.T) {
		parentID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Child Note", parentID, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes").
			WithArgs(userID).
			WillReturnRows(rows)

		notes, err := repo.GetNotes(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if *notes[0].ParentID != parentID {
			t.Errorf("expected ParentID %v, got %v", parentID, *notes[0].ParentID)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes").
			WithArgs(userID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetNotes(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("scan error", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow("not-a-uuid", userID, "Test Note", nil, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes").
			WithArgs(userID).
			WillReturnRows(rows)

		_, err := repo.GetNotes(context.Background(), userID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	t.Run("success without parent_id", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Test Note", nil, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id").
			WithArgs(noteID).
			WillReturnRows(rows)

		note, err := repo.GetNote(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if note.ID != noteID {
			t.Errorf("expected ID %v, got %v", noteID, note.ID)
		}
	})

	t.Run("success with parent_id", func(t *testing.T) {
		parentID := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Child Note", parentID, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id").
			WithArgs(noteID).
			WillReturnRows(rows)

		note, err := repo.GetNote(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if *note.ParentID != parentID {
			t.Errorf("expected ParentID %v, got %v", parentID, *note.ParentID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id").
			WithArgs(noteID).
			WillReturnError(sql.ErrNoRows)

		note, err := repo.GetNote(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if note != nil {
			t.Errorf("expected nil note, got %v", note)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id").
			WithArgs(noteID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetNote(context.Background(), noteID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBlocks(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	blockID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 0, "Content", now, now)

		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks").
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks) != 1 {
			t.Errorf("expected 1 block, got %d", len(blocks))
		}
	})

	t.Run("multiple blocks", func(t *testing.T) {
		blockID1 := uuid.New()
		blockID2 := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID1, noteID, 1, 0, "Content1", now, now).
			AddRow(blockID2, noteID, 2, 1, "Content2", now, now)

		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks").
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(blocks))
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks").
			WithArgs(noteID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlocks(context.Background(), noteID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBlockType(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockTypeID := 1

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "name"}).
			AddRow(1, "text")

		mock.ExpectQuery("SELECT id, name FROM block_types WHERE id").
			WithArgs(blockTypeID).
			WillReturnRows(rows)

		blockType, err := repo.GetBlockType(context.Background(), blockTypeID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if blockType.ID != 1 {
			t.Errorf("expected ID 1, got %d", blockType.ID)
		}
		if blockType.Name != "text" {
			t.Errorf("expected Name 'text', got '%s'", blockType.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name FROM block_types WHERE id").
			WithArgs(blockTypeID).
			WillReturnError(sql.ErrNoRows)

		blockType, err := repo.GetBlockType(context.Background(), blockTypeID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if blockType != nil {
			t.Errorf("expected nil, got %v", blockType)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, name FROM block_types WHERE id").
			WithArgs(blockTypeID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlockType(context.Background(), blockTypeID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestCreateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	userID := uuid.New()
	noteID := uuid.New()
	now := time.Now()

	t.Run("success without parent_id", func(t *testing.T) {
		note := models.Note{
			UserID:    userID,
			Title:     "New Note",
			CreatedAt: now,
			UpdatedAt: now,
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "New Note", nil, now, now)

		mock.ExpectQuery("INSERT INTO notes").
			WithArgs(userID, "New Note", sqlmock.AnyArg(), now, now).
			WillReturnRows(rows)

		created, err := repo.CreateNote(context.Background(), note)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if created.ID != noteID {
			t.Errorf("expected ID %v, got %v", noteID, created.ID)
		}
	})

	t.Run("success with parent_id", func(t *testing.T) {
		parentID := uuid.New()
		note := models.Note{
			UserID:    userID,
			Title:     "Child Note",
			ParentID:  &parentID,
			CreatedAt: now,
			UpdatedAt: now,
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Child Note", parentID, now, now)

		mock.ExpectQuery("INSERT INTO notes").
			WithArgs(userID, "Child Note", parentID, now, now).
			WillReturnRows(rows)

		created, err := repo.CreateNote(context.Background(), note)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if *created.ParentID != parentID {
			t.Errorf("expected ParentID %v, got %v", parentID, *created.ParentID)
		}
	})

	t.Run("query error", func(t *testing.T) {
		note := models.Note{
			UserID:    userID,
			Title:     "New Note",
			CreatedAt: now,
			UpdatedAt: now,
		}

		mock.ExpectQuery("INSERT INTO notes").
			WithArgs(userID, "New Note", sqlmock.AnyArg(), now, now).
			WillReturnError(errors.New("db error"))

		_, err := repo.CreateNote(context.Background(), note)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUpdateNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	t.Run("success without parent_id", func(t *testing.T) {
		note := models.Note{
			Title:     "Updated Title",
			UpdatedAt: now,
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Updated Title", nil, now, now)

		mock.ExpectQuery("UPDATE notes SET title").
			WithArgs(noteID, "Updated Title", sqlmock.AnyArg(), now).
			WillReturnRows(rows)

		updated, err := repo.UpdateNote(context.Background(), noteID, note)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if updated.Title != "Updated Title" {
			t.Errorf("expected title 'Updated Title', got '%s'", updated.Title)
		}
	})

	t.Run("success with parent_id", func(t *testing.T) {
		parentID := uuid.New()
		note := models.Note{
			Title:     "Updated Child",
			ParentID:  &parentID,
			UpdatedAt: now,
		}

		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID, userID, "Updated Child", parentID, now, now)

		mock.ExpectQuery("UPDATE notes SET title").
			WithArgs(noteID, "Updated Child", parentID, now).
			WillReturnRows(rows)

		updated, err := repo.UpdateNote(context.Background(), noteID, note)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if *updated.ParentID != parentID {
			t.Errorf("expected ParentID %v, got %v", parentID, *updated.ParentID)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		note := models.Note{
			Title:     "Updated Title",
			UpdatedAt: now,
		}

		mock.ExpectQuery("UPDATE notes SET title").
			WithArgs(noteID, "Updated Title", sqlmock.AnyArg(), now).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.UpdateNote(context.Background(), noteID, note)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		note := models.Note{
			Title:     "Updated Title",
			UpdatedAt: now,
		}

		mock.ExpectQuery("UPDATE notes SET title").
			WithArgs(noteID, "Updated Title", sqlmock.AnyArg(), now).
			WillReturnError(errors.New("db error"))

		_, err := repo.UpdateNote(context.Background(), noteID, note)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestDeleteNote(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id"}).AddRow(noteID)

		mock.ExpectQuery("DELETE FROM notes").
			WithArgs(noteID).
			WillReturnRows(rows)

		err := repo.DeleteNote(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("note not found", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM notes").
			WithArgs(noteID).
			WillReturnError(sql.ErrNoRows)

		err := repo.DeleteNote(context.Background(), noteID)

		if !errors.Is(err, notes.ErrNoteNotFound) {
			t.Errorf("expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM notes").
			WithArgs(noteID).
			WillReturnError(errors.New("db error"))

		err := repo.DeleteNote(context.Background(), noteID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestCreateBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	blockID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		block := models.Block{
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 0, "", now, now)

		mock.ExpectQuery("INSERT INTO blocks").
			WithArgs(noteID, 1, 0, "", now, now).
			WillReturnRows(rows)

		created, err := repo.CreateBlock(context.Background(), block)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if created.ID != blockID {
			t.Errorf("expected ID %v, got %v", blockID, created.ID)
		}
	})

	t.Run("query error", func(t *testing.T) {
		block := models.Block{
			NoteID:      noteID,
			BlockTypeID: 1,
			Position:    0,
			Content:     "",
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		mock.ExpectQuery("INSERT INTO blocks").
			WithArgs(noteID, 1, 0, "", now, now).
			WillReturnError(errors.New("db error"))

		_, err := repo.CreateBlock(context.Background(), block)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()
	noteID := uuid.New()
	now := time.Now()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 0, "Content", now, now)

		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks WHERE id").
			WithArgs(blockID).
			WillReturnRows(rows)

		block, err := repo.GetBlock(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block.ID != blockID {
			t.Errorf("expected ID %v, got %v", blockID, block.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks WHERE id").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		block, err := repo.GetBlock(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block != nil {
			t.Errorf("expected nil block, got %v", block)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT id, note_id, block_type_id, position, content, created_at, updated_at FROM blocks WHERE id").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlock(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUpdateBlockContent(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()
	noteID := uuid.New()
	now := time.Now()
	content := "New Content"

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
			AddRow(blockID, noteID, 1, 0, content, now, now)

		mock.ExpectQuery("UPDATE blocks SET content").
			WithArgs(blockID, content, now).
			WillReturnRows(rows)

		block, err := repo.UpdateBlockContent(context.Background(), blockID, content, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block.Content != content {
			t.Errorf("expected content '%s', got '%s'", content, block.Content)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mock.ExpectQuery("UPDATE blocks SET content").
			WithArgs(blockID, content, now).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.UpdateBlockContent(context.Background(), blockID, content, now)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("UPDATE blocks SET content").
			WithArgs(blockID, content, now).
			WillReturnError(errors.New("db error"))

		_, err := repo.UpdateBlockContent(context.Background(), blockID, content, now)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestMoveBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	blockID := uuid.New()
	now := time.Now()

	t.Run("success move down (old < new)", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE blocks SET position = position - 1").
			WithArgs(noteID, 0, 2, now).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectQuery("UPDATE blocks SET position").
			WithArgs(blockID, 2, now).
			WillReturnRows(sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
				AddRow(blockID, noteID, 1, 2, "Content", now, now))
		mock.ExpectCommit()

		block, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block.Position != 2 {
			t.Errorf("expected position 2, got %d", block.Position)
		}
	})

	t.Run("success move up (old > new)", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE blocks SET position = position \\+ 1").
			WithArgs(noteID, 2, 0, now).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectQuery("UPDATE blocks SET position").
			WithArgs(blockID, 0, now).
			WillReturnRows(sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
				AddRow(blockID, noteID, 1, 0, "Content", now, now))
		mock.ExpectCommit()

		block, err := repo.MoveBlock(context.Background(), noteID, blockID, 2, 0, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block.Position != 0 {
			t.Errorf("expected position 0, got %d", block.Position)
		}
	})

	t.Run("same position - no update", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectQuery("UPDATE blocks SET position").
			WithArgs(blockID, 0, now).
			WillReturnRows(sqlmock.NewRows([]string{"id", "note_id", "block_type_id", "position", "content", "created_at", "updated_at"}).
				AddRow(blockID, noteID, 1, 0, "Content", now, now))
		mock.ExpectCommit()

		block, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 0, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if block.Position != 0 {
			t.Errorf("expected position 0, got %d", block.Position)
		}
	})

	t.Run("begin transaction error", func(t *testing.T) {
		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		_, err := repo.MoveBlock(context.Background(), noteID, blockID, 0, 2, now)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestDeleteBlock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()
	noteID := uuid.New()

	t.Run("success", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"id", "note_id"}).AddRow(blockID, noteID)

		mock.ExpectQuery("DELETE FROM blocks").
			WithArgs(blockID).
			WillReturnRows(rows)

		returnedNoteID, err := repo.DeleteBlock(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if *returnedNoteID != noteID {
			t.Errorf("expected noteID %v, got %v", noteID, *returnedNoteID)
		}
	})

	t.Run("block not found", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM blocks").
			WithArgs(blockID).
			WillReturnError(sql.ErrNoRows)

		_, err := repo.DeleteBlock(context.Background(), blockID)

		if !errors.Is(err, notes.ErrBlockNotFound) {
			t.Errorf("expected ErrBlockNotFound, got %v", err)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("DELETE FROM blocks").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		_, err := repo.DeleteBlock(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestShiftBlockPositions(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	noteID := uuid.New()
	now := time.Now()

	t.Run("direction up (positive)", func(t *testing.T) {
		mock.ExpectExec("UPDATE blocks SET position = position \\+ 1").
			WithArgs(noteID, 1, now).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 1, 1, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("direction down (negative)", func(t *testing.T) {
		mock.ExpectExec("UPDATE blocks SET position = position - 1").
			WithArgs(noteID, 1, now).
			WillReturnResult(sqlmock.NewResult(0, 2))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 1, -1, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("direction zero - no operation", func(t *testing.T) {
		err := repo.ShiftBlockPositions(context.Background(), noteID, 1, 0, now)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("exec error on shift up", func(t *testing.T) {
		mock.ExpectExec("UPDATE blocks SET position = position \\+ 1").
			WithArgs(noteID, 1, now).
			WillReturnError(errors.New("db error"))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 1, 1, now)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBlockFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()

	t.Run("success with formatting ranges", func(t *testing.T) {
		boldTrue := true
		italicFalse := false
		underlineTrue := true
		textAlignCenter := 1

		rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 5, &boldTrue, &italicFalse, &underlineTrue, &textAlignCenter).
			AddRow(6, 10, nil, nil, nil, nil)

		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(rows)

		formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(formatting.Ranges) != 2 {
			t.Errorf("expected 2 ranges, got %d", len(formatting.Ranges))
		}
		if formatting.BlockID != blockID.String() {
			t.Errorf("expected BlockID %s, got %s", blockID.String(), formatting.BlockID)
		}
	})

	t.Run("no formatting", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(rows)

		formatting, err := repo.GetBlockFormatting(context.Background(), blockID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(formatting.Ranges) != 0 {
			t.Errorf("expected 0 ranges, got %d", len(formatting.Ranges))
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlockFormatting(context.Background(), blockID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestGetBlocksFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID1 := uuid.New()
	blockID2 := uuid.New()
	blockIDs := []uuid.UUID{blockID1, blockID2}

	t.Run("success with multiple blocks", func(t *testing.T) {
		boldTrue := true
		italicFalse := false

		rows := sqlmock.NewRows([]string{"block_id", "start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(blockID1.String(), 0, 5, &boldTrue, &italicFalse, nil, nil).
			AddRow(blockID1.String(), 6, 10, nil, nil, nil, nil).
			AddRow(blockID2.String(), 0, 3, nil, nil, nil, nil)

		mock.ExpectQuery("SELECT block_id, start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting WHERE block_id = ANY").
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(rows)

		result, err := repo.GetBlocksFormatting(context.Background(), blockIDs)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(result) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(result))
		}
		if len(result[blockID1.String()].Ranges) != 2 {
			t.Errorf("expected 2 ranges for block1, got %d", len(result[blockID1.String()].Ranges))
		}
		if len(result[blockID2.String()].Ranges) != 1 {
			t.Errorf("expected 1 range for block2, got %d", len(result[blockID2.String()].Ranges))
		}
	})

	t.Run("empty blockIDs", func(t *testing.T) {
		result, err := repo.GetBlocksFormatting(context.Background(), []uuid.UUID{})

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(result) != 0 {
			t.Errorf("expected empty result, got %d", len(result))
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT block_id, start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting WHERE block_id = ANY").
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlocksFormatting(context.Background(), blockIDs)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestUpdateBlockFormatting(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("cant create mock: %s", err)
	}
	defer db.Close()

	repo := NewNoteRepository(db)
	blockID := uuid.New()
	boldTrue := true

	t.Run("success - apply new formatting", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(existingRows)
		mock.ExpectExec("DELETE FROM block_formatting WHERE block_id").
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO block_formatting").
			WithArgs(blockID, 0, 10, &boldTrue, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		formattingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 10, &boldTrue, nil, nil, nil)

		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(formattingRows)

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("success - apply formatting with all fields", func(t *testing.T) {
		italicTrue := true
		underlineTrue := true
		textAlignCenter := 1

		formattingRange := models.FormattingRange{
			StartPos:  0,
			EndPos:    10,
			Bold:      &boldTrue,
			Italic:    &italicTrue,
			Underline: &underlineTrue,
			TextAlign: &textAlignCenter,
		}

		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(existingRows)
		mock.ExpectExec("DELETE FROM block_formatting WHERE block_id").
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO block_formatting").
			WithArgs(blockID, 0, 10, &boldTrue, &italicTrue, &underlineTrue, &textAlignCenter).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit()

		formattingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"}).
			AddRow(0, 10, &boldTrue, &italicTrue, &underlineTrue, &textAlignCenter)

		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(formattingRows)

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
	})

	t.Run("begin transaction error", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		mock.ExpectBegin().WillReturnError(errors.New("begin error"))

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("get formatting ranges error", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnError(errors.New("query error"))

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("delete formatting error", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(existingRows)
		mock.ExpectExec("DELETE FROM block_formatting WHERE block_id").
			WithArgs(blockID).
			WillReturnError(errors.New("delete error"))

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("insert formatting error", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(existingRows)
		mock.ExpectExec("DELETE FROM block_formatting WHERE block_id").
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO block_formatting").
			WithArgs(blockID, 0, 10, &boldTrue, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(errors.New("insert error"))

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("commit error", func(t *testing.T) {
		formattingRange := models.FormattingRange{
			StartPos: 0,
			EndPos:   10,
			Bold:     &boldTrue,
		}

		existingRows := sqlmock.NewRows([]string{"start_pos", "end_pos", "bold", "italic", "underline", "text_align"})

		mock.ExpectBegin()
		mock.ExpectQuery("SELECT start_pos, end_pos, bold, italic, underline, text_align FROM block_formatting").
			WithArgs(blockID).
			WillReturnRows(existingRows)
		mock.ExpectExec("DELETE FROM block_formatting WHERE block_id").
			WithArgs(blockID).
			WillReturnResult(sqlmock.NewResult(0, 1))
		mock.ExpectExec("INSERT INTO block_formatting").
			WithArgs(blockID, 0, 10, &boldTrue, sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectCommit().WillReturnError(errors.New("commit error"))

		_, err := repo.UpdateBlockFormatting(context.Background(), blockID, formattingRange)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}

func TestApplyFormattingToRanges(t *testing.T) {
	boldTrue := true
	italicTrue := true
	textAlignCenter := 1

	t.Run("merge overlapping ranges with same formatting", func(t *testing.T) {
		existing := []models.FormattingRange{
			{StartPos: 0, EndPos: 5, Bold: &boldTrue},
			{StartPos: 10, EndPos: 15, Bold: &boldTrue},
		}
		newRange := models.FormattingRange{StartPos: 3, EndPos: 12, Bold: &boldTrue}

		result := applyFormattingToRanges(existing, newRange)

		// ожидается один объединенный диапазон от 0 до 15
		if len(result) != 1 {
			t.Errorf("expected 1 range, got %d", len(result))
		}
		if result[0].StartPos != 0 {
			t.Errorf("expected start 0, got %d", result[0].StartPos)
		}
		if result[0].EndPos != 15 {
			t.Errorf("expected end 15, got %d", result[0].EndPos)
		}
	})

	t.Run("apply text align", func(t *testing.T) {
		existing := []models.FormattingRange{}
		newRange := models.FormattingRange{StartPos: 0, EndPos: 10, TextAlign: &textAlignCenter}

		result := applyFormattingToRanges(existing, newRange)

		if len(result) != 1 {
			t.Errorf("expected 1 range, got %d", len(result))
		}
		if *result[0].TextAlign != textAlignCenter {
			t.Errorf("expected text align %d, got %d", textAlignCenter, *result[0].TextAlign)
		}
	})

	t.Run("empty existing ranges", func(t *testing.T) {
		existing := []models.FormattingRange{}
		newRange := models.FormattingRange{StartPos: 0, EndPos: 10, Bold: &boldTrue}

		result := applyFormattingToRanges(existing, newRange)

		if len(result) != 1 {
			t.Errorf("expected 1 range, got %d", len(result))
		}
		if result[0].StartPos != 0 || result[0].EndPos != 10 {
			t.Errorf("expected range [0-10], got [%d-%d]", result[0].StartPos, result[0].EndPos)
		}
	})

	t.Run("multiple formatting properties", func(t *testing.T) {
		existing := []models.FormattingRange{
			{StartPos: 0, EndPos: 10, Bold: &boldTrue, Italic: &italicTrue},
		}
		newRange := models.FormattingRange{StartPos: 5, EndPos: 15, Underline: &boldTrue}

		result := applyFormattingToRanges(existing, newRange)

		// ожидается 3 диапазона с разными комбинациями свойств
		if len(result) != 3 {
			t.Errorf("expected 3 ranges, got %d", len(result))
		}
	})
}
