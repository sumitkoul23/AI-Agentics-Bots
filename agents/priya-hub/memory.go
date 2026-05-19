package main

import (
	"encoding/json"
	"os"
	"sync"
	"time"
)

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
}

type Turn struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

func NewMemory(path string) *Memory {
	m := &Memory{
		path: path,
		Data: MemoryData{
			Preferences: make(map[string]string),
			Facts:       make(map[string]string),
		},
	}
	raw, err := os.ReadFile(path)
	if err == nil {
		json.Unmarshal(raw, &m.Data)
	}
	return m
}

func (m *Memory) Save() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.UpdatedAt = time.Now()
	raw, _ := json.MarshalIndent(m.Data, "", "  ")
	os.WriteFile(m.path, raw, 0600)
}

func (m *Memory) Push(role, content string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Data.History = append(m.Data.History, Turn{role, content})
	if len(m.Data.History) > 80 {
		m.Data.History = m.Data.History[len(m.Data.History)-80:]
	}
}

func (m *Memory) Set(k, v string) { m.mu.Lock(); m.Data.Preferences[k] = v; m.mu.Unlock() }
func (m *Memory) Get(k string) string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.Data.Preferences[k]
}
func (m *Memory) Learn(k, v string) { m.mu.Lock(); m.Data.Facts[k] = v; m.mu.Unlock() }
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
