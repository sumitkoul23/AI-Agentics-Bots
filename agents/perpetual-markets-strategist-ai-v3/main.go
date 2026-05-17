package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/joho/godotenv"
)

type StrategistAgent struct{}

type Metadata struct {
	Name             string             `json:"name"`
	AgentID          string             `json:"agent_id"`
	ShortDescription string             `json:"short_description"`
	Description      string             `json:"description"`
	AgentType        string             `json:"agent_type"`
	Categories       []string           `json:"categories"`
	Capabilities     []types.Capability `json:"capabilities"`
	Commands         json.RawMessage    `json:"commands"`
	NLPFallback      bool               `json:"nlp_fallback"`
	FAQItems         json.RawMessage    `json:"faq_items"`
}

func (a *StrategistAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	trimmed := strings.TrimSpace(task)
	if trimmed == "" {
		return "Perpetual Markets Strategist AI is online. Ask about a crypto perpetual market, chart setup, sentiment, or trade plan.", nil
	}
	return "Perpetual Markets Strategist AI is online and ready. Runtime verification response: " + trimmed, nil
}

func main() {
	_ = godotenv.Load()

	raw, err := os.ReadFile("perpetual-markets-strategist-ai-v3-metadata.json")
	if err != nil {
		log.Fatal(err)
	}

	var meta Metadata
	if err := json.Unmarshal(raw, &meta); err != nil {
		log.Fatal(err)
	}

	capabilitiesJSON, err := json.Marshal(meta.Capabilities)
	if err != nil {
		log.Fatal(err)
	}
	categoriesJSON, err := json.Marshal(meta.Categories)
	if err != nil {
		log.Fatal(err)
	}

	result, err := deploy.DeployAgent(deploy.DeployConfig{
		PrivateKey:       os.Getenv("PRIVATE_KEY"),
		AgentID:          meta.AgentID,
		AgentName:        meta.Name,
		Description:      meta.Description,
		AgentType:        meta.AgentType,
		Capabilities:     capabilitiesJSON,
		Commands:         meta.Commands,
		NlpFallback:      meta.NLPFallback,
		Categories:       categoriesJSON,
		ShortDescription: meta.ShortDescription,
		FAQItems:         meta.FAQItems,
		MetadataVersion:  "2.4.0",
		StateFilePath:    ".teneo-deploy-state-v3.json",
	})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Agent ready - token_id=%d", result.TokenID)

	cfg := agent.DefaultConfig()
	if err := cfg.LoadFromEnv(); err != nil {
		log.Fatal(err)
	}
	cfg.AgentID = meta.AgentID
	cfg.Name = meta.Name
	cfg.Description = meta.Description
	cfg.ShortDescription = meta.ShortDescription
	cfg.CapabilityDetails = meta.Capabilities
	cfg.Capabilities = make([]string, 0, len(meta.Capabilities))
	for _, capability := range meta.Capabilities {
		cfg.Capabilities = append(cfg.Capabilities, capability.Name)
	}

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:          cfg,
		AgentHandler:    &StrategistAgent{},
		AgentID:         meta.AgentID,
		TokenID:         result.TokenID,
		SubmitForReview: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
