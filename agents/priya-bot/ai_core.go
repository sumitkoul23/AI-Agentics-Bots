package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// AICore wraps the Claude API and gives Priya her intelligence.
type AICore struct {
	client *anthropic.Client
	mem    *Memory
	model  anthropic.Model
}

func NewAICore(mem *Memory) (*AICore, error) {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY is not set")
	}
	client := anthropic.NewClient(option.WithAPIKey(key))
	return &AICore{
		client: client,
		mem:    mem,
		model:  anthropic.ModelClaude3_7SonnetLatest,
	}, nil
}

// Think sends the user message to Claude with full conversation context and returns Priya's reply.
func (ai *AICore) Think(ctx context.Context, userInput string) (string, error) {
	// Build system prompt enriched with learned context
	system := ai.enrichedSystemPrompt()

	// Build messages from recent history + current input
	history := ai.mem.RecentHistory(30)
	var msgs []anthropic.MessageParam
	for _, h := range history {
		switch h.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		case "assistant":
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(h.Content)))
		}
	}
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)))

	resp, err := ai.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     ai.model,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: system},
		},
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}

	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	reply := sb.String()

	// Persist to memory
	ai.mem.AddMessage("user", userInput)
	ai.mem.AddMessage("assistant", reply)
	_ = ai.mem.Save()

	return reply, nil
}

// ThinkWithContext injects extra context (e.g. market data, job listings) into a single call.
func (ai *AICore) ThinkWithContext(ctx context.Context, systemContext, userInput string) (string, error) {
	combined := ai.enrichedSystemPrompt() + "\n\n━━━ LIVE CONTEXT ━━━\n" + systemContext
	msgs := []anthropic.MessageParam{
		anthropic.NewUserMessage(anthropic.NewTextBlock(userInput)),
	}
	resp, err := ai.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     ai.model,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: combined},
		},
		Messages: msgs,
	})
	if err != nil {
		return "", fmt.Errorf("Claude API error: %w", err)
	}
	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	return sb.String(), nil
}

func (ai *AICore) enrichedSystemPrompt() string {
	base := PriyaSystemPrompt
	var extras []string

	// Inject learned user voice samples
	samples := ai.mem.Data.UserVoiceSamples
	if len(samples) > 0 {
		extras = append(extras, "━━━ USER'S WRITING VOICE (mirror this in content) ━━━\n"+strings.Join(samples, "\n---\n"))
	}

	// Inject learned facts
	if len(ai.mem.Data.LearnedFacts) > 0 {
		var facts []string
		for k, v := range ai.mem.Data.LearnedFacts {
			facts = append(facts, fmt.Sprintf("%s: %s", k, v))
		}
		extras = append(extras, "━━━ KNOWN FACTS ABOUT USER ━━━\n"+strings.Join(facts, "\n"))
	}

	// Inject preferences
	if len(ai.mem.Data.UserPreferences) > 0 {
		var prefs []string
		for k, v := range ai.mem.Data.UserPreferences {
			prefs = append(prefs, fmt.Sprintf("%s: %s", k, v))
		}
		extras = append(extras, "━━━ USER PREFERENCES ━━━\n"+strings.Join(prefs, "\n"))
	}

	if len(extras) > 0 {
		return base + "\n\n" + strings.Join(extras, "\n\n")
	}
	return base
}
