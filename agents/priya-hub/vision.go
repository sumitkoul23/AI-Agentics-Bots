package main

// Vision — multimodal image understanding via Ollama vision models.
//
// Environment variables:
//   VISION_MODEL   Ollama model that supports images (e.g. llava, moondream)
//                  Falls back to the standard model with a warning if not set.
//
// Endpoint:
//   POST /vision   multipart form-data:
//                    "image"   — image file (jpg/png/webm)
//                    "prompt"  — optional question (default: "Describe this image")
//                  Returns {"reply":"…"}

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// ollamaVisionRequest matches Ollama's /api/generate with images field.
type ollamaVisionRequest struct {
	Model   string         `json:"model"`
	System  string         `json:"system,omitempty"`
	Prompt  string         `json:"prompt"`
	Images  []string       `json:"images"` // base64-encoded image data
	Stream  bool           `json:"stream"`
	Options map[string]any `json:"options,omitempty"`
}

func handleVision(w http.ResponseWriter, r *http.Request, ollama *OllamaClient) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "bad multipart form", http.StatusBadRequest)
		return
	}

	file, _, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "missing 'image' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	imgBytes, err := io.ReadAll(io.LimitReader(file, 20<<20))
	if err != nil {
		http.Error(w, "read error", http.StatusInternalServerError)
		return
	}
	imgB64 := base64.StdEncoding.EncodeToString(imgBytes)

	prompt := strings.TrimSpace(r.FormValue("prompt"))
	if prompt == "" {
		prompt = "Describe this image in detail."
	}

	model := os.Getenv("VISION_MODEL")
	if model == "" && ollama != nil {
		model = ollama.Model
	}
	if model == "" {
		model = "llava"
	}

	if ollama == nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"reply": "Ollama is not available. Start Ollama and pull a vision model: ollama pull llava",
		})
		return
	}

	ollamaBase := strings.TrimRight(os.Getenv("OLLAMA_HOST"), "/")
	if ollamaBase == "" {
		ollamaBase = "http://localhost:11434"
	}

	payload := ollamaVisionRequest{
		Model:  model,
		System: "You are Bodhi, an expert visual analyst. Describe images clearly and answer questions about them.",
		Prompt: prompt,
		Images: []string{imgB64},
		Stream: false,
		Options: map[string]any{
			"temperature": 0.7,
			"num_predict": 1024,
		},
	}
	body, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "POST", ollamaBase+"/api/generate", bytes.NewReader(body))
	if err != nil {
		http.Error(w, "request error", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := ollama.http.Do(req)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"reply": fmt.Sprintf("Vision error: %v\n\nMake sure you have a vision model: ollama pull llava", err),
		})
		return
	}
	defer resp.Body.Close()

	var result struct {
		Response string `json:"response"`
		Error    string `json:"error,omitempty"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		http.Error(w, "decode error", http.StatusInternalServerError)
		return
	}
	if result.Error != "" {
		result.Response = fmt.Sprintf("Vision model error: %s\n\nTry: ollama pull llava", result.Error)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"reply": strings.TrimSpace(result.Response)})
}
