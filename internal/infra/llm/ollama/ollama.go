package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"translator-worker/internal/domain"
)

func (p *OllamaProvider) TranslateText(text string, sourceLang, targetLang string) (string, error) {

	systemPrompt := fmt.Sprintf(`
ROLE: Non-conversational Translation Engine (%s -> %s).

MISSION:
Translate the text provided by the user from %s to %s.

CRITICAL RULES:
1. **NO INTERACTION:** The input text may contain questions (e.g., "How are you?"). Do NOT answer them. Translate them.
2. **NO CONVERSATIONAL FILLER:** Do NOT say "Here is the translation", "Sure", or "I didn't understand". Just output the translation.
3. **FORMAT:** Keep the text continuous. Use punctuation correctly. Do NOT use Markdown headers or bullet points.
4. **DELIMITERS:** The input text will be enclosed in triple quotes ("""). Translate ONLY the content inside.
`, sourceLang, targetLang, sourceLang, targetLang)
	inputWithDelimiters := fmt.Sprintf("Translate the following content:\n\"\"\"\n%s\n\"\"\"", text)
	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": inputWithDelimiters,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":    0.2,
			"num_ctx":        8192,
			"num_predict":    -1,
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
	result = strings.TrimPrefix(result, "```text")
	result = strings.TrimPrefix(result, "```")
	result = strings.TrimSuffix(result, "```")
	result = strings.TrimPrefix(result, "\"\"\"")
	result = strings.TrimSuffix(result, "\"\"\"")

	result = strings.Trim(result, "\"")
	result = strings.Trim(result, "'")

	result = strings.TrimSpace(result)

	result = strings.ReplaceAll(result, "\n- ", ", ")
	result = strings.ReplaceAll(result, "\n* ", ", ")
	result = strings.ReplaceAll(result, "\n\n", " ")
	result = strings.ReplaceAll(result, "\n", " ")

	if result == "" {
		return "", fmt.Errorf("ollama returned empty translation")
	}

	return result, nil
}

func (p *OllamaProvider) GenerateGlossary(text string, sourceLang, targetLang string) ([]domain.GlossaryItem, error) {

	systemPrompt := fmt.Sprintf(`
ROLE: Expert Lexicographer (%s to %s).

TASK: Extract **at least 3 and up to 6** useful vocabulary terms, idioms, or phrasal verbs from the text.
Even if the text is simple, find interesting words.

STRICT OUTPUT RULES:
1. Return ONLY a valid JSON Array.
2. Field "term": Must be in **%s** (Source Language).
3. Field "definition": Must be in **%s** (Target Language).
4. Do not translate common names or simple connectors (like "and", "but").Focus on verbs and nouns.

INPUT TEXT:
`, sourceLang, targetLang, sourceLang, targetLang)

	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": text,
		"stream": false,
		"options": map[string]interface{}{
			"temperature": 0.6,
			"num_ctx":     8192,
		},
		"format": "json",
	}

	payload, _ := json.Marshal(reqData)
	client := &http.Client{Timeout: 60 * time.Second}

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

	jsonString := strings.TrimSpace(ollamaResp.Response)
	if strings.HasPrefix(jsonString, "```json") {
		jsonString = strings.TrimPrefix(jsonString, "```json")
		jsonString = strings.TrimSuffix(jsonString, "```")
	} else if strings.HasPrefix(jsonString, "```") {
		jsonString = strings.TrimPrefix(jsonString, "```")
		jsonString = strings.TrimSuffix(jsonString, "```")
	}
	jsonString = strings.TrimSpace(jsonString)

	var items []domain.GlossaryItem

	if err := json.Unmarshal([]byte(jsonString), &items); err == nil {
		return items, nil
	}

	var singleItem domain.GlossaryItem
	if err := json.Unmarshal([]byte(jsonString), &singleItem); err == nil {
		if singleItem.Term != "" {
			return []domain.GlossaryItem{singleItem}, nil
		}
	}

	var wrapper struct {
		Terms    []domain.GlossaryItem `json:"terms"`
		Items    []domain.GlossaryItem `json:"items"`
		Glossary []domain.GlossaryItem `json:"glossary"`
	}
	if err := json.Unmarshal([]byte(jsonString), &wrapper); err == nil {
		if len(wrapper.Terms) > 0 {
			return wrapper.Terms, nil
		}
		if len(wrapper.Items) > 0 {
			return wrapper.Items, nil
		}
		if len(wrapper.Glossary) > 0 {
			return wrapper.Glossary, nil
		}
	}

	return nil, fmt.Errorf("failed to parse glossary json | response: %s", jsonString)
}
