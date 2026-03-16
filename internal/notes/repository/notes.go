package repository

import (
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
)

type NoteRepository interface {
	GetNotesByUserID(userID uuid.UUID) ([]models.Note, error)
	GetNoteByID(noteID uuid.UUID) (*models.Note, error)
	GetBlocksWithStatesByNoteID(noteID uuid.UUID) ([]map[string]interface{}, error)
}

type noteRepository struct {
	db *sql.DB
}

func NewNoteRepository(db *sql.DB) NoteRepository {
	return &noteRepository{db: db}
}

func (r *noteRepository) GetNotesByUserID(userID uuid.UUID) ([]models.Note, error) {
	rows, err := r.db.Query(
		"SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE user_id = $1 ORDER BY updated_at DESC",
		userID,
	)
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

func (r *noteRepository) GetNoteByID(noteID uuid.UUID) (*models.Note, error) {
	var note models.Note
	var parentID sql.NullString

	err := r.db.QueryRow(
		"SELECT id, user_id, title, parent_id, created_at, updated_at FROM notes WHERE id = $1",
		noteID,
	).Scan(&note.ID, &note.UserID, &note.Title, &parentID, &note.CreatedAt, &note.UpdatedAt)
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

func (r *noteRepository) GetBlocksWithStatesByNoteID(noteID uuid.UUID) ([]map[string]interface{}, error) {
	rows, err := r.db.Query(`
		SELECT 
			b.id, b.note_id, b.block_type_id, b.position, b.content,
			bs.id, bs.formatting, bs.created_at, bs.updated_at
		FROM blocks b
		LEFT JOIN block_states bs ON b.id = bs.block_id
		WHERE b.note_id = $1
		ORDER BY b.position, bs.created_at
	`, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	blocksMap := make(map[string]map[string]interface{})

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

		blockKey := blockID.String()
		if _, exists := blocksMap[blockKey]; !exists {
			blocksMap[blockKey] = map[string]interface{}{
				"id":       blockID.String(),
				"note_id":  noteID.String(),
				"type_id":  blockTypeID,
				"position": position,
				"content":  content,
				"states":   []map[string]interface{}{},
			}
		}

		if stateID.Valid {
			var formattingData map[string]interface{}
			if err := json.Unmarshal([]byte(formatting.String), &formattingData); err != nil {
				formattingData = map[string]interface{}{"format": "text"}
			}

			state := map[string]interface{}{
				"ID":         stateID.String,
				"BlockID":    blockID.String(),
				"Formatting": formattingData,
				"CreatedAt":  stateCreatedAt.Time,
				"UpdatedAt":  stateUpdatedAt.Time,
			}

			states := blocksMap[blockKey]["states"].([]map[string]interface{})
			blocksMap[blockKey]["states"] = append(states, state)
		}
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	blocks := make([]map[string]interface{}, 0, len(blocksMap))
	for _, block := range blocksMap {
		blocks = append(blocks, block)
	}

	return blocks, nil
}
