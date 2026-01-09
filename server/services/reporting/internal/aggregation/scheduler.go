// Package aggregation provides scheduled aggregation jobs.
package aggregation

import (
	"context"
	"time"

	"github.com/MuhibNayem/Travio/server/pkg/logger"
	"github.com/MuhibNayem/Travio/server/services/reporting/internal/clickhouse"
	"github.com/robfig/cron/v3"
)

// Scheduler manages scheduled aggregation jobs.
type Scheduler struct {
	cron     *cron.Cron
	ch       *clickhouse.Client
	stopChan chan struct{}
}

// Job represents a scheduled aggregation job.
type Job struct {
	Name  string
	Cron  string
	Query string
}

// DefaultJobs returns the default aggregation jobs.
func DefaultJobs() []Job {
	return []Job{
		{
			Name: "daily_revenue_rollup",
			Cron: "0 1 * * *", // 1 AM daily
			Query: `
				INSERT INTO daily_revenue_mv
				SELECT
					organization_id,
					toDate(timestamp) AS date,
					count() AS order_count,
					sum(amount_paisa) AS total_revenue_paisa,
					avg(amount_paisa) AS avg_order_value
				FROM events
				WHERE event_type = 'order_completed'
				  AND toDate(timestamp) = yesterday()
				GROUP BY organization_id, date
			`,
		},
		{
			Name: "cleanup_old_events",
			Cron: "0 3 * * 0", // 3 AM Sunday
			Query: `
				ALTER TABLE events DELETE 
				WHERE event_date < today() - INTERVAL 730 DAY
			`,
		},
		{
			Name: "optimize_tables",
			Cron: "0 4 * * 0", // 4 AM Sunday
			Query: `
				OPTIMIZE TABLE events FINAL
			`,
		},
	}
}

// NewScheduler creates a new aggregation scheduler.
func NewScheduler(ch *clickhouse.Client) *Scheduler {
	return &Scheduler{
		cron:     cron.New(cron.WithSeconds()),
		ch:       ch,
		stopChan: make(chan struct{}),
	}
}

// Start begins the scheduler with default jobs.
func (s *Scheduler) Start() error {
	return s.StartWithJobs(DefaultJobs())
}

// StartWithJobs begins the scheduler with specified jobs.
func (s *Scheduler) StartWithJobs(jobs []Job) error {
	for _, job := range jobs {
		j := job // Capture for closure
		_, err := s.cron.AddFunc(j.Cron, func() {
			s.runJob(j)
		})
		if err != nil {
			logger.Error("Failed to schedule job", "job", j.Name, "error", err)
			continue
		}
		logger.Info("Scheduled aggregation job", "job", j.Name, "cron", j.Cron)
	}

	s.cron.Start()
	logger.Info("Aggregation scheduler started", "jobs", len(jobs))
	return nil
}

// Stop gracefully stops the scheduler.
func (s *Scheduler) Stop() {
	close(s.stopChan)
	ctx := s.cron.Stop()
	<-ctx.Done()
	logger.Info("Aggregation scheduler stopped")
}

// runJob executes an aggregation job.
func (s *Scheduler) runJob(job Job) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	logger.Info("Running aggregation job", "job", job.Name)
	start := time.Now()

	if err := s.ch.Exec(ctx, job.Query); err != nil {
		logger.Error("Aggregation job failed", "job", job.Name, "error", err, "duration", time.Since(start))
		return
	}

	logger.Info("Aggregation job completed", "job", job.Name, "duration", time.Since(start))
}

// RunNow immediately runs a specific job by name.
func (s *Scheduler) RunNow(jobName string) error {
	for _, job := range DefaultJobs() {
		if job.Name == jobName {
			s.runJob(job)
			return nil
		}
	}
	return nil
}
