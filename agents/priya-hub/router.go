package main

import (
	"log"
	"strings"
)

// Router classifies incoming messages via keyword matching and dispatches to the swarm.
// No external API calls — classification is pure Go string matching.
type Router struct {
	registry *Registry
	swarm    *Swarm
}

func NewRouter(registry *Registry, swarm *Swarm) *Router {
	return &Router{registry: registry, swarm: swarm}
}

// Route picks the best agent and returns the swarm response.
func (r *Router) Route(input string) string {
	lower := strings.ToLower(strings.TrimSpace(input))
	agent := r.registry.Route(lower)
	log.Printf("[Router] → %s", agent.ID)
	return r.swarm.Route(agent.ID, input)
}

// Dispatch forces a specific agent by ID. Returns (response, true) or ("", false).
func (r *Router) Dispatch(agentID, input string) (string, bool) {
	if r.registry.Get(agentID) == nil {
		return "", false
	}
	return r.swarm.Route(agentID, input), true
}
