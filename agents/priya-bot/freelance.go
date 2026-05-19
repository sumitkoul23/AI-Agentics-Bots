package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// FreelanceModule handles job search, proposals, applications, and skills.
type FreelanceModule struct {
	ai  *AICore
	mem *Memory
}

func NewFreelanceModule(ai *AICore, mem *Memory) *FreelanceModule {
	return &FreelanceModule{ai: ai, mem: mem}
}

func (f *FreelanceModule) Handle(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)

	switch {
	case strings.Contains(lower, "apply") || strings.Contains(lower, "proposal") || strings.Contains(lower, "cover letter") || strings.Contains(lower, "bid"):
		return f.proposal(ctx, input)
	case strings.Contains(lower, "track") || strings.Contains(lower, "application") || strings.Contains(lower, "status"):
		return f.tracker(ctx)
	case strings.Contains(lower, "skill") || strings.Contains(lower, "gap") || strings.Contains(lower, "profile"):
		return f.skillGap(ctx, input)
	case strings.Contains(lower, "rate") || strings.Contains(lower, "price") || strings.Contains(lower, "charge"):
		return f.rateStrategy(ctx, input)
	case strings.Contains(lower, "client") || strings.Contains(lower, "onboard"):
		return f.clientStrategy(ctx, input)
	case strings.Contains(lower, "niche") || strings.Contains(lower, "speciali"):
		return f.nicheStrategy(ctx, input)
	case strings.Contains(lower, "interview") || strings.Contains(lower, "vetting"):
		return f.interviewPrep(ctx, input)
	default:
		return f.jobSearch(ctx, input)
	}
}

func (f *FreelanceModule) jobSearch(ctx context.Context, query string) (string, error) {
	keywords := extractArgs(query, "/jobs")
	if keywords == "" {
		keywords = query
	}
	if keywords == "" {
		keywords = f.mem.GetPreference("primary_skill")
	}
	if keywords == "" {
		keywords = "software development"
	}
	prompt := fmt.Sprintf(`Search and rank freelance opportunities for: %s

Return top 5 opportunities across Upwork, Fiverr, Toptal, LinkedIn Jobs, and Contra.

For each:
• Platform + job title
• Budget/rate range
• Client quality indicators
• Why this is a good fit
• Match score (1–10)
• First line of a winning proposal

Then: the single best opportunity to prioritise today and why.

(Note: live search requires JOBS_API_KEY — showing AI-generated recommendations based on market knowledge.)`, keywords)

	reply, err := f.ai.Think(ctx, prompt)
	if err != nil {
		return "", err
	}
	// Log the search
	f.mem.Learn("last_job_search", keywords)
	_ = f.mem.Save()
	return reply, nil
}

func (f *FreelanceModule) proposal(ctx context.Context, details string) (string, error) {
	jobTitle := extractArgs(details, "/apply")
	if jobTitle == "" {
		jobTitle = details
	}
	prompt := fmt.Sprintf(`Write a winning freelance proposal for: %s

Structure:
━━ OPENING HOOK (1–2 sentences that prove I read their post)
━━ RELEVANT EXPERIENCE (2–3 sentences, specific and credible)
━━ MY APPROACH (how I'd solve their specific problem)
━━ SOCIAL PROOF (1 result or metric)
━━ CLEAR CTA + AVAILABILITY
━━ RATE/TIMELINE (if appropriate)

Tone: confident, specific, not generic. Under 250 words.
Make the client feel like I already understand their problem.

Also provide:
• Subject line (if email)
• 2 questions to ask the client in the first call
• Red flags to watch for in this type of job`, jobTitle)

	reply, err := f.ai.Think(ctx, prompt)
	if err != nil {
		return "", err
	}
	// Auto-track the application
	f.mem.AddTrackedJob(TrackedJob{
		Title:     jobTitle,
		Platform:  "Unknown",
		Status:    "Proposal Drafted",
		AppliedAt: time.Now(),
		FollowUp:  time.Now().AddDate(0, 0, 5),
	})
	_ = f.mem.Save()
	return reply, nil
}

func (f *FreelanceModule) tracker(ctx context.Context) (string, error) {
	jobs := f.mem.Data.TrackedJobs
	if len(jobs) == 0 {
		return `Application Tracker — No applications logged yet.

Use /apply <job title> to draft a proposal (auto-logged).
Or tell me: "Log application: [title] on [platform], status [applied/interview/offer]"`, nil
	}

	var sb strings.Builder
	sb.WriteString("Application Tracker\n\n")
	sb.WriteString(fmt.Sprintf("%-3s %-30s %-12s %-15s %s\n", "#", "Job", "Platform", "Status", "Follow-up"))
	sb.WriteString(strings.Repeat("─", 80) + "\n")
	for i, j := range jobs {
		sb.WriteString(fmt.Sprintf("%-3d %-30s %-12s %-15s %s\n",
			i+1,
			truncate(j.Title, 28),
			truncate(j.Platform, 10),
			j.Status,
			j.FollowUp.Format("2006-01-02"),
		))
	}

	// Ask AI for follow-up recommendations
	prompt := fmt.Sprintf("I have %d tracked job applications. Give me a quick prioritised follow-up plan for today — which to follow up on and what to say.", len(jobs))
	aiAdvice, err := f.ai.Think(ctx, prompt)
	if err == nil {
		sb.WriteString("\n\n" + aiAdvice)
	}
	return sb.String(), nil
}

func (f *FreelanceModule) skillGap(ctx context.Context, details string) (string, error) {
	currentSkills := f.mem.GetPreference("skills")
	prompt := fmt.Sprintf(`Perform a skill-gap analysis for the current freelance market.
My skills: %s
Context: %s

Deliver:
1. Top 10 in-demand skills RIGHT NOW (ranked by earning potential)
2. Which I likely already have vs. need to learn
3. Fastest path to skill #1 gap (resources, timeline, project to build)
4. Profile optimisation: exact keywords to add to Upwork/LinkedIn/Fiverr
5. One certification or credential that commands a rate premium

Be specific — no generic "learn Python" advice.`, currentSkills, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FreelanceModule) rateStrategy(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me set and negotiate my freelance rates. Context: %s

Cover:
1. Market rate benchmark for my skill set (hourly + project-based)
2. Value-based pricing framework (stop selling hours)
3. How to raise rates with existing clients
4. Handling "your rate is too high" objection (3 responses)
5. Proposal pricing structure (retainer / milestone / hourly — which to use when)
6. Package tiers to offer (Basic / Professional / Premium)

Give me exact numbers and scripts, not principles.`, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FreelanceModule) clientStrategy(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Build a client management strategy for: %s

Include:
1. Client onboarding checklist (first 48 hours)
2. Communication cadence (how often to update, which channel)
3. Scope creep prevention script
4. How to handle a difficult/demanding client
5. Turning one-off clients into retainers (script + timing)
6. Red flags to spot in new clients before signing

Give me scripts and templates, not advice.`, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FreelanceModule) nicheStrategy(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Help me niche down my freelance business: %s

Analyse:
1. Top 5 most profitable niches for my skills right now
2. Pros/cons of each (competition, rate ceiling, client quality)
3. My best-fit niche based on context
4. Positioning statement for chosen niche (1 sentence)
5. First 3 actions to establish authority in that niche this week

Niche selection changes everything — be honest if my current direction isn't optimal.`, details)
	return f.ai.Think(ctx, prompt)
}

func (f *FreelanceModule) interviewPrep(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Prepare me for a client call or vetting interview: %s

Provide:
1. Top 10 questions they'll likely ask — with ideal answers
2. 5 smart questions I should ask them (shows expertise)
3. How to discuss rates confidently
4. Red flags to listen for about the project/client
5. Closing script to end the call with clear next steps

Also: what NOT to say in a first client call.`, details)
	return f.ai.Think(ctx, prompt)
}

// AutonomousJobScan runs on schedule to find new opportunities.
func (f *FreelanceModule) AutonomousJobScan(ctx context.Context) string {
	reply, err := f.jobSearch(ctx, f.mem.GetPreference("primary_skill"))
	if err != nil {
		return ""
	}
	f.mem.AddMessage("assistant", "[Autonomous job scan]\n"+reply)
	_ = f.mem.Save()
	return reply
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-1] + "…"
}
