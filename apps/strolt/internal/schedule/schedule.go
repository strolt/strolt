// Package schedule runs backup and prune tasks on their cron schedules.
package schedule

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"
)

func start(processContext context.Context, serviceName string, taskName string, operation sctxt.OperationType) func() {
	return func() {
		log := logger.New().WithField("taskName", taskName).WithField("operation", operation)
		log.Info("started")

		if operation == sctxt.OpTypeBackup {
			backup(processContext, serviceName, taskName)
		}
		if operation == sctxt.OpTypePrune {
			prune(processContext, serviceName, taskName)
		}

		log.Info("finished")
	}
}

func runCron(processContext context.Context, cr *cron.Cron) {
	cr.Start()
	<-processContext.Done()
	<-cr.Stop().Done()
	logger.New().Debug("schedule manager stopped")
}

// Run starts the cron scheduler and waits for its active jobs to finish after cancellation.
func Run(processContext context.Context) {
	c := config.Get()
	cr := cron.New(cron.WithLocation(config.GetLocation()))

	for serviceName, service := range c.Services {
		for taskName, task := range service {
			log := logger.New().WithField("taskName", taskName)
			if task.Schedule.Backup != "" {
				if _, err := cr.AddFunc(task.Schedule.Backup, start(processContext, serviceName, taskName, sctxt.OpTypeBackup)); err != nil {
					log.Error(fmt.Errorf("error start schedule backup - %w", err))
				}
			}
			if task.Schedule.Prune != "" {
				if _, err := cr.AddFunc(task.Schedule.Prune, start(processContext, serviceName, taskName, sctxt.OpTypePrune)); err != nil {
					log.Error(fmt.Errorf("error start schedule prune - %w", err))
				}
			}
		}
	}

	runCron(processContext, cr)
}

func backup(processContext context.Context, serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)
	t, err := task.New(processContext, serviceName, taskName, sctxt.TSchedule, sctxt.OpTypeBackup)
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if err := t.Close(); err != nil {
			log.Error(err)
		}
	}()
	//nolint:contextcheck // Task retains the scheduled operation context through driver calls.
	if err := t.Backup(); err != nil {
		log.Error(err)
	}
}

func prune(processContext context.Context, serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)
	t, err := task.New(processContext, serviceName, taskName, sctxt.TSchedule, sctxt.OpTypePrune)
	if err != nil {
		log.Error(err)
		return
	}
	defer func() {
		if err := t.Close(); err != nil {
			log.Error(err)
		}
	}()
	//nolint:contextcheck // Task retains the scheduled operation context through driver calls.
	if err := t.PruneAll(); err != nil {
		log.Error(err)
	}
}
