package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Learner extracts key facts from conversations and stores them in shared memory,
// enabling all agents to build a growing model of the user over time.
type Learner struct {
	mem    *Memory
	ollama *OllamaClient
}

func NewLearner(mem *Memory, ollama *OllamaClient) *Learner {
	return &Learner{mem: mem, ollama: ollama}
}

// Learn extracts one key fact from a conversation exchange and persists it.
func (l *Learner) Learn(agentID, userInput, agentResponse string) {
	if l.ollama == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()

	system := `You are a concise fact extractor. From the conversation below, extract ONE specific piece of information the user revealed about themselves, their preferences, goals, or situation.
Output ONLY in the format:  key: value
If nothing specific was revealed, output exactly: none
Do not explain. Do not add punctuation after the value.`

	prompt := fmt.Sprintf("User: %s\nAgent: %s", truncate(userInput, 400), truncate(agentResponse, 200))

	fact, err := l.ollama.GenerateShort(ctx, system, prompt)
	if err != nil || strings.TrimSpace(fact) == "none" {
		return
	}

	parts := strings.SplitN(fact, ":", 2)
	if len(parts) != 2 {
		return
	}
	key := strings.TrimSpace(parts[0])
	value := strings.TrimSpace(parts[1])
	if key == "" || value == "" || strings.Contains(key, "\n") {
		return
	}

	// Namespace facts by agent so agents have domain-specific memory
	l.mem.Learn(agentID+":"+key, value)
	log.Printf("[Learner] %s/%s = %s", agentID, key, value)
}

// BuildContext assembles the learned context string to inject into a system prompt.
func (l *Learner) BuildContext(agentID string) string {
	facts := l.mem.GetFacts()
	prefs := l.mem.GetPreferences()

	var lines []string
	prefix := agentID + ":"

	for k, v := range facts {
		if strings.HasPrefix(k, prefix) {
			lines = append(lines, fmt.Sprintf("- %s: %s", strings.TrimPrefix(k, prefix), v))
		}
	}
	for k, v := range prefs {
		lines = append(lines, fmt.Sprintf("- %s: %s", k, v))
	}

	if len(lines) == 0 {
		return ""
	}
	return "Context known about this user:\n" + strings.Join(lines, "\n")
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
