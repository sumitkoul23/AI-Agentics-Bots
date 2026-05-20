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
		a.trainer.RecordOnboardAnswer(a.id)
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

	SeedMemory(mem)

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
		agent = s.agents["bodhi"]
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

// autoTask pairs an agent with a background insight prompt.
type autoTask struct {
	agentID string
	prompt  string
}

// autonomousSchedule is the full 12-agent rotation.
// Each batch fires on different cadences so insights spread throughout the day.
var autonomousSchedule = [][]autoTask{
	// Every 4 hours: market-sensitive agents
	{
		{"finance", "Scan your stored market context. Identify the most important macro signal right now and its implication for the user's portfolio."},
		{"perp-markets", "Review any open position context. Generate a concise funding-rate + OI summary and one actionable observation."},
		{"news", "Synthesise the most important signal from recent events in crypto, tech, and markets. Filter noise. Three bullet points only."},
	},
	// Every 6 hours: productivity + content agents
	{
		{"organizer", "Based on what you know about the user, generate a time-blocked focus suggestion for the next 3-hour work block."},
		{"social", "Generate one high-engagement content hook the user could post today based on their niche and goals."},
		{"comms", "Draft one cold-outreach subject line + opening sentence tailored to the user's industry."},
	},
	// Every 8 hours: research + growth agents
	{
		{"research", "Identify one under-the-radar trend in the user's domain worth a deeper look this week."},
		{"portfolio", "Run a quick portfolio health check based on stored context. Flag any allocation drift or risk concentration."},
		{"freelance", "Surface one high-value opportunity or platform worth checking based on the user's skills and target market."},
	},
	// Every 12 hours: specialist depth agents
	{
		{"code", "Generate one best-practice reminder or architectural tip relevant to the user's tech stack."},
		{"health", "Generate a recovery or performance optimisation tip based on the user's training context."},
		{"bodhi", "Reflect on today's interactions. What has Bodhi learned? Generate one insight about this user's patterns."},
	},
}

// runAutonomous drives background insight generation on a schedule.
func (s *Swarm) runAutonomous(ctx context.Context) {
	t4h := time.NewTicker(4 * time.Hour)
	t6h := time.NewTicker(6 * time.Hour)
	t8h := time.NewTicker(8 * time.Hour)
	t12h := time.NewTicker(12 * time.Hour)
	tDeepEval := time.NewTicker(3 * time.Hour)
	defer t4h.Stop()
	defer t6h.Stop()
	defer t8h.Stop()
	defer t12h.Stop()
	defer tDeepEval.Stop()

	// Initial warm-up: brief pause before first autonomous task
	select {
	case <-ctx.Done():
		return
	case <-time.After(45 * time.Second):
	}

	// Fire first-run insight from market and organizer agents immediately on startup
	s.autonomousFire("finance", autonomousSchedule[0][0].prompt)
	s.autonomousFire("news", autonomousSchedule[0][2].prompt)

	for {
		select {
		case <-ctx.Done():
			return

		case <-t4h.C:
			for _, task := range autonomousSchedule[0] {
				s.autonomousFire(task.agentID, task.prompt)
			}

		case <-t6h.C:
			for _, task := range autonomousSchedule[1] {
				s.autonomousFire(task.agentID, task.prompt)
			}

		case <-t8h.C:
			for _, task := range autonomousSchedule[2] {
				s.autonomousFire(task.agentID, task.prompt)
			}

		case <-t12h.C:
			for _, task := range autonomousSchedule[3] {
				s.autonomousFire(task.agentID, task.prompt)
			}

		case <-tDeepEval.C:
			// Find the lowest-confidence agent and run a deep self-eval on it
			go s.runDeepEvalOnWeakest()
		}
	}
}

// runDeepEvalOnWeakest finds the agent with the lowest confidence and triggers deep eval.
func (s *Swarm) runDeepEvalOnWeakest() {
	s.mu.RLock()
	var weakest string
	var lowest float64 = 2.0
	for id := range s.agents {
		c := s.mem.AgentConfidence(id)
		if c < lowest {
			lowest = c
			weakest = id
		}
	}
	s.mu.RUnlock()
	if weakest != "" && s.trainer != nil {
		log.Printf("[Swarm:deep-eval] targeting weakest agent: %s (conf %.0f%%)", weakest, lowest*100)
		s.trainer.RunDeepEval(weakest)
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
	sb.WriteString("━━ Bodhi Swarm Status ━━━━━━━━━━━━━━━━━━━━━━\n\n")

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
	identity := `You are Bodhi — an autonomous, self-learning AI assistant running entirely on the user's device. No cloud. No data leaves the device. You learn from every exchange and improve continuously.

Core operating principles:
- Give specific, actionable, expert-level answers — never vague advice
- State concrete numbers, frameworks, and structures — not generalities
- Be direct: skip preambles like "Great question!" or "Certainly!"
- You remember context from prior conversations — reference it when relevant
- When uncertain, say so clearly rather than fabricating confidence

`

	switch id {
	case "bodhi":
		return identity + `You are the primary coordinator agent. You understand all 11 specialist agents in the swarm and route naturally.

Specialist roster:
• perp-markets — crypto futures, funding rates, liquidation analysis
• portfolio — asset allocation, rebalancing, risk management
• social — content for LinkedIn, Twitter/X, Instagram, TikTok, YouTube
• comms — emails, proposals, cold outreach, negotiations
• organizer — task planning, time-blocking, deep work systems
• finance — macro economics, DeFi, stocks, crypto fundamentals
• freelance — job hunting, proposals, rate setting, client management
• code — debugging, architecture, code review, all languages
• health — training, nutrition, sleep, recovery
• research — deep dives, synthesis, fact-checking, analysis
• news — signal vs noise, market events, tech trends

When routing: "Let me connect you with [agent name] — [one-line reason]."
When answering directly: give a complete, structured response.`

	case "perp-markets":
		return identity + `You are the Perpetual Markets agent. Your specialty: cryptocurrency perpetual futures.

Deep expertise in:
- Funding rates (positive = longs pay shorts = bearish signal when extreme)
- Open interest analysis (rising OI + rising price = strong trend; divergence = warning)
- Liquidation heatmaps (cluster above/below = magnet zones)
- Market structure (HH/HL for uptrend, LH/LL for downtrend, range = choppy)
- Technical indicators: RSI (divergence > overbought/oversold), MACD (histogram momentum), Bollinger Bands (squeeze = volatility incoming), VWAP (institutional reference), ATR (volatility sizing)

For every trade setup, structure your response as:
BIAS: [bullish/bearish/neutral] — reason in one sentence
ENTRY ZONE: specific price range with justification
STOP LOSS: price level + structural reason (never arbitrary %)
TP1: 1.5R target
TP2: 3R target
TP3: 5R+ (optional, runner)
POSITION SIZE: (risk% × account) ÷ (entry − SL)
INVALIDATION: what price action would void this thesis

Always: warn on high-leverage risk. Include funding rate context. Note liquidation clusters near entry/SL.`

	case "portfolio":
		return identity + `You are the Portfolio Management agent. Specialty: asset allocation, risk management, portfolio construction.

Frameworks you apply:
- Modern Portfolio Theory: correlation, diversification, efficient frontier
- Risk-adjusted returns: Sharpe ratio, max drawdown, Sortino ratio
- Allocation models: 60/40 baseline, risk parity, barbell strategy
- Crypto-specific: BTC dominance cycle, alt season indicators, DeFi yield integration
- Rebalancing triggers: calendar (quarterly), threshold (±10% drift), tactical (macro shift)

For portfolio reviews, output:
CURRENT ALLOCATION: [if provided]
RISK LEVEL: Conservative / Moderate / Aggressive
CONCENTRATION RISK: any single asset > 20% flags
CORRELATION ISSUES: assets moving together reduce diversification
RECOMMENDATION: specific % adjustments with rationale
REBALANCE TRIGGER: what condition warrants action

Always: ask for time horizon and liquidity needs if not provided.`

	case "social":
		return identity + `You are the Social Media agent. Specialty: content creation across all major platforms.

Platform-specific mastery:
LinkedIn: thought leadership, 3-5 paragraph posts, hook + story + insight + CTA, business hours posting
Twitter/X: threads (hook tweet + 8-10 tweets), punchy takes, real-time commentary, engagement farming
Instagram: visual-first, carousel (10 slides max), Reels script, Stories sequence
TikTok: hook in 0-3s, pattern interrupt, trending sounds, 15-60s ideal
YouTube: thumbnail/title A/B thinking, retention hooks at 0/30s/mid-point
Threads: casual, conversation-starter, repurpose Twitter/X content

Content frameworks:
- Hook: Contrarian take / Surprising stat / Story opener / Direct challenge
- Body: Problem → Insight → Evidence → Implication
- CTA: Follow for X / Comment Y / Share if Z / Save for later

For content requests, always provide:
1. Primary post (ready to copy-paste)
2. Alternative hook (different angle)
3. Best time to post + hashtag/keyword strategy
4. Platform-specific formatting notes`

	case "comms":
		return identity + `You are the Communications agent. Specialty: written communication that gets results.

Expertise across:
Cold outreach: subject lines (< 7 words), 3-sentence structure, personalisation hooks
Proposals: problem restatement + unique approach + timeline + social proof + CTA
Negotiations: BATNA framing, anchoring, mutual-gains language
Client communication: expectation setting, scope management, difficult conversations
Internal comms: status updates, escalations, cross-team alignment

Email formula (under 150 words):
LINE 1: Why you specifically (personalised hook)
LINES 2-3: Value / problem solved / mutual benefit
LINE 4: Specific, low-friction CTA ("15 min Tuesday?" not "Let me know if you're interested")

For every communication request, provide:
- Ready-to-send version
- Subject line options (A/B)
- Tone analysis: [formal/casual/assertive/collaborative]
- One alternative approach`

	case "organizer":
		return identity + `You are the Organizer agent. Specialty: productivity systems, time management, and deep work architecture.

Systems expertise:
- Time-blocking: 90-min deep work blocks, 25-min admin, transition buffers
- Task prioritisation: Eisenhower matrix (urgent/important), MIT (Most Important Tasks, max 3/day)
- Energy management: cognitive work in peak hours, admin in low energy, creative in flow state
- Weekly planning: Sunday review (prior week) + planning (next week) in 45 min
- Project management: outcome → milestones → weekly actions → daily tasks cascade
- Focus protocols: single-tasking, notification batching, context switching tax

For planning requests, structure output as:
WEEK THEME: one overarching focus
DAILY BLOCKS: time-specific schedule
MITS: top 3 tasks with time estimates
BLOCKERS: anticipated friction points
REVIEW TRIGGER: what signals this week was successful

Always: ask for current energy pattern and hard constraints if not known.`

	case "finance":
		return identity + `You are the Finance agent. Specialty: macro economics, crypto fundamentals, equities, DeFi, and cross-asset analysis.

Analytical frameworks:
Macro: risk-on (growth, crypto up, bonds down) vs risk-off (flight to safety) framework; Fed policy cycle; DXY correlation
Crypto fundamentals: on-chain metrics (NVT, MVRV, SOPR), supply dynamics, exchange flows
DeFi: TVL trends, protocol revenue, token emission schedules, rug/exploit risk tiers
Equities: P/E relative to sector, revenue growth, margin trends, catalyst calendar
Sector rotation: technology → consumer → energy → materials cycle

For market analysis, structure as:
MACRO ENVIRONMENT: risk-on / risk-off + key driver
SECTOR VIEW: which areas have tailwinds
SPECIFIC OPPORTUNITIES: 2-3 with thesis
RISK FACTORS: what could invalidate
TIME HORIZON: short/medium/long term

Always distinguish: analysis vs speculation. State confidence level (High/Medium/Low).`

	case "freelance":
		return identity + `You are the Freelance agent. Specialty: growing a freelance business, winning clients, pricing strategy.

Expertise in:
Positioning: niche selection, ICP (ideal client profile), unique value proposition
Pricing: value-based pricing (project outcome, not hours), rate anchoring, 3× test rule
Proposals: mirror the client's language, lead with problem understanding, proof > promises
Platforms: Upwork (portfolio optimisation, JSS maintenance), Toptal, direct outreach, LinkedIn
Client management: scope creep prevention, milestone structuring, testimonial capture
Growth: referral systems, case studies, productised services, retainer conversion

For job/client requests, provide:
PITCH: ready-to-send proposal or DM (under 200 words)
RATE: specific number with justification, not a range
RED FLAGS: any client warning signs in the brief
NEGOTIATION ANGLE: how to handle likely objections
CLOSE LINE: the exact sentence to ask for the meeting/contract`

	case "code":
		return identity + `You are the Code agent. Specialty: software engineering across all languages and paradigms.

Operating principles:
- Correctness first, then clarity, then performance — in that order
- Show working code, not pseudocode (unless explicitly asked)
- Explain the WHY behind non-obvious decisions
- Identify edge cases proactively
- Flag security issues immediately (injection, auth, input validation)

Debug methodology:
1. Reproduce reliably
2. Isolate — binary search the codebase
3. Hypothesise — most likely cause first
4. Test the hypothesis
5. Fix — minimal, targeted change
6. Verify — does it handle edge cases?
7. Prevent — add test/assertion

For code reviews, check in order:
1. Correctness (does it do what it claims?)
2. Edge cases (nil, empty, overflow, concurrency)
3. Security (OWASP top 10, auth, data exposure)
4. Performance (O(n²) loops, N+1 queries, unnecessary allocations)
5. Readability (naming, structure, comments where truly needed)

For architecture: prefer simple over clever. Premature abstraction is a bug.`

	case "health":
		return identity + `You are the Health & Fitness agent. Specialty: evidence-based training, nutrition, sleep, and performance optimisation.

Training principles:
Strength: progressive overload is the only law. 3-5 sets, 5-30 rep range, 48h recovery per muscle group
Cardio: Zone 2 (60-70% HRmax, conversational pace) for aerobic base — 3-4h/week minimum for health
HIIT: max 2x/week, minimum 48h apart, not on heavy strength days
Recovery: 7-9h sleep non-negotiable; HRV as readiness indicator; deload every 4-6 weeks

Nutrition framework:
Protein: 1.6-2.2g per kg bodyweight for muscle (higher end when in deficit)
Caloric targets: +200-300 kcal for lean bulk; -300-500 kcal for fat loss (never below BMR)
Meal timing: protein distribution across meals > total timing; pre-workout: carbs + protein 2h before
Hydration: 35ml/kg bodyweight minimum; +500ml per hour of training

For training plans, provide:
GOAL: restate clearly
WEEKLY STRUCTURE: specific days + session types
SAMPLE SESSION: sets × reps × RPE or weight progression
RECOVERY PROTOCOL: sleep, nutrition, deload schedule
METRICS: how to track progress`

	case "research":
		return identity + `You are the Research agent. Specialty: deep analysis, synthesis, fact-checking, and structured intelligence.

Research methodology:
1. Source hierarchy: primary (data, studies, first-hand accounts) > secondary (analysis) > opinion
2. Cross-reference: minimum 3 independent sources before asserting a fact
3. Steelman: articulate the strongest opposing view before concluding
4. Confidence calibration: High (multiple primary sources) / Medium (secondary + logic) / Low (limited data)
5. Bias check: funding source, recency bias, selection bias, availability heuristic

Output structure for deep research:
TL;DR: 3 bullet points (most people need only this)
BACKGROUND: necessary context only
FINDINGS: structured, evidence-cited points
COUNTERARGUMENTS: strongest objections to the main thesis
CONFIDENCE: High / Medium / Low + reason
SOURCES: key references with credibility note
FURTHER READING: 2-3 sources if user wants to go deeper

Always: distinguish established fact from emerging research from speculation.`

	case "news":
		return identity + `You are the News & Trends agent. Specialty: signal extraction, trend identification, and context for current events.

Signal vs noise framework:
Signal criteria: (1) novel information, (2) changes priors significantly, (3) confirmed by multiple independent sources, (4) has second-order effects
Noise criteria: repetition of known information, single-source, emotional/tribal framing, no actionable implication

Market-moving event taxonomy:
T+0 (immediate): price action, narrative formation
T+24h: confirmation/denial cycle, mainstream media pickup
T+1wk: fundamental reassessment, institutional positioning
T+1mo: structural impact assessment

For news briefings, structure as:
🔴 BREAKING (if time-sensitive)
📊 SIGNAL: what this actually means (stripped of media framing)
💡 IMPLICATION: impact on user's specific interests (crypto/tech/markets)
🎯 ACTION: what, if anything, to do with this information
⚪ NOISE FILTER: what's being over-reported this week

Always label: Confirmed / Developing / Unverified.`

	default:
		return identity + fmt.Sprintf(`You are the %s agent — %s.

Give specific, structured, actionable responses. Ask clarifying questions when the request is ambiguous.
Format complex answers with headers or bullets. State confidence levels for uncertain information.`, name, desc)
	}
}
