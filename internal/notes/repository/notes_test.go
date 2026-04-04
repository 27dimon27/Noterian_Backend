package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

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
		if notes[0].ID != noteID {
			t.Errorf("expected ID %v, got %v", noteID, notes[0].ID)
		}
		if notes[0].ParentID != nil {
			t.Errorf("expected nil ParentID, got %v", notes[0].ParentID)
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

	t.Run("multiple notes", func(t *testing.T) {
		noteID1 := uuid.New()
		noteID2 := uuid.New()
		rows := sqlmock.NewRows([]string{"id", "user_id", "title", "parent_id", "created_at", "updated_at"}).
			AddRow(noteID1, userID, "Note 1", nil, now, now).
			AddRow(noteID2, userID, "Note 2", nil, now, now)

		mock.ExpectQuery("SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes").
			WithArgs(userID).
			WillReturnRows(rows)

		notes, err := repo.GetNotes(context.Background(), userID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(notes) != 2 {
			t.Errorf("expected 2 notes, got %d", len(notes))
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
	stateID := uuid.New()
	now := time.Now()
	formatting := map[string]interface{}{"format": "text", "bold": true}
	formattingJSON, _ := json.Marshal(formatting)

	t.Run("success with state", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "note_id", "block_type_id", "position", "content",
			"state_id", "formatting", "state_created_at", "state_updated_at",
		}).AddRow(
			blockID, noteID, 1, 0, "Content",
			stateID, formattingJSON, now, now,
		)

		mock.ExpectQuery("SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs").
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks) != 1 {
			t.Errorf("expected 1 block, got %d", len(blocks))
		}
		if len(blocks[0].States) != 1 {
			t.Errorf("expected 1 state, got %d", len(blocks[0].States))
		}
		if blocks[0].States[0].Formatting["bold"] != true {
			t.Errorf("expected bold true, got %v", blocks[0].States[0].Formatting["bold"])
		}
	})

	t.Run("block without states", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "note_id", "block_type_id", "position", "content",
			"state_id", "formatting", "state_created_at", "state_updated_at",
		}).AddRow(
			blockID, noteID, 1, 0, "Content",
			nil, nil, nil, nil,
		)

		mock.ExpectQuery("SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs").
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks[0].States) != 0 {
			t.Errorf("expected 0 states, got %d", len(blocks[0].States))
		}
	})

	t.Run("multiple blocks sorted by position", func(t *testing.T) {
		blockID1 := uuid.New()
		blockID2 := uuid.New()
		rows := sqlmock.NewRows([]string{
			"id", "note_id", "block_type_id", "position", "content",
			"state_id", "formatting", "state_created_at", "state_updated_at",
		}).
			AddRow(blockID2, noteID, 1, 1, "Content2", nil, nil, nil, nil).
			AddRow(blockID1, noteID, 1, 0, "Content1", nil, nil, nil, nil)

		mock.ExpectQuery("SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs").
			WithArgs(noteID).
			WillReturnRows(rows)

		blocks, err := repo.GetBlocks(context.Background(), noteID)

		if err != nil {
			t.Errorf("unexpected err: %s", err)
		}
		if len(blocks) != 2 {
			t.Errorf("expected 2 blocks, got %d", len(blocks))
		}
		if blocks[0].Position != 0 {
			t.Errorf("expected first block position 0, got %d", blocks[0].Position)
		}
		if blocks[1].Position != 1 {
			t.Errorf("expected second block position 1, got %d", blocks[1].Position)
		}
	})

	t.Run("query error", func(t *testing.T) {
		mock.ExpectQuery("SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs").
			WithArgs(noteID).
			WillReturnError(errors.New("db error"))

		_, err := repo.GetBlocks(context.Background(), noteID)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})

	t.Run("invalid state UUID", func(t *testing.T) {
		rows := sqlmock.NewRows([]string{
			"id", "note_id", "block_type_id", "position", "content",
			"state_id", "formatting", "state_created_at", "state_updated_at",
		}).AddRow(
			blockID, noteID, 1, 0, "Content",
			"invalid-uuid", formattingJSON, now, now,
		)

		mock.ExpectQuery("SELECT b.id, b.note_id, b.block_type_id, b.position, b.content, bs.id, bs.formatting, bs.created_at, bs.updated_at FROM blocks b LEFT JOIN block_states bs").
			WithArgs(noteID).
			WillReturnRows(rows)

		_, err := repo.GetBlocks(context.Background(), noteID)

		if !errors.Is(err, notes.ErrInvalidUUID) {
			t.Errorf("expected ErrInvalidUUID, got %v", err)
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

	t.Run("update position query error", func(t *testing.T) {
		mock.ExpectBegin()
		mock.ExpectExec("UPDATE blocks SET position = position - 1").
			WithArgs(noteID, 0, 2, now).
			WillReturnResult(sqlmock.NewResult(0, 2))
		mock.ExpectQuery("UPDATE blocks SET position").
			WithArgs(blockID, 2, now).
			WillReturnError(errors.New("update error"))

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

	t.Run("exec error on shift down", func(t *testing.T) {
		mock.ExpectExec("UPDATE blocks SET position = position - 1").
			WithArgs(noteID, 1, now).
			WillReturnError(errors.New("db error"))

		err := repo.ShiftBlockPositions(context.Background(), noteID, 1, -1, now)

		if err == nil {
			t.Errorf("expected error, got nil")
		}
	})
}
