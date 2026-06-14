package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"sort"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/types"
	"github.com/google/uuid"
	"github.com/lib/pq"
)

type noteRepository struct {
	db     *sql.DB
	logger *slog.Logger
}

func NewNoteRepository(db *sql.DB, logger *slog.Logger) *noteRepository {
	return &noteRepository{
		db:     db,
		logger: logger,
	}
}

func (r *noteRepository) GetNotes(ctx context.Context, userID uuid.UUID) (userNotes []models.Note, appErr types.AppErrorInterface) {
	rows, err := r.db.QueryContext(ctx, GET_NOTES_BY_USER, userID)
	if err != nil {
		return nil, &types.AppError{
			Err:        notes.ErrInternalServer,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetNotes",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetNotes",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetNotes",
				}
			}
		}
	}()

	for rows.Next() {
		var note models.Note
		var parentID sql.NullString

		if err := rows.Scan(&note.ID, &note.UserID, &note.Title, &parentID, &note.IsPublic, &note.IsFavorite, &note.Icon, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, &types.AppError{
				Err:        notes.ErrInternalServer,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetNotes",
			}
		}

		if parentID.Valid {
			pid, err := uuid.Parse(parentID.String)
			if err != nil {
				return nil, &types.AppError{
					Err:        err,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetNotes",
				}
			}
			note.ParentID = &pid
		}

		userNotes = append(userNotes, note)
	}

	if err = rows.Err(); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetNotes",
		}
	}

	return userNotes, nil
}

func (r *noteRepository) GetNote(ctx context.Context, noteID uuid.UUID) (*models.Note, types.AppErrorInterface) {
	var note models.Note
	var parentID sql.NullString

	if err := r.db.QueryRowContext(ctx, GET_NOTE_BY_ID, noteID).Scan(
		&note.ID, &note.UserID, &note.Title, &parentID, &note.IsPublic, &note.IsFavorite, &note.Icon, &note.CreatedAt, &note.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrNoteNotFound,
				PublicMsg:  notes.PublicMsgErrNoteNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetNote",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetNote",
		}
	}

	if parentID.Valid {
		pid, err := uuid.Parse(parentID.String)
		if err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetNote",
			}
		}
		note.ParentID = &pid
	}

	return &note, nil
}

func (r *noteRepository) GetBlocks(ctx context.Context, noteID uuid.UUID) (blocks []models.Block, appErr types.AppErrorInterface) {
	rows, err := r.db.QueryContext(ctx, GET_BLOCKS_BY_NOTE, noteID)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlocks",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlocks",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlocks",
				}
			}
		}
	}()

	for rows.Next() {
		var block models.Block

		if err := rows.Scan(&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetBlocks",
			}
		}

		blocks = append(blocks, block)
	}

	if err := rows.Err(); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlocks",
		}
	}

	return blocks, nil
}

func (r *noteRepository) GetBlockType(ctx context.Context, blockTypeID int) (*models.BlockType, types.AppErrorInterface) {
	var blockType models.BlockType

	if err := r.db.QueryRowContext(ctx, "SELECT id, name FROM block_types WHERE id = $1", blockTypeID).Scan(&blockType.ID, &blockType.Name); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrBlockTypeNotFound,
				PublicMsg:  notes.PublicMsgErrBlockTypeNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetBlockType",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlockType",
		}
	}
	return &blockType, nil
}

func (r *noteRepository) CreateNote(ctx context.Context, note models.Note) (*models.Note, types.AppErrorInterface) {
	parentID := sql.NullString{}

	if note.ParentID != nil {
		parentID = sql.NullString{
			String: note.ParentID.String(),
			Valid:  true,
		}
	}

	if err := r.db.QueryRowContext(ctx, CREATE_NOTE, note.UserID, note.Title, parentID, note.IsPublic, note.IsFavorite, note.Icon).Scan(
		&note.ID, &note.UserID, &note.Title, &note.ParentID, &note.IsPublic, &note.IsFavorite, &note.Icon, &note.CreatedAt, &note.UpdatedAt,
	); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "CreateNote",
		}
	}

	return &note, nil
}

func (r *noteRepository) UpdateNote(ctx context.Context, noteID uuid.UUID, note models.Note) (*models.Note, types.AppErrorInterface) {
	parentID := sql.NullString{}

	if note.ParentID != nil {
		parentID = sql.NullString{
			String: note.ParentID.String(),
			Valid:  true,
		}
	}

	updatedNote := &models.Note{}

	if err := r.db.QueryRowContext(ctx, UPDATE_NOTE, noteID, note.Title, parentID, note.IsPublic, note.IsFavorite, note.Icon).Scan(
		&updatedNote.ID,
		&updatedNote.UserID,
		&updatedNote.Title,
		&updatedNote.ParentID,
		&updatedNote.IsPublic,
		&updatedNote.IsFavorite,
		&updatedNote.Icon,
		&updatedNote.CreatedAt,
		&updatedNote.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrNoteNotFound,
				PublicMsg:  notes.PublicMsgErrNoteNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "UpdateNote",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateNote",
		}
	}

	return updatedNote, nil
}

func (r *noteRepository) DeleteNote(ctx context.Context, noteID uuid.UUID) types.AppErrorInterface {
	var id uuid.UUID

	if err := r.db.QueryRowContext(ctx, DELETE_NOTE, noteID).Scan(&id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &types.AppError{
				Err:        notes.ErrNoteNotFound,
				PublicMsg:  notes.PublicMsgErrNoteNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteNote",
			}
		}
		return &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteNote",
		}
	}

	return nil
}

func (r *noteRepository) CreateBlock(ctx context.Context, block models.Block) (*models.Block, types.AppErrorInterface) {
	if err := r.db.QueryRowContext(ctx, CREATE_BLOCK, block.NoteID, block.BlockTypeID, block.Position, block.Content).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "CreateBlock",
		}
	}

	return &block, nil
}

func (r *noteRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, types.AppErrorInterface) {
	var block models.Block

	if err := r.db.QueryRowContext(ctx, GET_BLOCK_BY_ID, blockID).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrBlockNotFound,
				PublicMsg:  notes.PublicMsgErrBlockNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "GetBlock",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlock",
		}
	}

	return &block, nil
}

func (r *noteRepository) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string) (*models.Block, types.AppErrorInterface) {
	var block models.Block

	if err := r.db.QueryRowContext(ctx, UPDATE_BLOCK_CONTENT, blockID, content).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrBlockNotFound,
				PublicMsg:  notes.PublicMsgErrBlockNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "UpdateBlockContent",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateBlockContent",
		}
	}

	return &block, nil
}

func (r *noteRepository) MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int) (updatedBlock *models.Block, appErr types.AppErrorInterface) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "MoveBlock",
		}
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), rollbackErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "MoveBlock",
				}
			} else {
				appErr = &types.AppError{
					Err:        rollbackErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "MoveBlock",
				}
			}
		}
	}()

	if oldPosition < newPosition {
		if _, err = tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_DOWN, noteID, oldPosition, newPosition); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "MoveBlock",
			}
		}
	} else if oldPosition > newPosition {
		if _, err = tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_UP, noteID, oldPosition, newPosition); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "MoveBlock",
			}
		}
	}

	if err = tx.QueryRowContext(ctx, UPDATE_BLOCK_POSITION, blockID, newPosition).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID, &updatedBlock.Position, &updatedBlock.Content,
		&updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
	); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "MoveBlock",
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "MoveBlock",
		}
	}

	return updatedBlock, nil
}

func (r *noteRepository) DeleteBlock(ctx context.Context, blockID uuid.UUID) (*uuid.UUID, types.AppErrorInterface) {
	var deletedBlockID uuid.UUID
	var noteID uuid.UUID

	if err := r.db.QueryRowContext(ctx, DELETE_BLOCK, blockID).Scan(&deletedBlockID, &noteID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, &types.AppError{
				Err:        notes.ErrBlockNotFound,
				PublicMsg:  notes.PublicMsgErrBlockNotFound,
				StatusCode: 404,
				Layer:      "repo",
				Op:         "DeleteBlock",
			}
		}
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "DeleteBlock",
		}
	}

	return &noteID, nil
}

func (r *noteRepository) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) types.AppErrorInterface {
	if direction > 0 {
		if _, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_UP, noteID, fromPosition); err != nil {
			return &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "ShiftBlockPositions",
			}
		}
		return nil
	} else if direction < 0 {
		if _, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_DOWN, noteID, fromPosition); err != nil {
			return &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "ShiftBlockPositions",
			}
		}
		return nil
	}

	return nil
}

func (r *noteRepository) GetBlockFormatting(ctx context.Context, blockID uuid.UUID) (formatting *models.BlockFormatting, appErr types.AppErrorInterface) {
	rows, err := r.db.QueryContext(ctx, GET_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlockFormatting",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlockFormatting",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlockFormatting",
				}
			}
		}
	}()

	formatting = &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	for rows.Next() {
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		if err := rows.Scan(&rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetBlockFormatting",
			}
		}

		if bold != nil {
			rng.Bold = bold
		}
		if italic != nil {
			rng.Italic = italic
		}
		if underline != nil {
			rng.Underline = underline
		}
		if textAlign != nil {
			rng.TextAlign = textAlign
		}

		formatting.Ranges = append(formatting.Ranges, rng)
	}

	return formatting, nil
}

func (r *noteRepository) GetBlocksFormatting(ctx context.Context, blockIDs []uuid.UUID) (formattings map[string]models.BlockFormatting, appErr types.AppErrorInterface) {
	if len(blockIDs) == 0 {
		return map[string]models.BlockFormatting{}, nil
	}

	rows, err := r.db.QueryContext(ctx, GET_BLOCKS_FORMATTING, pq.Array(blockIDs))
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetBlocksFormatting",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlocksFormatting",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetBlocksFormatting",
				}
			}
		}
	}()

	for rows.Next() {
		var blockIDStr string
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		err := rows.Scan(&blockIDStr, &rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign)
		if err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetBlocksFormatting",
			}
		}

		if bold != nil {
			rng.Bold = bold
		}
		if italic != nil {
			rng.Italic = italic
		}
		if underline != nil {
			rng.Underline = underline
		}
		if textAlign != nil {
			rng.TextAlign = textAlign
		}

		formatting, exists := formattings[blockIDStr]
		if !exists {
			formatting = models.BlockFormatting{
				BlockID: blockIDStr,
				Ranges:  []models.FormattingRange{},
			}
		}

		formatting.Ranges = append(formatting.Ranges, rng)
		formattings[blockIDStr] = formatting
	}

	for blockID, formatting := range formattings {
		sort.Slice(formatting.Ranges, func(i, j int) bool {
			if formatting.Ranges[i].StartPos != formatting.Ranges[j].StartPos {
				return formatting.Ranges[i].StartPos < formatting.Ranges[j].StartPos
			}
			return formatting.Ranges[i].EndPos < formatting.Ranges[j].EndPos
		})
		formattings[blockID] = formatting
	}

	return formattings, nil
}

func (r *noteRepository) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingRange models.FormattingRange) (formatting *models.BlockFormatting, appErr types.AppErrorInterface) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateBlockFormatting",
		}
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), rollbackErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "UpdateBlockFormatting",
				}
			} else {
				appErr = &types.AppError{
					Err:        rollbackErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "UpdateBlockFormatting",
				}
			}
		}
	}()

	existingRanges, customErr := r.getFormattingRangesInTx(ctx, tx, blockID)
	if customErr != nil {
		return nil, customErr
	}

	newRanges := applyFormattingToRanges(existingRanges, formattingRange)

	if _, err := tx.ExecContext(ctx, DELETE_BLOCK_FORMATTING, blockID); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateBlockFormatting",
		}
	}

	if len(newRanges) > 0 {
		for _, rng := range newRanges {
			if _, err := tx.ExecContext(
				ctx, INSERT_BLOCK_FORMATTING, blockID, rng.StartPos, rng.EndPos, rng.Bold, rng.Italic, rng.Underline, rng.TextAlign,
			); err != nil {
				return nil, &types.AppError{
					Err:        err,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "UpdateBlockFormatting",
				}
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "UpdateBlockFormatting",
		}
	}

	formatting, customErr = r.GetBlockFormatting(ctx, blockID)
	if customErr != nil {
		return nil, customErr
	}

	return formatting, nil
}

func (r *noteRepository) GetSubnotes(ctx context.Context, noteID uuid.UUID) (subnotes []models.Note, appErr types.AppErrorInterface) {
	rows, err := r.db.QueryContext(ctx, GET_SUBNOTES_BY_NOTE, noteID)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetSubnotes",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetSubnotes",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "GetSubnotes",
				}
			}
		}
	}()

	for rows.Next() {
		var subnote models.Note

		if err := rows.Scan(
			&subnote.ID, &subnote.UserID, &subnote.Title, &subnote.ParentID, &subnote.IsPublic, &subnote.IsFavorite, &subnote.Icon, &subnote.CreatedAt, &subnote.UpdatedAt,
		); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "GetSubnotes",
			}
		}

		subnotes = append(subnotes, subnote)
	}

	if err = rows.Err(); err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "GetSubnotes",
		}
	}

	return subnotes, nil
}

func (r *noteRepository) getFormattingRangesInTx(ctx context.Context, tx *sql.Tx, blockID uuid.UUID) (ranges []models.FormattingRange, appErr types.AppErrorInterface) {
	rows, err := tx.QueryContext(ctx, GET_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, &types.AppError{
			Err:        err,
			PublicMsg:  notes.PublicMsgErrInternalServer,
			StatusCode: 500,
			Layer:      "repo",
			Op:         "getFormattingRangesInTx",
		}
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			if appErr != nil {
				appErr = &types.AppError{
					Err:        errors.Join(appErr.Unwrap(), closeErr),
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "getFormattingRangesInTx",
				}
			} else {
				appErr = &types.AppError{
					Err:        closeErr,
					PublicMsg:  notes.PublicMsgErrInternalServer,
					StatusCode: 500,
					Layer:      "repo",
					Op:         "getFormattingRangesInTx",
				}
			}
		}
	}()

	for rows.Next() {
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		if err := rows.Scan(&rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign); err != nil {
			return nil, &types.AppError{
				Err:        err,
				PublicMsg:  notes.PublicMsgErrInternalServer,
				StatusCode: 500,
				Layer:      "repo",
				Op:         "getFormattingRangesInTx",
			}
		}

		if bold != nil {
			rng.Bold = bold
		}
		if italic != nil {
			rng.Italic = italic
		}
		if underline != nil {
			rng.Underline = underline
		}
		if textAlign != nil {
			rng.TextAlign = textAlign
		}

		ranges = append(ranges, rng)
	}

	sort.Slice(ranges, func(i, j int) bool {
		if ranges[i].StartPos != ranges[j].StartPos {
			return ranges[i].StartPos < ranges[j].StartPos
		}
		return ranges[i].EndPos < ranges[j].EndPos
	})

	return ranges, nil
}

func applyFormattingToRanges(existingRanges []models.FormattingRange, newRange models.FormattingRange) []models.FormattingRange {
	points := make(map[int]bool)

	for _, r := range existingRanges {
		points[r.StartPos] = true
		points[r.EndPos] = true
	}

	points[newRange.StartPos] = true
	points[newRange.EndPos] = true

	pointList := make([]int, 0, len(points))
	for p := range points {
		pointList = append(pointList, p)
	}

	sort.Ints(pointList)

	segments := make([]struct {
		start int
		end   int
	}, 0, len(pointList)-1)

	for i := 0; i < len(pointList)-1; i++ {
		if pointList[i] < pointList[i+1] {
			segments = append(segments, struct {
				start int
				end   int
			}{start: pointList[i], end: pointList[i+1]})
		}
	}

	result := make([]models.FormattingRange, 0, len(segments))

	for _, segment := range segments {
		var bold, italic, underline bool
		textAlign := 0
		hasTextAlign := false

		for _, r := range existingRanges {
			if segment.start >= r.StartPos && segment.end <= r.EndPos {
				if r.Bold != nil {
					bold = *r.Bold
				}
				if r.Italic != nil {
					italic = *r.Italic
				}
				if r.Underline != nil {
					underline = *r.Underline
				}

				if r.TextAlign != nil {
					textAlign = *r.TextAlign
					hasTextAlign = true
				}
			}
		}

		if segment.start >= newRange.StartPos && segment.end <= newRange.EndPos {
			if newRange.Bold != nil {
				bold = *newRange.Bold
			}
			if newRange.Italic != nil {
				italic = *newRange.Italic
			}
			if newRange.Underline != nil {
				underline = *newRange.Underline
			}

			if newRange.TextAlign != nil {
				textAlign = *newRange.TextAlign
				hasTextAlign = true
			}
		}

		if bold || italic || underline || hasTextAlign {
			result = append(result, models.FormattingRange{
				StartPos:  segment.start,
				EndPos:    segment.end,
				Bold:      &bold,
				Italic:    &italic,
				Underline: &underline,
				TextAlign: &textAlign,
			})
		}
	}

	merged := make([]models.FormattingRange, 0, len(result))
	for i := 0; i < len(result); i++ {
		if len(merged) == 0 {
			merged = append(merged, result[i])
			continue
		}

		last := &merged[len(merged)-1]
		current := result[i]

		sameBold := *last.Bold == *current.Bold
		sameItalic := *last.Italic == *current.Italic
		sameUnderline := *last.Underline == *current.Underline
		sameTextAlign := *last.TextAlign == *current.TextAlign

		if last.EndPos >= current.StartPos && sameBold && sameItalic && sameUnderline && sameTextAlign {
			if last.EndPos >= current.EndPos {
				continue
			}
			last.EndPos = current.EndPos
		} else {
			merged = append(merged, current)
		}
	}

	return merged
}
