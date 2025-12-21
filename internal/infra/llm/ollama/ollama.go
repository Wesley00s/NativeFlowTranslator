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
Translate the following text.
1. **Faithfulness:** Do not over-localize names or pronouns.
   - Eg: "because of him or her" -> Target: "**por causa dele ou dela**" (NEVER use "fulano ou sicrano").
2. Maintain the original tone and meaning.
3. Use natural, fluid %s.
4. Do NOT output explanations, just the translated text.
5. Do NOT use specific formatting like IDs or Timestamps.
`, sourceLang, targetLang, targetLang)

	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": text,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":    0.1,
			"num_ctx":        8192,
			"num_predict":    4096,
			"repeat_penalty": 1.1,
		},
	}

	payload, _ := json.Marshal(reqData)
	client := &http.Client{Timeout: 1000 * time.Second}

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

	if result == "" {
		return "", fmt.Errorf("ollama returned empty translation (model might be loading or failed)")
	}

	return result, nil
}
