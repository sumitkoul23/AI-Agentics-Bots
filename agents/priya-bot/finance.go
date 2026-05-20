package main

import (
	"context"
	"fmt"
	"strings"
)

// FinanceModule handles crypto, stocks, forex, and portfolio.
type FinanceModule struct {
	ai  *AICore
	mem *Memory
}

func NewFinanceModule(ai *AICore, mem *Memory) *FinanceModule {
	return &FinanceModule{ai: ai, mem: mem}
}

func (f *FinanceModule) Handle(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)
	args := extractArgs(input, "/trade")
	if args == "" {
		args = extractArgs(input, "/finance")
	}

	switch {
	case strings.Contains(lower, "portfolio") || strings.Contains(lower, "holdings"):
		return f.portfolio(ctx, args)
	case strings.Contains(lower, "defi") || strings.Contains(lower, "yield"):
		return f.defi(ctx, args)
	case strings.Contains(lower, "news") || strings.Contains(lower, "macro"):
		return f.macro(ctx, args)
	case strings.Contains(lower, "explain") || strings.Contains(lower, "what is"):
		return f.explain(ctx, args)
	case strings.Contains(lower, "screen") || strings.Contains(lower, "find stocks") || strings.Contains(lower, "find coins"):
		return f.screen(ctx, args)
	default:
		return f.tradePlan(ctx, args)
	}
}

func (f *FinanceModule) tradePlan(ctx context.Context, input string) (string, error) {
	if input == "" {
		input = "BTC — full analysis"
	}
	prompt := fmt.Sprintf(`Generate a complete trade plan for: %s

Structure:
━━ MARKET CONTEXT
• Current trend (higher timeframe)
• Key support & resistance levels
• Volume and momentum read
• Sentiment (funding, fear/greed, on-chain if crypto)

━━ TRADE SETUP
• Direction: Long / Short / No trade (with reason)
• Entry zone: [price range]
• Stop-loss: [price] — [%%] risk
• Target 1: [price] (+[%%])
• Target 2: [price] (+[%%])
• Invalidation: [condition]

━━ EXECUTION
• Preferred timeframe for entry signal
• Position size formula (1%% account risk)
• Risk/Reward ratio
• Timing note

━━ RISK FLAGS
List 2–3 things that could invalidate this setup immediately.

⚠️ Advisory mode — no live orders placed without explicit confirmation.`, input)
	return f.ai.Think(ctx, prompt)
}

func (f *FinanceModule) portfolio(ctx context.Context, details string) (string, error) {
	holdings := f.mem.GetPreference("portfolio_holdings")
	context_ := details
	if holdings != "" {
		context_ = "Holdings: " + holdings + ". " + details
	}
	prompt := fmt.Sprintf(`Analyse this portfolio and give actionable recommendations: %s

Cover:
1. Current allocation breakdown (%%per asset if holdings known)
2. Risk assessment (concentration, volatility, correlation)
3. Top 2 actions to take this week (rebalance, add, reduce, hedge)
4. Tax efficiency note
5. Suggested target allocation

If no holdings provided, describe what an ideal diversified portfolio looks like for someone in crypto + stocks + freelance income.`, context_)
	return f.ai.Think(ctx, prompt)
}

func (f *FinanceModule) defi(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`DeFi yield opportunities analysis: %s

Cover:
• Top 5 yield strategies by risk tier (low / medium / high)
• Protocol risk assessment (smart contract, liquidity, team)
• Current APY ranges (approximate)
• Entry steps for best risk-adjusted opportunity
• Exit signals to watch

Flag any impermanent loss risks clearly.`, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FinanceModule) macro(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Macro and crypto market news briefing. %s

Format:
1. Top 3 market-moving stories today (1 sentence each + impact)
2. Key economic events this week (dates + expected impact)
3. Overall market sentiment (risk-on / risk-off / neutral)
4. What this means for my portfolio — one concrete action
5. Assets to watch this week

Be direct. No padding.`, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FinanceModule) explain(ctx context.Context, topic string) (string, error) {
	prompt := fmt.Sprintf(`Explain this financial concept clearly: %s

Use:
• Plain language first (no jargon)
• A real-world analogy
• A crypto/stock example
• Why it matters for trading or investing
• One common mistake people make with this concept

Keep it mobile-friendly — concise but complete.`, topic)
	return f.ai.Think(ctx, prompt)
}

func (f *FinanceModule) screen(ctx context.Context, criteria string) (string, error) {
	prompt := fmt.Sprintf(`Screen for investment opportunities matching: %s

Return:
• Top 5 candidates with brief thesis for each
• Key metric that makes each one interesting
• Risk level (1–5)
• Suggested entry approach (lump sum / DCA / wait for pullback)
• The single best pick and why

⚠️ Educational only — verify with live data before trading.`, criteria)
	return f.ai.Think(ctx, prompt)
}

// AutonomousScan runs on a schedule to flag market alerts.
func (f *FinanceModule) AutonomousScan(ctx context.Context) string {
	reply, err := f.macro(ctx, "")
	if err != nil {
		return ""
	}
	f.mem.AddMessage("assistant", "[Autonomous market scan]\n"+reply)
	_ = f.mem.Save()
	return reply
}
