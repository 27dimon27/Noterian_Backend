package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"
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

	blocksMap := make(map[uuid.UUID]*models.Block)

	for rows.Next() {
		var (
			blockID        uuid.UUID
			noteID         uuid.UUID
			blockTypeID    int
			position       int
			content        string
			stateID        sql.NullString
			formatting     sql.NullString
			stateCreatedAt sql.NullTime
			stateUpdatedAt sql.NullTime
		)

		err := rows.Scan(&blockID, &noteID, &blockTypeID, &position, &content,
			&stateID, &formatting, &stateCreatedAt, &stateUpdatedAt)
		if err != nil {
			return nil, err
		}

		block, exists := blocksMap[blockID]
		if !exists {
			block = &models.Block{
				ID:          blockID,
				NoteID:      noteID,
				BlockTypeID: blockTypeID,
				Position:    position,
				Content:     content,
				States:      []models.BlockState{},
			}
			blocksMap[blockID] = block
		}

		if stateID.Valid {
			var formattingData map[string]interface{}
			if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
				formattingData = map[string]interface{}{
					"format": "text",
				}
			}

			stateUUID, err := uuid.Parse(stateID.String)
			if err != nil {
				return nil, notes.ErrInvalidUUID
			}

			state := models.BlockState{
				ID:         stateUUID,
				BlockID:    blockID,
				Formatting: formattingData,
				CreatedAt:  stateCreatedAt.Time,
				UpdatedAt:  stateUpdatedAt.Time,
			}

			block.States = append(block.States, state)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	blocks := make([]models.Block, 0, len(blocksMap))
	for _, block := range blocksMap {
		blocks = append(blocks, *block)
	}

	sort.Slice(blocks, func(i, j int) bool {
		return blocks[i].Position < blocks[j].Position
	})

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
	err := r.db.QueryRowContext(ctx, CREATE_BLOCK,
		block.NoteID, block.BlockTypeID, block.Position, block.Content, block.CreatedAt, block.UpdatedAt,
	).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	block.States = []models.BlockState{}
	return &block, nil
}

func (r *noteRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	var block models.Block

	err := r.db.QueryRowContext(ctx, GET_BLOCK_BY_ID, blockID).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	block.States = []models.BlockState{}
	return &block, nil
}

func (r *noteRepository) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string, updatedAt time.Time) (*models.Block, error) {
	var block models.Block

	err := r.db.QueryRowContext(ctx, UPDATE_BLOCK_CONTENT, blockID, content, updatedAt).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	block.States = []models.BlockState{}
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

	err = tx.QueryRowContext(ctx, UPDATE_BLOCK_POSITION, blockID, newPosition, updatedAt).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID, &updatedBlock.Position, &updatedBlock.Content, &updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
	)
	if err != nil {
		return nil, err
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
