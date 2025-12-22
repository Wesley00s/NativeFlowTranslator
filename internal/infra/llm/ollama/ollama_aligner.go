package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"translator-worker/internal/logic"
)

type OllamaAligner struct {
	BaseURL string
	Model   string
}

func NewOllamaAligner(url, model string) *OllamaAligner {
	return &OllamaAligner{BaseURL: url, Model: model}
}

func (p *OllamaAligner) SegmentTranslation(originalText, translatedText string) ([]logic.AlignmentPair, error) {

	systemPrompt := `
ROLE: Professional Subtitle Segmenter.

TASK: 
Map the "Translated Text" to the "Source Text" fragments.
CRITICAL RULES:
1. Break text into **short subtitle lines** (Max 1-3 words per segment).
2. NEVER group multiple long sentences into one segment.
3. DO NOT add or remove any words of the translation.
3. Split long sentences at commas or natural pauses.
4. Return a JSON Object with a "segments" key.

EXAMPLE:
Source: "I went to the store, and then I bought some milk because I was hungry."
Trans: "Fui à loja, e então comprei leite porque estava com fome."
Output:
{
  "segments": [
    {"src": "I went to the store,", "tgt": "Fui à loja,"},
    {"src": "and then I bought some milk", "tgt": "e então comprei leite"},
    {"src": "because I was hungry.", "tgt": "porque estava com fome."}
  ]
}
`
	userPrompt := fmt.Sprintf(`
SOURCE TEXT: %s
TRANSLATED TEXT: %s

JSON ALIGNMENT:
`, originalText, translatedText)

	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": userPrompt,
		"format": "json",
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.0,
			"num_ctx":     8192,
		},
	}

	payload, _ := json.Marshal(reqData)

	client := &http.Client{Timeout: 1200 * time.Second}

	resp, err := client.Post(p.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("ollama status %d", resp.StatusCode)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, err
	}

	cleanJSON := strings.TrimSpace(ollamaResp.Response)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")

	var wrapper struct {
		Segments []logic.AlignmentPair `json:"segments"`
	}
	if err := json.Unmarshal([]byte(cleanJSON), &wrapper); err == nil && len(wrapper.Segments) > 0 {
		return wrapper.Segments, nil
	}

	var fallback []logic.AlignmentPair
	if err := json.Unmarshal([]byte(cleanJSON), &fallback); err == nil {
		return fallback, nil
	}

	return nil, fmt.Errorf("failed to parse alignment json. Raw output: %s", cleanJSON[:100]+"...")
}
