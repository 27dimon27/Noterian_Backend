package dto

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToNoteDTO(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	note := &models.Note{
		ID:        id,
		UserID:    userID,
		Title:     "Test Note",
		ParentID:  &parentID,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	result := ToNoteDTO(note)

	assert.Equal(t, id, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "Test Note", result.Title)
	assert.Equal(t, &parentID, result.ParentID)
	assert.Equal(t, createdAt, result.CreatedAt)
	assert.Equal(t, updatedAt, result.UpdatedAt)
}

func TestFromNoteRequestDTO(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()
	req := NoteRequest{
		UserID:   userID,
		Title:    "New Note",
		ParentID: &parentID,
	}

	result := FromNoteRequestDTO(req)

	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "New Note", result.Title)
	assert.Equal(t, &parentID, result.ParentID)
}

func TestToBlockDTO(t *testing.T) {
	id := uuid.New()
	noteID := uuid.New()
	createdAt := time.Now()
	updatedAt := time.Now()

	block := models.Block{
		ID:          id,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Content",
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
	}

	result := ToBlockDTO(block)

	assert.Equal(t, id, result.ID)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, 1, result.BlockTypeID)
	assert.Equal(t, 0, result.Position)
	assert.Equal(t, "Content", result.Content)
	assert.Equal(t, createdAt, result.CreatedAt)
	assert.Equal(t, updatedAt, result.UpdatedAt)
}

func TestFromBlockRequestDTO(t *testing.T) {
	noteID := uuid.New()
	req := BlockRequest{
		NoteID:      noteID,
		BlockTypeID: 2,
		Position:    1,
		Content:     "Block Content",
	}

	result := FromBlockRequestDTO(req)

	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, 2, result.BlockTypeID)
	assert.Equal(t, 1, result.Position)
	assert.Equal(t, "Block Content", result.Content)
}

func TestToNotesResponse(t *testing.T) {
	notes := []models.Note{
		{ID: uuid.New(), Title: "Note 1"},
		{ID: uuid.New(), Title: "Note 2"},
	}

	result := ToNotesResponse(notes)

	assert.Equal(t, 2, result.Total)
	assert.Len(t, result.Notes, 2)
	assert.Equal(t, "Note 1", result.Notes[0].Title)
	assert.Equal(t, "Note 2", result.Notes[1].Title)
}

func TestToFormattingRangeDTO(t *testing.T) {
	bold := true
	italic := false
	textAlign := 1

	rng := models.FormattingRange{
		StartPos:  0,
		EndPos:    10,
		Bold:      &bold,
		Italic:    &italic,
		Underline: nil,
		TextAlign: &textAlign,
	}

	result := ToFormattingRangeDTO(rng)

	assert.Equal(t, 0, result.StartPos)
	assert.Equal(t, 10, result.EndPos)
	assert.Equal(t, &bold, result.Bold)
	assert.Equal(t, &italic, result.Italic)
	assert.Nil(t, result.Underline)
	assert.Equal(t, &textAlign, result.TextAlign)
}

func TestFromFormattingRangeDTO(t *testing.T) {
	bold := true
	textAlign := 2

	dto := FormattingRange{
		StartPos:  5,
		EndPos:    15,
		Bold:      &bold,
		Italic:    nil,
		Underline: nil,
		TextAlign: &textAlign,
	}

	result := FromFormattingRangeDTO(dto)

	assert.Equal(t, 5, result.StartPos)
	assert.Equal(t, 15, result.EndPos)
	assert.Equal(t, &bold, result.Bold)
	assert.Nil(t, result.Italic)
	assert.Equal(t, &textAlign, result.TextAlign)
}

func TestToBlockFormattingDTO(t *testing.T) {
	blockID := uuid.New()
	bold := true
	italic := false

	formatting := models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{
				StartPos: 0,
				EndPos:   5,
				Bold:     &bold,
				Italic:   &italic,
			},
			{
				StartPos: 6,
				EndPos:   10,
				Bold:     nil,
				Italic:   nil,
			},
		},
	}

	result := ToBlockFormattingDTO(formatting)

	assert.Equal(t, blockID.String(), result.BlockID)
	assert.Len(t, result.Ranges, 2)
	assert.Equal(t, 0, result.Ranges[0].StartPos)
	assert.Equal(t, 5, result.Ranges[0].EndPos)
	assert.Equal(t, &bold, result.Ranges[0].Bold)
	assert.Equal(t, &italic, result.Ranges[0].Italic)
	assert.Equal(t, 6, result.Ranges[1].StartPos)
	assert.Equal(t, 10, result.Ranges[1].EndPos)
}

func TestToNoteResponse(t *testing.T) {
	note := &models.Note{
		ID:    uuid.New(),
		Title: "Test Note",
	}
	blocks := []models.Block{
		{ID: uuid.New(), Content: "Block 1"},
		{ID: uuid.New(), Content: "Block 2"},
	}
	formattings := make(map[string]models.BlockFormatting)

	result := ToNoteResponse(note, blocks, formattings)

	assert.Equal(t, "Test Note", result.Note.Title)
	assert.Len(t, result.Blocks, 2)
	assert.Equal(t, "Block 1", result.Blocks[0].Block.Content)
	assert.Equal(t, "Block 2", result.Blocks[1].Block.Content)
}
