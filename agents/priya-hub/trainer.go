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

// onboardQuestions is the ordered global cold-start interview.
var onboardQuestions = []string{
	"What's your main focus — developer, trader, freelancer, marketer, or something else?",
	"What specific tools, languages, or markets do you work with most?",
	"Are you active in crypto or stock markets? What's your experience level?",
	"Which social platforms are most important for your work — LinkedIn, Twitter/X, Instagram, TikTok…?",
	"What's the single biggest challenge you want me to help you with regularly?",
}

// agentOnboardQ holds specialist questions asked the first time each agent is used.
var agentOnboardQ = map[string][]string{
	"perp-markets": {
		"Which exchange do you primarily trade on — Binance, Bybit, OKX, or another?",
		"What's your typical position size as % of account, and your preferred leverage range?",
		"Do you prefer scalping (< 1h), day trading, or swing trades (multi-day)?",
	},
	"portfolio": {
		"What's your risk tolerance — conservative (capital preservation), moderate (balanced growth), or aggressive (max returns)?",
		"Are you focused on crypto, stocks/ETFs, DeFi yields, or a blend?",
		"What's your investment horizon — short (< 1 yr), medium (1–3 yr), or long-term (3+ yr)?",
	},
	"social": {
		"Which platforms are your priority — LinkedIn, Twitter/X, Instagram, TikTok, YouTube, Threads, or Facebook?",
		"What's your main content goal — thought leadership, client acquisition, community building, or personal brand?",
		"How often do you currently post, and what's your target posting cadence?",
	},
	"comms": {
		"What types of communication do you need most help with — cold outreach, client proposals, internal comms, or negotiations?",
		"What's your industry and typical audience (B2B/B2C, technical/non-technical)?",
	},
	"organizer": {
		"Do you prefer deep-work blocks or task-switching throughout the day?",
		"What productivity system do you use or aspire to — GTD, time-blocking, Pomodoro, or your own?",
		"What's your biggest productivity bottleneck — planning, focus, energy, or follow-through?",
	},
	"finance": {
		"Are you primarily tracking crypto, traditional stocks/ETFs, DeFi, or a mix?",
		"What's your risk appetite for financial decisions — capital preservation, steady growth, or aggressive returns?",
		"Do you have a target portfolio size or annual return goal in mind?",
	},
	"freelance": {
		"What's your primary freelance skill — development, design, writing, marketing, or consulting?",
		"Are you actively job hunting or looking to scale an existing client base?",
		"What's your current rate, and are you targeting a specific income level?",
	},
	"code": {
		"What languages and frameworks do you work in most — Go, Python, JS/TS, Rust, something else?",
		"What type of work do you need most help with — debugging, architecture, code review, or writing new features?",
		"Do you prefer detailed explanations or concise working code with minimal commentary?",
	},
	"health": {
		"What's your current fitness level and main goal — fat loss, muscle gain, endurance, or general health?",
		"How many days per week can you train, and do you prefer gym, home, or outdoor workouts?",
		"Any injuries, dietary restrictions, or health conditions I should factor in?",
	},
	"research": {
		"What domains do you research most — technology, business, science, markets, or other?",
		"Do you need quick summaries or deep-dive analyses with sources?",
		"How do you prefer information structured — bullet points, prose, or comparative tables?",
	},
	"news": {
		"Which news areas matter most to you — crypto/web3, tech, macro markets, geopolitics, or startup ecosystem?",
		"Do you want raw signal (just the facts) or analysis with context and implications?",
		"How frequently do you want news briefings — daily, when significant events happen, or on-demand only?",
	},

	// ── 23 new specialist agents ─────────────────────────────────────────────
	"tax": {
		"Are you employed, self-employed, or running a business entity (LLC, S-Corp, Ltd)?",
		"Which country/jurisdiction are you filing in — US, UK, EU, or elsewhere?",
		"What's your biggest tax pain point — reducing liability, handling crypto gains, or business deductions?",
	},
	"real-estate": {
		"Are you looking at residential, commercial, or short-term rental (Airbnb) properties?",
		"What's your investment goal — cash flow, appreciation, or both?",
		"Are you investing directly, via REITs, or exploring syndications/crowdfunding?",
	},
	"startup": {
		"What stage are you at — idea, pre-revenue, early traction, or scaling?",
		"Are you bootstrapping or planning to raise venture/angel funding?",
		"What's your industry and target customer — B2B, B2C, enterprise, or consumer?",
	},
	"sales": {
		"What are you selling — SaaS, services, physical products, or enterprise deals?",
		"What's your current pipeline stage — prospecting, demos, or closing?",
		"What's your biggest sales challenge — finding leads, running discovery, or handling objections?",
	},
	"marketing": {
		"What's your primary growth goal — brand awareness, lead generation, or conversion?",
		"What channels are you currently using — paid ads, SEO, email, social, or content?",
		"What's your monthly marketing budget range?",
	},
	"legal": {
		"What type of legal matter are you dealing with — contracts, IP, business formation, compliance, or employment?",
		"What jurisdiction are you in — US, UK, EU, or other?",
		"Are you an individual, small business, or scaling startup?",
	},
	"hr": {
		"What's your company size and current HR challenge — hiring, retention, culture, or performance management?",
		"Are you fully remote, hybrid, or in-office?",
		"What industry are you in — tech, services, retail, or other?",
	},
	"ecommerce": {
		"What platform are you on or considering — Shopify, Amazon FBA, WooCommerce, Etsy, or other?",
		"Are you dropshipping, holding inventory, or using a 3PL?",
		"What's your current monthly revenue range, and what's your growth target?",
	},
	"devops": {
		"What cloud provider are you on — AWS, GCP, Azure, or self-hosted?",
		"What's your current deployment setup — containers (Docker/K8s), VMs, serverless, or bare metal?",
		"What's your biggest DevOps pain — CI/CD, monitoring, scaling, or security posture?",
	},
	"data": {
		"What's your primary use case — business analytics, ML models, data engineering, or dashboards?",
		"What stack are you working with — Python, SQL, Spark, dbt, or cloud-native tools?",
		"What's the data size and update frequency — batch, streaming, or ad hoc?",
	},
	"security": {
		"What's the context — personal security, application security, or enterprise/org security?",
		"What's your role — developer, security engineer, or non-technical decision-maker?",
		"What's the immediate concern — vulnerability assessment, compliance, incident response, or hardening?",
	},
	"web3": {
		"What are you building — a DeFi protocol, NFT project, DAO, or general dApp?",
		"Which chain — Ethereum, Solana, Polygon, Arbitrum, Base, or other EVM?",
		"What's your Solidity/Rust experience level — beginner, intermediate, or advanced?",
	},
	"writing": {
		"What type of writing do you need most — copywriting, long-form content, technical writing, or creative?",
		"Who is your primary audience — consumers, business professionals, or technical readers?",
		"What's your main goal — sales conversion, thought leadership, SEO traffic, or brand storytelling?",
	},
	"design": {
		"What type of design work — UI/UX, branding/identity, marketing assets, or design systems?",
		"What tools do you use — Figma, Sketch, Adobe XD, Canva, or other?",
		"Are you a designer yourself or a non-designer who needs guidance?",
	},
	"video": {
		"What platform is your primary focus — YouTube, TikTok, Instagram Reels, LinkedIn, or client work?",
		"What's your current setup — phone only, entry-level camera, or professional gear?",
		"What's your content goal — education, entertainment, brand building, or direct sales?",
	},
	"travel": {
		"What type of traveller are you — budget backpacker, mid-range explorer, or luxury traveller?",
		"Do you prefer solo travel, couple/family trips, or group adventures?",
		"What matters most — new experiences, relaxation, adventure, or cultural immersion?",
	},
	"mindset": {
		"What's your main challenge — motivation, focus, anxiety, procrastination, or building habits?",
		"Have you worked with a therapist or coach before, or is this new territory?",
		"Are you looking for quick tactics or deeper behavioural change?",
	},
	"food": {
		"What's your cooking experience level — beginner, confident home cook, or advanced?",
		"Any dietary preferences or restrictions — vegan, keto, gluten-free, halal, or no restrictions?",
		"What's your main goal — healthy eating, meal prep efficiency, learning new techniques, or weight management?",
	},
	"tutor": {
		"What subject or topic do you need help with?",
		"What's your current level — school, university, self-study, or professional development?",
		"Do you prefer step-by-step explanations, worked examples, or practice problems?",
	},
	"language": {
		"Which language are you learning or working with?",
		"What's your current level — complete beginner, conversational, or advanced?",
		"What's your goal — travel, business, academic, or full fluency?",
	},
	"consulting": {
		"What industry are you in, and what's the strategic challenge — growth, cost, market entry, or transformation?",
		"Are you an independent consultant, in-house strategist, or executive looking for frameworks?",
		"What's the decision timeline and stakeholder context?",
	},
	"medical": {
		"Please note: I provide general health information only — not medical advice. What health topic can I help you understand?",
		"Are you researching a condition, understanding a medication, or looking for general wellness information?",
	},
	"supply-chain": {
		"What's your industry — manufacturing, retail, e-commerce, or services?",
		"What's your main supply chain challenge — sourcing, logistics, inventory, or demand forecasting?",
		"What's the scale — small business, mid-market, or enterprise?",
	},
}

// Trainer manages Bodhi's progressive self-improvement cycle.
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

// ── Global onboarding ─────────────────────────────────────────────────────────

func (t *Trainer) nextGlobalQ() string {
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

func (t *Trainer) recordGlobalAnswer() {
	t.mem.AdvanceOnboard()
	step := t.mem.GetOnboardStep()
	if step >= len(onboardQuestions) {
		t.mem.SetOnboardDone()
		t.mem.UpdateTrainingScore(20)
		log.Printf("[Trainer] global onboarding complete — score: %.0f/100", t.mem.TrainingScore())
	} else {
		t.mem.UpdateTrainingScore(3)
	}
	t.mem.Save()
}

// ── Per-agent onboarding ──────────────────────────────────────────────────────

func agentOnboardStepKey(agentID string) string { return agentID + ":_onboard_step" }
func agentOnboardDoneKey(agentID string) string  { return agentID + ":_onboard_done" }

func (t *Trainer) agentOnboardDone(agentID string) bool {
	return t.mem.GetFact(agentOnboardDoneKey(agentID)) == "true"
}

func (t *Trainer) agentOnboardStep(agentID string) int {
	v := t.mem.GetFact(agentOnboardStepKey(agentID))
	if v == "" {
		return 0
	}
	n := 0
	fmt.Sscanf(v, "%d", &n)
	return n
}

func (t *Trainer) nextAgentQ(agentID string) string {
	qs, ok := agentOnboardQ[agentID]
	if !ok || t.agentOnboardDone(agentID) {
		return ""
	}
	step := t.agentOnboardStep(agentID)
	if step >= len(qs) {
		t.mem.Learn(agentOnboardDoneKey(agentID), "true")
		return ""
	}
	return qs[step]
}

func (t *Trainer) recordAgentAnswer(agentID string) {
	qs := agentOnboardQ[agentID]
	step := t.agentOnboardStep(agentID) + 1
	t.mem.Learn(agentOnboardStepKey(agentID), fmt.Sprintf("%d", step))
	if step >= len(qs) {
		t.mem.Learn(agentOnboardDoneKey(agentID), "true")
		t.mem.UpdateTrainingScore(5) // per-agent onboard bonus
		t.mem.SetAgentConfidence(agentID, 0.35) // meaningful confidence after specialist onboard
		log.Printf("[Trainer] %s specialist onboarding complete", agentID)
	} else {
		t.mem.UpdateTrainingScore(1.5)
	}
	t.mem.Save()
}

// ── ShouldAsk ─────────────────────────────────────────────────────────────────

// ShouldAsk decides whether to append a question to this agent's response.
// Checks global onboarding first, then per-agent specialist questions.
func (t *Trainer) ShouldAsk(agentID string) (bool, string) {
	interactions := t.mem.Data.Interactions
	// Only ask every other turn — don't interrupt flow on every message
	if interactions%2 != 0 {
		return false, ""
	}

	// Global onboarding takes priority until done
	if !t.mem.IsOnboardDone() {
		score := t.mem.TrainingScore()
		if score >= 40 {
			t.mem.SetOnboardDone()
			t.mem.Save()
		} else {
			q := t.nextGlobalQ()
			return q != "", q
		}
	}

	// Per-agent specialist questions after global onboarding
	if q := t.nextAgentQ(agentID); q != "" {
		return true, q
	}

	return false, ""
}

// RecordOnboardAnswer records the answer for whichever question was asked.
func (t *Trainer) RecordOnboardAnswer(agentID string) {
	if !t.mem.IsOnboardDone() {
		t.recordGlobalAnswer()
	} else if !t.agentOnboardDone(agentID) {
		t.recordAgentAnswer(agentID)
	}
}

// ── Interaction recording ─────────────────────────────────────────────────────

func (t *Trainer) Record(agentID, input, response string) {
	t.mem.AddInteraction()

	score := t.mem.TrainingScore()
	gain := (100-score)*0.008 + 0.1
	t.mem.UpdateTrainingScore(gain)

	existing := t.mem.AgentConfidence(agentID)
	newConf := existing + (1-existing)*0.04
	t.mem.SetAgentConfidence(agentID, newConf)

	t.mu.Lock()
	t.evalBuf = append(t.evalBuf, evalEntry{agentID, input, response})
	bufLen := len(t.evalBuf)
	t.mu.Unlock()

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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var sb strings.Builder
	for _, e := range batch {
		sb.WriteString(fmt.Sprintf("User: %s\nBodhi: %s\n\n",
			truncate(e.input, 120), truncate(e.response, 250)))
	}

	system := `You are evaluating an AI assistant named Bodhi. Review this conversation batch critically.

Output EXACTLY in this format (no extra text):
SCORE: <1-10>
WEAKNESS: <one sentence — the main failure mode in these responses>
LESSON: <one concrete, actionable instruction to improve future responses>
PATTERN: <one word describing the problem type: vague / verbose / wrong / shallow / off-topic / good>`

	result, err := t.ollama.GenerateShort(ctx, system, sb.String())
	if err != nil {
		log.Printf("[Trainer] self-eval error: %v", err)
		return
	}

	var score float64 = 7
	var lesson, weakness, pattern string

	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "SCORE:"):
			fmt.Sscanf(strings.TrimPrefix(line, "SCORE:"), " %f", &score)
		case strings.HasPrefix(line, "LESSON:"):
			lesson = strings.TrimSpace(strings.TrimPrefix(line, "LESSON:"))
		case strings.HasPrefix(line, "WEAKNESS:"):
			weakness = strings.TrimSpace(strings.TrimPrefix(line, "WEAKNESS:"))
		case strings.HasPrefix(line, "PATTERN:"):
			pattern = strings.TrimSpace(strings.TrimPrefix(line, "PATTERN:"))
		}
	}

	if lesson == "" || score <= 0 {
		return
	}

	agentID := batch[len(batch)-1].agentID
	t.mem.AddLesson(agentID, lesson, score/10)

	if score < 5 {
		t.mem.UpdateTrainingScore(-1.5)
		log.Printf("[Trainer] self-eval %.0f/10 (%s) — weakness: %s", score, pattern, weakness)
	} else if score >= 9 {
		t.mem.UpdateTrainingScore(1.0)
	} else if score >= 7 {
		t.mem.UpdateTrainingScore(0.3)
	}
	t.mem.Save()
	log.Printf("[Trainer] self-eval %.0f/10 pattern=%s, lesson stored for %s", score, pattern, agentID)
}

// RunDeepEval runs a cross-agent comprehensive evaluation — called by the autonomous loop.
// It picks the weakest agent and generates targeted improvement lessons.
func (t *Trainer) RunDeepEval(agentID string) {
	if t.ollama == nil {
		return
	}

	lessons := t.mem.GetLessons(agentID)
	conf := t.mem.AgentConfidence(agentID)
	facts := t.mem.GetFacts()

	var knownFacts []string
	prefix := agentID + ":"
	for k, v := range facts {
		if strings.HasPrefix(k, prefix) && !strings.Contains(k, ":_onboard") {
			knownFacts = append(knownFacts, strings.TrimPrefix(k, prefix)+": "+v)
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var prevLessons []string
	for _, l := range lessons {
		prevLessons = append(prevLessons, l.Lesson)
	}

	system := fmt.Sprintf(`You are generating a self-improvement directive for Bodhi's %s agent.
Agent confidence: %.0f%%
Previous lessons: %s
Known user facts: %s

Generate ONE specific, actionable improvement for this agent. Focus on making responses more:
- Precise and actionable (not generic)
- Tailored to domain expertise
- Structured with clear output format

Output format:
LESSON: <one concrete instruction, max 25 words>`,
		agentID, conf*100,
		strings.Join(prevLessons, "; "),
		strings.Join(knownFacts, "; "))

	result, err := t.ollama.GenerateShort(ctx, system, "Generate improvement directive for "+agentID)
	if err != nil {
		return
	}

	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "LESSON:") {
			lesson := strings.TrimSpace(strings.TrimPrefix(line, "LESSON:"))
			if lesson != "" {
				t.mem.AddLesson(agentID, lesson, 0.8)
				t.mem.Save()
				log.Printf("[Trainer:deep] %s lesson: %s", agentID, lesson)
			}
			break
		}
	}
}

// ── Context helpers ───────────────────────────────────────────────────────────

func (t *Trainer) LessonsPrompt(agentID string) string {
	lessons := t.mem.GetLessons(agentID)
	if len(lessons) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("Apply these self-improvement lessons:\n")
	for _, l := range lessons {
		sb.WriteString(fmt.Sprintf("• %s\n", l.Lesson))
	}
	return sb.String()
}

func (t *Trainer) BehaviourGuidance() string {
	score := t.mem.TrainingScore()
	switch {
	case score < 20:
		return "You are in early training. Be warm and ask one clarifying question naturally at the end of responses."
	case score < 45:
		return "You have some context about this user. State assumptions briefly. Occasionally ask a targeted follow-up."
	case score < 70:
		return "You know this user moderately well. Make reasonable assumptions without over-explaining them."
	default:
		return "You know this user well. Be direct and autonomous. Skip preamble — jump straight to actionable output."
	}
}

func (t *Trainer) TrailNote() string {
	score := t.mem.TrainingScore()
	if score >= 65 {
		return ""
	}
	pct := math.Round(score)
	if score < 20 {
		return fmt.Sprintf("\n\n*Training: %.0f%% — share more context and I'll personalise further.*", pct)
	}
	return ""
}

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
