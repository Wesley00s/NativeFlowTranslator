package service

import (
	"encoding/json"
	"log"
	"translator-worker/internal/domain"
	"translator-worker/internal/ports"
)

const GlossaryResultQueue = "video.glossary.result"

type GlossaryProcessor struct {
	LLM       ports.LLMProvider
	Publisher ports.MessagePublisher
}

func NewGlossaryProcessor(llm ports.LLMProvider, p ports.MessagePublisher) *GlossaryProcessor {
	return &GlossaryProcessor{
		LLM:       llm,
		Publisher: p,
	}
}

func (p *GlossaryProcessor) ProcessMessage(body []byte) {
	var input domain.GlossaryRequest
	if err := json.Unmarshal(body, &input); err != nil {
		log.Printf("❌ Invalid Glossary JSON: %s", err)
		return
	}
	p.Execute(input)
}

func (p *GlossaryProcessor) Execute(input domain.GlossaryRequest) {
	log.Printf("📖 Generating Glossary for Video %s...", input.VideoID)

	textToAnalyze := input.Text

	if len(textToAnalyze) > 4000 {
		textToAnalyze = textToAnalyze[:4000]
	}

	items, err := p.LLM.GenerateGlossary(textToAnalyze, input.SourceLang, input.TargetLang)
	if err != nil {
		log.Printf("❌ Glossary generation failed: %v", err)

		return
	}

	result := domain.GlossaryResult{
		VideoID: input.VideoID,
		Items:   items,
	}

	payload, _ := json.Marshal(result)
	err = p.Publisher.Publish(GlossaryResultQueue, payload)
	if err != nil {
		log.Printf("❌ Failed to publish glossary: %v", err)
		return
	}
	log.Printf("✅ Glossary published with %d items!", len(items))
}
