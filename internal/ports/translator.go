package ports

import "translator-worker/internal/domain"

type LLMProvider interface {
	TranslateText(text string, src, tgt string) (string, error)
	GenerateGlossary(text string, sourceLang, targetLang string) ([]domain.GlossaryItem, error)
}
