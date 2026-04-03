package models

type Formatting struct {
	Bold      bool `json:"bold"`
	Italic    bool `json:"italic"`
	Underline bool `json:"underline"`
	TextAlign int  `json:"text_align"`
}
