package main

// nlp_training.go — NLP training data pipeline for SkyAgents-hub.
//
// On startup, the hub can ingest one or more JSONL corpora to pre-seed agent
// memory with domain knowledge, improving NLP routing and response quality
// from the very first conversation.
//
// JSONL format (one JSON object per line):
//   {"agent":"code","key":"golang-idioms","value":"Use table-driven tests, defer for cleanup, errors as values not panics."}
//   {"agent":"perp-markets","key":"btc-halving","value":"BTC halving reduces block reward 50% every ~4 years; historically bullish 6-18m after."}
//
// Supported env vars:
//   NLP_TRAIN_DIR     path to a directory — loads all *.jsonl files
//   NLP_TRAIN_FILE    path to a single .jsonl file
//   NLP_TRAIN_ENDPOINT=1   enables POST /train for remote data push over tunnel

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// TrainingRecord is a single NLP training entry.
type TrainingRecord struct {
	Agent string `json:"agent"` // agent ID (e.g. "code", "perp-markets") or "" for global
	Key   string `json:"key"`   // fact key
	Value string `json:"value"` // fact value
}

// LoadNLPTrainingData loads JSONL training corpora into agent memory.
// Called from main() after memory is initialised but before the swarm starts.
func LoadNLPTrainingData(mem *Memory) {
	loaded := 0

	// Single file mode
	if f := os.Getenv("NLP_TRAIN_FILE"); f != "" {
		n, err := loadJSONL(f, mem)
		if err != nil {
			log.Printf("[NLP] error loading %s: %v", f, err)
		} else {
			loaded += n
		}
	}

	// Directory mode — loads all *.jsonl files
	if dir := os.Getenv("NLP_TRAIN_DIR"); dir != "" {
		entries, err := filepath.Glob(filepath.Join(dir, "*.jsonl"))
		if err != nil {
			log.Printf("[NLP] error scanning %s: %v", dir, err)
		} else {
			for _, path := range entries {
				n, err := loadJSONL(path, mem)
				if err != nil {
					log.Printf("[NLP] error loading %s: %v", path, err)
					continue
				}
				loaded += n
			}
		}
	}

	if loaded > 0 {
		mem.Save()
		log.Printf("[NLP] loaded %d training records into agent memory", loaded)
	}
}

// loadJSONL reads a JSONL file and injects each record into memory.
func loadJSONL(path string, mem *Memory) (int, error) {
	f, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer f.Close()

	count := 0
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024) // 1 MB line buffer
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}
		var rec TrainingRecord
		if err := json.Unmarshal([]byte(line), &rec); err != nil {
			log.Printf("[NLP] skipping malformed line in %s: %v", path, err)
			continue
		}
		if err := injectRecord(rec, mem); err != nil {
			log.Printf("[NLP] skipping invalid record: %v", err)
			continue
		}
		count++
	}
	if err := scanner.Err(); err != nil {
		return count, err
	}
	log.Printf("[NLP] loaded %d records from %s", count, filepath.Base(path))
	return count, nil
}

// injectRecord validates and stores a training record into memory.
func injectRecord(rec TrainingRecord, mem *Memory) error {
	if strings.TrimSpace(rec.Key) == "" {
		return fmt.Errorf("empty key")
	}
	if strings.TrimSpace(rec.Value) == "" {
		return fmt.Errorf("empty value")
	}
	key := strings.TrimSpace(rec.Key)
	value := strings.TrimSpace(rec.Value)
	agent := strings.TrimSpace(rec.Agent)

	if agent != "" {
		// Namespace by agent (same format as Learner and seed.go)
		mem.Learn(agent+":"+key, value)
	} else {
		// Global fact — shared across all agents
		mem.Learn(key, value)
	}
	return nil
}

// RegisterTrainEndpoint adds POST /train if NLP_TRAIN_ENDPOINT=1.
// This endpoint lets remote devices (connected via IP tunnel) push training
// data to the hub in real time — enabling distributed learning.
func RegisterTrainEndpoint(mux *http.ServeMux, mem *Memory) {
	if os.Getenv("NLP_TRAIN_ENDPOINT") != "1" {
		return
	}

	mux.HandleFunc("/train", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST only", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(io.LimitReader(r.Body, 4*1024*1024)) // 4 MB max
		if err != nil {
			http.Error(w, "read error", http.StatusBadRequest)
			return
		}

		// Accept either a single JSON object or a JSONL body (multiple lines)
		lines := strings.Split(strings.TrimSpace(string(body)), "\n")
		loaded := 0
		skipped := 0
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			var rec TrainingRecord
			if err := json.Unmarshal([]byte(line), &rec); err != nil {
				skipped++
				continue
			}
			if err := injectRecord(rec, mem); err != nil {
				skipped++
				continue
			}
			loaded++
		}

		if loaded > 0 {
			mem.Save()
		}

		log.Printf("[NLP /train] loaded=%d skipped=%d from %s", loaded, skipped, r.RemoteAddr)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"loaded":  loaded,
			"skipped": skipped,
			"total":   loaded + skipped,
		})
	})

	// GET /train/status — report current training state
	mux.HandleFunc("/train/status", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		mem.mu.RLock()
		facts := len(mem.Data.Facts)
		prefs := len(mem.Data.Preferences)
		score := mem.Data.TrainingScore
		interactions := mem.Data.Interactions
		lessons := len(mem.Data.Lessons)
		mem.mu.RUnlock()

		json.NewEncoder(w).Encode(map[string]any{
			"training_score": score,
			"interactions":   interactions,
			"facts":          facts,
			"preferences":    prefs,
			"lessons":        lessons,
			"endpoint":       "POST /train  (JSONL body)",
			"format":         `{"agent":"<id>","key":"<fact>","value":"<detail>"}`,
		})
	})

	log.Printf("[NLP] /train endpoint enabled — remote training active")
}
