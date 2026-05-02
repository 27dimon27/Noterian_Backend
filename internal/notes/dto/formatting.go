package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type FormattingRange struct {
	StartPos  int   `json:"start_pos"` // inclusive
	EndPos    int   `json:"end_pos"`   // exclusive
	Bold      *bool `json:"bold,omitempty"`
	Italic    *bool `json:"italic,omitempty"`
	Underline *bool `json:"underline,omitempty"`
	TextAlign *int  `json:"text_align,omitempty"`
}

type BlockFormatting struct {
	BlockID string            `json:"block_id"`
	Ranges  []FormattingRange `json:"ranges"`
}

func ToFormattingRangeDTO(rng models.FormattingRange) FormattingRange {
	return FormattingRange{
		StartPos:  rng.StartPos,
		EndPos:    rng.EndPos,
		Bold:      rng.Bold,
		Italic:    rng.Italic,
		Underline: rng.Underline,
		TextAlign: rng.TextAlign,
	}
}

func FromFormattingRangeDTO(dto FormattingRange) models.FormattingRange {
	return models.FormattingRange{
		StartPos:  dto.StartPos,
		EndPos:    dto.EndPos,
		Bold:      dto.Bold,
		Italic:    dto.Italic,
		Underline: dto.Underline,
		TextAlign: dto.TextAlign,
	}
}

func ToBlockFormattingDTO(formatting models.BlockFormatting) BlockFormatting {
	ranges := make([]FormattingRange, len(formatting.Ranges))
	for i, r := range formatting.Ranges {
		ranges[i] = ToFormattingRangeDTO(r)
	}
	return BlockFormatting{
		BlockID: formatting.BlockID,
		Ranges:  ranges,
	}
}
