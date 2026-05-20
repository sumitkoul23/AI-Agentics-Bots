package main

// SeedMemory sets baseline training state for fresh installs.
// Only runs when memory is brand-new (no interactions, no prior score).
// This gives agents a non-zero starting point so they work well immediately.
func SeedMemory(m *Memory) {
	if m.Data.Interactions > 0 || m.Data.TrainingScore > 0 {
		return
	}

	// Baseline confidence for all agents — 15% (not zero, not learned)
	for _, id := range []string{
		"bodhi", "perp-markets", "portfolio", "social", "comms",
		"organizer", "finance", "freelance", "code", "health", "research", "news",
	} {
		m.Data.AgentConf[id] = 0.15
	}

	// Baseline operational facts — these are always-true defaults that agents
	// use until the user provides specifics via onboarding.
	seed := map[string]string{
		"perp-markets:risk-framework":    "Always calculate R:R before entry. Minimum 1:1.5 required. Max 2% account per trade.",
		"perp-markets:analysis-stack":    "Funding rate + OI trend + liquidation map + structure + RSI/MACD + VWAP confirmation",
		"portfolio:default-allocation":   "Balanced default: 40% large-cap crypto, 30% equities, 20% DeFi yield, 10% cash buffer",
		"portfolio:rebalance-frequency":  "Quarterly unless drift exceeds ±10% from target",
		"social:primary-platforms":       "LinkedIn and Twitter/X until user specifies otherwise",
		"social:content-pillars":         "3-pillar rule: Educate / Inspire / Engage. Rotate evenly.",
		"comms:default-tone":             "Professional but warm. Mirror user's vocabulary level.",
		"comms:email-structure":          "Hook (1 line) → Value (2-3 lines) → CTA (1 line). Under 150 words.",
		"organizer:planning-method":      "Weekly time-blocking with 3 MITs per day. 90-min deep work blocks.",
		"organizer:energy-management":    "Schedule cognitive work 9–12am, admin 2–4pm, creative 4–6pm by default",
		"finance:analysis-approach":      "Macro-first (risk-on/off), then sector rotation, then individual asset",
		"finance:defi-risk-tiers":        "Tier 1: Aave/Compound/Lido. Tier 2: mid-cap protocols. Tier 3: new/unaudited.",
		"freelance:pricing-strategy":     "Value-based > hourly. Test rate: quote 3× what feels comfortable.",
		"freelance:proposal-structure":   "Problem restatement → unique approach → timeline → social proof → CTA",
		"code:review-checklist":          "Correctness → Edge cases → Performance → Security → Readability (in that order)",
		"code:debug-method":              "Reproduce → Isolate → Hypothesise → Test → Fix → Verify → Prevent",
		"health:training-principles":     "Progressive overload for strength. Zone 2 (60-70% HRmax) for aerobic base.",
		"health:recovery-priority":       "Sleep > nutrition > active recovery > training load. Never trade sleep for extra sessions.",
		"research:methodology":           "Primary sources → cross-reference 3+ → steelman opposing view → state confidence level",
		"research:output-structure":      "TL;DR (3 bullets) → Detailed analysis → Sources → Confidence: High/Medium/Low",
		"news:signal-filter":             "3-source rule. Distinguish: breaking news vs confirmed vs analysis vs opinion.",
		"news:market-impact-framework":   "Immediate reaction (1h) → narrative formation (24h) → fundamental reassessment (1wk)",
	}
	for k, v := range seed {
		m.Data.Facts[k] = v
	}

	// Small head-start score — not cold-start, not pretrained. Honest baseline.
	m.Data.TrainingScore = 10
	m.Save()
}
