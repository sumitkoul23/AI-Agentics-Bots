package main

import (
	"context"
	"log"

	"github.com/robfig/cron/v3"
)

// Scheduler runs Priya's autonomous background tasks.
type Scheduler struct {
	c         *cron.Cron
	social    *SocialModule
	finance   *FinanceModule
	freelance *FreelanceModule
	organizer *OrganizerModule
	mem       *Memory
}

func NewScheduler(
	social *SocialModule,
	finance *FinanceModule,
	freelance *FreelanceModule,
	organizer *OrganizerModule,
	mem *Memory,
) *Scheduler {
	return &Scheduler{
		c:         cron.New(),
		social:    social,
		finance:   finance,
		freelance: freelance,
		organizer: organizer,
		mem:       mem,
	}
}

// Start registers all autonomous tasks and begins the scheduler.
func (s *Scheduler) Start() {
	ctx := context.Background()

	// Daily morning briefing at 8:00 AM
	s.c.AddFunc("0 8 * * *", func() {
		log.Println("[Priya Scheduler] Running morning briefing...")
		reply, err := s.organizer.dailyBriefing(ctx)
		if err != nil {
			log.Printf("[Priya Scheduler] Morning briefing error: %v", err)
			return
		}
		s.mem.AddMessage("assistant", "[Morning Briefing]\n"+reply)
		_ = s.mem.Save()
		log.Println("[Priya Scheduler] Morning briefing complete.")
	})

	// Market scan every 4 hours
	s.c.AddFunc("0 */4 * * *", func() {
		log.Println("[Priya Scheduler] Running market scan...")
		result := s.finance.AutonomousScan(ctx)
		if result != "" {
			log.Println("[Priya Scheduler] Market scan saved to memory.")
		}
	})

	// Job scan every 6 hours
	s.c.AddFunc("0 */6 * * *", func() {
		log.Println("[Priya Scheduler] Running job scan...")
		result := s.freelance.AutonomousJobScan(ctx)
		if result != "" {
			log.Println("[Priya Scheduler] Job scan saved to memory.")
		}
	})

	// Social media trend check at 9:00 AM daily
	s.c.AddFunc("0 9 * * *", func() {
		log.Println("[Priya Scheduler] Running social trends check...")
		s.social.AutonomousPost(ctx)
		log.Println("[Priya Scheduler] Social post queued.")
	})

	// Weekly plan every Monday at 7:30 AM
	s.c.AddFunc("30 7 * * 1", func() {
		log.Println("[Priya Scheduler] Generating weekly plan...")
		reply, err := s.organizer.weeklyPlan(ctx, "autonomous weekly planning")
		if err != nil {
			log.Printf("[Priya Scheduler] Weekly plan error: %v", err)
			return
		}
		s.mem.AddMessage("assistant", "[Weekly Plan]\n"+reply)
		_ = s.mem.Save()
		log.Println("[Priya Scheduler] Weekly plan saved.")
	})

	s.c.Start()
	log.Println("[Priya Scheduler] All autonomous tasks running.")
}

func (s *Scheduler) Stop() {
	s.c.Stop()
}
