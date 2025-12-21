package service

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"translator-worker/internal/domain"
	"translator-worker/internal/ports"
	"translator-worker/internal/utils"
)

const ResultQueue = "video.translation.result"

type TranslationProcessor struct {
	Translator ports.LLMProvider
	Publisher  ports.MessagePublisher
}

func NewTranslationProcessor(t ports.LLMProvider, p ports.MessagePublisher) *TranslationProcessor {
	return &TranslationProcessor{
		Translator: t,
		Publisher:  p,
	}
}

func (p *TranslationProcessor) ProcessMessage(body []byte) {
	var input domain.TranscriptionData
	if err := json.Unmarshal(body, &input); err != nil {
		log.Printf("❌ Invalid JSON: %s", err)

		return
	}

	videoID := input.VideoID
	if videoID == "" {
		videoID = input.MongoID
	}

	fullText := input.FullText
	if fullText == "" {
		var sb strings.Builder
		for _, item := range input.Transcription {
			sb.WriteString(item.Text + " ")
		}
		fullText = sb.String()
	}

	log.Printf("📄 Processing Translation for Video %s | Lang: %s", videoID, input.TargetLang)

	const ChunkSize = 6000
	chunks := splitTextByChars(fullText, ChunkSize)

	var finalTranslationBuilder strings.Builder
	var translationErr error

	for i, chunk := range chunks {
		log.Printf("   🔄 Translating Chunk %d/%d...", i+1, len(chunks))

		var translatedChunk string
		var err error

		for attempt := 1; attempt <= 3; attempt++ {
			translatedChunk, err = p.Translator.TranslateText(chunk, input.SourceLang, input.TargetLang)
			if err == nil {
				break
			}
			log.Printf("      ⚠️ Attempt %d failed: %v", attempt, err)
			time.Sleep(2 * time.Second)
		}

		if err != nil {
			log.Printf("      ❌ Failed to translate chunk %d.", i)
			translationErr = err
			break
		}

		finalTranslationBuilder.WriteString(translatedChunk + " ")
	}

	if translationErr != nil || len(strings.TrimSpace(finalTranslationBuilder.String())) < 10 {
		log.Printf("❌ Translation failed. Sending Error Event.")
		p.publishError(videoID, input.TargetLang, "Translation failed or result empty")
		return
	}

	finalText := finalTranslationBuilder.String()
	mongoLangCode := utils.NormalizeLangCode(input.TargetLang)

	p.publishSuccess(videoID, mongoLangCode, finalText)
}

func (p *TranslationProcessor) publishSuccess(videoID, langCode, text string) {
	result := domain.TranslationResult{
		VideoID:        videoID,
		Status:         "SUCCESS",
		TargetLang:     langCode,
		TranslatedText: text,
	}

	payload, _ := json.Marshal(result)

	log.Printf("📤 Publishing SUCCESS to '%s'...", ResultQueue)
	if err := p.Publisher.Publish(ResultQueue, payload); err != nil {
		log.Printf("❌ Failed to publish result: %v", err)
	} else {
		log.Println("✅ Result published successfully!")
	}
}

func (p *TranslationProcessor) publishError(videoID, langCode, msg string) {
	result := domain.TranslationResult{
		VideoID:      videoID,
		Status:       "ERROR",
		TargetLang:   langCode,
		ErrorMessage: msg,
	}
	payload, _ := json.Marshal(result)
	err := p.Publisher.Publish(ResultQueue, payload)
	if err != nil {
		return
	}
}

func splitTextByChars(text string, limit int) []string {
	if len(text) <= limit {
		return []string{text}
	}
	var chunks []string
	for len(text) > limit {
		cut := strings.LastIndex(text[:limit], " ")
		if cut == -1 {
			cut = limit
		}
		chunks = append(chunks, text[:cut])
		text = text[cut:]
	}
	chunks = append(chunks, text)
	return chunks
}
