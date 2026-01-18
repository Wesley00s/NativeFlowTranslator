package domain

type GlossaryRequest struct {
	VideoID    string `json:"videoId"`
	Text       string `json:"text"`
	SourceLang string `json:"sourceLang"`
	TargetLang string `json:"targetLang"`
}

type GlossaryItem struct {
	Term       string `json:"term"`
	Definition string `json:"definition"`
}

type GlossaryResult struct {
	VideoID string         `json:"videoId"`
	Items   []GlossaryItem `json:"items"`
}
