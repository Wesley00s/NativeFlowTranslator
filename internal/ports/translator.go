package ports

type LLMProvider interface {
	TranslateText(text string, src, tgt string) (string, error)
}
