package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

type OllamaClient struct {
	baseURL string
	Model   string
	http    *http.Client
}

type ollamaGenRequest struct {
	Model   string         `json:"model"`
	System  string         `json:"system,omitempty"`
	Prompt  string         `json:"prompt"`
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

type ollamaGenResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
	Error    string `json:"error,omitempty"`
}

func NewOllamaClient() *OllamaClient {
	base := os.Getenv("OLLAMA_HOST")
	if base == "" {
		base = "http://localhost:11434"
	}
	return &OllamaClient{
		baseURL: strings.TrimRight(base, "/"),
		Model:   "llama3.2",
		http:    &http.Client{Timeout: 120 * time.Second},
	}
}

// IsAvailable checks whether the Ollama server is reachable.
func (c *OllamaClient) IsAvailable() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return false
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}

// AutoModel selects the best available model from the running Ollama server.
func (c *OllamaClient) AutoModel() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	var result struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Models) == 0 {
		return
	}

	preferred := []string{"llama3.2", "llama3.1", "llama3", "mistral", "gemma2", "phi3", "gemma", "qwen"}
	for _, p := range preferred {
		for _, m := range result.Models {
			if strings.HasPrefix(m.Name, p) {
				c.Model = m.Name
				return
			}
		}
	}
	c.Model = result.Models[0].Name
}

// Generate sends a prompt to Ollama and returns the response text.
func (c *OllamaClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	payload := ollamaGenRequest{
		Model:  c.Model,
		System: system,
		Prompt: prompt,
		Stream: false,
		Options: map[string]any{
			"temperature": 0.7,
			"num_predict": 2048,
		},
	}
	return c.post(ctx, payload)
}

// GenerateShort uses a low token limit — good for classification and extraction.
func (c *OllamaClient) GenerateShort(ctx context.Context, system, prompt string) (string, error) {
	payload := ollamaGenRequest{
		Model:  c.Model,
		System: system,
		Prompt: prompt,
		Stream: false,
		Options: map[string]any{
			"temperature": 0.1,
			"num_predict": 60,
		},
	}
	return c.post(ctx, payload)
}

func (c *OllamaClient) post(ctx context.Context, payload ollamaGenRequest) (string, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("ollama: %w", err)
	}
	defer resp.Body.Close()

	var result ollamaGenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama decode: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("ollama: %s", result.Error)
	}
	return strings.TrimSpace(result.Response), nil
}
