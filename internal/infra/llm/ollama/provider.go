package ollama

type OllamaProvider struct {
	BaseURL string
	Model   string
}

func NewOllamaProvider(url, model string) *OllamaProvider {
	return &OllamaProvider{
		BaseURL: url,
		Model:   model,
	}
}
