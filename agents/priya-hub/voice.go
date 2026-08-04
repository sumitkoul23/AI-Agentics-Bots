package main

// Voice — Speech-to-Text (Whisper) and Text-to-Speech (Piper) support.
//
// Environment variables:
//   WHISPER_BIN    path to whisper-cli or whisper.cpp main binary
//   WHISPER_MODEL  path to .gguf model file
//   PIPER_BIN      path to piper binary
//   PIPER_MODEL    path to piper .onnx voice model
//
// Endpoints:
//   POST /transcribe   multipart with "audio" field → {"text":"…"}
//   GET  /speak?text=… streams WAV audio
//   GET  /voice/status returns JSON capabilities

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// ── Whisper STT ───────────────────────────────────────────────────────────────

func handleTranscribe(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	whisperBin := resolveWhisperBin()
	if whisperBin == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "WHISPER_BIN not set — install whisper.cpp and set WHISPER_BIN=/path/to/whisper-cli",
		})
		return
	}

	if err := r.ParseMultipartForm(25 << 20); err != nil {
		http.Error(w, "bad multipart form", http.StatusBadRequest)
		return
	}
	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "missing 'audio' field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".wav"
	}
	tmp, err := os.CreateTemp("", "bodhi-audio-*"+ext)
	if err != nil {
		http.Error(w, "temp file error", http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, file); err != nil {
		tmp.Close()
		http.Error(w, "write error", http.StatusInternalServerError)
		return
	}
	tmp.Close()

	inputPath := tmp.Name()
	if ext != ".wav" {
		wavPath := strings.TrimSuffix(tmp.Name(), ext) + ".wav"
		ffmpeg := exec.Command("ffmpeg", "-y", "-i", inputPath, "-ar", "16000", "-ac", "1", wavPath)
		if ffmpeg.Run() == nil {
			defer os.Remove(wavPath)
			inputPath = wavPath
		}
	}

	args := []string{"-f", inputPath, "--output-txt", "--no-prints", "-np", "-nt"}
	if model := os.Getenv("WHISPER_MODEL"); model != "" {
		args = append(args, "-m", model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, whisperBin, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	_ = cmd.Run()

	text := ""
	txtPath := inputPath + ".txt"
	if data, err := os.ReadFile(txtPath); err == nil {
		text = strings.TrimSpace(string(data))
		os.Remove(txtPath)
	}
	if text == "" {
		text = strings.TrimSpace(buf.String())
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"text": text})
}

func resolveWhisperBin() string {
	if b := os.Getenv("WHISPER_BIN"); b != "" {
		return b
	}
	for _, name := range []string{"whisper-cli", "whisper", "main"} {
		if p, err := exec.LookPath(name); err == nil {
			return p
		}
	}
	return ""
}

// ── Piper TTS ─────────────────────────────────────────────────────────────────

func handleSpeak(w http.ResponseWriter, r *http.Request) {
	text := strings.TrimSpace(r.URL.Query().Get("text"))
	if text == "" {
		http.Error(w, "?text= required", http.StatusBadRequest)
		return
	}
	if len(text) > 4096 {
		text = text[:4096]
	}

	piperBin := resolvePiperBin()
	if piperBin == "" {
		http.Error(w, "PIPER_BIN not set — install piper-tts", http.StatusServiceUnavailable)
		return
	}

	args := []string{"--output-raw"}
	if model := os.Getenv("PIPER_MODEL"); model != "" {
		args = append(args, "--model", model)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, piperBin, args...)
	cmd.Stdin = strings.NewReader(text)
	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		http.Error(w, "piper error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	wav := rawPCMtoWAV(out.Bytes(), 22050, 1, 16)
	w.Header().Set("Content-Type", "audio/wav")
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(wav)))
	w.Header().Set("Cache-Control", "no-store")
	w.Write(wav)
}

func resolvePiperBin() string {
	if b := os.Getenv("PIPER_BIN"); b != "" {
		return b
	}
	if p, err := exec.LookPath("piper"); err == nil {
		return p
	}
	return ""
}

// rawPCMtoWAV wraps raw 16-bit signed LE PCM in a RIFF/WAV header.
func rawPCMtoWAV(pcm []byte, sampleRate, channels, bitsPerSample int) []byte {
	dataSize := len(pcm)
	byteRate := sampleRate * channels * bitsPerSample / 8
	blockAlign := channels * bitsPerSample / 8

	var buf bytes.Buffer
	le := func(v uint32, n int) {
		b := make([]byte, n)
		for i := 0; i < n; i++ {
			b[i] = byte(v >> (8 * i))
		}
		buf.Write(b)
	}
	buf.WriteString("RIFF")
	le(uint32(36+dataSize), 4)
	buf.WriteString("WAVE")
	buf.WriteString("fmt ")
	le(16, 4)
	le(1, 2)
	le(uint32(channels), 2)
	le(uint32(sampleRate), 4)
	le(uint32(byteRate), 4)
	le(uint32(blockAlign), 2)
	le(uint32(bitsPerSample), 2)
	buf.WriteString("data")
	le(uint32(dataSize), 4)
	buf.Write(pcm)
	return buf.Bytes()
}

// ── Voice status ──────────────────────────────────────────────────────────────

func handleVoiceStatus(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"stt_ready":     resolveWhisperBin() != "",
		"tts_ready":     resolvePiperBin() != "",
		"whisper_model": os.Getenv("WHISPER_MODEL"),
		"piper_model":   os.Getenv("PIPER_MODEL"),
		"install_hint":  "whisper.cpp: https://github.com/ggerganov/whisper.cpp | piper: https://github.com/rhasspy/piper",
	})
}
