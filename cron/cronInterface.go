package cron

import (
	"context"
	"vaxtrack/scheduler"
)

type CronImplementer interface {
	ScheduleJobs(ctx context.Context, schedules map[string]scheduler.Scheduler) error
}
