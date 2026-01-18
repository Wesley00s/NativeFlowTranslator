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
ROLE: Expert Translator (%s -> %s).

MISSION:
Translate the input text in %s to %s.

CRITICAL INSTRUCTIONS ON LANGUAGES:
**OUTPUT ENFORCEMENT:** The output MUST be in the **%s**. Do NOT output the Source Language.

STRICT CONSTRAINTS:
1. **FULL TRANSLATION:** Translate EVERY word. Do NOT leave English words (like "those", "the", "and") in the middle of the sentence.
2. **NO MARKDOWN:** Do NOT use bold (**text**), headers (##), or italics.
3. **NO LISTS:** Do NOT use bullet points (-), dashes, or numbered lists. 
4. **CONTINUOUS TEXT:** Keep text continuous (prosa). Use commas instead of new lines.
5. **NO NEWLINES:** Do not add line breaks (\n).
6. **Pure Output:** Do NOT output explanations.
`, sourceLang, targetLang, sourceLang, targetLang, targetLang)

	reqData := map[string]interface{}{
		"model":  p.Model,
		"system": systemPrompt,
		"prompt": text,
		"stream": false,
		"options": map[string]interface{}{
			"temperature":    0.3,
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

TASK: Analyze the input text and extract **3 to 6 useful vocabulary terms**.
Do not extract only the main topic. Look for variety:
- Interesting Verbs
- Specific Nouns or Objects
- Idioms or Expressions
- Adjectives

STRICT OUTPUT RULES:
1. Return ONLY a JSON Array.
2. Field "term": Must be in **%s** (Source Language). Copy exactly from text.
3. Field "definition": Must be in **%s** (Target Language). Explain the meaning.
4. **NO MARKDOWN:** Do not use **bold** or italics in the JSON values.

EXAMPLE FORMAT:
[
  {"term": "Source_Word_1", "definition": "Target_Definition_1"},
  {"term": "Source_Word_2", "definition": "Target_Definition_2"}
]

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
