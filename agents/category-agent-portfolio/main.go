package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/agent"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/deploy"
	"github.com/TeneoProtocolAI/teneo-agent-sdk/pkg/types"
	"github.com/joho/godotenv"
)

type PortfolioAgent struct {
	meta Metadata
}

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
	HealthPort       int                `json:"health_port"`
	OpeningLine      string             `json:"opening_line"`
	OutputStyle      string             `json:"output_style"`
}

func (a *PortfolioAgent) ProcessTask(ctx context.Context, task string) (string, error) {
	trimmed := strings.TrimSpace(task)
	if trimmed == "" {
		return a.meta.OpeningLine, nil
	}

	return fmt.Sprintf(
		"%s\n\nRequest received: %s\n\nResponse style: %s",
		a.meta.OpeningLine,
		trimmed,
		a.meta.OutputStyle,
	), nil
}

func main() {
	_ = godotenv.Load()

	metadataFile := os.Getenv("AGENT_METADATA_FILE")
	if len(os.Args) > 1 {
		metadataFile = os.Args[1]
	}
	if metadataFile == "" {
		log.Fatal("AGENT_METADATA_FILE or metadata path argument is required")
	}

	raw, err := os.ReadFile(metadataFile)
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
		StateFilePath:    fmt.Sprintf(".teneo-deploy-state-%s.json", meta.AgentID),
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
	cfg.HealthPort = meta.HealthPort

	a, err := agent.NewEnhancedAgent(&agent.EnhancedAgentConfig{
		Config:          cfg,
		AgentHandler:    &PortfolioAgent{meta: meta},
		AgentID:         meta.AgentID,
		TokenID:         result.TokenID,
		SubmitForReview: true,
		StateFilePath:   filepath.Base(fmt.Sprintf(".teneo-runtime-state-%s.json", meta.AgentID)),
	})
	if err != nil {
		log.Fatal(err)
	}

	if err := a.Run(); err != nil {
		log.Fatal(err)
	}
}
