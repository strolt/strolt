package schedule

import (
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
	"github.com/strolt/strolt/shared/logger"

	ctx "context"
)

func start(serviceName string, taskName string, operation sctxt.OperationType) func() {
	return func() {
		log := logger.New().WithField("taskName", taskName).WithField("operation", operation)

		log.Info("started")

		if operation == sctxt.OpTypeBackup {
			backup(serviceName, taskName)
		}

		if operation == sctxt.OpTypePrune {
			prune(serviceName, taskName)
		}

		log.Info("finished")
	}
}

func Run(ctx ctx.Context) {
	log := logger.New()
	c := config.Get()
	cr := cron.New(cron.WithLocation(config.GetLocation()))

	go func() {
		<-ctx.Done()

		log.Debug("stop schedule manager...")

		ctxCron := cr.Stop()
		<-ctxCron.Done()

		log.Debug("schedule manager stopped")
	}()

	for serviceName, service := range c.Services {
		for taskName, task := range service {
			taskName, task := taskName, task
			log := log.WithField("taskName", taskName)

			if task.Schedule.Backup != "" {
				if _, err := cr.AddFunc(task.Schedule.Backup, start(serviceName, taskName, sctxt.OpTypeBackup)); err != nil {
					log.Error(fmt.Errorf("error start schedule backup - %w", err))
				}
			}

			if task.Schedule.Prune != "" {
				if _, err := cr.AddFunc(task.Schedule.Prune, start(serviceName, taskName, sctxt.OpTypePrune)); err != nil {
					log.Error(fmt.Errorf("error start schedule prune - %w", err))
				}
			}
		}
	}

	cr.Run()
}

func backup(serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)
	t, err := task.New(serviceName, taskName, sctxt.TSchedule, sctxt.OpTypeBackup)

	if err != nil {
		log.Error(err)
	}

	defer t.Close()

	if err := t.Backup(); err != nil {
		log.Error(err)
	}
}

func prune(serviceName string, taskName string) {
	log := logger.New().WithField("serviceName", serviceName).WithField("taskName", taskName).WithField("trigger", sctxt.TSchedule)

	t, err := task.New(serviceName, taskName, sctxt.TSchedule, sctxt.OpTypePrune)
	if err != nil {
		log.Error(err)
	}

	defer t.Close()

	if err := t.PruneAll(); err != nil {
		log.Error(err)
	}
}
