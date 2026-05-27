// pdf/generator_test.go
package pdf

import (
	"testing"
	"time"

	"github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratePDF_BasicNote(t *testing.T) {
	noteID := uuid.New()
	userID := uuid.New()
	now := time.Now()

	noteContent := &NoteContent{
		Note: &models.Note{
			ID:        noteID,
			UserID:    userID,
			Title:     "Test PDF Note",
			Icon:      "📝",
			CreatedAt: now,
			UpdatedAt: now,
		},
		Blocks: []models.Block{
			{
				ID:          uuid.New(),
				BlockTypeID: 1,
				Position:    0,
				Content:     "Hello, this is a text block.",
			},
			{
				ID:          uuid.New(),
				BlockTypeID: 1,
				Position:    1,
				Content:     "This is another paragraph with some **bold** and *italic* text.",
			},
		},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
	assert.Greater(t, pdfBuffer.Len(), 0)
}

func TestGeneratePDF_WithHeader(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Note with Header",
		},
		Blocks:     []models.Block{},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "https://example.com/header.jpg",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_WithImageBlock(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Note with Image",
		},
		Blocks: []models.Block{
			{
				ID:          uuid.New(),
				BlockTypeID: 2,
				Position:    0,
				Content:     "https://example.com/image.jpg",
			},
		},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_WithSubnote(t *testing.T) {
	subnoteID := uuid.New()
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Parent Note with Subnote",
		},
		Blocks: []models.Block{
			{
				ID:          uuid.New(),
				BlockTypeID: 5,
				Position:    0,
				Content:     subnoteID.String(),
			},
		},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes: map[string]models.Note{
			subnoteID.String(): {
				ID:    subnoteID,
				Title: "Embedded Subnote Content",
				Icon:  "📌",
			},
		},
		HeaderURL: "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_WithFormatting(t *testing.T) {
	blockID := uuid.New()
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Note with Formatting",
		},
		Blocks: []models.Block{
			{
				ID:          blockID,
				BlockTypeID: 1,
				Position:    0,
				Content:     "This text has various formatting applied to it.",
			},
		},
		Formatting: map[string]models.BlockFormatting{
			blockID.String(): {
				BlockID: blockID.String(),
				Ranges: []models.FormattingRange{
					{StartPos: 0, EndPos: 4, Bold: boolPtr(true)},
					{StartPos: 5, EndPos: 9, Italic: boolPtr(true)},
					{StartPos: 10, EndPos: 14, Underline: boolPtr(true)},
				},
			},
		},
		Subnotes:  map[string]models.Note{},
		HeaderURL: "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_EmptyNote(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Empty Note",
		},
		Blocks:     []models.Block{},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
	assert.Greater(t, pdfBuffer.Len(), 0)
}

func TestGeneratePDF_EmptyTitle(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "",
		},
		Blocks:     []models.Block{},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_InvalidImageURL(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Note with Invalid Image",
		},
		Blocks: []models.Block{
			{
				ID:          uuid.New(),
				BlockTypeID: 2,
				Position:    0,
				Content:     "not-a-valid-url",
			},
		},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

func TestGeneratePDF_SubnoteNotFound(t *testing.T) {
	noteContent := &NoteContent{
		Note: &models.Note{
			ID:    uuid.New(),
			Title: "Note with Missing Subnote",
		},
		Blocks: []models.Block{
			{
				ID:          uuid.New(),
				BlockTypeID: 5,
				Position:    0,
				Content:     uuid.New().String(),
			},
		},
		Formatting: map[string]models.BlockFormatting{},
		Subnotes:   map[string]models.Note{},
		HeaderURL:  "",
	}

	pdfBuffer, err := GeneratePDF(noteContent)

	require.NoError(t, err)
	assert.NotNil(t, pdfBuffer)
}

// Helper functions
func boolPtr(b bool) *bool {
	return &b
}

func intPtr(i int) *int {
	return &i
}
