package models

type FormattingRange struct {
	ID        string `json:"id"`
	StartPos  int    `json:"start_pos"` // inclusive
	EndPos    int    `json:"end_pos"`   // exclusive
	Bold      *bool  `json:"bold,omitempty"`
	Italic    *bool  `json:"italic,omitempty"`
	Underline *bool  `json:"underline,omitempty"`
	TextAlign *int   `json:"text_align,omitempty"` // -1: по умолчанию, 0: лево, 1: центр, 2: право
}

type BlockFormatting struct {
	BlockID string            `json:"block_id"`
	Ranges  []FormattingRange `json:"ranges"`
}
