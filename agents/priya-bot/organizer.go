package main

import (
	"context"
	"fmt"
	"strings"
)

// OrganizerModule cleans messes, prioritises tasks, and manages the user's work.
type OrganizerModule struct {
	ai  *AICore
	mem *Memory
}

func NewOrganizerModule(ai *AICore, mem *Memory) *OrganizerModule {
	return &OrganizerModule{ai: ai, mem: mem}
}

func (o *OrganizerModule) Handle(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "brain dump") || strings.Contains(lower, "braindump") || strings.Contains(lower, "my mess"):
		return o.brainDump(ctx, input)
	case strings.Contains(lower, "prioriti") || strings.Contains(lower, "urgent") || strings.Contains(lower, "todo"):
		return o.prioritise(ctx, input)
	case strings.Contains(lower, "calendar") || strings.Contains(lower, "schedule") || strings.Contains(lower, "block"):
		return o.calendarBlock(ctx, input)
	case strings.Contains(lower, "delegate") || strings.Contains(lower, "outsource"):
		return o.delegate(ctx, input)
	case strings.Contains(lower, "daily") || strings.Contains(lower, "morning") || strings.Contains(lower, "routine"):
		return o.dailyBriefing(ctx)
	case strings.Contains(lower, "week") || strings.Contains(lower, "plan"):
		return o.weeklyPlan(ctx, input)
	case strings.Contains(lower, "stuck") || strings.Contains(lower, "procrastinat") || strings.Contains(lower, "blocked"):
		return o.unstuck(ctx, input)
	default:
		return o.brainDump(ctx, input)
	}
}

func (o *OrganizerModule) brainDump(ctx context.Context, dump string) (string, error) {
	prompt := fmt.Sprintf(`Take this brain dump and turn it into a clean, actionable system: %s

Output:
━━ IMMEDIATE (do today)
[List with estimated time per task]

━━ THIS WEEK
[List with deadline suggestions]

━━ BACKLOG (no deadline yet)
[List]

━━ DELETE / IGNORE
[Things not worth doing]

━━ QUICK WINS (under 10 min)
[List — do these first to build momentum]

Then: give me the ONE thing I should start with right now and why.`, dump)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) prioritise(ctx context.Context, tasks string) (string, error) {
	prompt := fmt.Sprintf(`Prioritise these tasks using urgency × impact matrix: %s

Format as:
Q1 (Urgent + High Impact) — DO NOW
Q2 (Not Urgent + High Impact) — SCHEDULE
Q3 (Urgent + Low Impact) — DELEGATE or BATCH
Q4 (Not Urgent + Low Impact) — DELETE

After the matrix:
• Top 3 for today (with time blocks)
• What I should NOT work on today
• Energy management tip for the task list`, tasks)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) calendarBlock(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Create an optimised daily time-block schedule. Context: %s

Build a full day schedule with:
• Deep work blocks (protect from meetings)
• Admin/email batching windows
• Social media posting times
• Market check-ins (if trading)
• Energy-aware scheduling (hard tasks in peak hours)
• Buffer blocks for unexpected items
• End-of-day review slot

Output as a clean table: Time | Activity | Duration | Notes

Assume I work [flexible/remote] unless specified otherwise.`, details)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) delegate(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me identify what to delegate or outsource: %s

For each delegatable item:
• Suggested platform (Fiverr, VA, automation tool)
• Estimated cost and time to brief
• What to include in the brief
• Risk of delegating vs doing it myself

Then: build me a short brief template I can post immediately on Fiverr/Upwork for the highest-priority item.`, details)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) dailyBriefing(ctx context.Context) (string, error) {
	jobs := o.mem.Data.TrackedJobs
	posts := o.mem.Data.ScheduledPosts

	var contextParts []string
	if len(jobs) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("%d tracked job applications", len(jobs)))
	}
	if len(posts) > 0 {
		contextParts = append(contextParts, fmt.Sprintf("%d scheduled social posts", len(posts)))
	}
	context_ := strings.Join(contextParts, ", ")
	if context_ == "" {
		context_ = "general daily briefing"
	}

	prompt := fmt.Sprintf(`Generate my morning briefing. Context: %s

Include:
1. Top 3 priorities for today (based on memory)
2. Social media to-do (what to post today)
3. Finance check (market sentiment, anything to watch)
4. Freelance/jobs (applications to follow up, new opportunities)
5. Quick win I can complete in the first 30 minutes
6. Energy/mindset note

Keep it crisp — this is my morning overview, not a novel.`, context_)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) weeklyPlan(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Build my weekly work plan. %s

Structure:
━━ WEEK THEME / MAIN GOAL
━━ MONDAY–FRIDAY breakdown
  Each day: 2–3 key tasks + time estimate
━━ SOCIAL MEDIA (what to post each day + platform)
━━ FINANCE (trading check-in days + portfolio review)
━━ FREELANCE (proposals to send, clients to follow up)
━━ PERSONAL DEVELOPMENT (1 thing to learn or do)
━━ DO NOT DO this week (intentional focus)

End with: the single metric that will tell me this was a good week.`, details)
	return o.ai.Think(ctx, prompt)
}

func (o *OrganizerModule) unstuck(ctx context.Context, situation string) (string, error) {
	prompt := fmt.Sprintf(`I'm stuck or procrastinating on: %s

Help me:
1. Name the real reason I'm stuck (fear, unclear next step, wrong time, energy?)
2. Break it into the smallest possible first action (2-minute start)
3. Address the mental block directly
4. Give me a 25-minute focused work sprint plan for just this task
5. What happens if I keep avoiding it? (honest impact)

Be direct — no motivational fluff. Just help me start.`, situation)
	return o.ai.Think(ctx, prompt)
}
