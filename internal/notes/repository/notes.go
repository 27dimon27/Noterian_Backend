package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
)

type noteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) *noteRepository {
	return &noteRepository{
		db: db,
	}
}

func (r *noteRepository) GetNotes(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
	rows, err := r.db.QueryContext(ctx, GET_NOTES_BY_USER, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []models.Note
	for rows.Next() {
		var note models.Note
		var parentID sql.NullString

		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &parentID, &note.CreatedAt, &note.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if parentID.Valid {
			pid, err := uuid.Parse(parentID.String)
			if err != nil {
				return nil, err
			}
			note.ParentID = &pid
		}

		notes = append(notes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notes, nil
}

func (r *noteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
	var note models.Note
	var parentID sql.NullString

	err := r.db.QueryRowContext(ctx, GET_NOTE_BY_ID, noteID).Scan(
		&note.ID, &note.UserID, &note.Title, &parentID, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return nil, err
		}
		note.ParentID = &pid
	}

	return &note, nil
}

func (r *noteRepository) GetBlocks(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
	rows, err := r.db.QueryContext(ctx, GET_BLOCKS_BY_NOTE, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var blocks []models.Block

	for rows.Next() {
		var (
			block      models.Block
			formatting sql.NullString
		)

		err := rows.Scan(&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
			&formatting, &block.CreatedAt, &block.UpdatedAt)
		if err != nil {
			return nil, err
		}

		if formatting.Valid {
			var formattingData models.Formatting
			if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
				return nil, err
			}
			block.Formatting = formattingData
		} else {
			block.Formatting = getDefaultFormatting()
		}

		blocks = append(blocks, block)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (r *noteRepository) CreateNote(ctx context.Context, note models.Note) (*models.Note, error) {
	parentID := sql.NullString{}
	if note.ParentID != nil {
		parentID = sql.NullString{
			String: note.ParentID.String(),
			Valid:  true,
		}
	}

	err := r.db.QueryRowContext(ctx, CREATE_NOTE,
		note.UserID, note.Title, parentID, note.CreatedAt, note.UpdatedAt,
	).Scan(
		&note.ID, &note.UserID, &note.Title, &note.ParentID, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &note, nil
}

func (r *noteRepository) UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, error) {
	parentID := sql.NullString{}
	if note.ParentID != nil {
		parentID = sql.NullString{
			String: note.ParentID.String(),
			Valid:  true,
		}
	}

	updatedNote := &models.Note{}

	err := r.db.QueryRowContext(ctx, UPDATE_NOTE, noteID, note.Title, parentID, note.UpdatedAt).Scan(
		&updatedNote.ID,
		&updatedNote.UserID,
		&updatedNote.Title,
		&updatedNote.ParentID,
		&updatedNote.CreatedAt,
		&updatedNote.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrNoteNotFound
		}
		return nil, err
	}

	return updatedNote, nil
}

func (r *noteRepository) DeleteNote(ctx context.Context, noteID uuid.UUID) error {
	var id uuid.UUID

	err := r.db.QueryRowContext(ctx, DELETE_NOTE, noteID).Scan(&id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return notes.ErrNoteNotFound
		}
		return err
	}

	return nil
}

func (r *noteRepository) CreateBlock(ctx context.Context, block models.Block) (*models.Block, error) {
	formattingJSON, err := json.Marshal(getDefaultFormatting())
	if err != nil {
		return nil, err
	}

	err = r.db.QueryRowContext(ctx, CREATE_BLOCK,
		block.NoteID, block.BlockTypeID, block.Position, block.Content, formattingJSON, block.CreatedAt, block.UpdatedAt,
	).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&formattingJSON, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	var formattingData models.Formatting
	if err := json.Unmarshal(formattingJSON, &formattingData); err != nil {
		return nil, err
	}

	block.Formatting = formattingData
	return &block, nil
}

func (r *noteRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	var block models.Block
	var formatting sql.NullString

	err := r.db.QueryRowContext(ctx, GET_BLOCK_BY_ID, blockID).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&formatting, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if formatting.Valid {
		var formattingData models.Formatting
		if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
			return nil, err
		}
		block.Formatting = formattingData
	} else {
		block.Formatting = getDefaultFormatting()
	}

	return &block, nil
}

func (r *noteRepository) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string, updatedAt time.Time) (*models.Block, error) {
	var block models.Block
	var formatting sql.NullString

	err := r.db.QueryRowContext(ctx, UPDATE_BLOCK_CONTENT, blockID, content, updatedAt).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&formatting, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	if formatting.Valid {
		var formattingData models.Formatting
		if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
			return nil, err
		}
		block.Formatting = formattingData
	} else {
		block.Formatting = getDefaultFormatting()
	}

	return &block, nil
}

func (r *noteRepository) MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int, updatedAt time.Time) (*models.Block, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			if err == nil {
				err = rollbackErr
			}
		}
	}()

	if oldPosition < newPosition {
		_, err := tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_DOWN, noteID, oldPosition, newPosition, updatedAt)
		if err != nil {
			return nil, err
		}
	} else if oldPosition > newPosition {
		_, err := tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_UP, noteID, oldPosition, newPosition, updatedAt)
		if err != nil {
			return nil, err
		}
	}

	var updatedBlock models.Block
	var formatting sql.NullString

	err = tx.QueryRowContext(ctx, UPDATE_BLOCK_POSITION, blockID, newPosition, updatedAt).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID, &updatedBlock.Position, &updatedBlock.Content,
		&formatting, &updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if formatting.Valid {
		var formattingData models.Formatting
		if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
			return nil, err
		}
		updatedBlock.Formatting = formattingData
	} else {
		updatedBlock.Formatting = getDefaultFormatting()
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &updatedBlock, nil
}

func (r *noteRepository) DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, error) {
	var deletedBlockID uuid.UUID
	var noteID uuid.UUID

	err := r.db.QueryRowContext(ctx, DELETE_BLOCK, blockID).Scan(&deletedBlockID, &noteID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	return &noteID, nil
}

func (r *noteRepository) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int, updatedAt time.Time) error {
	if direction > 0 {
		_, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_UP, noteID, fromPosition, updatedAt)
		return err
	} else if direction < 0 {
		_, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_DOWN, noteID, fromPosition, updatedAt)
		return err
	}
	return nil
}

func (r *noteRepository) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingData models.Formatting) (*models.Block, error) {
	formattingJSON, err := json.Marshal(formattingData)
	if err != nil {
		return nil, err
	}

	var updatedBlock models.Block
	var formatting sql.NullString

	err = r.db.QueryRowContext(ctx, UPDATE_BLOCK_FORMATTING,
		blockID, formattingJSON, time.Now(),
	).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID,
		&updatedBlock.Position, &updatedBlock.Content, &formatting,
		&updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	if formatting.Valid {
		var updatedFormattingData models.Formatting
		if err := json.Unmarshal([]byte(formatting.String), &updatedFormattingData); err != nil {
			return nil, err
		}
		updatedBlock.Formatting = updatedFormattingData
	} else {
		updatedBlock.Formatting = getDefaultFormatting()
	}

	return &updatedBlock, nil
}

func (r *noteRepository) ResetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	defaultFormatting := getDefaultFormatting()

	formattingJSON, err := json.Marshal(defaultFormatting)
	if err != nil {
		return nil, err
	}

	var updatedBlock models.Block
	var formatting sql.NullString

	err = r.db.QueryRowContext(ctx, UPDATE_BLOCK_FORMATTING,
		blockID, formattingJSON, time.Now(),
	).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID,
		&updatedBlock.Position, &updatedBlock.Content, &formatting,
		&updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	updatedBlock.Formatting = defaultFormatting
	return &updatedBlock, nil
}

func getDefaultFormatting() models.Formatting {
	return models.Formatting{
		Bold:      false,
		Italic:    false,
		Underline: false,
		TextAlign: -1,
	}
}
