// dto/dto_test.go
package dto

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestToBlockDTO(t *testing.T) {
	now := time.Now()
	blockID := uuid.New()
	noteID := uuid.New()

	block := models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    2,
		Content:     "Test content",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result := ToBlockDTO(block)

	assert.Equal(t, blockID, result.ID)
	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, 1, result.BlockTypeID)
	assert.Equal(t, 2, result.Position)
	assert.Equal(t, "Test content", result.Content)
	assert.Equal(t, now, result.CreatedAt)
	assert.Equal(t, now, result.UpdatedAt)
}

func TestToBlockWithFormattingDTO(t *testing.T) {
	now := time.Now()
	blockID := uuid.New()
	noteID := uuid.New()

	block := models.Block{
		ID:          blockID,
		NoteID:      noteID,
		BlockTypeID: 1,
		Position:    0,
		Content:     "Content",
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	formatting := models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{StartPos: 0, EndPos: 5, Bold: boolPtr(true)},
		},
	}

	result := ToBlockWithFormattingDTO(block, formatting)

	assert.Equal(t, blockID, result.Block.ID)
	assert.Equal(t, noteID, result.Block.NoteID)
	assert.Equal(t, formatting.BlockID, result.Formatting.BlockID)
	assert.Len(t, result.Formatting.Ranges, 1)
}

func TestFromBlockRequestDTO(t *testing.T) {
	noteID := uuid.New()
	req := BlockRequest{
		NoteID:      noteID,
		BlockTypeID: 2,
		Position:    5,
		Content:     "Request content",
	}

	result := FromBlockRequestDTO(req)

	assert.Equal(t, noteID, result.NoteID)
	assert.Equal(t, 2, result.BlockTypeID)
	assert.Equal(t, 5, result.Position)
	assert.Equal(t, "Request content", result.Content)
}

func TestToFormattingRangeDTO(t *testing.T) {
	rng := models.FormattingRange{
		StartPos:  10,
		EndPos:    20,
		Bold:      boolPtr(true),
		Italic:    boolPtr(false),
		Underline: boolPtr(true),
		TextAlign: intPtr(1),
	}

	result := ToFormattingRangeDTO(rng)

	assert.Equal(t, 10, result.StartPos)
	assert.Equal(t, 20, result.EndPos)
	assert.Equal(t, true, *result.Bold)
	assert.Equal(t, false, *result.Italic)
	assert.Equal(t, true, *result.Underline)
	assert.Equal(t, 1, *result.TextAlign)
}

func TestFromFormattingRangeDTO(t *testing.T) {
	dto := FormattingRange{
		StartPos:  5,
		EndPos:    15,
		Bold:      boolPtr(false),
		Italic:    boolPtr(true),
		Underline: boolPtr(false),
		TextAlign: intPtr(2),
	}

	result := FromFormattingRangeDTO(dto)

	assert.Equal(t, 5, result.StartPos)
	assert.Equal(t, 15, result.EndPos)
	assert.Equal(t, false, *result.Bold)
	assert.Equal(t, true, *result.Italic)
	assert.Equal(t, false, *result.Underline)
	assert.Equal(t, 2, *result.TextAlign)
}

func TestToBlockFormattingDTO(t *testing.T) {
	blockID := uuid.New()
	formatting := models.BlockFormatting{
		BlockID: blockID.String(),
		Ranges: []models.FormattingRange{
			{StartPos: 0, EndPos: 3, Bold: boolPtr(true)},
			{StartPos: 4, EndPos: 7, Italic: boolPtr(true)},
		},
	}

	result := ToBlockFormattingDTO(formatting)

	assert.Equal(t, blockID.String(), result.BlockID)
	assert.Len(t, result.Ranges, 2)
}

func TestToNoteDTO(t *testing.T) {
	now := time.Now()
	noteID := uuid.New()
	userID := uuid.New()
	parentID := uuid.New()

	note := &models.Note{
		ID:         noteID,
		UserID:     userID,
		Title:      "Test Note",
		ParentID:   &parentID,
		IsPublic:   true,
		IsFavorite: false,
		Icon:       "📝",
		HeaderURL:  "http://example.com/header.jpg",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	result := ToNoteDTO(note)

	assert.Equal(t, noteID, result.ID)
	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "Test Note", result.Title)
	assert.Equal(t, parentID, *result.ParentID)
	assert.True(t, result.IsPublic)
	assert.False(t, result.IsFavorite)
	assert.Equal(t, "📝", result.Icon)
	assert.Equal(t, "http://example.com/header.jpg", result.HeaderURL)
}

func TestToNoteDTO_NilParentID(t *testing.T) {
	note := &models.Note{
		ID:       uuid.New(),
		UserID:   uuid.New(),
		Title:    "No Parent",
		ParentID: nil,
	}

	result := ToNoteDTO(note)

	assert.Nil(t, result.ParentID)
}

func TestToSubnoteDTO(t *testing.T) {
	note := models.Note{
		ID:    uuid.New(),
		Title: "Subnote",
	}
	blockID := uuid.New()

	result := ToSubnoteDTO(note, blockID)

	assert.Equal(t, note.ID, result.Note.ID)
	assert.Equal(t, note.Title, result.Note.Title)
	assert.Equal(t, blockID, result.BlockID)
}

func TestToSubnotesDTO(t *testing.T) {
	subnotes := []models.Note{
		{ID: uuid.New(), Title: "Subnote 1"},
		{ID: uuid.New(), Title: "Subnote 2"},
	}

	result := ToSubnotesDTO(subnotes)

	assert.Len(t, result, 2)
	assert.Equal(t, subnotes[0].ID, result[0].ID)
	assert.Equal(t, subnotes[1].ID, result[1].ID)
}

func TestFromNoteRequestDTO(t *testing.T) {
	userID := uuid.New()
	parentID := uuid.New()
	req := NoteRequest{
		UserID:     userID,
		Title:      "New Note",
		ParentID:   &parentID,
		IsPublic:   true,
		IsFavorite: true,
		Icon:       "⭐",
	}

	result := FromNoteRequestDTO(req)

	assert.Equal(t, userID, result.UserID)
	assert.Equal(t, "New Note", result.Title)
	assert.Equal(t, parentID, *result.ParentID)
	assert.True(t, result.IsPublic)
	assert.True(t, result.IsFavorite)
	assert.Equal(t, "⭐", result.Icon)
}

func TestToNoteResponse(t *testing.T) {
	note := &models.Note{
		ID:    uuid.New(),
		Title: "Note with Blocks",
	}
	blocks := []models.Block{
		{ID: uuid.New(), Content: "Block 1"},
		{ID: uuid.New(), Content: "Block 2"},
	}
	blockFormattings := map[string]models.BlockFormatting{
		blocks[0].ID.String(): {BlockID: blocks[0].ID.String(), Ranges: []models.FormattingRange{}},
		blocks[1].ID.String(): {BlockID: blocks[1].ID.String(), Ranges: []models.FormattingRange{}},
	}

	result := ToNoteResponse(note, blocks, blockFormattings)

	assert.Equal(t, note.ID, result.Note.ID)
	assert.Len(t, result.Blocks, 2)
}

func TestToNotesResponse(t *testing.T) {
	notes := []models.Note{
		{ID: uuid.New(), Title: "Note 1"},
		{ID: uuid.New(), Title: "Note 2"},
		{ID: uuid.New(), Title: "Note 3"},
	}

	result := ToNotesResponse(notes)

	assert.Len(t, result.Notes, 3)
	assert.Equal(t, 3, result.Total)
}

func TestToPublicNoteResponse(t *testing.T) {
	note := models.Note{
		ID:    uuid.New(),
		Title: "Public Note",
		Icon:  "🌐",
	}

	result := ToPublicNoteResponse(note)

	assert.Equal(t, note.ID, result.ID)
	assert.Equal(t, "Public Note", result.Title)
	assert.Equal(t, "🌐", result.Icon)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
