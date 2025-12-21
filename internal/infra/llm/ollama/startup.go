package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

func EnsureModelLoaded(host, modelName string) error {
	log.Printf("🔍 Checking Ollama model '%s' at %s...", modelName, host)

	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(host + "/api/tags")
	if err != nil {
		return fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(resp.Body)

	var tagsResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&tagsResp); err != nil {
		return err
	}

	for _, m := range tagsResp.Models {

		if m.Name == modelName || m.Name == modelName+":latest" {
			log.Println("✅ Model already loaded in Ollama.")
			return nil
		}
	}

	log.Printf("⚠️ Model not found. Starting download of '%s'... (This can take minutes)", modelName)

	pullClient := &http.Client{Timeout: 0}

	payload := map[string]string{"name": modelName}
	jsonPayload, _ := json.Marshal(payload)

	pullResp, err := pullClient.Post(host+"/api/pull", "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return fmt.Errorf("failed to trigger pull: %w", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing response body:", err)
		}
	}(pullResp.Body)

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(pullResp.Body)
	if err != nil {
		return fmt.Errorf("error reading pull response: %w", err)
	}

	if pullResp.StatusCode != 200 {
		return fmt.Errorf("pull failed with status: %d", pullResp.StatusCode)
	}

	log.Println("🎉 Model downloaded successfully!")
	return nil
}
