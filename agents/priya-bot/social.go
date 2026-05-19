package main

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// SocialModule handles all social media domains.
type SocialModule struct {
	ai  *AICore
	mem *Memory
}

func NewSocialModule(ai *AICore, mem *Memory) *SocialModule {
	return &SocialModule{ai: ai, mem: mem}
}

// Handle routes a social request to the right handler.
func (s *SocialModule) Handle(ctx context.Context, input string) (string, error) {
	lower := strings.ToLower(input)
	args := extractArgs(input, "/social")

	switch {
	case strings.Contains(lower, "trends") || strings.Contains(lower, "trending"):
		return s.trends(ctx, args)
	case strings.Contains(lower, "calendar") || strings.Contains(lower, "schedule"):
		return s.contentCalendar(ctx, args)
	case strings.Contains(lower, "reply") || strings.Contains(lower, "respond"):
		return s.engagementReply(ctx, args)
	case strings.Contains(lower, "image") || strings.Contains(lower, "graphic") || strings.Contains(lower, "visual") || strings.Contains(lower, "design"):
		return s.visualBrief(ctx, args)
	case strings.Contains(lower, "strategy") || strings.Contains(lower, "plan") || strings.Contains(lower, "growth"):
		return s.strategy(ctx, args)
	case strings.Contains(lower, "twitter") || strings.Contains(lower, "tweet") || strings.Contains(lower, "x.com"):
		return s.createContent(ctx, "Twitter/X", args)
	case strings.Contains(lower, "linkedin"):
		return s.createContent(ctx, "LinkedIn", args)
	case strings.Contains(lower, "instagram") || strings.Contains(lower, "ig") || strings.Contains(lower, "reel"):
		return s.createContent(ctx, "Instagram", args)
	case strings.Contains(lower, "tiktok") || strings.Contains(lower, "tik tok"):
		return s.createContent(ctx, "TikTok", args)
	case strings.Contains(lower, "youtube") || strings.Contains(lower, "yt"):
		return s.createContent(ctx, "YouTube", args)
	case strings.Contains(lower, "facebook") || strings.Contains(lower, "fb"):
		return s.createContent(ctx, "Facebook", args)
	case strings.Contains(lower, "pinterest"):
		return s.createContent(ctx, "Pinterest", args)
	case strings.Contains(lower, "all") || strings.Contains(lower, "every"):
		return s.createForAllPlatforms(ctx, args)
	default:
		// Let Priya decide the best platform
		return s.ai.Think(ctx, "Social media request: "+input)
	}
}

func (s *SocialModule) createContent(ctx context.Context, platform, topic string) (string, error) {
	if topic == "" {
		topic = "general content about my niche"
	}
	prompt := fmt.Sprintf(`Create high-performing %s content about: %s

Deliver:
1. The complete post/caption (ready to copy-paste)
2. Hashtag set (if applicable to platform)
3. Best time to post (day + hour)
4. One A/B variant
5. A matching image/graphic brief (Midjourney or DALL-E prompt)

Make it feel authentic, not corporate. Use my voice.`, platform, topic)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) createForAllPlatforms(ctx context.Context, topic string) (string, error) {
	if topic == "" {
		topic = "my latest project or insight"
	}
	prompt := fmt.Sprintf(`Create a full cross-platform content pack for: %s

Deliver one version for each:
• Twitter/X (thread + standalone tweet)
• LinkedIn (long-form post)
• Instagram (caption + reel script hook)
• TikTok (hook + script outline)
• YouTube (title + description + tags)
• Facebook (community post)

Keep each version native to the platform. Include hashtags where relevant.
End with a unified image brief that works across all platforms.`, topic)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) trends(ctx context.Context, niche string) (string, error) {
	if niche == "" {
		niche = s.mem.GetPreference("niche")
	}
	if niche == "" {
		niche = "tech, AI, freelancing, and finance"
	}
	prompt := fmt.Sprintf(`Identify the top trending topics, formats, and conversations right now in: %s

For each trend:
• What is it and why is it hot
• Best platform(s) to post about it
• Angle I should take (contrarian / insider / tutorial / story)
• A ready-to-use hook line

Then recommend the single best topic for me to post about today and why.`, niche)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) contentCalendar(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Build a 7-day social media content calendar. %s

Format as a table:
Day | Platform | Content Type | Topic/Hook | Best Time | Status

Include:
• Mix of content pillars (educate, entertain, inspire, promote)
• Platform-native formats (threads, reels, carousels, shorts)
• Engagement prompts (questions, polls, CTAs)
• One piece of personal/behind-the-scenes content

After the table, flag the 2 highest-priority posts to create first.`, details)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) engagementReply(ctx context.Context, context_ string) (string, error) {
	prompt := fmt.Sprintf(`Draft 3 reply options for this comment/DM: %s

Each reply should:
• Feel human, not automated
• Advance the relationship (not just say thanks)
• Be appropriate for the platform tone
• Include a soft next step where natural

Label them: [Warm] [Professional] [Bold]`, context_)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) visualBrief(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Create a detailed visual content brief for: %s

Include:
1. Concept and message
2. Midjourney prompt (detailed, v6 style)
3. DALL-E 3 prompt alternative
4. Canva template instructions (layout, colours, fonts, text overlays)
5. Brand consistency notes
6. Dimensions for each platform (Instagram 1:1, Story 9:16, LinkedIn 1.91:1, etc.)`, details)
	return s.ai.Think(ctx, prompt)
}

func (s *SocialModule) strategy(ctx context.Context, details string) (string, error) {
	prompt := fmt.Sprintf(`Create a social media growth strategy. Context: %s

Cover:
1. Content pillars (3-4 core themes)
2. Platform priority ranking and why
3. Posting frequency per platform
4. Growth tactics for 0–1K and 1K–10K follower stages
5. Monetisation milestones
6. Weekly time commitment estimate
7. The single biggest mistake to avoid

Be direct and specific — no generic advice.`, details)
	return s.ai.Think(ctx, prompt)
}

// AutonomousPost is called by the scheduler to autonomously draft and queue posts.
func (s *SocialModule) AutonomousPost(ctx context.Context) {
	niche := s.mem.GetPreference("niche")
	if niche == "" {
		return
	}
	reply, err := s.trends(ctx, niche)
	if err != nil {
		return
	}
	s.mem.AddScheduledPost(ScheduledPost{
		Platform: "all",
		Content:  reply,
		PostAt:   time.Now().Add(2 * time.Hour),
		Status:   "pending",
	})
	_ = s.mem.Save()
}
