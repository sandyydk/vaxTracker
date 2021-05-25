package cron

import (
	"context"
	"fmt"
	"vaxtrack/config"
	"vaxtrack/notifier"
	"vaxtrack/scheduler"

	robfig "github.com/robfig/cron/v3"
)

type CronManager struct {
	r                   *robfig.Cron
	cfg                 config.Config
	notificationManager *notifier.NotificationManager
}

func New(cfg config.Config) CronImplementer {

	var cManager CronManager
	cManager.cfg = cfg
	cManager.notificationManager = notifier.NewManager()
	cManager.r = robfig.New(robfig.WithChain(robfig.SkipIfStillRunning(robfig.DefaultLogger), robfig.Recover(robfig.DefaultLogger)))
	return &cManager
}

func (c *CronManager) notify(appointments []scheduler.Appointment) {
	for _, appointment := range appointments {
		c.notificationManager.HandleNotifications(appointment.String())
	}
}

func (c *CronManager) wrap(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.r.Stop()
			return
		}
	}
}

func (c *CronManager) ScheduleJobs(ctx context.Context, schedules map[string]scheduler.Scheduler) error {

	go c.wrap(ctx)

	go c.notificationManager.Start(ctx)

	if len(schedules) == 0 {
		return fmt.Errorf("no schedules found")
	}

	fmt.Printf("Schedules: %#v\n", schedules)

	for schedule, scheduler := range schedules {
		if schedule != "" || len(schedule) > 0 {
			// TODO :- Check if the schedule is a valid one either using regex or some other way
			_, err := c.r.AddFunc(schedule, func() {
				fmt.Println("Adding Schedule method")
				err := scheduler.Initialize()
				if err != nil {
					fmt.Printf("Failed to initialize scheduler: %#v\n", err)
					return
				}
				appointments, err := scheduler.GetAppointments(ctx)
				if err != nil {
					fmt.Printf("Failed to obtain appointments: %#v\n", err)
					return
				}

				if appointments != nil && len(appointments) > 0 {
					c.notify(appointments)
				}

			})

			if err != nil {
				// Skip the job if configured wrongly
				fmt.Printf("Error parsing the cron notation: %#v\n", err)
				continue
			}

		} else {
			_, err := c.r.AddFunc(c.cfg.DefaultCron, func() {
				fmt.Println("Calling GetAppointments")
				err := scheduler.Initialize()
				if err != nil {
					fmt.Printf("Failed to initialize scheduler: %#v\n", err)
					return
				}
				appointments, err := scheduler.GetAppointments(ctx)
				if err != nil {
					fmt.Printf("Failed to obtain appointments: %#v\n", err)
					return
				}

				fmt.Printf("Appointments found: %#v\n", len(appointments))

				if appointments != nil && len(appointments) > 0 {
					fmt.Printf("Notifying\n")
					c.notify(appointments)
				}
			})

			if err != nil {
				// Skip the job if configured wrongly
				fmt.Printf("Error parsing the default cron notation: %#v\n", err)
				continue
			}
		}
	}

	c.r.Start()

	return nil
}
