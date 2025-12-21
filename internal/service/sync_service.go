package service

import (
	"encoding/json"
	"log"
	"translator-worker/internal/domain"
	"translator-worker/internal/infra/llm/ollama"
	"translator-worker/internal/logic"
	"translator-worker/internal/ports"
	"translator-worker/internal/utils"
)

const SyncResultQueue = "video.sync.result"

type SyncService struct {
	Aligner   *ollama.OllamaAligner
	Publisher ports.MessagePublisher
}

func NewSyncService(aligner *ollama.OllamaAligner, pub ports.MessagePublisher) *SyncService {
	return &SyncService{
		Aligner:   aligner,
		Publisher: pub,
	}
}

func (s *SyncService) ProcessSync(body []byte) {
	var input domain.SyncCommand
	if err := json.Unmarshal(body, &input); err != nil {
		log.Printf("❌ [Sync] JSON Error: %v", err)
		return
	}

	log.Printf("🔄 [Sync] Processing Video: %s | Lang: %s", input.VideoID, input.TargetLang)

	if input.TranslatedText == "" || input.OriginalText == "" {
		log.Printf("❌ [Sync] Missing text data. Original: %d chars, Translated: %d chars",
			len(input.OriginalText), len(input.TranslatedText))
		s.publishError(input.VideoID, input.TargetLang, "Missing text for sync")
		return
	}

	log.Println("🤖 [Sync] Asking AI to segment alignment...")

	segments, err := s.Aligner.SegmentTranslation(input.OriginalText, input.TranslatedText)
	if err != nil {
		log.Printf("❌ [Sync] Aligner AI failed: %v", err)
		s.publishError(input.VideoID, input.TargetLang, "AI Alignment failed")
		return
	}

	log.Printf("✅ [Sync] AI returned %d segments. Mapping timestamps...", len(segments))

	finalSubtitles := logic.MapTimestamps(input.Transcription, segments)

	s.publishSuccess(input.VideoID, input.TargetLang, finalSubtitles)
}

func (s *SyncService) publishSuccess(videoID, lang string, subtitles []domain.SubtitleItem) {
	langCode := utils.NormalizeLangCode(lang)

	result := domain.SyncResult{
		VideoID:    videoID,
		Status:     "SUCCESS",
		TargetLang: langCode,
		Subtitles:  subtitles,
	}

	payload, _ := json.Marshal(result)

	log.Printf("📤 [Sync] Publishing SUCCESS to '%s'...", SyncResultQueue)
	if err := s.Publisher.Publish(SyncResultQueue, payload); err != nil {
		log.Printf("❌ [Sync] Failed to publish result: %v", err)
	}
}

func (s *SyncService) publishError(videoID, lang, msg string) {
	result := domain.SyncResult{
		VideoID:      videoID,
		Status:       "ERROR",
		TargetLang:   lang,
		ErrorMessage: msg,
		Subtitles:    []domain.SubtitleItem{},
	}
	payload, _ := json.Marshal(result)
	s.Publisher.Publish(SyncResultQueue, payload)
}
