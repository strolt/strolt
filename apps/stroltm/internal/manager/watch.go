package manager

import (
	"context"
	"time"

	"github.com/strolt/strolt/shared/logger"
)

func (m *Manager) Watch(ctx context.Context, cancel func()) {
	log := logger.New()
	ticker := time.NewTicker(5 * time.Second) //nolint:gomnd
	quit := make(chan struct{})

	go func() {
		<-ctx.Done()

		log.Debug("stop manager...")

		close(quit)

		log.Warn("manager stopped")
	}()

	go m.pingInstances()

	for {
		select {
		case <-ticker.C:
			go m.pingInstances()
		case <-quit:
			ticker.Stop()
			return
		}
	}
}

func (s *Instance) isPingAllowed() bool {
	s.RLock()
	defer s.RUnlock()

	return !s.Watch.IsPingInProcess
}

func (m *Manager) pingInstances() {
	for _, instance := range m.Instances {
		go func(instance *Instance) {
			isPingAllowed := instance.isPingAllowed()

			if isPingAllowed {
				instance.ping()
			}
		}(instance)
	}
}

func (s *Instance) updateConfig() {
	requestedAt := time.Now()

	s.Lock()
	s.Config.UpdateRequestedAt = requestedAt
	s.Unlock()

	result, err := s.sdk.GetConfig()
	if err != nil {
		s.log.Debugf("error get config %s", err)
		return
	}

	s.Lock()

	if s.Config.UpdateRequestedAt.Equal(requestedAt) {
		s.Config.Data = result.Payload
	}

	s.Config.IsInitialized = true
	s.Config.UpdatedAt = time.Now()
	s.Unlock()
}

func (s *Instance) updateTaskStatus() {
	requestedAt := time.Now()

	s.Lock()
	s.TaskStatus.UpdateRequestedAt = requestedAt
	s.Unlock()

	result, err := s.sdk.GetStatus()
	if err != nil || result.Payload == nil {
		s.log.Debugf("error get task status %s", err)
		return
	}

	s.Lock()

	if s.TaskStatus.UpdateRequestedAt.Equal(requestedAt) {
		s.TaskStatus.Data = result.Payload.Data
	}

	s.TaskStatus.IsInitialized = true
	s.TaskStatus.UpdatedAt = time.Now()
	s.Unlock()
}

func (s *Instance) ping() {
	pingAt := time.Now()

	s.Lock()
	s.Watch.IsPingInProcess = true
	s.Unlock()

	result, err := s.sdk.GetInfo()
	if err != nil {
		s.log.Debugf("ping error: %s", err)

		s.Lock()
		s.Watch.IsPingInProcess = false

		if s.IsOnline {
			s.Watch.LatestSuccessPingAt = s.Watch.LatestPingAt
		}

		s.Watch.LatestPingAt = pingAt
		s.IsOnline = false
		s.Unlock()
	}

	if err == nil {
		s.Lock()
		{
			s.Watch.IsPingInProcess = false
			s.IsOnline = true
			s.Watch.LatestPingAt = pingAt
			s.Watch.LatestSuccessPingAt = pingAt
		}

		// check and update instance config
		if !s.Config.IsInitialized {
			go s.updateConfig()
		} else {
			infoConfigUpdatedAt, err := time.Parse(time.RFC3339, s.Info.ConfigUpdatedAt)
			if err != nil {
				s.log.Errorf("parse time s.Info.ConfigUpdatedAt %s", s.Info.ConfigUpdatedAt)
			} else {
				rConfigUpdatedAt, err := time.Parse(time.RFC3339, result.Payload.ConfigUpdatedAt)
				if err != nil {
					s.log.Errorf("parse time result.Payload.ConfigUpdatedAt %s", result.Payload.ConfigUpdatedAt)
				} else if rConfigUpdatedAt.Unix() > infoConfigUpdatedAt.Unix() {
					go s.updateConfig()
				}
			}
		}

		// check and update instance task status
		if !s.TaskStatus.IsInitialized {
			go s.updateTaskStatus()
		} else {
			infoTaskStatusUpdatedAt, err := time.Parse(time.RFC3339, s.Info.TaskStatusUpdatedAt)
			if err != nil {
				s.log.Errorf("parse time s.Info.TaskStatusUpdatedAt %s", s.Info.TaskStatusUpdatedAt)
			} else {
				rTaskStatusUpdatedAt, err := time.Parse(time.RFC3339, result.Payload.TaskStatusUpdatedAt)
				if err != nil {
					s.log.Errorf("parse time result.Payload.TaskStatusUpdatedAt %s", result.Payload.TaskStatusUpdatedAt)
				} else if rTaskStatusUpdatedAt.Unix() > infoTaskStatusUpdatedAt.Unix() {
					go s.updateTaskStatus()
				}
			}
		}

		s.Info = result.Payload
		s.Unlock()
	}
}
