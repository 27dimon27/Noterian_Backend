package repository

import (
	"context"
	"database/sql"
	"errors"
	"sort"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/notes"
	"github.com/google/uuid"
	"github.com/lib/pq"
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

		err := rows.Scan(&note.ID, &note.UserID, &note.Title, &parentID, &note.IsPublic, &note.CreatedAt, &note.UpdatedAt)
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
		&note.ID, &note.UserID, &note.Title, &parentID, &note.IsPublic, &note.CreatedAt, &note.UpdatedAt,
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
		var block models.Block

		err := rows.Scan(&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content, &block.CreatedAt, &block.UpdatedAt)
		if err != nil {
			return nil, err
		}

		blocks = append(blocks, block)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return blocks, nil
}

func (r *noteRepository) GetBlockType(ctx context.Context, blockTypeID int) (*models.BlockType, error) {
	var blockType models.BlockType
	err := r.db.QueryRowContext(ctx, "SELECT id, name FROM block_types WHERE id = $1", blockTypeID).Scan(&blockType.ID, &blockType.Name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &blockType, nil
}

func (r *noteRepository) CreateNote(ctx context.Context, note models.Note) (*models.Note, error) {
	parentID := sql.NullString{}
	if note.ParentID != nil {
		parentID = sql.NullString{
			String: note.ParentID.String(),
			Valid:  true,
		}
	}

	err := r.db.QueryRowContext(ctx, CREATE_NOTE, note.UserID, note.Title, parentID).Scan(
		&note.ID, &note.UserID, &note.Title, &note.ParentID, &note.IsPublic, &note.CreatedAt, &note.UpdatedAt,
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

	err := r.db.QueryRowContext(ctx, UPDATE_NOTE, noteID, note.Title, parentID, note.IsPublic).Scan(
		&updatedNote.ID,
		&updatedNote.UserID,
		&updatedNote.Title,
		&updatedNote.ParentID,
		&updatedNote.IsPublic,
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
	err := r.db.QueryRowContext(ctx, CREATE_BLOCK, block.NoteID, block.BlockTypeID, block.Position, block.Content).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &block, nil
}

func (r *noteRepository) GetBlock(ctx context.Context, blockID uuid.UUID) (*models.Block, error) {
	var block models.Block

	err := r.db.QueryRowContext(ctx, GET_BLOCK_BY_ID, blockID).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &block, nil
}

func (r *noteRepository) UpdateBlockContent(ctx context.Context, blockID uuid.UUID, content string) (*models.Block, error) {
	var block models.Block

	err := r.db.QueryRowContext(ctx, UPDATE_BLOCK_CONTENT, blockID, content).Scan(
		&block.ID, &block.NoteID, &block.BlockTypeID, &block.Position, &block.Content,
		&block.CreatedAt, &block.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, notes.ErrBlockNotFound
		}
		return nil, err
	}

	return &block, nil
}

func (r *noteRepository) MoveBlock(ctx context.Context, noteID uuid.UUID, blockID uuid.UUID, oldPosition int, newPosition int) (*models.Block, error) {
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
		_, err := tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_DOWN, noteID, oldPosition, newPosition)
		if err != nil {
			return nil, err
		}
	} else if oldPosition > newPosition {
		_, err := tx.ExecContext(ctx, UPDATE_BLOCKS_POSITION_UP, noteID, oldPosition, newPosition)
		if err != nil {
			return nil, err
		}
	}

	var updatedBlock models.Block

	err = tx.QueryRowContext(ctx, UPDATE_BLOCK_POSITION, blockID, newPosition).Scan(
		&updatedBlock.ID, &updatedBlock.NoteID, &updatedBlock.BlockTypeID, &updatedBlock.Position, &updatedBlock.Content,
		&updatedBlock.CreatedAt, &updatedBlock.UpdatedAt,
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

func (r *noteRepository) ShiftBlockPositions(ctx context.Context, noteID uuid.UUID, fromPosition int, direction int) error {
	if direction > 0 {
		_, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_UP, noteID, fromPosition)
		return err
	} else if direction < 0 {
		_, err := r.db.ExecContext(ctx, UPDATE_ALL_BLOCKS_POSITION_DOWN, noteID, fromPosition)
		return err
	}
	return nil
}

func (r *noteRepository) GetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.BlockFormatting, error) {
	rows, err := r.db.QueryContext(ctx, GET_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	formatting := &models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges:  []models.FormattingRange{},
	}

	for rows.Next() {
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		err := rows.Scan(&rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign)
		if err != nil {
			return nil, err
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

func (r *noteRepository) GetBlocksFormatting(ctx context.Context, blockIDs []uuid.UUID) (map[string]models.BlockFormatting, error) {
	if len(blockIDs) == 0 {
		return map[string]models.BlockFormatting{}, nil
	}

	rows, err := r.db.QueryContext(ctx, GET_BLOCKS_FORMATTING, pq.Array(blockIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]models.BlockFormatting)

	for rows.Next() {
		var blockIDStr string
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		err := rows.Scan(&blockIDStr, &rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign)
		if err != nil {
			return nil, err
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

		formatting, exists := result[blockIDStr]
		if !exists {
			formatting = models.BlockFormatting{
				BlockID: blockIDStr,
				Ranges:  []models.FormattingRange{},
			}
		}
		formatting.Ranges = append(formatting.Ranges, rng)
		result[blockIDStr] = formatting
	}

	for blockID, formatting := range result {
		sort.Slice(formatting.Ranges, func(i, j int) bool {
			if formatting.Ranges[i].StartPos != formatting.Ranges[j].StartPos {
				return formatting.Ranges[i].StartPos < formatting.Ranges[j].StartPos
			}
			return formatting.Ranges[i].EndPos < formatting.Ranges[j].EndPos
		})
		result[blockID] = formatting
	}

	return result, nil
}

func (r *noteRepository) UpdateBlockFormatting(ctx context.Context, blockID uuid.UUID, formattingRange models.FormattingRange) (*models.BlockFormatting, error) {
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

	existingRanges, err := r.getFormattingRangesInTx(ctx, tx, blockID)
	if err != nil {
		return nil, err
	}

	newRanges := applyFormattingToRanges(existingRanges, formattingRange)

	_, err = tx.ExecContext(ctx, DELETE_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, err
	}

	if len(newRanges) > 0 {
		for _, rng := range newRanges {
			_, err = tx.ExecContext(ctx, INSERT_BLOCK_FORMATTING,
				blockID, rng.StartPos, rng.EndPos, rng.Bold, rng.Italic, rng.Underline, rng.TextAlign)
			if err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetBlockFormatting(ctx, blockID)
}

func (r *noteRepository) ResetBlockFormatting(ctx context.Context, blockID uuid.UUID) (*models.BlockFormatting, error) {
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

	_, err = tx.ExecContext(ctx, DELETE_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetBlockFormatting(ctx, blockID)
}

func (r *noteRepository) GetSubnotes(ctx context.Context, noteID uuid.UUID) ([]models.Note, error) {
	rows, err := r.db.QueryContext(ctx, GET_SUBNOTES_BY_NOTE, noteID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subnotes []models.Note

	for rows.Next() {
		var subnote models.Note

		err := rows.Scan(&subnote.ID, &subnote.UserID, &subnote.Title, &subnote.ParentID, &subnote.CreatedAt, &subnote.UpdatedAt)
		if err != nil {
			return nil, err
		}

		subnotes = append(subnotes, subnote)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return subnotes, nil
}

func (r *noteRepository) getFormattingRangesInTx(ctx context.Context, tx *sql.Tx, blockID uuid.UUID) ([]models.FormattingRange, error) {
	rows, err := tx.QueryContext(ctx, GET_BLOCK_FORMATTING, blockID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ranges []models.FormattingRange

	for rows.Next() {
		var rng models.FormattingRange
		var bold, italic, underline *bool
		var textAlign *int

		err := rows.Scan(&rng.StartPos, &rng.EndPos, &bold, &italic, &underline, &textAlign)
		if err != nil {
			return nil, err
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
