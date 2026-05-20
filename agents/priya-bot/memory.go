package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// Memory gives Priya persistent state across restarts.
type Memory struct {
	mu   sync.RWMutex
	path string
	Data MemoryData
}

type MemoryData struct {
	UserVoiceSamples    []string            `json:"user_voice_samples"`
	UserPreferences     map[string]string   `json:"user_preferences"`
	KnownPlatforms      []string            `json:"known_platforms"`
	TrackedJobs         []TrackedJob        `json:"tracked_jobs"`
	ScheduledPosts      []ScheduledPost     `json:"scheduled_posts"`
	ConversationHistory []Message           `json:"conversation_history"`
	LearnedFacts        map[string]string   `json:"learned_facts"`
	OAuthTokens         map[string]OAuthToken `json:"oauth_tokens"`
	LastUpdated         time.Time           `json:"last_updated"`
}

type TrackedJob struct {
	Title     string    `json:"title"`
	Platform  string    `json:"platform"`
	Status    string    `json:"status"`
	AppliedAt time.Time `json:"applied_at"`
	FollowUp  time.Time `json:"follow_up"`
	Notes     string    `json:"notes"`
}

type ScheduledPost struct {
	Platform  string    `json:"platform"`
	Content   string    `json:"content"`
	PostAt    time.Time `json:"post_at"`
	Status    string    `json:"status"` // pending | posted | failed
}

type Message struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func NewMemory(path string) *Memory {
	m := &Memory{path: path, Data: MemoryData{
		UserPreferences: make(map[string]string),
		LearnedFacts:    make(map[string]string),
	}}
	_ = m.load()
	return m
}

func (m *Memory) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &m.Data)
}

func (m *Memory) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.LastUpdated = time.Now()
	raw, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, raw, 0600)
}

func (m *Memory) AddMessage(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.ConversationHistory = append(m.Data.ConversationHistory, Message{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	// Keep last 100 messages to avoid unbounded growth
	if len(m.Data.ConversationHistory) > 100 {
		m.Data.ConversationHistory = m.Data.ConversationHistory[len(m.Data.ConversationHistory)-100:]
	}
}

func (m *Memory) Learn(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.LearnedFacts[key] = value
}

func (m *Memory) SetPreference(key, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UserPreferences[key] = value
}

func (m *Memory) GetPreference(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.UserPreferences[key]
}

func (m *Memory) AddVoiceSample(sample string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UserVoiceSamples = append(m.Data.UserVoiceSamples, sample)
	if len(m.Data.UserVoiceSamples) > 20 {
		m.Data.UserVoiceSamples = m.Data.UserVoiceSamples[len(m.Data.UserVoiceSamples)-20:]
	}
}

func (m *Memory) RecentHistory(n int) []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.Data.ConversationHistory
	if len(h) <= n {
		return h
	}
	return h[len(h)-n:]
}

func (m *Memory) AddTrackedJob(job TrackedJob) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.TrackedJobs = append(m.Data.TrackedJobs, job)
}

func (m *Memory) AddScheduledPost(post ScheduledPost) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.ScheduledPosts = append(m.Data.ScheduledPosts, post)
}
