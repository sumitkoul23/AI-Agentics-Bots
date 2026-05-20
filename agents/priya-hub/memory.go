package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// ── Types ──────────────────────────────────────────────────────────────────────

type Memory struct {
	mu   sync.RWMutex
	path string
	Data MemoryData
}

type MemoryData struct {
	VoiceSamples []string          `json:"voice_samples"`
	Preferences  map[string]string `json:"preferences"`
	Facts        map[string]string `json:"facts"`
	History      []Turn            `json:"history"`
	UpdatedAt    time.Time         `json:"updated_at"`

	// Training & self-improvement state
	TrainingScore  float64            `json:"training_score"`   // 0–100
	Interactions   int                `json:"interactions"`     // lifetime turns
	OnboardDone    bool               `json:"onboard_done"`
	OnboardStep    int                `json:"onboard_step"`     // which question is next
	AgentConf      map[string]float64 `json:"agent_conf"`       // per-agent confidence 0–1
	DecisionLog    []StoredDecision   `json:"decision_log"`     // last 50 decisions
	Lessons        []StoredLesson     `json:"lessons"`          // self-eval lessons
}

type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// StoredDecision records why Bodhi decided what she decided.
type StoredDecision struct {
	At         time.Time `json:"at"`
	AgentID    string    `json:"agent_id"`
	Input      string    `json:"input"`
	Confidence float64   `json:"confidence"`
	Reasoning  string    `json:"reasoning,omitempty"`
}

// StoredLesson is a self-evaluation result that Bodhi learns from.
type StoredLesson struct {
	At      time.Time `json:"at"`
	AgentID string    `json:"agent_id"`
	Lesson  string    `json:"lesson"`
	Score   float64   `json:"score"` // 0–1, lower = more room to improve
}

// ── Constructor ────────────────────────────────────────────────────────────────

func NewMemory(path string) *Memory {
	m := &Memory{
		path: path,
		Data: MemoryData{
			Preferences: make(map[string]string),
			Facts:       make(map[string]string),
			AgentConf:   make(map[string]float64),
		},
	}
	raw, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(raw, &m.Data)
	}
	// Init nil maps after JSON load
	if m.Data.Preferences == nil {
		m.Data.Preferences = make(map[string]string)
	}
	if m.Data.Facts == nil {
		m.Data.Facts = make(map[string]string)
	}
	if m.Data.AgentConf == nil {
		m.Data.AgentConf = make(map[string]float64)
	}
	return m
}

// ── Persistence ────────────────────────────────────────────────────────────────

func (m *Memory) Save() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UpdatedAt = time.Now()
	raw, _ := json.MarshalIndent(m.Data, "", "  ")
	os.WriteFile(m.path, raw, 0600)
}

// ── Conversation history ───────────────────────────────────────────────────────

func (m *Memory) Push(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.History = append(m.Data.History, Turn{role, content})
	if len(m.Data.History) > 80 {
		m.Data.History = m.Data.History[len(m.Data.History)-80:]
	}
}

func (m *Memory) RecentHistory(n int) []Turn {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.Data.History
	if len(h) > n {
		h = h[len(h)-n:]
	}
	result := make([]Turn, len(h))
	copy(result, h)
	return result
}

// ── Key-value stores ───────────────────────────────────────────────────────────

func (m *Memory) Set(k, v string)  { m.mu.Lock(); m.Data.Preferences[k] = v; m.mu.Unlock() }
func (m *Memory) Get(k string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.Preferences[k]
}
func (m *Memory) Learn(k, v string) { m.mu.Lock(); m.Data.Facts[k] = v; m.mu.Unlock() }

func (m *Memory) GetFacts() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.Data.Facts))
	for k, v := range m.Data.Facts {
		out[k] = v
	}
	return out
}

func (m *Memory) GetPreferences() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	out := make(map[string]string, len(m.Data.Preferences))
	for k, v := range m.Data.Preferences {
		out[k] = v
	}
	return out
}

func (m *Memory) AddVoice(s string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.VoiceSamples = append(m.Data.VoiceSamples, s)
	if len(m.Data.VoiceSamples) > 20 {
		m.Data.VoiceSamples = m.Data.VoiceSamples[len(m.Data.VoiceSamples)-20:]
	}
}

// ── Training & confidence ─────────────────────────────────────────────────────

func (m *Memory) TrainingScore() float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.TrainingScore
}

func (m *Memory) UpdateTrainingScore(delta float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.TrainingScore += delta
	if m.Data.TrainingScore < 0 {
		m.Data.TrainingScore = 0
	}
	if m.Data.TrainingScore > 100 {
		m.Data.TrainingScore = 100
	}
}

func (m *Memory) AddInteraction() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.Interactions++
}

func (m *Memory) AgentConfidence(agentID string) float64 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.AgentConf[agentID]
}

func (m *Memory) SetAgentConfidence(agentID string, v float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	m.Data.AgentConf[agentID] = v
}

// ── Onboarding ─────────────────────────────────────────────────────────────────

func (m *Memory) GetOnboardStep() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.OnboardStep
}

func (m *Memory) AdvanceOnboard() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.OnboardStep++
}

func (m *Memory) IsOnboardDone() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.OnboardDone
}

func (m *Memory) SetOnboardDone() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.OnboardDone = true
}

// ── Decision log ──────────────────────────────────────────────────────────────

func (m *Memory) AddDecision(d StoredDecision) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.DecisionLog = append(m.Data.DecisionLog, d)
	if len(m.Data.DecisionLog) > 50 {
		m.Data.DecisionLog = m.Data.DecisionLog[len(m.Data.DecisionLog)-50:]
	}
}

// ── Self-eval lessons ─────────────────────────────────────────────────────────

func (m *Memory) AddLesson(agentID, lesson string, score float64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.Lessons = append(m.Data.Lessons, StoredLesson{
		At: time.Now(), AgentID: agentID, Lesson: lesson, Score: score,
	})
	if len(m.Data.Lessons) > 40 {
		m.Data.Lessons = m.Data.Lessons[len(m.Data.Lessons)-40:]
	}
}

func (m *Memory) GetLessons(agentID string) []StoredLesson {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []StoredLesson
	for _, l := range m.Data.Lessons {
		if l.AgentID == agentID || l.AgentID == "" {
			out = append(out, l)
		}
	}
	// Return most recent 5
	if len(out) > 5 {
		out = out[len(out)-5:]
	}
	return out
}
