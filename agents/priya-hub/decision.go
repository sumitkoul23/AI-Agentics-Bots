package main

import (
	"fmt"
	"strings"
	"time"
)

// DecisionEngine gives Bodhi explicit control over her decision-making.
//
// It decides:
//   - Should she ask a clarifying/onboarding question?
//   - How confident is she for this agent+input combination?
//   - What assumption is she making (and should she state it)?
//   - What context should be injected into this agent's system prompt?
//
// Decisions are logged so Bodhi can review and improve her own reasoning.
type DecisionEngine struct {
	mem     *Memory
	trainer *Trainer
}

func NewDecisionEngine(mem *Memory, trainer *Trainer) *DecisionEngine {
	return &DecisionEngine{mem: mem, trainer: trainer}
}

// Decide is called before every agent invocation.
// Returns:
//
//	appendQ    — whether to append an onboarding question to the response
//	question   — the question text (empty if appendQ is false)
//	sysAppend  — extra text to append to the agent's system prompt
func (d *DecisionEngine) Decide(agentID, input string) (appendQ bool, question, sysAppend string) {
	appendQ, question = d.trainer.ShouldAsk()

	var parts []string

	// Lessons from self-evaluation
	if lp := d.trainer.LessonsPrompt(agentID); lp != "" {
		parts = append(parts, lp)
	}

	// Behaviour guidance based on current training level
	if bg := d.trainer.BehaviourGuidance(); bg != "" {
		parts = append(parts, bg)
	}

	// If we're going to ask a question, hint to the agent to weave it in naturally
	if appendQ && question != "" {
		parts = append(parts, fmt.Sprintf(
			"After your response, naturally ask this question to better understand the user: \"%s\"", question))
	}

	// Low-confidence assumption note
	score := d.mem.TrainingScore()
	if score < 35 && !isVagueInput(input) {
		assumption := d.buildAssumption(agentID)
		if assumption != "" {
			parts = append(parts, assumption)
		}
	}

	sysAppend = strings.Join(parts, "\n\n")
	return
}

// Confidence returns a 0–1 confidence score for this agent on this user.
func (d *DecisionEngine) Confidence(agentID string) float64 {
	overall := d.mem.TrainingScore() / 100.0
	agentConf := d.mem.AgentConfidence(agentID)
	if agentConf == 0 {
		return overall
	}
	return overall*0.6 + agentConf*0.4
}

// Log records a decision for later analysis and self-improvement.
func (d *DecisionEngine) Log(agentID, input string) {
	conf := d.Confidence(agentID)
	d.mem.AddDecision(StoredDecision{
		At:         time.Now(),
		AgentID:    agentID,
		Input:      truncate(input, 100),
		Confidence: conf,
	})
}

// AppendOnboardQ appends the onboarding question to a response in a natural way.
func (d *DecisionEngine) AppendOnboardQ(response, question string) string {
	if question == "" {
		return response
	}
	score := d.mem.TrainingScore()
	// Personalise the wrapper based on how far along we are
	if score < 10 {
		return fmt.Sprintf("%s\n\n---\n🌸 *One quick question so I can personalise my help:* **%s**", response, question)
	}
	return fmt.Sprintf("%s\n\n---\n*To sharpen my responses:* **%s**", response, question)
}

// buildAssumption creates a context note when Bodhi has partial knowledge.
func (d *DecisionEngine) buildAssumption(agentID string) string {
	prefs := d.mem.GetPreferences()
	facts := d.mem.GetFacts()

	var known []string
	for k, v := range prefs {
		if k != "" && v != "" {
			known = append(known, k+": "+v)
		}
	}
	prefix := agentID + ":"
	for k, v := range facts {
		if strings.HasPrefix(k, prefix) {
			known = append(known, strings.TrimPrefix(k, prefix)+": "+v)
		}
	}

	if len(known) == 0 {
		return ""
	}
	return fmt.Sprintf("Known context about the user: %s\nUse this to personalise your answer. If you make an assumption, state it briefly.", strings.Join(known, " | "))
}

// isVagueInput returns true for very short or content-free messages.
func isVagueInput(input string) bool {
	l := strings.ToLower(strings.TrimSpace(input))
	if len(l) < 8 {
		return true
	}
	for _, v := range []string{"ok", "okay", "yes", "no", "sure", "hmm", "hi", "hello", "hey"} {
		if l == v {
			return true
		}
	}
	return false
}
