package main

import (
	"context"
	"fmt"
	"strings"
)

// CommsModule handles all communication drafting and management.
type CommsModule struct {
	ai  *AICore
	mem *Memory
}

func NewCommsModule(ai *AICore, mem *Memory) *CommsModule {
	return &CommsModule{ai: ai, mem: mem}
}

func (c *CommsModule) Handle(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "email") || strings.Contains(lower, "mail"):
		return c.email(ctx, input)
	case strings.Contains(lower, "dm") || strings.Contains(lower, "direct message"):
		return c.directMessage(ctx, input)
	case strings.Contains(lower, "follow up") || strings.Contains(lower, "followup"):
		return c.followUp(ctx, input)
	case strings.Contains(lower, "negotiat"):
		return c.negotiate(ctx, input)
	case strings.Contains(lower, "decline") || strings.Contains(lower, "reject") || strings.Contains(lower, "say no"):
		return c.decline(ctx, input)
	case strings.Contains(lower, "onboard") || strings.Contains(lower, "welcome"):
		return c.onboard(ctx, input)
	case strings.Contains(lower, "complaint") || strings.Contains(lower, "issue") || strings.Contains(lower, "unhappy"):
		return c.handleComplaint(ctx, input)
	case strings.Contains(lower, "clean") || strings.Contains(lower, "inbox") || strings.Contains(lower, "triage"):
		return c.triageInbox(ctx, input)
	default:
		return c.ai.Think(ctx, "Communication request: "+input)
	}
}

func (c *CommsModule) email(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Draft a professional email based on this context: %s

Deliver:
• Subject line (2 options — one direct, one curiosity-driven)
• Full email body (3–4 short paragraphs max)
• Tone: [infer from context — professional / warm / urgent]
• Signature placeholder
• Follow-up timing suggestion

Write in my voice. No corporate fluff. Get to the point quickly.`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) directMessage(ctx context.Context, details string) (string, error) {
	platform := "LinkedIn"
	if strings.Contains(strings.ToLower(details), "twitter") || strings.Contains(strings.ToLower(details), "x.com") {
		platform = "Twitter/X"
	} else if strings.Contains(strings.ToLower(details), "upwork") {
		platform = "Upwork"
	}
	prompt := fmt.Sprintf(`Write a %s DM for this situation: %s

Make it:
• Short (under 100 words for cold, up to 200 for warm)
• Opening that references something specific about them
• Clear value proposition or ask in one sentence
• Soft CTA — no pressure
• Feel like a real human wrote it, not a template

Provide 2 versions: [Direct] and [Conversational]`, platform, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) followUp(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Write a follow-up message for: %s

Include:
• A brief callback to the original message (1 sentence)
• New value or angle — don't just say "following up"
• Clear next step or question
• Keep it under 80 words

Also suggest: how many days since original message is ideal to send this, and what to do if no response after this follow-up.`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) negotiate(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me negotiate this situation: %s

Provide:
1. My ideal outcome and a reasonable floor
2. Opening message to start the negotiation
3. 3 counter-offer responses for common pushbacks
4. When to walk away signal
5. Closing message to seal the deal

Tone: confident, collaborative, not desperate.`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) decline(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Write a professional, kind decline message for: %s

Requirements:
• Express genuine appreciation
• Give a brief, honest reason (no over-explaining)
• Keep the door open for future collaboration if appropriate
• Under 80 words
• Leaves the other person feeling respected, not rejected

Draft 2 versions: [Firm] and [Warm/Open-door]`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) onboard(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Create a client/user onboarding communication sequence for: %s

Deliver:
1. Welcome message (Day 0)
2. Day 3 check-in
3. Day 7 value recap
4. Day 14 milestone check + upsell opener

Each message: subject + body (under 150 words). Warm, human, not robotic.`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) handleComplaint(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me handle this complaint or difficult situation: %s

Provide:
1. What NOT to say (common mistakes)
2. Empathy-first opening
3. Full response message
4. Resolution offer options
5. Internal note: was this our fault? What to fix?

Tone: calm, accountable, solution-focused.`, details)
	return c.ai.Think(ctx, prompt)
}

func (c *CommsModule) triageInbox(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me triage and clean up my inbox/messages. Context: %s

I need you to:
1. Categorise by: [Urgent Action] [Reply Needed] [FYI] [Archive/Delete]
2. Draft quick replies for any [Reply Needed] items
3. Flag anything I might be procrastinating on
4. Suggest what can be batch-replied vs needs personal attention
5. Give me a 15-minute inbox-zero plan

If no specific messages provided, give me a general inbox-triage framework.`, details)
	return c.ai.Think(ctx, prompt)
}
