package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (p *OllamaProvider) TranslateText(text string, sourceLang, targetLang string) (string, error) {

	systemPrompt := fmt.Sprintf(`
ROLE: Expert Translator (%s -> %s).

MISSION:
Translate the input text preserving its original structure as a continuous stream of text.

STRICT CONSTRAINTS:
1. **FULL TRANSLATION:** Translate EVERY word. Do NOT leave English words (like "those", "the", "and") in the middle of the Portuguese sentence.
2. **Faithfulness:** Maintain original tone. Do not over-localize proper names.
3. **NO MARKDOWN:** Do NOT use bold (**text**), headers (##), or italics.
4. **NO LISTS:** Do NOT use bullet points (-), dashes, or numbered lists. 
5. **CONTINUOUS TEXT:** If the input is a paragraph, the output MUST be a paragraph. Use commas or semicolons to separate items, NOT new lines.
6. **NO NEWLINES:** Do not add line breaks (\n) unless explicitly present in the source context as a paragraph break.
7. **Pure Output:** Do NOT output explanations, only the translated text.
8. **NATIVE GRAMMAR:** Ensure strict noun-gender agreement and correct article usage for the specific target dialect. The text must sound 100%% natural to a native speaker, correcting any source-text ambiguity.
`, sourceLang, targetLang)

	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": text,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":    0.3,
			"num_ctx":        8192,
			"num_predict":    4096,
			"repeat_penalty": 1.1,
		},
	}

	payload, _ := json.Marshal(reqData)
	client := &http.Client{Timeout: 1200 * time.Second}

	resp, err := client.Post(p.BaseURL+"/api/generate", "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama error: status %d (check if model '%s' is pulled)", resp.StatusCode, p.Model)
	}

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return "", err
	}

	result := strings.TrimSpace(ollamaResp.Response)
	result = strings.ReplaceAll(result, "\n- ", ", ")
	result = strings.ReplaceAll(result, "\n* ", ", ")
	result = strings.ReplaceAll(result, "\n\n", " ")
	result = strings.ReplaceAll(result, "\n", " ")

	if result == "" {
		return "", fmt.Errorf("ollama returned empty translation (model might be loading or failed)")
	}

	return result, nil
}
