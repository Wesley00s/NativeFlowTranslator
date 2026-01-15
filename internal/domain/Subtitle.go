package domain

type SubtitleItem struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Conf  float64 `json:"conf"`
}

type TranscriptionData struct {
	VideoID       string         `json:"videoId"`
	OriginalText  string         `json:"originalText"`
	SourceLang    string         `json:"sourceLang"`
	TargetLang    string         `json:"targetLang"`
	Transcription []SubtitleItem `json:"transcription"`
}

type TranslationResult struct {
	VideoID        string `json:"videoId"`
	Status         string `json:"status"`
	TargetLang     string `json:"targetLang"`
	TranslatedText string `json:"translatedText"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
}
