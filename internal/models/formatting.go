package models

type FormattingRange struct {
	ID        string `json:"id"`
	StartPos  int    `json:"start_pos"` // inclusive
	EndPos    int    `json:"end_pos"`   // exclusive
	Bold      *bool  `json:"bold,omitempty"`
	Italic    *bool  `json:"italic,omitempty"`
	Underline *bool  `json:"underline,omitempty"`
	TextAlign *int   `json:"text_align,omitempty"` // 0: left, 1: center, 2: right
}

type BlockFormatting struct {
	BlockID string            `json:"block_id"`
	Ranges  []FormattingRange `json:"ranges"`
}
