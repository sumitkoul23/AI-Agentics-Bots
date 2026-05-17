package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/joho/godotenv"
)

type StrategistAgent struct{}

func (a *StrategistAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	trimmed := strings.TrimSpace(task)
	if trimmed == "" {
		return "Perpetual Markets Strategist AI is online. Ask about a crypto perpetual market, chart setup, sentiment, or trade plan.", nil
	}
	return "Perpetual Markets Strategist AI is online and ready. Runtime verification response: " + trimmed, nil
}

func main() {
	_ = godotenv.Load()

	raw, err := os.ReadFile("perpetual-markets-strategist-ai-metadata.json")
	if err != nil {
		log.Fatal(err)
	}

	var meta struct {
		Name        string `json:"name"`
		AgentID     string `json:"agent_id"`
		Description string `json:"description"`
	}
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal(err)
	}

	cfg := agent.DefaultConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		log.Fatal(err)
	}
	cfg.AgentID = meta.AgentID
	cfg.Name = meta.Name
	cfg.Description = meta.Description

	tokenID, err := strconv.ParseUint(os.Getenv("NFT_TOKEN_ID"), 10, 64)
	if err != nil || tokenID == 0 {
		log.Fatal("NFT_TOKEN_ID must be set to the existing minted token ID")
	}
	log.Printf("Agent ready - existing_token_id=%d", tokenID)

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:       cfg,
		AgentHandler: &StrategistAgent{},
		AgentID:      meta.AgentID,
		TokenID:      tokenID,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
