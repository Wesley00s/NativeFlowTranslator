package domain

type SubtitleItem struct {
	Text  string  `json:"text"`
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Conf  float64 `json:"conf"`
}

type TranscriptionData struct {
	VideoID          string         `json:"videoId"`
	MongoID          string         `json:"mongoId"`
	OriginalFilename string         `json:"originalFilename"`
	FullText         string         `json:"fullText"`
	TargetLang       string         `json:"targetLang"`
	Transcription    []SubtitleItem `json:"transcription"`
	SourceLang       string         `json:"sourceLang"`
}

type TranslationResult struct {
	VideoID        string `json:"videoId"`
	Status         string `json:"status"`
	TargetLang     string `json:"targetLang"`
	TranslatedText string `json:"translatedText"`
	ErrorMessage   string `json:"errorMessage,omitempty"`
}
