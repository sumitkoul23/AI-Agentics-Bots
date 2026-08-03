package main

// Reasoning — chain-of-thought / tree-of-thought wrapper for the LLM.
//
// Environment variable:
//   REASONING_MODE=chain-of-thought   (or cot) — adds a hidden thinking pass
//   REASONING_MODE=off|""             disabled (default)
//
// When enabled, every agent call first runs a "think" prompt that asks the LLM
// to reason step by step. The reasoning trace is stored in memory and the
// cleaned final answer is returned to the user.

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

const reasoningSystem = `You are a rigorous thinker. Given the user's question and the specialist's perspective, 
reason step by step — consider multiple angles, check for errors, weigh evidence. 
Output your raw chain of thought between <think> and </think> tags.
Then output the final, clean answer after </think>.
Do NOT include the thinking tags or reasoning traces in the final answer — only the answer.`

// ApplyReasoning wraps an LLM call with a chain-of-thought pre-pass if
// REASONING_MODE is set. Returns the final cleaned answer and the reasoning
// trace (which may be empty if reasoning is disabled).
func ApplyReasoning(ollama *OllamaClient, agentSystem, userPrompt string) (answer, trace string) {
	mode := strings.ToLower(os.Getenv("REASONING_MODE"))
	if mode == "" || mode == "off" || mode == "false" || ollama == nil {
		return "", "" // caller should fall back to direct generation
	}

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	// Step 1: combined think + answer in one pass
	hybridSystem := agentSystem + "\n\n" + reasoningSystem
	combined, err := ollama.Generate(ctx, hybridSystem, userPrompt)
	if err != nil {
		return "", ""
	}

	// Parse out <think>…</think>
	trace, answer = splitThinkAnswer(combined)
	if answer == "" {
		answer = combined // fallback: use entire output
	}
	return strings.TrimSpace(answer), strings.TrimSpace(trace)
}

// splitThinkAnswer splits combined output into (trace, answer).
func splitThinkAnswer(text string) (trace, answer string) {
	start := strings.Index(text, "<think>")
	end := strings.Index(text, "</think>")
	if start == -1 || end == -1 || end <= start {
		return "", text
	}
	trace = strings.TrimSpace(text[start+7 : end])
	answer = strings.TrimSpace(text[end+8:])
	return trace, answer
}

// ReasoningModeActive reports whether reasoning is enabled.
func ReasoningModeActive() bool {
	m := strings.ToLower(os.Getenv("REASONING_MODE"))
	return m != "" && m != "off" && m != "false"
}

// FormatReasoningTrace returns a user-facing block showing the reasoning trace.
func FormatReasoningTrace(trace string) string {
	if trace == "" {
		return ""
	}
	return fmt.Sprintf("\n\n<details><summary>💭 Reasoning trace</summary>\n\n%s\n\n</details>", trace)
}

// ReasoningInfo returns a human-readable status string.
func ReasoningInfo() string {
	if !ReasoningModeActive() {
		return "Reasoning: off (set REASONING_MODE=chain-of-thought to enable)"
	}
	return fmt.Sprintf("Reasoning: %s active — step-by-step thinking before every answer", os.Getenv("REASONING_MODE"))
}

// timeoutSecs is a helper used by voice.go and other files in this package.
var _ = time.Second // ensure time import is used
