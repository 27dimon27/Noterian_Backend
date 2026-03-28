package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sort"

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

func (r *noteRepository) GetNotesByUserID(ctx context.Context, userID uuid.UUID) ([]models.Note, error) {
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

func (r *noteRepository) GetNoteByID(ctx context.Context, noteID uuid.UUID) (*models.Note, error) {
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

func (r *noteRepository) GetBlocksByNoteID(ctx context.Context, noteID uuid.UUID) ([]models.Block, error) {
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
