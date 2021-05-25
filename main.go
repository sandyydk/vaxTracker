package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vaxtrack/config"
	"vaxtrack/cron"
	"vaxtrack/scheduler"
)

var (
	configFile string
)

func main() {

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conf := config.Conf
	cronManager := cron.New(*conf)

	err := cronManager.ScheduleJobs(ctx, scheduler.SchedulerMap)
	if err != nil {
		log.Fatalf("failed to schedule jobs: %#v", err)
		return
	}

	setupSignalCatcher(cancel)
	<-ctx.Done()
	// give some time for exiting go routines
	time.Sleep(2 * time.Second)

}

func setupSignalCatcher(cancel context.CancelFunc) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGINT)
	go func(cancel context.CancelFunc) {
		<-c
		log.Println("Interrupt signal received. Stopping...")
		cancel()
	}(cancel)
}
