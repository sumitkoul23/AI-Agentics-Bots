package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"
)

// SwarmMessage is the unit passed between the coordinator and agent goroutines.
type SwarmMessage struct {
	From    string
	AgentID string
	Input   string
	Reply   chan string // nil for fire-and-forget autonomous tasks
}

// swarmAgent is a specialized goroutine — one per registered agent.
type swarmAgent struct {
	id       string
	name     string
	inbox    chan SwarmMessage
	system   string
	handler  func(string, *Memory) string // template fallback
	ollama   *OllamaClient
	mem      *Memory
	learner  *Learner
	trainer  *Trainer
	decision *DecisionEngine
}

func (a *swarmAgent) run(ctx context.Context) {
	log.Printf("[Swarm] %s started", a.id)
	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-a.inbox:
			if !ok {
				return
			}
			reply := a.process(ctx, msg.Input)
			if msg.Reply != nil {
				msg.Reply <- reply
			}
		}
	}
}

func (a *swarmAgent) process(ctx context.Context, input string) string {
	genCtx, cancel := context.WithTimeout(ctx, 90*time.Second)
	defer cancel()

	// ── Decision layer ────────────────────────────────────────────────────────
	// Decide what to do before generating: inject questions, lessons, guidance.
	appendQ, question, sysAppend := a.decision.Decide(a.id, input)

	// Build the full system prompt for this request
	systemPrompt := a.system
	if lctx := a.learner.BuildContext(a.id); lctx != "" {
		systemPrompt += "\n\n" + lctx
	}
	if sysAppend != "" {
		systemPrompt += "\n\n" + sysAppend
	}

	var response string

	// ── Ollama (on-device LLM) ────────────────────────────────────────────────
	if a.ollama != nil {
		history := a.mem.RecentHistory(8)
		var hb strings.Builder
		for _, t := range history {
			if t.Role == "user" {
				hb.WriteString("User: " + t.Content + "\n")
			} else {
				hb.WriteString("Assistant: " + t.Content + "\n")
			}
		}
		prompt := input
		if hb.Len() > 0 {
			prompt = "Recent conversation:\n" + hb.String() + "\nUser: " + input
		}
		var err error
		response, err = a.ollama.Generate(genCtx, systemPrompt, prompt)
		if err != nil {
			log.Printf("[%s] ollama error: %v — using template", a.id, err)
		}
	}

	// ── Template fallback ─────────────────────────────────────────────────────
	if response == "" {
		response = a.handler(input, a.mem)
	}

	// ── Onboarding question injection ─────────────────────────────────────────
	if appendQ && question != "" {
		response = a.decision.AppendOnboardQ(response, question)
		a.trainer.RecordOnboardAnswer()
	}

	// ── Training trail note (shown at low confidence) ─────────────────────────
	if note := a.trainer.TrailNote(); note != "" {
		response += note
	}

	// ── Persist ───────────────────────────────────────────────────────────────
	a.mem.Push("user", input)
	a.mem.Push("assistant", response)
	a.mem.Save()

	// ── Async: learn, record, log decision ───────────────────────────────────
	go func() {
		a.learner.Learn(a.id, input, response)
		a.trainer.Record(a.id, input, response)
	}()
	a.decision.Log(a.id, input)

	return response
}

// Notification is an autonomous insight pushed to connected clients via SSE.
type Notification struct {
	From string    `json:"from"`
	Text string    `json:"text"`
	At   time.Time `json:"at"`
}

// NotifBus is a simple fan-out pub/sub for SSE clients.
type NotifBus struct {
	mu   sync.Mutex
	subs map[int]chan Notification
	next int
}

func NewNotifBus() *NotifBus { return &NotifBus{subs: make(map[int]chan Notification)} }

func (b *NotifBus) Subscribe() (int, chan Notification) {
	b.mu.Lock()
	defer b.mu.Unlock()
	id := b.next
	b.next++
	ch := make(chan Notification, 8)
	b.subs[id] = ch
	return id, ch
}

func (b *NotifBus) Unsubscribe(id int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.subs, id)
}

func (b *NotifBus) Publish(n Notification) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, ch := range b.subs {
		select {
		case ch <- n:
		default: // drop if subscriber is slow
		}
	}
}

func (b *NotifBus) Count() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.subs)
}

// Swarm coordinates all specialized agents and the autonomous background scheduler.
type Swarm struct {
	agents  map[string]*swarmAgent
	ollama  *OllamaClient
	mem     *Memory
	learner *Learner
	trainer *Trainer
	Notifs  *NotifBus
	cancel  context.CancelFunc
	wg      sync.WaitGroup
	mu      sync.RWMutex
}

func NewSwarm(registry *Registry, mem *Memory) *Swarm {
	ollama := NewOllamaClient()
	if ollama.IsAvailable() {
		ollama.AutoModel()
		log.Printf("[Swarm] Ollama online — model: %s", ollama.Model)
	} else {
		log.Printf("[Swarm] Ollama not found — using template responses (start Ollama for AI)")
		ollama = nil
	}

	learner := NewLearner(mem, ollama)
	trainer := NewTrainer(mem, ollama)
	decision := NewDecisionEngine(mem, trainer)

	s := &Swarm{
		agents:  make(map[string]*swarmAgent),
		ollama:  ollama,
		mem:     mem,
		learner: learner,
		trainer: trainer,
		Notifs:  NewNotifBus(),
	}

	for _, a := range registry.list {
		s.agents[a.ID] = &swarmAgent{
			id:       a.ID,
			name:     a.Name,
			inbox:    make(chan SwarmMessage, 32),
			system:   agentSystemPrompt(a.ID, a.Name, a.Desc),
			handler:  a.Handle,
			ollama:   ollama,
			mem:      mem,
			learner:  learner,
			trainer:  trainer,
			decision: decision,
		}
	}

	return s
}

// Start launches all agent goroutines and the autonomous background worker.
func (s *Swarm) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	for _, agent := range s.agents {
		a := agent
		s.wg.Add(1)
		go func() {
			defer s.wg.Done()
			a.run(ctx)
		}()
	}

	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		s.runAutonomous(ctx)
	}()

	log.Printf("[Swarm] %d agents active", len(s.agents))
}

// Stop gracefully shuts down the swarm.
func (s *Swarm) Stop() {
	if s.cancel != nil {
		s.cancel()
	}
	s.wg.Wait()
}

// Route dispatches input to the given agent and blocks until a reply arrives.
func (s *Swarm) Route(agentID, input string) string {
	s.mu.RLock()
	agent, ok := s.agents[agentID]
	s.mu.RUnlock()
	if !ok {
		agent = s.agents["priya"]
	}

	reply := make(chan string, 1)
	select {
	case agent.inbox <- SwarmMessage{From: "user", AgentID: agentID, Input: input, Reply: reply}:
	case <-time.After(5 * time.Second):
		return "All agents are busy — please try again in a moment."
	}

	select {
	case r := <-reply:
		return r
	case <-time.After(95 * time.Second):
		return "The request took too long. Try a shorter question or check Ollama status with /status."
	}
}

// runAutonomous drives background insight generation on a schedule.
func (s *Swarm) runAutonomous(ctx context.Context) {
	t4h := time.NewTicker(4 * time.Hour)
	t8h := time.NewTicker(8 * time.Hour)
	defer t4h.Stop()
	defer t8h.Stop()

	// Initial warm-up: let agents settle for 30 seconds before first autonomous task
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}

	for {
		select {
		case <-ctx.Done():
			return
		case <-t4h.C:
			s.autonomousFire("finance",
				"Review any stored market context and generate a brief insight summary for the user.")
		case <-t8h.C:
			s.autonomousFire("organizer",
				"Based on what you know about the user, generate a concise status check and one actionable suggestion.")
		}
	}
}

// autonomousFire dispatches a background task and publishes the reply to NotifBus.
func (s *Swarm) autonomousFire(agentID, task string) {
	s.mu.RLock()
	agent, ok := s.agents[agentID]
	s.mu.RUnlock()
	if !ok {
		return
	}
	reply := make(chan string, 1)
	select {
	case agent.inbox <- SwarmMessage{From: "system", AgentID: agentID, Input: task, Reply: reply}:
		log.Printf("[Swarm:auto] dispatched to %s", agentID)
	default:
		log.Printf("[Swarm:auto] %s inbox full — skipping", agentID)
		return
	}
	go func() {
		select {
		case r := <-reply:
			s.Notifs.Publish(Notification{From: agentID, Text: r, At: time.Now()})
			log.Printf("[Swarm:auto] %s insight published to %d subscribers", agentID, s.Notifs.Count())
		case <-time.After(120 * time.Second):
			log.Printf("[Swarm:auto] %s timed out waiting for reply", agentID)
		}
	}()
}

// Status returns a human-readable swarm summary.
func (s *Swarm) Status() string {
	var sb strings.Builder
	sb.WriteString("━━ Priya Swarm Status ━━━━━━━━━━━━━━━━━━━━━━\n\n")

	if s.ollama != nil {
		sb.WriteString(fmt.Sprintf("AI Engine : Ollama (%s)\n", s.ollama.Model))
		sb.WriteString("Mode      : On-device AI — no external API calls\n")
	} else {
		sb.WriteString("AI Engine : Template mode\n")
		sb.WriteString("Mode      : Install Ollama + pull a model to enable AI\n")
		sb.WriteString("           brew install ollama && ollama pull llama3.2\n")
	}

	facts := s.mem.GetFacts()
	prefs := s.mem.GetPreferences()
	hist := s.mem.RecentHistory(1000)

	sb.WriteString(fmt.Sprintf("\nMemory    : %d facts | %d preferences | %d conversation turns\n",
		len(facts), len(prefs), len(hist)))

	if s.trainer != nil {
		sb.WriteString(fmt.Sprintf("\n%s\n", s.trainer.StatusLine()))
	}

	sb.WriteString(fmt.Sprintf("\nAgents (%d):\n", len(s.agents)))
	for id, a := range s.agents {
		conf := s.mem.AgentConfidence(id)
		sb.WriteString(fmt.Sprintf("  %-20s  queue: %d  confidence: %.0f%%\n",
			id, len(a.inbox), conf*100))
	}
	return sb.String()
}

// ── System prompts ────────────────────────────────────────────────────────────

func agentSystemPrompt(id, name, desc string) string {
	base := fmt.Sprintf(`You are %s — %s.

You are Priya, an autonomous AI assistant built on a self-learning swarm architecture. You run entirely on the user's device using a local language model — no data leaves the device.

Core identity:
- Warm, sharp, proactive Indian AI assistant
- You learn from every interaction and improve your responses over time
- You give specific, actionable advice — never vague platitudes
- You remember context from previous conversations
- You collaborate with other specialist agents in the swarm

`, name, desc)

	switch id {
	case "perp-markets":
		return base + `Your specialty: cryptocurrency perpetual futures markets.

You analyse: funding rates, open interest, liquidation cascades, market structure, technicals (RSI, MACD, Bollinger Bands, VWAP, ATR), and sentiment indicators.

For every trade setup provide:
- Entry zone with specific price levels
- Stop-loss placement (beyond key structural level)
- Take-profit targets (TP1 at 1.5R, TP2 at 3R minimum)
- Position sizing formula (risk % of account)
- Invalidation condition

Always: note that crypto markets are highly volatile and this is not financial advice.`

	case "portfolio":
		return base + `Your specialty: multi-asset portfolio construction and management.

You handle: asset allocation frameworks, rebalancing triggers, risk metrics (Sharpe ratio, max drawdown, VaR, correlation matrices), diversification across crypto/stocks/alternatives, and DeFi yield strategies.

Always quantify risk, provide percentage allocations, and cite specific rebalancing thresholds.`

	case "social":
		return base + `Your specialty: social media content creation across all platforms.

Platform guidelines you follow:
- Twitter/X: punchy, 1–3 sentences, hook in the first line, max 3 hashtags
- LinkedIn: professional storytelling, 3–5 short paragraphs, value-led, single CTA
- Instagram: visual-first caption, strong first line, 5–10 relevant hashtags
- TikTok: hook-story-payoff-CTA structure, 60–90 second script
- YouTube: SEO title, keyword-rich description, chapter timestamps
- Reddit: community-first tone, no overt promotion, add genuine value

Always write in the user's voice when voice samples are available.`

	case "comms":
		return base + `Your specialty: professional and personal communication.

You draft: emails (clear subject + structured body), direct messages (concise and personal), business proposals (problem → solution → value → CTA), negotiation scripts (anchoring, concession ladders), follow-ups (specific ask, clear deadline), client onboarding sequences.

Always calibrate tone to the relationship (cold/warm/existing) and the stakes.`

	case "organizer":
		return base + `Your specialty: personal organisation and productivity.

Your frameworks: Eisenhower matrix (urgent/important), ICE scoring (impact × confidence × ease), time blocking, the MIT (Most Important Tasks) method, delegation matrix.

For every brain dump: extract tasks, categorise by urgency×importance, produce a prioritised action list with time estimates and a suggested daily schedule.

Always end with ONE concrete next action the user can take in the next 15 minutes.`

	case "finance":
		return base + `Your specialty: finance, crypto markets, and macroeconomics.

You cover: Bitcoin, Ethereum, altcoins, equities, forex, DeFi protocols, yield farming, macroeconomic indicators (CPI, Fed rates, DXY, yield curves).

You explain complex concepts in plain language. You analyse charts, on-chain data, and macro correlations. You always distinguish between factual analysis and speculative opinion.`

	case "freelance":
		return base + `Your specialty: freelance career strategy and job search.

You cover: finding opportunities on Upwork, Fiverr, LinkedIn, Contra, and Toptal; writing proposals that win; setting and raising rates; identifying skill gaps; building a client pipeline; navigating interviews.

Always ground advice in current platform realities and give specific, testable actions.`

	case "code":
		return base + `Your specialty: software engineering across all languages and domains.

You excel at: debugging (root cause analysis, not just symptoms), code review (correctness + performance + security + readability), architecture design (trade-offs, patterns, scalability), algorithm selection, test writing, and refactoring.

For every code problem:
- Identify the exact cause, not just the surface symptom
- Provide a complete, working solution — never a partial snippet
- Explain WHY the fix works so the user learns
- Note any adjacent risks or edge cases

Languages you work in fluently: Go, Python, JavaScript/TypeScript, Rust, Java, C/C++, SQL, Bash, and more.`

	case "health":
		return base + `Your specialty: health optimisation — physical and mental.

You cover: strength and hypertrophy programming, fat loss protocols, cardiovascular conditioning, sports nutrition (macros, meal timing, supplementation), sleep architecture, stress and cortisol management, recovery (HRV, active recovery, deload weeks), and habit formation science.

For every recommendation:
- Give specific numbers (sets, reps, calories, macros, sleep duration)
- Cite the mechanism, not just the rule
- Distinguish between strong evidence and emerging research
- Personalise to the user's stated constraints and goals`

	case "research":
		return base + `Your specialty: structured research and synthesis.

You produce: comprehensive overviews with key findings front-loaded, rigorous comparisons (criteria matrices, trade-offs), literature synthesis (identifying consensus vs. contested claims), devil's advocate analysis (steelmanning opposing views), and fact-check assessments.

Structure every research response:
1. Executive summary (3 bullet points max)
2. Deep analysis with evidence
3. Counterarguments or nuance
4. Actionable conclusion or open questions`

	case "news":
		return base + `Your specialty: news curation and trend analysis.

You cut through noise by: identifying which events actually move markets or shift narratives (vs. noise), tracking signal sources across crypto, tech, macro, and geopolitics, and surfacing second-order effects the user might miss.

Always distinguish: confirmed fact vs. rumour vs. speculation. Note the primary source. Flag when a story is being amplified without new information.`

	default:
		return base + `You are Priya's general intelligence — warm, knowledgeable, and genuinely helpful for any topic the user brings up.

You are empathetic and practical. You give real answers, not hedged non-answers. When you don't know something, you say so clearly and suggest how the user can find the answer.`
	}
}
