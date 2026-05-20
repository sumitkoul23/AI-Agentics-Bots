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

// ── Gemini 2.0 Flash (Google) — FREE tier ────────────────────────────────────

type GeminiClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGeminiClient(key string) *GeminiClient {
	model := os.Getenv("GEMINI_MODEL")
	if model == "" {
		model = "gemini-2.0-flash"
	}
	return &GeminiClient{apiKey: key, model: model, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c *GeminiClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	url := fmt.Sprintf(
		"https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		c.model, c.apiKey,
	)
	payload := map[string]any{
		"system_instruction": map[string]any{
			"parts": []map[string]string{{"text": system}},
		},
		"contents": []map[string]any{
			{"role": "user", "parts": []map[string]string{{"text": prompt}}},
		},
		"generationConfig": map[string]any{
			"temperature":     0.7,
			"maxOutputTokens": 4096,
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("gemini: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("gemini decode: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("gemini: %s", result.Error.Message)
	}
	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("gemini: empty response")
	}
	return strings.TrimSpace(result.Candidates[0].Content.Parts[0].Text), nil
}

// ── Groq (Llama 3.3 70B) — FREE tier ─────────────────────────────────────────

type GroqClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewGroqClient(key string) *GroqClient {
	model := os.Getenv("GROQ_MODEL")
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &GroqClient{apiKey: key, model: model, http: &http.Client{Timeout: 60 * time.Second}}
}

func (c *GroqClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	return openAICompat(ctx, c.http, "https://api.groq.com/openai/v1/chat/completions", c.apiKey, c.model, system, prompt)
}

// ── Claude (Anthropic) — optional paid ───────────────────────────────────────

type ClaudeClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewClaudeClient(key string) *ClaudeClient {
	model := os.Getenv("CLAUDE_MODEL")
	if model == "" {
		model = "claude-sonnet-4-6"
	}
	return &ClaudeClient{apiKey: key, model: model, http: &http.Client{Timeout: 120 * time.Second}}
}

func (c *ClaudeClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	payload := map[string]any{
		"model":      c.model,
		"max_tokens": 4096,
		"system":     system,
		"messages":   []map[string]string{{"role": "user", "content": prompt}},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("claude: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Content []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("claude decode: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("claude: %s", result.Error.Message)
	}
	for _, part := range result.Content {
		if part.Type == "text" {
			return strings.TrimSpace(part.Text), nil
		}
	}
	return "", fmt.Errorf("claude: no text in response")
}

// ── OpenAI GPT-4o mini — optional paid (Codex successor) ─────────────────────

type OpenAIClient struct {
	apiKey string
	model  string
	http   *http.Client
}

func NewOpenAIClient(key string) *OpenAIClient {
	model := os.Getenv("OPENAI_MODEL")
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &OpenAIClient{apiKey: key, model: model, http: &http.Client{Timeout: 120 * time.Second}}
}

func (c *OpenAIClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	return openAICompat(ctx, c.http, "https://api.openai.com/v1/chat/completions", c.apiKey, c.model, system, prompt)
}

// openAICompat calls any OpenAI-compatible endpoint (Groq, OpenAI, etc.)
func openAICompat(ctx context.Context, client *http.Client, url, apiKey, model, system, prompt string) (string, error) {
	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": system},
			{"role": "user", "content": prompt},
		},
		"temperature": 0.7,
		"max_tokens":  4096,
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai-compat: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Error *struct{ Message string `json:"message"` } `json:"error"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("openai-compat decode: %w", err)
	}
	if result.Error != nil {
		return "", fmt.Errorf("openai-compat: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("openai-compat: empty response")
	}
	return strings.TrimSpace(result.Choices[0].Message.Content), nil
}

// ── Ollama — local, 100% free ─────────────────────────────────────────────────

type OllamaClient struct {
	baseURL string
	Model   string
	http    *http.Client
}

func NewOllamaClient() *OllamaClient {
	base := os.Getenv("OLLAMA_HOST")
	if base == "" {
		base = "http://localhost:11434"
	}
	return &OllamaClient{
		baseURL: strings.TrimRight(base, "/"),
		Model:   "llama3.2",
		http:    &http.Client{Timeout: 180 * time.Second},
	}
}

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
		Models []struct{ Name string `json:"name"` } `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil || len(result.Models) == 0 {
		return
	}
	// Prefer code-capable models when available
	preferred := []string{"codellama", "deepseek-coder", "llama3.2", "llama3.1", "llama3", "mistral", "gemma2", "phi3", "qwen"}
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

func (c *OllamaClient) Generate(ctx context.Context, system, prompt string) (string, error) {
	payload := map[string]any{
		"model":  c.Model,
		"system": system,
		"prompt": prompt,
		"stream": false,
		"options": map[string]any{
			"temperature": 0.7,
			"num_predict": 2048,
		},
	}
	body, _ := json.Marshal(payload)
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

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("ollama decode: %w", err)
	}
	if result.Error != "" {
		return "", fmt.Errorf("ollama: %s", result.Error)
	}
	return strings.TrimSpace(result.Response), nil
}
