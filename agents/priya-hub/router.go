package main

import (
	"context"
	"log"
	"os"
	"strings"

	anthropic "github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Router classifies messages and dispatches to the right AgentDef via Claude.
type Router struct {
	client   *anthropic.Client
	registry *Registry
	mem      *HubMemory
}

func NewRouter(registry *Registry, mem *HubMemory) *Router {
	key := os.Getenv("ANTHROPIC_API_KEY")
	if key == "" {
		log.Println("[Router] ANTHROPIC_API_KEY not set — running in keyword-only mode")
		return &Router{registry: registry, mem: mem}
	}
	client := anthropic.NewClient(option.WithAPIKey(key))
	return &Router{client: client, registry: registry, mem: mem}
}

// Route selects the best agent and returns its response.
func (r *Router) Route(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(strings.TrimSpace(input))

	// 1. Fast keyword routing — no API call needed
	if id := r.registry.FastRoute(lower); id != "" {
		log.Printf("[Router] fast-route → %s", id)
		return r.dispatch(ctx, id, input)
	}

	// 2. Claude-powered classification for ambiguous input
	if r.client != nil {
		id := r.classify(ctx, input)
		if id != "" {
			log.Printf("[Router] claude-route → %s", id)
			return r.dispatch(ctx, id, input)
		}
	}

	// 3. Fallback to Priya
	log.Printf("[Router] fallback → priya")
	return r.dispatch(ctx, "priya", input)
}

// classify asks Claude (Haiku — fast + cheap) to pick the best agent ID.
func (r *Router) classify(ctx context.Context, input string) string {
	resp, err := r.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3HaikuLatest,
		MaxTokens: 20,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(r.registry.ClassifyPrompt(input))),
		},
	})
	if err != nil {
		log.Printf("[Router] classify error: %v", err)
		return ""
	}
	for _, block := range resp.Content {
		if block.Type == "text" {
			id := strings.ToLower(strings.TrimSpace(block.Text))
			// Strip any surrounding punctuation
			id = strings.Trim(id, `"'.,:; `)
			if r.registry.Get(id) != nil {
				return id
			}
		}
	}
	return ""
}

// dispatch sends the input to the agent and returns its response.
func (r *Router) dispatch(ctx context.Context, agentID, input string) (string, error) {
	agent := r.registry.Get(agentID)
	if agent == nil {
		agent = r.registry.Get("priya")
	}

	if r.client == nil {
		return "Set ANTHROPIC_API_KEY to enable AI responses. Agent selected: " + agent.Name, nil
	}

	// Build conversation context from memory
	history := r.mem.RecentHistory(24)
	var msgs []anthropic.MessageParam
	for _, h := range history {
		switch h.Role {
		case "user":
			msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(h.Content)))
		case "assistant":
			// Strip the [Agent Name] prefix before sending back to Claude
			content := h.Content
			if idx := strings.Index(content, "] "); idx > 0 && strings.HasPrefix(content, "[") {
				content = content[idx+2:]
			}
			msgs = append(msgs, anthropic.NewAssistantMessage(anthropic.NewTextBlock(content)))
		}
	}
	msgs = append(msgs, anthropic.NewUserMessage(anthropic.NewTextBlock(input)))

	resp, err := r.client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.ModelClaude3_7SonnetLatest,
		MaxTokens: 2048,
		System: []anthropic.TextBlockParam{
			{Text: agent.SystemPrompt},
		},
		Messages: msgs,
	})
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, block := range resp.Content {
		if block.Type == "text" {
			sb.WriteString(block.Text)
		}
	}
	reply := sb.String()

	// Persist with agent tag so future context knows which agent handled it
	r.mem.AddMessage("user", input)
	r.mem.AddMessage("assistant", "["+agent.Name+"] "+reply)
	_ = r.mem.Save()

	return reply, nil
}

// IsReady returns true when the Claude client is available.
func (r *Router) IsReady() bool {
	return r.client != nil
}
