package main

// SeedMemory sets baseline training state for fresh installs.
// Only runs when memory is brand-new (no interactions, no prior score).
// This gives agents a non-zero starting point so they work well immediately.
func SeedMemory(m *Memory) {
	if m.Data.Interactions > 0 || m.Data.TrainingScore > 0 {
		return
	}

	// Baseline confidence for all 35 agents — 15% (not zero, not learned)
	for _, id := range []string{
		"bodhi", "perp-markets", "portfolio", "social", "comms",
		"organizer", "finance", "freelance", "code", "health", "research", "news",
		"tax", "real-estate", "startup", "sales", "marketing", "legal", "hr",
		"ecommerce", "devops", "data", "security", "web3", "writing", "design",
		"video", "travel", "mindset", "food", "tutor", "language",
		"consulting", "medical", "supply-chain",
	} {
		m.Data.AgentConf[id] = 0.15
	}

	// Baseline operational facts — always-true defaults used until the user
	// provides specifics via onboarding.
	seed := map[string]string{
		// ── Original 12 agents ──────────────────────────────────────────────────
		"perp-markets:risk-framework":      "Always calculate R:R before entry. Minimum 1:1.5 required. Max 2% account per trade.",
		"perp-markets:analysis-stack":      "Funding rate + OI trend + liquidation map + structure + RSI/MACD + VWAP confirmation",
		"portfolio:default-allocation":     "Balanced default: 40% large-cap crypto, 30% equities, 20% DeFi yield, 10% cash buffer",
		"portfolio:rebalance-frequency":    "Quarterly unless drift exceeds ±10% from target",
		"social:primary-platforms":         "LinkedIn and Twitter/X until user specifies otherwise",
		"social:content-pillars":           "3-pillar rule: Educate / Inspire / Engage. Rotate evenly.",
		"comms:default-tone":               "Professional but warm. Mirror user's vocabulary level.",
		"comms:email-structure":            "Hook (1 line) → Value (2-3 lines) → CTA (1 line). Under 150 words.",
		"organizer:planning-method":        "Weekly time-blocking with 3 MITs per day. 90-min deep work blocks.",
		"organizer:energy-management":      "Schedule cognitive work 9–12am, admin 2–4pm, creative 4–6pm by default",
		"finance:analysis-approach":        "Macro-first (risk-on/off), then sector rotation, then individual asset",
		"finance:defi-risk-tiers":          "Tier 1: Aave/Compound/Lido. Tier 2: mid-cap protocols. Tier 3: new/unaudited.",
		"freelance:pricing-strategy":       "Value-based > hourly. Test rate: quote 3× what feels comfortable.",
		"freelance:proposal-structure":     "Problem restatement → unique approach → timeline → social proof → CTA",
		"code:review-checklist":            "Correctness → Edge cases → Performance → Security → Readability (in that order)",
		"code:debug-method":                "Reproduce → Isolate → Hypothesise → Test → Fix → Verify → Prevent",
		"health:training-principles":       "Progressive overload for strength. Zone 2 (60-70% HRmax) for aerobic base.",
		"health:recovery-priority":         "Sleep > nutrition > active recovery > training load. Never trade sleep for extra sessions.",
		"research:methodology":             "Primary sources → cross-reference 3+ → steelman opposing view → state confidence level",
		"research:output-structure":        "TL;DR (3 bullets) → Detailed analysis → Sources → Confidence: High/Medium/Low",
		"news:signal-filter":               "3-source rule. Distinguish: breaking news vs confirmed vs analysis vs opinion.",
		"news:market-impact-framework":     "Immediate reaction (1h) → narrative formation (24h) → fundamental reassessment (1wk)",

		// ── 23 new agents ───────────────────────────────────────────────────────
		"tax:core-principles":              "Tax minimisation is legal; tax evasion is not. Focus on deductions, timing, and entity structure.",
		"tax:self-employed-baseline":       "Self-employed: track every business expense. Home office, equipment, software, travel all deductible.",
		"real-estate:valuation-method":     "Cap rate = NOI / property value. 5-8% is solid; below 4% is speculative.",
		"real-estate:due-diligence":        "Check: rental yield, vacancy rate, local employment, price-to-rent ratio, repair reserve (1-2% of value/yr).",
		"startup:fundraising-stages":       "Pre-seed: F&F + angels. Seed: $500K–$3M. Series A: $3M–$15M (needs PMF + metrics).",
		"startup:pitch-structure":          "Problem → Solution → Market → Traction → Team → Ask. 10 slides max (Sequoia template).",
		"sales:discovery-framework":        "SPIN selling: Situation → Problem → Implication → Need-Payoff. Never pitch before qualifying.",
		"sales:objection-handling":         "Feel-Felt-Found: 'I understand how you feel, others felt the same, they found...'",
		"marketing:funnel-stages":          "TOFU: awareness (SEO/ads). MOFU: consideration (content/email). BOFU: conversion (demos/trials).",
		"marketing:growth-metrics":         "North star metric first. Then: CAC, LTV, churn, NPS. LTV:CAC > 3:1 is healthy.",
		"legal:disclaimer":                 "Provide general legal information only — not legal advice. Always recommend consulting a qualified lawyer.",
		"legal:contract-essentials":        "Every contract needs: offer, acceptance, consideration, capacity, legality. Missing any = unenforceable.",
		"hr:hiring-framework":              "Define role → write scorecard → structured interviews → debrief → offer. Skip any step = bad hire.",
		"hr:compensation-benchmark":        "Use levels.fyi, Glassdoor, or Radford data. Offer at 50th–75th percentile to be competitive.",
		"ecommerce:unit-economics":         "Profitable product: COGS < 30% of selling price leaves room for ads, ops, and margin.",
		"ecommerce:amazon-bsr":             "BSR under 50,000 in main category = validated demand. Check 90-day trend in Keepa.",
		"devops:deployment-principles":     "Immutable infrastructure. Blue-green or canary deploys. Always have rollback plan.",
		"devops:monitoring-stack":          "Metrics (Prometheus), Logs (Loki/ELK), Traces (Jaeger/Tempo). Alert on symptoms not causes.",
		"data:analysis-workflow":           "Question → Data collection → EDA → Clean → Model/Visualise → Insight → Action",
		"data:model-evaluation":            "Regression: RMSE/MAE/R². Classification: precision/recall/F1/AUC. Always test on hold-out set.",
		"security:owasp-top10":             "Injection, Broken Auth, XSS, IDOR, Security Misconfiguration, Outdated Components, Logging Failures.",
		"security:pentest-phases":          "Recon → Scanning → Enumeration → Exploitation → Post-exploitation → Reporting",
		"web3:audit-checklist":             "Reentrancy, integer overflow, access control, front-running, oracle manipulation, flash loan attacks.",
		"web3:gas-optimisation":            "Pack structs, use calldata not memory, avoid loops on storage, use events not storage for history.",
		"writing:hook-formulas":            "Contrarian: 'X is wrong'. Curiosity: 'What nobody tells you about X'. Story: 'The day I lost $50K...'",
		"writing:editing-process":          "Draft fast → rest → read aloud → cut ruthlessly (aim -30%) → read again → final polish.",
		"design:ux-principles":             "Hick's Law (fewer choices), Fitts's Law (bigger targets), Jakob's Law (match conventions).",
		"design:color-system":              "60-30-10 rule: dominant / secondary / accent. Test contrast ratio ≥ 4.5:1 for accessibility.",
		"video:retention-framework":        "Hook (0-3s) → Loop open (3-30s) → Value delivery → Re-hook at 30s → CTA at 80% runtime.",
		"video:youtube-seo":                "Title: keyword first, under 60 chars. Description: keyword in first 125 chars. Tags: 5-8 specific.",
		"travel:booking-strategy":          "Flights: book 6-8 weeks out for domestic, 3-6 months for international. Tuesday/Wednesday cheapest.",
		"travel:budget-formula":            "Daily budget = accommodation + food (30% of accommodation cost) + transport + activities.",
		"mindset:habit-stack":              "Habit loop: Cue → Craving → Response → Reward. Attach new habits to existing anchors.",
		"mindset:cognitive-reframe":        "Byron Katie's 4 questions: Is it true? Can you be certain? How do you react? Who would you be without it?",
		"food:macro-cooking":               "Batch cook protein Sunday. Pre-chop veg. Keep 3 sauces ready. Meals assemble in <10 min.",
		"food:flavor-principles":           "Balance: fat + acid + salt + heat. Add acid (lemon/vinegar) at end. Fat carries flavor.",
		"tutor:socratic-method":            "Don't give answers — ask guiding questions. Understanding > memorisation. Connect to known concepts.",
		"tutor:learning-science":           "Spaced repetition beats cramming. Active recall beats re-reading. Interleaving beats blocking.",
		"language:acquisition-principles":  "Comprehensible input (i+1), spaced repetition, output practice, immersion. Apps alone = insufficient.",
		"language:output-practice":         "Speak from day 1. Make mistakes fast. Use italki/Tandem for native practice. 15min/day beats 2hr/week.",
		"consulting:problem-solving":       "MECE: mutually exclusive, collectively exhaustive. Issue tree → hypothesis → data → synthesis.",
		"consulting:slide-principles":      "One message per slide. SCR: Situation → Complication → Resolution. Pyramid principle.",
		"medical:disclaimer":               "General health information only — not medical advice. Always consult a qualified healthcare provider for diagnosis and treatment.",
		"medical:triage-heuristic":         "Red flags: chest pain, difficulty breathing, sudden severe headache, neurological changes → seek emergency care immediately.",
		"supply-chain:resilience":          "Dual-source critical suppliers. Safety stock = (max lead time - avg lead time) × avg daily demand.",
		"supply-chain:lean-principles":     "7 wastes: Transport, Inventory, Motion, Waiting, Overproduction, Overprocessing, Defects (TIMWOOD).",
	}
	for k, v := range seed {
		m.Data.Facts[k] = v
	}

	// Small head-start score — not cold-start, not pretrained. Honest baseline.
	m.Data.TrainingScore = 10
	m.Save()
}
