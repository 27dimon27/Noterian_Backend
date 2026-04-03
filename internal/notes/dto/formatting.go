package dto

import "github.com/go-park-mail-ru/2026_1_WHITECROWSOFT/internal/models"

type Formatting struct {
	Bold      bool `json:"bold"`
	Italic    bool `json:"italic"`
	Underline bool `json:"underline"`
	TextAlign int  `json:"text_align"`
}

func ToFormattingDTO(formatting models.Formatting) Formatting {
	return Formatting{
		Bold:      formatting.Bold,
		Italic:    formatting.Italic,
		Underline: formatting.Underline,
		TextAlign: formatting.TextAlign,
	}
}

func FromFormattingDTO(dto Formatting) models.Formatting {
	return models.Formatting{
		Bold:      dto.Bold,
		Italic:    dto.Italic,
		Underline: dto.Underline,
		TextAlign: dto.TextAlign,
	}
}
