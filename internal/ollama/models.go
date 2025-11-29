package ollama

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type Model struct {
	Name       string    `json:"name"`
	Size       int64     `json:"size"`
	ModifiedAt time.Time `json:"modified_at"`
}

type ModelList struct {
	Models []Model `json:"models"`
}

type PullProgress struct {
	Status    string `json:"status"`
	Digest    string `json:"digest,omitempty"`
	Total     int64  `json:"total,omitempty"`
	Completed int64  `json:"completed,omitempty"`
}

type RecommendedModel struct {
	Name        string
	Description string
	Size        string
}

func GetRecommendedModels() []RecommendedModel {
	return []RecommendedModel{
		{
			Name:        "qwen2.5-coder:7b",
			Description: "Great for coding, good balance of speed and quality",
			Size:        "4.7 GB",
		},
		{
			Name:        "qwen2.5-coder:14b",
			Description: "Better reasoning, needs more RAM",
			Size:        "9.0 GB",
		},
		{
			Name:        "deepseek-coder-v2:16b",
			Description: "Excellent at code generation and explanation",
			Size:        "8.9 GB",
		},
		{
			Name:        "llama3.2:3b",
			Description: "Fast and lightweight, good for quick tasks",
			Size:        "2.0 GB",
		},
		{
			Name:        "mistral:7b",
			Description: "Solid general purpose model",
			Size:        "4.1 GB",
		},
	}
}

func ListModels(baseURL string) ([]Model, error) {
	client := &http.Client{Timeout: 10 * time.Second}

	resp, err := client.Get(baseURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Ollama: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-ok status: %d", resp.StatusCode)
	}

	var list ModelList
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return list.Models, nil
}

func PullModel(baseURL, modelName string, onProgress func(PullProgress)) error {
	reqBody, _ := json.Marshal(map[string]string{"name": modelName})

	resp, err := http.Post(baseURL+"/api/pull", "application/json", bytes.NewBuffer(reqBody))

	if err != nil {
		return fmt.Errorf("failed to start pull: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("non-ok status: %d", resp.StatusCode)
	}

	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		var progress PullProgress
		if err := json.Unmarshal(scanner.Bytes(), &progress); err != nil {
			continue
		}
		if onProgress != nil {
			onProgress(progress)
		}
	}
	return scanner.Err()
}

func HasModels(baseURL string) (bool, error) {
	models, err := ListModels(baseURL)
	if err != nil {
		return false, err
	}
	return len(models) > 0, nil
}
