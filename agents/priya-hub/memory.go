package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

// HubMemory gives the hub persistent conversation state across restarts.
type HubMemory struct {
	mu   sync.RWMutex
	path string
	Data HubMemoryData
}

type HubMemoryData struct {
	UserVoiceSamples    []string          `json:"user_voice_samples"`
	UserPreferences     map[string]string `json:"user_preferences"`
	ConversationHistory []HubMessage      `json:"conversation_history"`
	LearnedFacts        map[string]string `json:"learned_facts"`
	LastUpdated         time.Time         `json:"last_updated"`
}

type HubMessage struct {
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

func NewHubMemory(path string) *HubMemory {
	m := &HubMemory{
		path: path,
		Data: HubMemoryData{
			UserPreferences: make(map[string]string),
			LearnedFacts:    make(map[string]string),
		},
	}
	_ = m.load()
	return m
}

func (m *HubMemory) load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	raw, err := os.ReadFile(m.path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &m.Data)
}

func (m *HubMemory) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.LastUpdated = time.Now()
	raw, err := json.MarshalIndent(m.Data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.path, raw, 0600)
}

func (m *HubMemory) AddMessage(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.ConversationHistory = append(m.Data.ConversationHistory, HubMessage{
		Role:      role,
		Content:   content,
		Timestamp: time.Now(),
	})
	if len(m.Data.ConversationHistory) > 100 {
		m.Data.ConversationHistory = m.Data.ConversationHistory[len(m.Data.ConversationHistory)-100:]
	}
}

func (m *HubMemory) AddVoiceSample(sample string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UserVoiceSamples = append(m.Data.UserVoiceSamples, sample)
	if len(m.Data.UserVoiceSamples) > 20 {
		m.Data.UserVoiceSamples = m.Data.UserVoiceSamples[len(m.Data.UserVoiceSamples)-20:]
	}
}

func (m *HubMemory) SetPreference(key, val string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UserPreferences[key] = val
}

func (m *HubMemory) GetPreference(key string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.UserPreferences[key]
}

func (m *HubMemory) Learn(key, val string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.LearnedFacts[key] = val
}

func (m *HubMemory) RecentHistory(n int) []HubMessage {
	m.mu.RLock()
	defer m.mu.RUnlock()
	h := m.Data.ConversationHistory
	if len(h) <= n {
		return append([]HubMessage(nil), h...)
	}
	return append([]HubMessage(nil), h[len(h)-n:]...)
}
