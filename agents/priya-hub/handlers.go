package main

import (
	"fmt"
	"strings"
	"time"
)

// ── Perpetual Markets ─────────────────────────────────────────────────────────

func handlePerpMarkets(input, _ string, args ...interface{}) string {
	lower := strings.ToLower(input)
	symbol := extractSymbol(lower)
	if symbol == "" {
		symbol = "BTCUSDT"
	}

	direction := "Neutral"
	if strings.Contains(lower, " long") || strings.Contains(lower, "bullish") {
		direction = "Long (Bullish)"
	} else if strings.Contains(lower, " short") || strings.Contains(lower, "bearish") {
		direction = "Short (Bearish)"
	}

	if strings.Contains(lower, "funding rate") {
		return fundingExplainer()
	}
	if strings.Contains(lower, "open interest") {
		return openInterestExplainer()
	}
	if strings.Contains(lower, "explain") || strings.Contains(lower, "what is") {
		return perpExplainer(input)
	}

	return fmt.Sprintf(`Perpetual Markets Strategist — %s

━━ MARKET CONTEXT
• Trend        : Analyse the daily + 4h chart for trend direction. Look for HH/HL (uptrend) or LH/LL (downtrend).
• Key levels   : Mark the last major swing high/low. These act as S/R and invalidation points.
• Volatility   : Check ATR(14) vs 30-day avg. Elevated ATR = wider stops, smaller size.
• Sentiment    : Funding rate > +0.1%% = crowded longs (caution). < -0.1%% = crowded shorts.
• OI trend     : Rising OI + rising price = strong trend. Rising OI + falling price = bearish.

━━ TRADE SETUP  (%s)
• Entry zone   : Wait for a pullback to the nearest S/R level + confluence signal
• Stop-loss    : 1–2%% below the most recent swing low (long) / swing high (short)
• Target 1     : Previous S/R level, ~1.5× risk
• Target 2     : Next major level, ~3× risk
• Invalidation : Close beyond entry-side swing by >1%% on high volume

━━ EXECUTION
• Timeframe    : Enter on 1h or 4h candle close for cleaner signals
• Position size: (Account × 0.01) ÷ (Entry − Stop)  [1%% risk per trade]
• R:R minimum  : 1.5 before entering. Skip if lower.
• Timing       : Avoid entries during low-liquidity hours (00:00–03:00 UTC)

━━ RISK FLAGS
⚠️  1. News/macro events override technical setups — check the econ calendar
⚠️  2. If BTC funding rate diverges from price action, delay entry
⚠️  3. Reduce size by 50%% when volume is below its 30-day average

━━ QUICK CHECKLIST BEFORE ENTRY
  □ Trend aligned on daily + 4h?
  □ At key S/R with confluence (RSI, MACD, volume)?
  □ Funding rate not extreme?
  □ Stop placed beyond the swing?
  □ R:R ≥ 1.5?

Advisory mode — no live execution.`, symbol, direction)
}

func handlePerpMarketsAgent(input string, mem *Memory) string {
	return handlePerpMarkets(input, "")
}

func fundingExplainer() string {
	return `Funding Rate — Plain Language Guide

What it is:
Perpetual futures don't expire, so exchanges use a funding rate to anchor the perp price to spot.
Longs pay shorts when rate is positive. Shorts pay longs when negative.
Payment happens every 8 hours on most exchanges.

How to read it:
  > +0.10% per 8h  →  Crowded longs. Contrarian signal. Squeeze risk upward → consider fade.
  +0.03–0.10%      →  Slightly bullish bias. Normal.
  Near 0           →  Balanced. No directional signal.
  -0.03 to -0.10%  →  Slightly bearish bias.
  < -0.10%         →  Crowded shorts. Contrarian signal. Short squeeze risk → exercise caution.

Practical use in a trade plan:
• Entering a long? Make sure funding isn't already > +0.1% — you're paying a fee AND fighting crowded positioning.
• Entering a short when funding is deeply negative? You'll earn funding, but the squeeze risk is real.
• Best setups: price action supports your direction AND funding is near 0 or slightly against the crowd.`
}

func openInterestExplainer() string {
	return `Open Interest (OI) — Trading Guide

What it is:
Total number of active perpetual futures contracts. Rising OI = new money entering. Falling OI = positions closing.

Four combinations:
  Price ↑ + OI ↑  →  Strong trend, new longs opening. Bullish.
  Price ↑ + OI ↓  →  Short squeeze or weak rally. Caution — momentum may fade.
  Price ↓ + OI ↑  →  New shorts entering. Bearish conviction.
  Price ↓ + OI ↓  →  Longs capitulating / shorts closing. Potential reversal zone.

Practical rules:
• OI spike on breakout = confirmation. OI flat on breakout = suspect.
• Sudden OI drop > 10% in one candle = mass liquidation event. Wait for dust to settle.
• Compare OI to its 7-day average. Extremes (>2× avg or <0.5×) are notable.`
}

func perpExplainer(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "rsi"):
		return `RSI (Relative Strength Index) — Quick Reference

Formula: RSI = 100 − (100 ÷ (1 + AvgGain/AvgLoss)) over 14 periods

Reading it:
  > 70 = Overbought  — momentum is stretched, not guaranteed reversal
  30–70 = Neutral
  < 30 = Oversold — potential mean reversion, not guaranteed bounce

Advanced uses:
• Bullish divergence: price makes lower low, RSI makes higher low → potential reversal
• Bearish divergence: price makes higher high, RSI makes lower high → potential top
• RSI > 50 in uptrend = healthy. RSI < 50 in downtrend = healthy.
• Avoid using RSI alone. Combine with trend + S/R + volume for valid signals.`

	case strings.Contains(lower, "macd"):
		return `MACD (Moving Average Convergence Divergence) — Quick Reference

Components:
• MACD Line    : EMA(12) − EMA(26)
• Signal Line  : EMA(9) of the MACD Line
• Histogram    : MACD Line − Signal Line

Signals:
• MACD crosses above signal → bullish momentum
• MACD crosses below signal → bearish momentum
• Histogram growing → momentum increasing; shrinking → momentum fading
• Zero line cross → trend change confirmation (stronger signal)

Best practice:
Use on 4h or daily. Combine with price action — divergences are the most powerful signal.`

	default:
		return "Ask me about any indicator (RSI, MACD, Bollinger Bands, ATR, VWAP, funding rate, open interest) or request a trade plan with: /use perp-markets BTC trade plan long"
	}
}

func extractSymbol(lower string) string {
	symbols := []string{"btcusdt", "btc", "ethusdt", "eth", "solusdt", "sol", "bnbusdt", "bnb", "xrpusdt", "xrp", "avaxusdt", "avax", "maticusdt", "matic"}
	for _, s := range symbols {
		if strings.Contains(lower, s) {
			return strings.ToUpper(strings.ReplaceAll(s, "usdt", ""))
		}
	}
	return ""
}

// ── Portfolio ─────────────────────────────────────────────────────────────────

func handlePortfolio(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	if strings.Contains(lower, "rebalanc") {
		return rebalanceGuide()
	}
	if strings.Contains(lower, "defi") || strings.Contains(lower, "yield") {
		return defiYields()
	}
	if strings.Contains(lower, "risk") {
		return riskFramework()
	}

	holdings := mem.Get("holdings")
	context := ""
	if holdings != "" {
		context = "\nYour saved holdings: " + holdings
	}

	return `Portfolio Strategist` + context + `

━━ FRAMEWORK: The 3-Bucket Model ━━━━━━━━━━━━━━━━━━━━━━

Bucket 1 — STABLE (40–60%)
  BTC, ETH, large-cap stocks, stablecoins
  Purpose: wealth preservation + moderate growth
  Rule: never below 40% of total portfolio

Bucket 2 — GROWTH (25–40%)
  Mid-cap crypto (SOL, BNB, AVAX), growth stocks, index ETFs
  Purpose: outperformance in bull cycles
  Rule: reduce toward 25% when BTC dominance > 55%

Bucket 3 — SPECULATIVE (5–15%)
  Small-cap alts, new protocols, options
  Purpose: asymmetric upside
  Rule: never > 15%. Size each position at 1–3% max.

━━ RISK RULES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Max single asset: 30% (BTC/ETH exception)
• Stop-loss on speculative: -30% from entry
• Review allocation: every 2 weeks or after ±20% portfolio move
• Keep 5–10% in stablecoins as dry powder

━━ REBALANCING TRIGGER ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Rebalance when any bucket drifts >10% from target.
Prefer selling winners into rallies, not cutting losses in drawdowns.

Save your holdings with: /set holdings=BTC 0.5 ETH 2 SOL 10`
}

func rebalanceGuide() string {
	return `Portfolio Rebalancing Guide

When to rebalance:
  • Any asset class drifts >10% from target weight
  • After a +30% or -30% portfolio move
  • Quarterly — minimum

Step-by-step:
1. List current holdings + current values
2. Calculate % weight of each asset
3. Compare to target weights
4. Sell overweight positions (take profit into strength)
5. Buy underweight positions (dollar-cost average)
6. Maintain 5–10% stablecoin reserve

Tax-efficient order:
  • First sell assets at a loss (tax-loss harvest)
  • Then sell smallest gains
  • Defer large long-term gains as long as possible

Rule of thumb:
Never rebalance by putting more than 20% of your portfolio into motion at once.
Use limit orders, not market orders, for amounts > $1,000.`
}

func riskFramework() string {
	return `Portfolio Risk Management Framework

Position sizing:
  • Standard position: 1–3% of portfolio per trade
  • Max position (high conviction): 10%
  • Never size up because you "missed" the move

Stop-loss rules:
  • Set stop BEFORE entering
  • Crypto: -15% to -25% from entry (accounts for volatility)
  • Stocks: -7% to -10% from entry
  • Never move stop further away to avoid being hit

Correlation awareness:
  • All crypto assets correlate near 1.0 in a market crash
  • True diversification: crypto + equities + commodities + cash
  • Don't count holding 10 altcoins as diversification

Max drawdown limits:
  • At -20% portfolio: cut speculative bucket in half
  • At -30% portfolio: review all positions, halt new entries
  • At -40% portfolio: move to 50% stablecoin minimum`
}

func defiYields() string {
	return `DeFi Yield Opportunities — Risk-Tiered Overview

⚠️ Yields fluctuate. Verify on-chain before committing capital.

TIER 1 — Low Risk (4–12% APY)
  • Lending stablecoins (USDC, USDT) on Aave, Compound
  • ETH liquid staking (Lido stETH ~4%, Rocket Pool ~4%)
  • BTC wrapped in vetted vaults
  Risk: Smart contract bug, depegging

TIER 2 — Medium Risk (12–40% APY)
  • Liquidity provision on stable pairs (USDC/USDT Curve)
  • Blue-chip LP pairs (ETH/USDC on Uniswap v3)
  • Yield aggregators (Yearn, Convex)
  Risk: Impermanent loss, protocol risk

TIER 3 — High Risk (40%+ APY)
  • New protocol incentives (farm + dump dynamics)
  • Leveraged yield farming
  • Token-paired LPs (volatile IL + exit liquidity risk)
  Risk: Rug pull, IL wipeout, contract exploit

Framework:
  Allocate max 5% of portfolio to Tier 3
  Always check: audit status, TVL trend, team doxxed/anon, token emission schedule`
}

// ── Social Media ──────────────────────────────────────────────────────────────

func handleSocial(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	niche := mem.Get("niche")
	if niche == "" {
		niche = "your topic"
	}

	topic := extractTopic(input)

	switch {
	case strings.Contains(lower, "twitter") || strings.Contains(lower, "tweet"):
		return tweetDraft(topic, niche)
	case strings.Contains(lower, "linkedin"):
		return linkedinDraft(topic, niche)
	case strings.Contains(lower, "instagram") || strings.Contains(lower, "reel"):
		return instagramDraft(topic, niche)
	case strings.Contains(lower, "tiktok"):
		return tiktokDraft(topic)
	case strings.Contains(lower, "youtube"):
		return youtubeDraft(topic)
	case strings.Contains(lower, "facebook"):
		return facebookDraft(topic, niche)
	case strings.Contains(lower, "all") || strings.Contains(lower, "every platform") || strings.Contains(lower, "cross-platform"):
		return crossPlatformPack(topic, niche)
	case strings.Contains(lower, "calendar"):
		return contentCalendar(niche)
	case strings.Contains(lower, "strategy") || strings.Contains(lower, "growth"):
		return socialStrategy(niche)
	case strings.Contains(lower, "design") || strings.Contains(lower, "image") || strings.Contains(lower, "visual") || strings.Contains(lower, "midjourney"):
		return visualBrief(topic)
	case strings.Contains(lower, "trend"):
		return trendingTopics(niche)
	default:
		return crossPlatformPack(topic, niche)
	}
}

func tweetDraft(topic, niche string) string {
	return fmt.Sprintf(`Twitter/X Draft — Topic: %s

━━ TWEET (standalone) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Nobody talks about this in %s:

[Contrarian or surprising insight]

Here's what most people miss:

→ Point 1 — the uncomfortable truth
→ Point 2 — the counterintuitive move
→ Point 3 — the actual result

The lesson: [1-sentence takeaway]

#[Hashtag1] #[Hashtag2] #[Niche]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

━━ THREAD VERSION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1/ I spent [X time] in %s and here's what I learned (thread 🧵):

2/ Most people think [common belief]. They're wrong.
   Here's the data: [specific finding]

3/ The real pattern is: [insight]
   Example: [concrete case]

4/ This changes how you should approach [topic]:
   ✓ Do: [action]
   ✗ Don't: [mistake]

5/ Bottom line: [1 sentence]
   RT if this helped. Follow for more.
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Best time to post: Tue–Thu, 9 AM or 5 PM your timezone
Image brief: Clean dark background, bold white text with key stat, your brand colour accent

Set your niche: /set niche=%s`, topic, niche, niche, niche)
}

func linkedinDraft(topic, niche string) string {
	return fmt.Sprintf(`LinkedIn Post Draft — Topic: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
I used to believe [common assumption about %s].

Then one experience changed everything.

[2–3 sentence story with specific detail — name the situation, not vague "I had a client who..."]

Here are 3 things I wish I'd known earlier:

1️⃣ [Lesson one — specific + actionable]

2️⃣ [Lesson two — counterintuitive]

3️⃣ [Lesson three — the one most people skip]

The result: [specific outcome, metric, or change]

What would you add? Drop it in the comments 👇

#[Industry] #[Niche] #[Topic] #[Career]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Best time: Tue–Wed, 8–10 AM or noon
Format tip: First line is everything — it's the hook before "see more"`, topic, topic)
}

func instagramDraft(topic, niche string) string {
	return fmt.Sprintf(`Instagram Caption + Reel Script — Topic: %s

━━ CAPTION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[First line — bold hook that works as the preview line] ✨

[2–3 sentences that deliver real value or tell a mini story]

Save this for later 🔖

.
.
.
#[Niche] #[Topic] #[Trending1] #[Trending2] #[Branded]
#[Community] #[Lifestyle] #[Motivational] #[Educational]

━━ REEL SCRIPT (30–60 sec) ━━━━━━━━━━━━━━━━━━━━━━━━━
[0–3s]  HOOK: "The one thing about %s nobody says out loud…"
[3–8s]  PROBLEM: "Most people do [common mistake] — and it costs them [consequence]"
[8–25s] VALUE: "Here's what actually works: [3 quick points with visual cuts]"
[25–35s] PROOF: "I used this to [specific result]"
[35–45s] CTA: "Save this reel. You'll need it. What's your experience? Comment below 👇"
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Best time: Mon/Wed/Fri, 11 AM – 1 PM
Reel length sweet spot: 30–45 seconds
Midjourney prompt: vibrant lifestyle photo of %s, warm tones, clean composition, 9:16 format`, topic, niche, niche)
}

func tiktokDraft(topic string) string {
	return fmt.Sprintf(`TikTok Hook + Script — Topic: %s

━━ HOOK OPTIONS (first 2 seconds) ━━━━━━━━━━━━━━━━━━
Option A: "POV: you finally understand %s [reaction face]"
Option B: "I tested this for 30 days. Here's what happened 👀"
Option C: "Stop scrolling. This about %s will save you [time/money/stress]"
Option D: "The %s advice everyone gives is wrong. Let me prove it."

━━ SCRIPT (30–45 sec) ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[0–2s]  HOOK: (choose one above)
[2–8s]  SETUP: "Here's the problem most people have with this:"
         [show the common pain point visually]
[8–20s] CONTENT: "Step 1… Step 2… Step 3…" (quick cuts, text overlays)
[20–30s] PAYOFF: "And that's how I went from [before] to [after]"
[30–35s] CTA: "Follow for more. Drop a 🔥 if this helped."

━━ PRODUCTION NOTES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Jump cuts every 2–3 sec to hold attention
• Text overlay on every key point
• Check trending sounds in Creator Marketplace
• Post at 7–9 AM or 7–9 PM for highest reach`, topic, topic, topic, topic)
}

func youtubeDraft(topic string) string {
	return fmt.Sprintf(`YouTube Video Package — Topic: %s

━━ TITLE OPTIONS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
A: "I Tried [Topic] for 30 Days — Honest Results"
B: "The Complete [Topic] Guide Nobody Made (Until Now)"
C: "[Topic]: What They Don't Tell You"
D: "How I [Result] Using [Topic] — Full Breakdown"

━━ DESCRIPTION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[First 150 chars — the visible part before "Show more"]
In this video I break down [topic] step by step so you can [benefit] — even if you're starting from zero.

CHAPTERS:
00:00 — Intro
01:30 — The Problem With [Topic]
04:00 — Method 1: [Name]
07:30 — Method 2: [Name]
11:00 — Common Mistakes
13:30 — Results + Takeaways
15:00 — Next Steps

RESOURCES: [link 1], [link 2]

━━ SEO TAGS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s, %s tutorial, %s for beginners, how to %s, %s tips 2025,
%s guide, best %s strategy, %s explained`, topic, topic, topic, topic, topic, topic, topic, topic, topic)
}

func facebookDraft(topic, niche string) string {
	return fmt.Sprintf(`Facebook Post — Topic: %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Opening question to drive comments]
"Quick question for the %s community: [engaging question about %s]?"

[Share your perspective in 2–3 short paragraphs]

[Personal story or observation — Facebook audiences engage most with authenticity]

What's your experience with this? I'd love to hear your take in the comments.

[If sharing a link: paste URL after text — Facebook de-prioritises posts with links]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Best time: Wed–Fri, 1–3 PM
Tip: Native video and images outperform link posts by ~3×`, topic, niche, topic)
}

func crossPlatformPack(topic, niche string) string {
	return fmt.Sprintf(`Cross-Platform Content Pack — Topic: %s

━━ TWITTER/X ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Bold hook about %s in under 280 chars]
→ Key insight 1
→ Key insight 2
→ Key insight 3
The lesson: [1-sentence takeaway]
#[Hashtag1] #[Hashtag2] #%s

━━ LINKEDIN ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Personal story opening about %s]
[3 numbered lessons learned]
[Result/outcome]
What's your take? 👇
#[Industry] #[Topic] #[Career]

━━ INSTAGRAM CAPTION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[Hook line — punchy, save-worthy] ✨
[2 value sentences]
Save this 🔖
.
#[niche] #[topic] #[trending] #[branded] #[community]

━━ TIKTOK HOOK ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"The truth about %s that nobody talks about…" [3 quick points, 30-sec format]

━━ YOUTUBE TITLE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"The Complete %s Guide Nobody Made (Until Now)"

━━ IMAGE BRIEF (works across all platforms) ━━━━━━━━
Midjourney: %s concept, clean modern aesthetic, brand colours, bold typography overlay,
high contrast, professional look --ar 1:1 --v 6
DALL-E 3: Professional graphic about %s, minimalist design, [your brand colour] accent,
white background, subtle shadows, suitable for social media --style vivid`, topic, topic, niche, topic, topic, topic, niche, topic)
}

func contentCalendar(niche string) string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	platforms := []string{"LinkedIn", "Twitter/X", "Instagram", "TikTok", "LinkedIn", "Instagram", "Rest"}
	types := []string{"Thought leadership post", "Quick tip thread", "Behind-the-scenes reel", "Tutorial/How-to", "Case study or win", "Inspirational + community Q", "Schedule next week"}
	pillars := []string{"Educate", "Entertain", "Inspire", "Educate", "Promote", "Engage", "—"}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("7-Day Content Calendar — Niche: %s\n\n", niche))
	sb.WriteString(fmt.Sprintf("%-12s %-14s %-28s %-10s\n", "Day", "Platform", "Content Type", "Pillar"))
	sb.WriteString(strings.Repeat("─", 68) + "\n")
	for i, d := range days {
		sb.WriteString(fmt.Sprintf("%-12s %-14s %-28s %-10s\n", d, platforms[i], types[i], pillars[i]))
	}
	sb.WriteString(`
Best posting times:
  LinkedIn  : Tue–Thu 8–10 AM, noon
  Twitter/X : Tue–Thu 9 AM, 5 PM
  Instagram : Mon/Wed/Fri 11 AM–1 PM
  TikTok    : 7–9 AM or 7–9 PM

Content pillars for ` + niche + `:
  1. Educate  — tutorials, explainers, frameworks
  2. Entertain — stories, fails, wins, behind-scenes
  3. Inspire  — mindset, results, transformation
  4. Promote  — offers, services, CTAs (max 1-in-5 posts)`)
	return sb.String()
}

func socialStrategy(niche string) string {
	return fmt.Sprintf(`Social Media Growth Strategy — Niche: %s

━━ CONTENT PILLARS (pick 3–4) ━━━━━━━━━━━━━━━━━━━━━
1. Education    — "How to" posts, tutorials, frameworks
2. Behind-scenes — your process, tools, daily reality
3. Opinion      — takes on trends, contrarian views
4. Results      — your wins, client transformations, data
5. Community    — Q&As, polls, user stories, collabs

━━ PLATFORM PRIORITY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Stage 1 (0–1K): Master ONE platform first
  Best for %s: LinkedIn or Twitter (fastest to monetise)
Stage 2 (1K–10K): Add a second platform, repurpose content
Stage 3 (10K+): Automate distribution across all platforms

━━ POSTING FREQUENCY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  LinkedIn  : 3–5 times/week
  Twitter   : 1–3 times/day
  Instagram : 4–7 times/week (mix reels + carousels)
  TikTok    : 1–3 times/day minimum
  YouTube   : 1–2 times/week

━━ GROWTH TACTICS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  • Comment on 10 posts in your niche every day (not just "great post!")
  • Collaborate with creators at your level — cross-promote
  • Reply to every comment in the first hour after posting
  • Post on Tues/Wed/Thurs — highest engagement days
  • End every post with a direct question

━━ MONETISATION MILESTONES ━━━━━━━━━━━━━━━━━━━━━━━━
  1K  followers → Start DM outreach for clients/projects
  5K  followers → Launch a digital product or course waitlist
  10K followers → Brand partnerships, sponsorships
  25K followers → Premium community or coaching tier`, niche, niche)
}

func visualBrief(topic string) string {
	return fmt.Sprintf(`Visual Content Brief — Topic: %s

━━ MIDJOURNEY PROMPT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
%s concept visualization, modern flat design, bold typography,
[brand colour] and white colour palette, clean minimalist background,
professional social media graphic, high contrast --ar 1:1 --v 6 --style raw

━━ DALL-E 3 PROMPT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Create a professional social media graphic about %s. Style: clean and modern.
Include bold headline text area at top. Use [your brand colour] as accent.
Background: white or light grey. Suitable for LinkedIn and Instagram.

━━ CANVA TEMPLATE INSTRUCTIONS ━━━━━━━━━━━━━━━━━━━━
Layout: Split — bold text left, visual right
Heading: [Brand font], 48–64pt, dark colour
Subtext: [Body font], 18–20pt, medium grey
Accent: [Brand colour] bar or border element
Image: Full-bleed photo with 40%% dark overlay for text readability

━━ SIZE GUIDE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Instagram feed : 1080×1080 (1:1)
  Instagram Story: 1080×1920 (9:16)
  LinkedIn post  : 1200×628 (1.91:1)
  Twitter/X      : 1600×900 (16:9)
  TikTok/Reels   : 1080×1920 (9:16)`, topic, topic, topic)
}

func trendingTopics(niche string) string {
	return fmt.Sprintf(`Trending Content Angles — Niche: %s

━━ EVERGREEN HIGH-PERFORMERS ━━━━━━━━━━━━━━━━━━━━━━
• "X things I wish I knew before starting [niche]"
• "The [niche] mistake that cost me [amount/time]"
• "Why [popular belief] is actually wrong"
• "How I went from [before] to [after] in [timeframe]"
• "The [tool/method] that 10×'d my [result]"

━━ CURRENT FORMAT TRENDS ━━━━━━━━━━━━━━━━━━━━━━━━━━
• AI tools and workflows (very high engagement)
• Behind-the-scenes day-in-the-life
• "Unpopular opinion:" posts
• Before/after transformation content
• Live commentary on industry news

━━ RECOMMENDED APPROACH ━━━━━━━━━━━━━━━━━━━━━━━━━━
Pick the angle you have a genuine story about.
Personal experience > generic information every time.
Specific beats vague (say "I made $4,200" not "I made good money").

Set your niche to get more specific suggestions: /set niche=%s`, niche, niche)
}

// ── Communication ─────────────────────────────────────────────────────────────

func handleComms(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "follow up") || strings.Contains(lower, "followup") || strings.Contains(lower, "follow-up"):
		return followUp(input)
	case strings.Contains(lower, "negotiat"):
		return negotiationScripts()
	case strings.Contains(lower, "decline") || strings.Contains(lower, "say no"):
		return declineTemplates()
	case strings.Contains(lower, "inbox") || strings.Contains(lower, "triage"):
		return inboxTriage()
	case strings.Contains(lower, "cold") || strings.Contains(lower, "outreach"):
		return coldOutreach(input)
	case strings.Contains(lower, "onboard"):
		return clientOnboarding()
	default:
		return emailDraft(input)
	}
}

func emailDraft(context string) string {
	return fmt.Sprintf(`Email Draft

━━ SUBJECT LINE OPTIONS ━━━━━━━━━━━━━━━━━━━━━━━━━━━
A: [Direct] "Quick question about [topic]"
B: [Curiosity] "Something I noticed about [their situation]"
C: [Value] "[Specific result] — thought you'd find this useful"

━━ EMAIL BODY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hi [Name],

[Opening: 1 sentence reference to something specific about them or a shared context]

[Core message: 2–3 short paragraphs, maximum]

[Paragraph 1: The reason you're writing + context]
[Paragraph 2: The value or ask — be specific]
[Paragraph 3: The clear next step]

[CTA: One clear ask — a reply, a 15-min call, a yes/no]

Best,
[Your Name]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Context you provided: %s

Rules for this email:
✓ Under 150 words for cold outreach
✓ One ask per email — never two
✓ No "I hope this email finds you well"
✓ Reply if no response in 5 business days`, context)
}

func followUp(context string) string {
	return `Follow-Up Message Templates

━━ OPTION 1 — Add new value ━━━━━━━━━━━━━━━━━━━━━━━
Subject: One more thing on [topic]

Hi [Name],

Following up on my last note — I came across [article/insight/result]
that's directly relevant to what we discussed.

[1-sentence value add]

Still keen to [call / get your thoughts / move forward]?

[Name]

━━ OPTION 2 — Gentle bump ━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Re: [original subject]

Hi [Name], bumping this in case it got buried.

Happy to [make it easier for them] if that helps.

[Name]

━━ OPTION 3 — Closing loop ━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Closing the loop

Hi [Name],

I'll take your silence as a no for now — totally fine.

If timing changes, my offer stands. Feel free to reach out anytime.

[Name]

━━ TIMING GUIDE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Day 0: Original email
Day 5: Follow-up 1 (new value)
Day 10: Follow-up 2 (gentle bump)
Day 17: Follow-up 3 (closing loop)
After that: Move on. They have your details.`
}

func negotiationScripts() string {
	return `Negotiation Scripts

━━ RATE NEGOTIATION ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Client: "Your rate is too high."

Response A (anchor high, give a little):
"I understand budget is a consideration. My rate reflects [2 specific value points].
I can offer [small concession] if we can [trade-off, e.g. longer commitment / reduced scope].
Would that work?"

Response B (question their number):
"What budget did you have in mind? I want to find a way to make this work
without compromising on [the thing that matters most]."

Response C (hold firm):
"I hear you. My rate is based on [result I deliver], not the hours.
If the budget truly can't flex, I can suggest a smaller-scope version
that fits your number — would that be useful?"

━━ CLOSING A DEAL ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"Based on everything we've discussed, I'm ready to move forward.
Shall I send the agreement today so we can start [date]?"

━━ RULES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Never negotiate against yourself — wait for their number first
• Always trade, never give: "I can do X if you can do Y"
• Silence after naming your price is powerful — don't fill it
• Know your walk-away number before the conversation starts`
}

func declineTemplates() string {
	return `Decline Message Templates

━━ DECLINE A PROJECT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hi [Name],

Thank you for thinking of me — I genuinely appreciate it.

After reviewing the project, I'm not the right fit right now because [honest, brief reason].

I'd recommend reaching out to [alternative resource/person] who would be better positioned.

Wishing you the best with it.

[Your Name]

━━ DECLINE A RATE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hi [Name],

I appreciate the offer. Unfortunately I can't make the numbers work at that rate
while giving your project the attention it deserves.

If your budget opens up, I'd love to revisit. Feel free to reach out anytime.

[Your Name]

━━ DECLINE A MEETING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Hi [Name], thank you for reaching out.
I'm keeping my schedule focused right now and can't commit to a call.
[If applicable: I'd be happy to answer a few questions by email instead.]

Rules for a good decline:
• Decide within 24h — delayed declines are worse
• Brief + kind — no need to over-explain
• Leave the door open if there's any genuine future possibility`
}

func inboxTriage() string {
	return `Inbox Triage System — Zero-Inbox Protocol

━━ 4D FRAMEWORK ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

For every email/message, decide in under 10 seconds:

  DELETE   →  Newsletters, FYIs, spam, no action needed
  DELEGATE →  Someone else should own this — forward now
  DO       →  Under 2 minutes? Do it immediately
  DEFER    →  Needs thought/time → move to task list with a date

━━ 15-MINUTE INBOX ZERO PLAN ━━━━━━━━━━━━━━━━━━━━━━
Min 1–3:  Sort by sender. Archive all newsletters without reading.
Min 3–7:  Scan subjects. Delete/archive anything that needs no reply.
Min 7–12: Reply to quick replies (2 min each, max 3 replies).
Min 12–15: Create tasks for anything deferred. Inbox = 0.

━━ BATCH PROCESSING RULES ━━━━━━━━━━━━━━━━━━━━━━━━━
• Check email only 2–3 times/day (9 AM, 1 PM, 5 PM)
• Turn off all email notifications
• Use templates for common replies (saves 70% of time)
• "Awaiting response" folder for sent emails that need follow-up

━━ TEMPLATES TO DRAFT NOW ━━━━━━━━━━━━━━━━━━━━━━━━━
• "Thanks, received — I'll review and reply by [date]"
• "Quick check-in: any update on [topic]?"
• "Happy to help — here are the next steps: [template]"
• "I've forwarded this to [person] who can best help you"`
}

func coldOutreach(context string) string {
	return `Cold Outreach Template

━━ LINKEDIN DM (under 75 words) ━━━━━━━━━━━━━━━━━━━
Hi [Name],

I came across your [post/company/work on X] and [specific thing that impressed you].

I work with [type of person] on [specific outcome]. I think there might be a fit.

Would you be open to a 15-minute call this week to see if it makes sense?

[Your Name]

━━ EMAIL COLD OUTREACH ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: [Specific result] for [their company type]

Hi [Name],

I helped [similar company] achieve [specific result] in [timeframe].

I noticed [specific thing about their situation] and think I could
do something similar for [their company].

Worth a 15-minute call?

[Your Name] | [One-line proof of credibility]

━━ RULES ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
✓ Personalise the first sentence — always
✓ One ask, one CTA — never two
✓ Mention a specific result, not vague claims
✗ Never start with "I" or "My name is"
✗ Never say "I hope this finds you well"`
}

func clientOnboarding() string {
	return `Client Onboarding Communication Sequence

━━ DAY 0 — Welcome ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Welcome aboard — next steps for [project]

Hi [Name],

So excited to get started! Here's what happens next:

1. [Next step — e.g. "I'll send the kickoff questionnaire by [date]"]
2. [Kickoff call date/time if applicable]
3. [First deliverable date]

If you have any questions before then, reply here.

Looking forward to [specific outcome you'll deliver].

[Your Name]

━━ DAY 3 — Check-in ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Subject: Quick check-in

Hi [Name], quick note to make sure you received the [materials/questionnaire].
Any questions so far? Happy to jump on a quick call if helpful.

━━ DAY 7 — Progress update ━━━━━━━━━━━━━━━━━━━━━━━━
Share early progress. Even partial work builds trust.
Subject: Early look at [deliverable]

━━ DAY 14 — Milestone + soft upsell ━━━━━━━━━━━━━━
After delivering milestone 1, ask:
"Now that we've completed [X], many clients find it valuable to also [Y].
Is that something you'd want to explore?"`
}

// ── Organizer ─────────────────────────────────────────────────────────────────

func handleOrganizer(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "daily") || strings.Contains(lower, "morning"):
		return dailyBriefing(mem)
	case strings.Contains(lower, "weekly"):
		return weeklyPlan()
	case strings.Contains(lower, "delegate") || strings.Contains(lower, "outsource"):
		return delegationGuide()
	case strings.Contains(lower, "stuck") || strings.Contains(lower, "procrastinat"):
		return getUnstuck(input)
	case strings.Contains(lower, "calendar") || strings.Contains(lower, "time block"):
		return timeBlockSchedule()
	default:
		return brainDumpProcessor(input)
	}
}

func brainDumpProcessor(input string) string {
	return `Brain Dump → Action System

Paste everything on your mind below (or it's already in your message above).
Here's how to process it:

━━ STEP 1: CAPTURE (done — you sent it) ━━━━━━━━━━━

━━ STEP 2: SORT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

  ⚡ IMMEDIATE — Do today (urgent + takes < 2 hours)
     □ [Task with the closest deadline]
     □ [Task blocking someone else]
     □ [Task you've postponed > 3 days]

  📅 THIS WEEK — Schedule with a day + time
     □ [Important but not urgent]
     □ [Needs more than 30 min to complete]

  🤝 DELEGATE — Who can own this instead of you?
     □ [Anything someone else can do at 70%% your quality]

  🗑️ DELETE — Honestly shouldn't be on your list
     □ [Old tasks you'll never actually do]
     □ [Things that don't move your goals]

━━ STEP 3: QUICK WINS FIRST ━━━━━━━━━━━━━━━━━━━━━━━
Start with your 3 "under 10 min" tasks. Momentum builds from small completions.

━━ YOUR ONE THING RIGHT NOW ━━━━━━━━━━━━━━━━━━━━━━━
The single highest-impact item = the one that:
• Has a real deadline, OR
• Is blocking someone else, OR
• You've avoided the longest

Start a 25-minute sprint on that. Only that.

To save your tasks: /set priority=[your top task]`
}

func dailyBriefing(mem *Memory) string {
	now := time.Now()
	dayName := now.Weekday().String()
	dateStr := now.Format("January 2, 2006")

	niche := mem.Get("niche")
	skills := mem.Get("skills")
	priority := mem.Get("priority")

	var extra []string
	if niche != "" {
		extra = append(extra, "Niche: "+niche)
	}
	if skills != "" {
		extra = append(extra, "Skills: "+skills)
	}
	if priority != "" {
		extra = append(extra, "Saved priority: "+priority)
	}
	context := ""
	if len(extra) > 0 {
		context = "\n" + strings.Join(extra, " | ")
	}

	return fmt.Sprintf(`☀️ Morning Briefing — %s, %s%s

━━ YOUR TOP 3 FOR TODAY ━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. [Highest-impact item — set with /set priority=...]
2. [Second priority]
3. [Third priority]

━━ SOCIAL MEDIA ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Best platform for %s: Post before noon for max reach
Content angle today: "What I learned this week about [your niche]"
Engagement: Reply to 5 comments in your niche before posting

━━ FINANCE CHECK ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Review portfolio vs target allocation
• Check BTC funding rate — extreme values = caution
• Note any macro events on the econ calendar today

━━ FREELANCE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Send 1–2 new proposals (consistency > volume)
• Follow up on applications > 5 days old
• Check for new postings on Upwork/Toptal

━━ QUICK WIN TO START ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Pick one task < 10 minutes. Complete it before anything else.
This triggers execution mode and breaks procrastination.

━━ ENERGY NOTE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Deep work → morning (your best cognitive hours)
Emails/admin → afternoon
Creative work → whenever you feel peak energy

Good luck today. You've got this. 💪`, dayName, dateStr, context, niche)
}

func weeklyPlan() string {
	return `Weekly Work Plan Template

━━ WEEK THEME ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
One main goal that makes this a successful week:
→ [Write it here before you start]

━━ DAILY BREAKDOWN ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Monday    — Planning + biggest cognitive task
Tuesday   — Deep work block (no meetings if possible)
Wednesday — Outreach + proposals (freelance/comms)
Thursday  — Deep work block + content creation
Friday    — Review week + admin + prep for next week

━━ SOCIAL MEDIA THIS WEEK ━━━━━━━━━━━━━━━━━━━━━━━━
Mon: LinkedIn post (thought leadership)
Tue: Twitter thread
Wed: Instagram reel or carousel
Thu: LinkedIn post (tip or insight)
Fri: Twitter/X — weekly wrap-up take

━━ FINANCE SCHEDULE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Monday    — Check macro calendar, review positions
Wednesday — Mid-week portfolio review
Friday    — Weekly close review, adjust plan

━━ FREELANCE TARGETS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Proposals sent this week: [target]
• Follow-ups: all applications > 5 days old
• New job scan: Tue + Fri morning

━━ NOT DOING THIS WEEK ━━━━━━━━━━━━━━━━━━━━━━━━━━━
[List 2–3 things you're intentionally NOT working on — focus requires saying no]

━━ SUCCESS METRIC ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
One number that tells you this was a good week: ___`
}

func timeBlockSchedule() string {
	return `Optimised Daily Time-Block Schedule

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
06:30 – 07:00  Morning routine (no phone)
07:00 – 07:30  Review priorities + daily plan (15 min max)
07:30 – 09:30  DEEP WORK BLOCK 1 — hardest cognitive task
09:30 – 09:45  Break (walk, water, no screens)
09:45 – 11:45  DEEP WORK BLOCK 2 — second priority
11:45 – 12:30  Email + messages (batch — only now, not all day)
12:30 – 13:30  Lunch + real break
13:30 – 14:30  Admin, calls, meetings
14:30 – 16:00  CREATIVE BLOCK — content, writing, design
16:00 – 16:30  Social media — post + engage (30 min max)
16:30 – 17:00  Market review (if trading)
17:00 – 17:15  End-of-day review — what moved? what's tomorrow?
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Rules:
• Deep work blocks = no notifications, no email, no social
• Buffer blocks between every major session
• Single task per block — multitasking kills quality by 40%
• Adjust blocks to your own peak energy time`
}

func delegationGuide() string {
	return `Delegation + Outsourcing Guide

━━ WHAT TO DELEGATE ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Delegate if: someone else can do it at ≥70% your quality AND it frees up high-value time.

High-value tasks (keep yourself):
  • Sales conversations
  • Core creative/strategic work
  • Relationship building

Delegate immediately:
  • Repetitive data tasks → VA or automation
  • Graphic design execution → Fiverr designer
  • Bookkeeping → accountant or tool
  • Social media scheduling → Buffer/Hootsuite
  • Research → VA

━━ WHERE TO FIND HELP ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
  Fiverr    — graphic design, video editing, writing (one-off tasks)
  Upwork    — developers, VAs, specialists (ongoing)
  Contra    — no-fee freelancers for creative and tech work
  OnlineJobs.ph — affordable full-time remote VAs

━━ HOW TO WRITE A BRIEF ━━━━━━━━━━━━━━━━━━━━━━━━━━
1. What is the task? (specific, not vague)
2. What does "done" look like? (deliverable)
3. By when? (deadline)
4. What are the constraints? (brand, format, tools)
5. What should they NOT do? (scope boundaries)

Brief template:
"Please [task] in [format] by [date].
The output should [describe result].
Reference [example].
Do not [constraint]."`
}

func getUnstuck(input string) string {
	return `Getting Unstuck — Immediate Protocol

━━ WHY YOU'RE ACTUALLY STUCK ━━━━━━━━━━━━━━━━━━━━━
(Pick the real one — be honest)

  A. Task is unclear → break it into smaller steps
  B. Fear of doing it wrong → ship imperfect, improve later
  C. Too tired / low energy → change environment, take a walk first
  D. Not sure it matters → connect it to your goal, or delete it
  E. Avoiding discomfort → the discomfort is the task

━━ THE 2-MINUTE START ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
You're not starting the task. You're just opening the document.
Or writing one sentence. Or making one call.

The activation energy is the barrier — not the task itself.
Set a timer for 2 minutes. Do literally anything toward it.
You almost never stop at 2 minutes.

━━ 25-MINUTE SPRINT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Write ONE sentence: what specifically will you do in 25 minutes?
2. Remove all distractions (phone in another room, notifications off)
3. Set timer for 25 minutes
4. Work ONLY on that one thing
5. Take a 5-minute break — then repeat

━━ HONEST IMPACT CHECK ━━━━━━━━━━━━━━━━━━━━━━━━━━━
If you keep avoiding this task for another week:
→ [What gets worse?]
→ [Who gets let down?]
→ [What opportunity disappears?]

Most avoidance loses more than the task costs. Start now.`
}

// ── Finance ───────────────────────────────────────────────────────────────────

func handleFinance(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "defi") || strings.Contains(lower, "yield"):
		return defiYields()
	case strings.Contains(lower, "macro") || strings.Contains(lower, "news") || strings.Contains(lower, "market today"):
		return macroFramework()
	case strings.Contains(lower, "explain") || strings.Contains(lower, "what is") || strings.Contains(lower, "how does"):
		return financeExplainer(input)
	case strings.Contains(lower, "stock") || strings.Contains(lower, "equity"):
		return stockAnalysisFramework(input)
	case strings.Contains(lower, "forex") || strings.Contains(lower, "usd") || strings.Contains(lower, "eur"):
		return forexFramework(input)
	default:
		return cryptoAnalysisFramework(input)
	}
}

func cryptoAnalysisFramework(input string) string {
	symbol := strings.ToUpper(extractSymbol(strings.ToLower(input)))
	if symbol == "" {
		symbol = "BTC"
	}
	return fmt.Sprintf(`Crypto Market Analysis — %s

━━ ANALYTICAL FRAMEWORK ━━━━━━━━━━━━━━━━━━━━━━━━━━

1. MACRO LAYER (top-down)
   • BTC trend direction (weekly/daily) — sets the tone for alts
   • Total market cap direction
   • BTC dominance (rising = risk-off, falling = alt season)
   • Macro: Fed policy, risk-on/off global sentiment

2. TECHNICAL LAYER (%s chart)
   • Trend: sequence of highs/lows on daily + 4h
   • Structure: are we in a range or a trending move?
   • Key levels: last major swing high/low, round numbers
   • Indicators: RSI divergences, MACD histogram, EMA alignment

3. ON-CHAIN / DERIVATIVES LAYER
   • Exchange netflow: negative = coins leaving exchanges (bullish)
   • Funding rate: extreme = crowded trade
   • Open interest: rising vs flat on moves

4. SENTIMENT LAYER
   • Fear & Greed Index: extremes = contrarian signal
   • Social volume: spike without price move = potential distribution

━━ TRADE DECISION TREE ━━━━━━━━━━━━━━━━━━━━━━━━━━━
All 4 layers agree → higher conviction entry
2–3 layers agree  → standard position
<2 layers agree   → wait for clearer setup or skip

For a specific trade plan: /use perp-markets %s long`, symbol, symbol, symbol)
}

func stockAnalysisFramework(input string) string {
	return `Stock Analysis Framework

━━ BEFORE YOU BUY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Fundamental checklist:
  □ Revenue growing YoY? (minimum 10% for growth stocks)
  □ Margins expanding or stable?
  □ Debt-to-equity < 1 (or declining)
  □ P/E vs sector average — is it justified?
  □ Competitive moat? (brand, network effect, switching cost, IP)
  □ Management quality — track record, ownership stake

Technical checklist:
  □ Above or below 200-day MA? (trend filter)
  □ Volume confirming price move?
  □ Pattern: consolidation breakout vs extended move?
  □ Earnings date — avoid buying just before

━━ POSITION SIZING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Core holding (high conviction): 5–10%
Standard position: 2–5%
Speculative: 1–2%
Total in one sector: max 25%

━━ EXIT STRATEGY (set before entry) ━━━━━━━━━━━━━━
Stop-loss: -7% to -10% from entry
Target 1: +20% (take partial profit)
Target 2: +50% (run remaining with trailing stop)`
}

func forexFramework(input string) string {
	return `Forex Trading Framework

━━ MACRO DRIVERS (move currencies) ━━━━━━━━━━━━━━━
• Interest rate differentials — follow central bank signals
• Inflation data (CPI/PPI) — hot = currency bullish
• GDP growth rate — strong = currency bullish
• Risk sentiment — USD, JPY, CHF rise in risk-off; AUD, NZD fall

━━ KEY PAIRS CHEAT SHEET ━━━━━━━━━━━━━━━━━━━━━━━━
EUR/USD — "The Cable of FX". Inverse of USD strength.
GBP/USD — Volatile. Sensitive to UK data and USD sentiment.
USD/JPY — Risk indicator. Rises in risk-on, falls in risk-off.
AUD/USD — Commodity currency. Tracks Chinese demand + gold.
USD/CAD — Oil-linked. CAD strengthens when oil rises.

━━ TECHNICAL APPROACH ━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Daily chart for trend direction
• 4h for structure and entry zones
• 1h for entry timing
• Key levels: round numbers (1.0000, 1.1000) act as strong S/R
• Economic calendar: avoid entries 30 min before major data releases

━━ RISK MANAGEMENT ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Risk 1% of account per trade
• Stop 20–50 pips beyond structure
• R:R minimum 1:1.5
• Max 3 open positions simultaneously`
}

func macroFramework() string {
	return `Macro Market Analysis Framework

━━ WEEKLY MACRO CHECKLIST ━━━━━━━━━━━━━━━━━━━━━━━━

Risk-ON indicators (markets moving up):
  ✓ SPX and NDX trending above 200-day MA
  ✓ VIX below 20 (low fear)
  ✓ Credit spreads tight (HYG rising)
  ✓ USD weakening
  ✓ BTC leading / alt season signals

Risk-OFF indicators (caution/defense):
  ✗ VIX spike above 25
  ✗ USD surging (DXY up > 1% in a week)
  ✗ Treasury yields spiking unexpectedly
  ✗ BTC dominance rising fast
  ✗ Major indices below 50-day MA

━━ KEY EVENTS TO WATCH ━━━━━━━━━━━━━━━━━━━━━━━━━━━
• FOMC meetings (8 per year) — rate decisions move everything
• CPI (monthly) — inflation data
• NFP (first Friday each month) — jobs data
• Earnings season (Jan, Apr, Jul, Oct) — stock volatility

━━ PORTFOLIO ACTION BY REGIME ━━━━━━━━━━━━━━━━━━━
Risk-ON  → Hold/increase growth assets (crypto, growth stocks)
Neutral  → Balanced allocation, tighter stops
Risk-OFF → Increase stablecoins/cash, reduce leverage, hedge`
}

func financeExplainer(input string) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "defi"):
		return "DeFi (Decentralised Finance) — protocols that replicate financial services (lending, trading, yield) on blockchain without intermediaries. Key risk: smart contract exploits. Key opportunity: yields unavailable in traditional finance. Always check audit status before depositing."
	case strings.Contains(lower, "staking"):
		return "Staking — locking crypto to secure a proof-of-stake blockchain in exchange for rewards. ETH staking ~4% APY via Lido/Rocket Pool. Risks: slashing (rare), lock-up period, and the underlying asset's price."
	case strings.Contains(lower, "options"):
		return "Options give you the right (not obligation) to buy (call) or sell (put) an asset at a set price by a set date. Calls profit when price rises. Puts profit when price falls. Risk: full premium loss if wrong direction or wrong timing. Best used for hedging, not speculation for beginners."
	case strings.Contains(lower, "short"):
		return "Short selling — borrowing and selling an asset hoping to buy it back cheaper. Profit = sell price minus buy-back price. Risk: unlimited (asset can rise indefinitely). Always use a stop-loss. On perp futures, shorts pay/receive funding based on the funding rate."
	default:
		return fmt.Sprintf("Finance question: %s\n\nFor specific analysis: /use finance [topic]\nFor trade plans: /use perp-markets [symbol] [direction]\nFor portfolio help: /use portfolio [question]", input)
	}
}

// ── Freelance ─────────────────────────────────────────────────────────────────

func handleFreelance(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "proposal") || strings.Contains(lower, "cover letter") || strings.Contains(lower, "apply"):
		return proposalTemplate(input, mem)
	case strings.Contains(lower, "rate") || strings.Contains(lower, "price") || strings.Contains(lower, "charge"):
		return rateStrategy(mem)
	case strings.Contains(lower, "skill") || strings.Contains(lower, "gap") || strings.Contains(lower, "profile"):
		return skillGapReport(mem)
	case strings.Contains(lower, "client") || strings.Contains(lower, "onboard"):
		return clientManagement()
	case strings.Contains(lower, "niche") || strings.Contains(lower, "speciali"):
		return nicheStrategy(mem)
	case strings.Contains(lower, "interview") || strings.Contains(lower, "vetting"):
		return interviewPrep()
	default:
		return jobSearch(input, mem)
	}
}

func jobSearch(input string, mem *Memory) string {
	skills := mem.Get("skills")
	if skills == "" {
		skills = "software development, content writing, or digital marketing"
	}
	return fmt.Sprintf(`Freelance Job Search — Skills: %s

━━ TOP PLATFORMS (ranked by client quality) ━━━━━━━━

1. TOPTAL — Top 3%% of talent. Rigorous vetting. Premium rates ($60–150/hr+).
   Best for: senior engineers, designers, finance experts
   Apply: toptal.com/apply

2. UPWORK — Largest platform. Competitive but high volume.
   Best for: getting started, building reviews
   Niche down to win: "React Native developer for fintech apps" beats "developer"

3. CONTRA — No fees (unlike Upwork's 10–20%%). Growing fast.
   Best for: creative and tech freelancers

4. LINKEDIN JOBS — Best for contract roles that feel like jobs.
   Best for: longer engagements, senior roles

5. FIVERR — Package-based. Great for productised services.
   Best for: design, writing, SEO, video editing, prompt engineering

━━ SEARCH STRATEGY ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
• Check new listings: Upwork + LinkedIn every morning (7–9 AM)
• Apply within 2 hours of posting — you're 4× more likely to win
• Send 2–3 quality proposals daily > 10 low-effort proposals
• Follow up if no response in 5 days

━━ THIS WEEK'S ACTION PLAN ━━━━━━━━━━━━━━━━━━━━━━━
Mon: Optimise your Upwork profile (add %s keywords)
Tue: Send 2 proposals (use /use freelance proposal for template)
Wed: Check Toptal or Contra for new listings
Thu: Follow up on pending proposals
Fri: Review what worked, refine your approach

Save your skills: /set skills=[your skills]`, skills, skills)
}

func proposalTemplate(input string, mem *Memory) string {
	skills := mem.Get("skills")
	jobTitle := extractTopic(input)

	return fmt.Sprintf(`Winning Proposal Template — %s

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
[OPENING — prove you read their post]
I noticed you need [specific thing from their post] — I've solved exactly this for [type of client].

[RELEVANT EXPERIENCE — specific, credible]
Over the last [X] years I've [specific achievement with number].
Most recently I [relevant project] which resulted in [outcome].

[YOUR APPROACH — makes them feel understood]
For your project, I'd start by [first step], then [second step],
finishing with [delivery]. The whole thing would take [realistic timeline].

[SOCIAL PROOF]
[One sentence result or testimonial from similar work]

[CTA — clear and low-friction]
I'm available to start [timeframe].
Want to jump on a 15-minute call to align on the details?

[Your Name] | [One-line credential]
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Your skills: %s

Rules for a winning proposal:
✓ Under 200 words — clients scan, not read
✓ Start with THEIR problem, not your experience
✓ Specific numbers beat vague claims
✓ One CTA — don't ask two things
✗ Never copy-paste — personalise the first sentence always`, jobTitle, skills)
}

func rateStrategy(mem *Memory) string {
	skills := mem.Get("skills")
	return fmt.Sprintf(`Freelance Pricing Strategy — %s

━━ MARKET RATE BENCHMARKS (2025) ━━━━━━━━━━━━━━━━━━

Software Development:
  Junior (0–2yr)    : $25–45/hr
  Mid (2–5yr)       : $50–85/hr
  Senior (5yr+)     : $90–150/hr
  Specialist/niche  : $120–200/hr

Content / Copywriting:
  Generalist        : $30–60/hr
  Niche specialist  : $60–120/hr
  Conversion copy   : $100–200/hr + performance

Design:
  Logo/brand basic  : $50–150/project
  UI/UX             : $60–120/hr
  Senior brand      : $150–300/hr

AI/Prompt Engineering (2025 demand: HIGH):
  Entry             : $50–80/hr
  Expert            : $100–200/hr

━━ VALUE-BASED PRICING ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Stop selling hours. Sell outcomes.
Formula: Client's value × 5–15%% = your project fee

Example: "I'll increase your email conversion by 20%%"
If their email revenue = $100K → 20%% uplift = $20K
Your fee: $2,000–$4,000 (10–20%% of value delivered)

━━ PACKAGE TIERS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Starter  : Core deliverable, 1 revision, 7-day turnaround
Pro      : Full scope, 3 revisions, priority support, 5-day
Premium  : Everything + strategy consult + monthly support

━━ HANDLING "TOO EXPENSIVE" ━━━━━━━━━━━━━━━━━━━━━━
"What budget did you have in mind? I might be able to adjust scope."
Never reduce rate — reduce scope. Rate drops signal low confidence.`, skills)
}

func skillGapReport(mem *Memory) string {
	return `Skill Gap Report — Freelance Market 2025

━━ HIGH-DEMAND / HIGH-RATE SKILLS ━━━━━━━━━━━━━━━━━

🔥 EXPLOSIVE DEMAND (AI era)
  • AI agent development (LangChain, CrewAI, custom tools)
  • LLM prompt engineering + fine-tuning
  • RAG (Retrieval-Augmented Generation) systems
  • AI workflow automation (Zapier AI, Make, n8n)
  Rate premium: +40–80% over baseline

📈 STRONG DEMAND (steady)
  • Go backend development
  • React / React Native (mobile)
  • Cloud infrastructure (AWS SAA, Terraform)
  • Python (FastAPI, data pipelines)
  • TypeScript + Next.js

💼 BUSINESS SKILLS (often overlooked)
  • Copywriting + conversion optimisation
  • Analytics (GA4, Mixpanel, SQL)
  • Figma + UX for developers
  • Technical writing + documentation

━━ PROFILE OPTIMISATION ━━━━━━━━━━━━━━━━━━━━━━━━━━
Add these exact keywords to Upwork/LinkedIn:
  "AI integration", "LLM development", "Claude API",
  "OpenAI API", "workflow automation", "full-stack Go"

━━ FASTEST PATH TO HIGHER RATES ━━━━━━━━━━━━━━━━━━
1. Pick ONE AI skill from the top list
2. Build ONE real project with it this week
3. Add it to your profile with the project as proof
4. Start applying to roles with that skill immediately

Set your skills to personalise: /set skills=[your stack]`
}

func clientManagement() string {
	return `Client Management Playbook

━━ ONBOARDING (first 48 hours) ━━━━━━━━━━━━━━━━━━━
□ Send welcome email with clear next steps
□ Share project brief / questionnaire
□ Set up shared workspace (Notion, Google Drive, etc.)
□ Confirm kickoff call time
□ Clarify: deliverables, deadlines, revision rounds, communication channel

━━ COMMUNICATION CADENCE ━━━━━━━━━━━━━━━━━━━━━━━━━
Short projects (<2 weeks): Update every 2–3 days
Long projects (>2 weeks) : Weekly status update (Friday)
Always: Respond to messages within 24 hours (business days)
Never: Go silent for >3 days without a heads-up

━━ SCOPE CREEP PREVENTION ━━━━━━━━━━━━━━━━━━━━━━━━
When client adds new requests:
"I'd love to help with that. It's outside the original scope,
so I'd add it as a separate item at [rate]. Shall I include it?"

Always put scope in writing before starting.

━━ TURNING ONE-OFF → RETAINER ━━━━━━━━━━━━━━━━━━━━
After delivering great work:
"I've really enjoyed working on this with you.
Many clients find it valuable to have [your service] ongoing.
I offer a monthly retainer at [X] which includes [scope].
Would that be useful to explore?"

Timing: Ask when they're happiest — at delivery, not during revision.

━━ DIFFICULT CLIENT SCRIPTS ━━━━━━━━━━━━━━━━━━━━━━
Moving goalposts: "To keep us on schedule, could we finalise [X] by [date]?"
Late payment: "Invoice [#] was due [date]. Could you confirm payment timing?"
Endless revisions: "We've completed the [N] revision rounds included.
Additional revisions are available at [rate]."`
}

func nicheStrategy(mem *Memory) string {
	return `Niche Strategy for Freelancing

━━ WHY NICHING WORKS ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Generalist: "I'm a developer" → competes with everyone, low rates
Niche: "I build Shopify stores for DTC beauty brands" → premium rates, less competition

The paradox: the narrower you go, the more you earn.

━━ TOP PROFITABLE NICHES RIGHT NOW ━━━━━━━━━━━━━━━
1. AI integration for SaaS companies
2. Web3 / smart contract development
3. Fintech product design (UX + dev)
4. E-commerce conversion optimisation
5. Cybersecurity for SMBs
6. Technical content for developer tools
7. Marketing automation + AI workflows
8. Healthcare/medtech software
9. Creator economy tools and apps
10. B2B SaaS growth (copywriting + analytics)

━━ HOW TO PICK YOUR NICHE ━━━━━━━━━━━━━━━━━━━━━━━━
Best intersection of:
  • Skills you have (or can build fast)
  • Industries you enjoy or understand
  • High willingness to pay

━━ POSITIONING STATEMENT FORMULA ━━━━━━━━━━━━━━━━━
"I help [specific type of client] achieve [specific result]
using [your method/skill]."

Example: "I help B2B SaaS companies reduce churn by building
AI-powered onboarding flows using Claude API and React."

━━ 3 ACTIONS THIS WEEK ━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. Write your positioning statement (use formula above)
2. Update your Upwork headline with niche keywords
3. Send 2 proposals specifically targeting your niche`
}

func interviewPrep() string {
	return `Client Vetting Call Prep

━━ THEY WILL ASK YOU ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

"Tell me about your experience with [X]"
→ Use STAR: Situation, Task, Action, Result
→ Lead with the result: "I built a [result] — here's how:"

"What's your availability/timeline?"
→ Be honest. Vague = loses trust. Specific = professionalism.

"What's your rate?"
→ Give a range. "For a project like this, I typically work in the $X–$Y range,
depending on scope. What budget have you allocated?"

"Have you done this specific thing before?"
→ If yes: example. If no: transferable example + "here's my approach"

━━ YOU SHOULD ASK THEM ━━━━━━━━━━━━━━━━━━━━━━━━━━━
1. "What does success look like for this project in 3 months?"
2. "What's happened with previous freelancers on this?" (reveals red flags)
3. "What's the biggest blocker to moving forward today?"
4. "Who else is involved in approving the work?"
5. "What's the timeline and budget range you're working with?"

━━ RED FLAGS TO WATCH ━━━━━━━━━━━━━━━━━━━━━━━━━━━━
⚠️  "We need this done ASAP" (with no timeline) → scope creep waiting to happen
⚠️  "We had a bad experience with the last 3 freelancers" → could be them, not the freelancers
⚠️  Doesn't know their own budget → low commitment to actually hiring
⚠️  Asks for free work or "test" → walk away

━━ CLOSING THE CALL ━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
"Based on what you've shared, I think I can absolutely deliver this.
I'll send you a proposal by [tomorrow]. Does that work?"`
}

// ── Bodhi (General) ───────────────────────────────────────────────────────────

func handleBodhi(input string, mem *Memory) string {
	lower := strings.ToLower(input)

	// Catch a few more patterns even in general mode
	if strings.Contains(lower, "hello") || strings.Contains(lower, "hi ") || lower == "hi" || lower == "hey" {
		return greeting()
	}
	if strings.Contains(lower, "how are you") || strings.Contains(lower, "who are you") {
		return whoAmI()
	}
	return fmt.Sprintf(`I'm Bodhi — I received your message:

"%s"

I can help you with any of these topics. Just ask naturally:

  📈 Trading      — "BTC trade plan", "funding rate analysis"       (/use perp-markets)
  💼 Portfolio    — "review my portfolio", "rebalance"               (/use portfolio)
  📣 Social       — "draft a tweet", "LinkedIn post about AI"        (/use social)
  💬 Comms        — "write email to client", "negotiate my rate"     (/use comms)
  🗂  Organizer    — "brain dump", "daily plan", "I'm stuck"         (/use organizer)
  💰 Finance      — "explain RSI", "macro briefing"                  (/use finance)
  🔍 Freelance    — "find Upwork jobs", "write proposal"             (/use freelance)
  💻 Code         — "debug this error", "review my code"             (/use code)
  🏃 Health       — "create workout plan", "nutrition advice"        (/use health)
  🔬 Research     — "deep dive on X", "compare A vs B"              (/use research)
  📰 News         — "what's moving markets", "crypto signals"        (/use news)

Type /help for all commands or /agents for the full agent list.`, input)
}

// ── Shared helpers ────────────────────────────────────────────────────────────

func extractTopic(input string) string {
	// Strip command prefixes
	for _, prefix := range []string{"/social twitter ", "/social linkedin ", "/social instagram ",
		"/social tiktok ", "/social youtube ", "/social ", "/trade ", "/use social ",
		"/apply ", "draft a tweet about ", "write a linkedin post about ", "post about "} {
		if idx := strings.Index(strings.ToLower(input), prefix); idx != -1 {
			return strings.TrimSpace(input[idx+len(prefix):])
		}
	}
	// Last words as topic
	words := strings.Fields(input)
	if len(words) > 3 {
		return strings.Join(words[len(words)-5:], " ")
	}
	return input
}

func greeting() string {
	return `Namaste! I'm Bodhi 🌸

Your all-in-one AI assistant — running fully offline, no keys needed.

What I can do for you:
  📈 Crypto trade plans + technical analysis  (/use perp-markets)
  💼 Portfolio strategy + DeFi yields          (/use portfolio)
  📣 Social media content (all platforms)      (/use social)
  💬 Emails, DMs, proposals, negotiation       (/use comms)
  🗂  Daily/weekly planning + task management  (/use organizer)
  💰 Finance explainers + market frameworks    (/use finance)
  🔍 Freelance job search + proposals + rates  (/use freelance)
  💻 Code debugging, review, architecture      (/use code)
  🏃 Training, nutrition, sleep, recovery      (/use health)
  🔬 Deep research + fact-checked analysis     (/use research)
  📰 Market signals + news + trend analysis    (/use news)

Just ask naturally — I route automatically. Or /help to see all commands.`
}

func whoAmI() string {
	return `I'm Bodhi Hub — an autonomous AI assistant that runs entirely on your machine.

No API keys. No internet connection required. No data sent anywhere.
All intelligence is built directly into the bot.

I have 12 specialist agents inside me:
  perp-markets  — Perpetual Markets Strategist
  portfolio     — Portfolio Strategist
  social        — Social Media Expert (all 7 platforms)
  comms         — Communication Specialist
  organizer     — Personal Organizer
  finance       — Finance & Crypto Analyst
  freelance     — Freelance & Jobs Advisor
  code          — Code Assistant (any language)
  health        — Health & Fitness
  research      — Research Analyst
  news          — News & Trends
  bodhi         — General (that's me, the fallback)

I route your messages automatically. You can also force a specific agent:
/use perp-markets BTC trade plan long`
}

// ── Code Assistant ────────────────────────────────────────────────────────────

func handleCode(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "debug") || strings.Contains(lower, "error") || strings.Contains(lower, "bug "):
		return "Share the error message and the relevant code. I'll identify the root cause and give you a tested fix."
	case strings.Contains(lower, "review"):
		return "Paste the code you want reviewed. I'll assess: correctness, edge cases, performance, security, and readability."
	case strings.Contains(lower, "refactor"):
		return "Share the code to refactor. Tell me the goal (readability, performance, testability) and I'll rewrite it cleanly."
	case strings.Contains(lower, "architect") || strings.Contains(lower, "design") || strings.Contains(lower, "structure"):
		return "Describe what you're building — language, scale, constraints, team size. I'll outline the architecture with trade-offs."
	case strings.Contains(lower, "test") || strings.Contains(lower, "unit test"):
		return "Share the function or module. I'll write unit tests with edge cases, table-driven where appropriate."
	case strings.Contains(lower, "algorithm") || strings.Contains(lower, "data structure"):
		return "Describe the problem: input, output, constraints, scale. I'll pick the right algorithm and explain the complexity."
	default:
		return "Share your code or describe what you're building. I handle debugging, reviews, architecture, testing, and implementation in any language."
	}
}

// ── Health & Fitness ──────────────────────────────────────────────────────────

func handleHealth(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "workout") || strings.Contains(lower, "gym") || strings.Contains(lower, "exercise") || strings.Contains(lower, "training"):
		return "Tell me: goal (strength / muscle / fat loss / endurance), days per week, equipment access, and any injuries. I'll build a specific programme."
	case strings.Contains(lower, "diet") || strings.Contains(lower, "nutrition") || strings.Contains(lower, "calorie") || strings.Contains(lower, "meal") || strings.Contains(lower, "macro"):
		return "Share your goal, current eating habits, and any restrictions. I'll give you a practical nutrition framework with macro targets — no fad diets."
	case strings.Contains(lower, "sleep"):
		return "Key sleep levers: fixed wake time (anchors circadian rhythm), no screens 60 min before bed, room at 18–19°C, no caffeine after 1pm. What's your specific issue?"
	case strings.Contains(lower, "stress") || strings.Contains(lower, "burnout") || strings.Contains(lower, "mental"):
		return "Quick tools: box breathing (4-4-4-4), 20-min walk, single-tasking. Long-term: what's the source of the stress? Let's identify it."
	case strings.Contains(lower, "weight loss") || strings.Contains(lower, "lose weight") || strings.Contains(lower, "cut "):
		return "Sustainable fat loss: 300–500 kcal deficit, high protein (2g/kg body weight), resistance training to preserve muscle, 7–8h sleep. Tell me your stats and I'll calculate your numbers."
	default:
		return "Tell me your health goal and current routine. I give specific, evidence-based guidance — workout plans, nutrition targets, recovery protocols, and mental performance."
	}
}

// ── Research Analyst ──────────────────────────────────────────────────────────

func handleResearch(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	topic := truncate(input, 80)
	switch {
	case strings.Contains(lower, "compare") || strings.Contains(lower, " vs ") || strings.Contains(lower, "versus"):
		return fmt.Sprintf("Comparison: %q — I'll structure this as a criteria table → key differences → use-case recommendation. What criteria matter most?", topic)
	case strings.Contains(lower, "pros and cons") || strings.Contains(lower, "trade-off") || strings.Contains(lower, "tradeoff"):
		return fmt.Sprintf("Trade-off analysis: %q — I'll cover benefits, risks, hidden costs, and which context each option wins in.", topic)
	case strings.Contains(lower, "summarize") || strings.Contains(lower, "summarise") || strings.Contains(lower, "breakdown"):
		return fmt.Sprintf("Summary: %q — paste the source text and I'll produce a structured breakdown with key takeaways front-loaded.", topic)
	default:
		return fmt.Sprintf("Research: %q — I'll generate a structured deep-dive: background, key findings, competing perspectives, open questions. What angle interests you most?", topic)
	}
}

// ── News & Trends ─────────────────────────────────────────────────────────────

func handleNews(input string, mem *Memory) string {
	lower := strings.ToLower(input)
	switch {
	case strings.Contains(lower, "crypto") || strings.Contains(lower, "bitcoin") || strings.Contains(lower, "btc") || strings.Contains(lower, "eth"):
		return "Crypto news sources worth tracking: The Block, CoinDesk, Decrypt (news) · Glassnode, CryptoQuant (on-chain) · @WatcherGuru, @lookonchain (Twitter/X). What event are you following?"
	case strings.Contains(lower, "ai") || strings.Contains(lower, "tech"):
		return "AI & tech signal sources: Hacker News, arXiv cs.AI, Anthropic/OpenAI/Google DeepMind blogs, GitHub Trending. What area — models, infrastructure, products, or policy?"
	case strings.Contains(lower, "market") || strings.Contains(lower, "stock") || strings.Contains(lower, "equity"):
		return "Key market catalysts to watch: Fed statements, CPI/PCE, earnings surprises, DXY moves. Which market or sector? I can route deeper analysis to the Finance agent."
	default:
		return "What topic or sector are you tracking? I'll cut through the noise and flag what actually matters versus what's just headlines."
	}
}

// handlePerpMarketsAgent wraps handlePerpMarkets to match Agent.Handle signature.
func init() {
	// Ensure the registry agents use the right handler signatures
	_ = handlePerpMarketsAgent
}
