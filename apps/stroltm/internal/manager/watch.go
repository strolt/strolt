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

func (s *Strolt) isPingAllowed() bool {
	return !s.Watch.IsPingInProcess
}

func (s *Strolt) updateConfig() {
	result, err := s.sdk.GetConfig()
	if err != nil {
		s.log.Debugf("error get config %s", err)
		return
	}

	s.Lock()
	s.Config = result.Payload
	s.Unlock()
}

func (s *Strolt) updateInfo() {
	result, err := s.sdk.GetInfo()
	if err != nil || result.Payload == nil {
		return
	}

	s.Lock()
	s.Info.Version = result.Payload.Version
	s.Unlock()
}

func (s *Strolt) updateStatus() {
	result, err := s.sdk.GetStatus()
	if err != nil || result.Payload == nil {
		return
	}

	s.Lock()
	s.Status = result.Payload.Data
	s.Unlock()
}

func (s *Strolt) setIsOnline(isOnline bool) {
	if !s.IsOnline && isOnline {
		s.updateInfo()
	}

	s.Lock()
	s.IsOnline = isOnline
	s.Unlock()
}

func (s *Strolt) setLatestSuccessPingAt(at time.Time) {
	s.Lock()
	s.Watch.LatestSuccessPingAt = at
	s.Unlock()
}

func (s *Strolt) setIsPingInProcess(isProcess bool) {
	s.Lock()
	s.Watch.IsPingInProcess = isProcess
	s.Unlock()
}

func (s *Strolt) setLatestPingAt(at time.Time) {
	s.Lock()
	s.Watch.LatestPingAt = at
	s.Unlock()
}

func (s *Strolt) ping() {
	pingAt := time.Now()
	isError := false

	s.setIsPingInProcess(true)
	s.setLatestPingAt(pingAt)

	result, err := s.sdk.Ping()
	if err != nil {
		s.log.Debugf("ping error: %s", err)

		isError = true
	} else {
		rConfigLoadedAt, err := time.Parse(time.RFC3339, result.Payload.ConfigLoadedAt)
		if err != nil {
			s.log.Debugf("ping error: %s", err)
			isError = true
		} else if !s.ConfigLoadedAt.Equal(rConfigLoadedAt) {
			s.Lock()
			s.ConfigLoadedAt = rConfigLoadedAt
			s.Unlock()

			go s.updateConfig()
		}

		{
			timeFromRequest, err := time.Parse(time.RFC3339, result.Payload.TaskManagerUpdatedAt)
			if err == nil && timeFromRequest.Unix() > s.Watch.LatestSuccessUpdateStatusAt.Unix() {
				s.Watch.LatestSuccessUpdateStatusAt = timeFromRequest
				s.updateStatus()
			}
		}
	}

	if isError {
		s.setIsOnline(false)
	} else {
		s.setLatestSuccessPingAt(pingAt)
		s.setIsOnline(true)
	}

	s.setIsPingInProcess(false)
}

func (m *Manager) pingInstances() {
	for _, instance := range m.Strolt {
		isPingAllowed := instance.isPingAllowed()

		if isPingAllowed {
			go instance.ping()
		}
	}
}
