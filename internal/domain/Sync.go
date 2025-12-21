package domain

type SyncCommand struct {
	VideoID        string         `json:"videoId"`
	OriginalText   string         `json:"originalText"`
	TranslatedText string         `json:"translatedText"`
	SourceLang     string         `json:"sourceLang"`
	TargetLang     string         `json:"targetLang"`
	Transcription  []SubtitleItem `json:"transcription"`
}

type SyncResult struct {
	VideoID        string         `json:"videoId"`
	Status         string         `json:"status"`
	TargetLang     string         `json:"targetLang"`
	TranslatedText string         `json:"translatedText"`
	Subtitles      []SubtitleItem `json:"subtitles"`
	ErrorMessage   string         `json:"errorMessage,omitempty"`
}
