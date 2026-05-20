package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"strings"
	"sync"
	"time"
)

// onboardQuestions is the ordered cold-start interview.
// Bodhi asks these one at a time at the start of each response until done.
var onboardQuestions = []string{
	"What's your main focus — developer, trader, freelancer, marketer, or something else?",
	"What specific tools, languages, or markets do you work with most?",
	"Are you active in crypto or stock markets? What's your experience level?",
	"Which social platforms are most important for your work — LinkedIn, Twitter/X, Instagram, TikTok…?",
	"What's the single biggest challenge you want me to help you with regularly?",
}

// Trainer manages Bodhi's progressive self-improvement cycle:
//   - Onboarding (cold start questions → user profile)
//   - Confidence scoring (how well she knows this user)
//   - Self-evaluation (review her own outputs, extract lessons)
//   - Score updates (track improvement over time)
type Trainer struct {
	mem     *Memory
	ollama  *OllamaClient
	mu      sync.Mutex
	evalBuf []evalEntry
}

type evalEntry struct {
	agentID  string
	input    string
	response string
}

func NewTrainer(mem *Memory, ollama *OllamaClient) *Trainer {
	return &Trainer{mem: mem, ollama: ollama}
}

// ── Onboarding ─────────────────────────────────────────────────────────────────

// NextOnboardQuestion returns the next question Bodhi should ask, or "".
func (t *Trainer) NextOnboardQuestion() string {
	if t.mem.IsOnboardDone() {
		return ""
	}
	step := t.mem.GetOnboardStep()
	if step >= len(onboardQuestions) {
		t.mem.SetOnboardDone()
		t.mem.Save()
		return ""
	}
	return onboardQuestions[step]
}

// RecordOnboardAnswer marks the current question answered and advances.
func (t *Trainer) RecordOnboardAnswer() {
	t.mem.AdvanceOnboard()
	step := t.mem.GetOnboardStep()
	if step >= len(onboardQuestions) {
		t.mem.SetOnboardDone()
		t.mem.UpdateTrainingScore(20) // onboarding completion bonus
		log.Printf("[Trainer] onboarding complete — score: %.0f/100", t.mem.TrainingScore())
	} else {
		t.mem.UpdateTrainingScore(3) // small per-answer bonus
	}
	t.mem.Save()
}

// ShouldAsk decides whether Bodhi should append an onboarding question.
// Returns (inject bool, question string).
// Early on: every other exchange. Later: never.
func (t *Trainer) ShouldAsk() (bool, string) {
	if t.mem.IsOnboardDone() {
		return false, ""
	}
	score := t.mem.TrainingScore()
	if score >= 40 {
		// User has shared enough context; stop asking
		t.mem.SetOnboardDone()
		t.mem.Save()
		return false, ""
	}
	// Ask every other turn (not every single one — let the user breathe)
	interactions := t.mem.Data.Interactions
	if interactions%2 != 0 {
		return false, ""
	}
	q := t.NextOnboardQuestion()
	return q != "", q
}

// ── Interaction recording ─────────────────────────────────────────────────────

// Record is called after every exchange. It updates training score and buffers
// the exchange for batch self-evaluation.
func (t *Trainer) Record(agentID, input, response string) {
	t.mem.AddInteraction()

	// Incremental score gain — diminishing returns as score approaches 100
	score := t.mem.TrainingScore()
	gain := (100-score)*0.008 + 0.1 // slows down near the top
	t.mem.UpdateTrainingScore(gain)

	// Agent-specific confidence nudge
	existing := t.mem.AgentConfidence(agentID)
	newConf := existing + (1-existing)*0.03
	t.mem.SetAgentConfidence(agentID, newConf)

	// Buffer for batch eval
	t.mu.Lock()
	t.evalBuf = append(t.evalBuf, evalEntry{agentID, input, response})
	bufLen := len(t.evalBuf)
	t.mu.Unlock()

	// Every 8 interactions trigger a background self-eval cycle
	if bufLen >= 8 {
		go t.runSelfEval()
	}
}

// ── Self-evaluation loop ──────────────────────────────────────────────────────

func (t *Trainer) runSelfEval() {
	if t.ollama == nil {
		return
	}

	t.mu.Lock()
	batch := make([]evalEntry, len(t.evalBuf))
	copy(batch, t.evalBuf)
	t.evalBuf = nil
	t.mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 25*time.Second)
	defer cancel()

	// Build transcript
	var sb strings.Builder
	for _, e := range batch {
		sb.WriteString(fmt.Sprintf("User: %s\nBodhi: %s\n\n",
			truncate(e.input, 120), truncate(e.response, 250)))
	}

	system := `You are evaluating an AI assistant named Bodhi. Review the conversation batch.

Output EXACTLY in this format (no extra text):
SCORE: <1-10>
WEAKNESS: <one sentence about the main weakness>
LESSON: <one concrete instruction to improve future responses>`

	result, err := t.ollama.GenerateShort(ctx, system, sb.String())
	if err != nil {
		log.Printf("[Trainer] self-eval error: %v", err)
		return
	}

	var score float64 = 7
	var lesson string
	var weakness string

	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SCORE:") {
			fmt.Sscanf(strings.TrimPrefix(line, "SCORE:"), " %f", &score)
		}
		if strings.HasPrefix(line, "LESSON:") {
			lesson = strings.TrimSpace(strings.TrimPrefix(line, "LESSON:"))
		}
		if strings.HasPrefix(line, "WEAKNESS:") {
			weakness = strings.TrimSpace(strings.TrimPrefix(line, "WEAKNESS:"))
		}
	}

	if lesson == "" || score <= 0 {
		return
	}

	agentID := batch[len(batch)-1].agentID
	t.mem.AddLesson(agentID, lesson, score/10)

	// Penalise low scores, reward excellent ones
	if score < 5 {
		t.mem.UpdateTrainingScore(-1.5)
		log.Printf("[Trainer] self-eval %.0f/10 — weakness: %s", score, weakness)
	} else if score >= 9 {
		t.mem.UpdateTrainingScore(0.5)
	}
	t.mem.Save()
	log.Printf("[Trainer] self-eval complete — score %.0f/10, lesson stored", score)
}

// ── Context helpers ───────────────────────────────────────────────────────────

// LessonsPrompt returns a formatted lessons block for injection into system prompts.
func (t *Trainer) LessonsPrompt(agentID string) string {
	lessons := t.mem.GetLessons(agentID)
	if len(lessons) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Apply these lessons from your previous self-evaluations:\n")
	for _, l := range lessons {
		sb.WriteString(fmt.Sprintf("• %s\n", l.Lesson))
	}
	return sb.String()
}

// BehaviourGuidance returns guidance based on training level.
// At low score: stay curious and ask follow-ups.
// At high score: act decisively, skip preamble.
func (t *Trainer) BehaviourGuidance() string {
	score := t.mem.TrainingScore()
	switch {
	case score < 20:
		return "You are still learning about this user. Be warm, ask clarifying questions naturally. End responses with one question."
	case score < 45:
		return "You have some context. State your assumptions briefly. Occasionally ask a follow-up to build your model of the user."
	case score < 70:
		return "You know this user moderately well. Make reasonable assumptions without explaining them. Ask only when genuinely needed."
	default:
		return "You know this user well. Be decisive and autonomous. Skip preamble, jump straight to the answer."
	}
}

// TrailNote returns a subtle footer note for low-confidence responses.
// Returns "" once the user is well-known.
func (t *Trainer) TrailNote() string {
	score := t.mem.TrainingScore()
	if score >= 65 {
		return ""
	}
	pct := math.Round(score)
	if score < 20 {
		return fmt.Sprintf("\n\n*Training: %.0f%% — I'm still learning your context. The more we talk, the better I'll get.*", pct)
	}
	return ""
}

// StatusLine returns a one-line summary for the /status endpoint.
func (t *Trainer) StatusLine() string {
	score := t.mem.TrainingScore()
	interactions := t.mem.Data.Interactions
	onboard := "in progress"
	if t.mem.IsOnboardDone() {
		onboard = "complete"
	}
	return fmt.Sprintf("Training: %.0f/100 | %d interactions | Onboarding: %s",
		score, interactions, onboard)
}
