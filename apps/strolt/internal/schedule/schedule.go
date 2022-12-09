package schedule

import (
	"fmt"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"

	"github.com/robfig/cron/v3"

	ctx "context"
)

//nolint:revive
type ScheduleManager struct {
	sync.Mutex
	m map[string]bool
}

var scheduleManager = ScheduleManager{
	m: make(map[string]bool),
}

func start(serviceName string, taskName string, operation sctxt.OperationType) {
	log := logger.New().WithField("taskName", taskName).WithField("operation", operation)
	if scheduleManager.m[taskName] {
		log.Warn("skip start operation for task")
		return
	}

	log.Info("started")
	scheduleManager.Lock()
	scheduleManager.m[taskName] = true
	scheduleManager.Unlock()

	defer func() {
		scheduleManager.Lock()
		scheduleManager.m[taskName] = false
		scheduleManager.Unlock()
	}()

	if operation == sctxt.OpTypeBackup {
		backup(serviceName, taskName)
	}

	if operation == sctxt.OpTypePrune {
		prune(serviceName, taskName)
	}

	log.Info("finished")
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

			scheduleManager.Lock()
			scheduleManager.m[taskName] = false
			scheduleManager.Unlock()

			if task.Schedule.Backup != "" {
				if _, err := cr.AddFunc(task.Schedule.Backup, func() { start(serviceName, taskName, sctxt.OpTypeBackup) }); err != nil {
					log.Error(fmt.Errorf("error start schedule backup - %w", err))
				}
			}

			if task.Schedule.Prune != "" {
				if _, err := cr.AddFunc(task.Schedule.Prune, func() { start(serviceName, taskName, sctxt.OpTypePrune) }); err != nil {
					log.Error(fmt.Errorf("error start schedule prune - %w", err))
				}
			}
		}
	}

	cr.Run()
}
