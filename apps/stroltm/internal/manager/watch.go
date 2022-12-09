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
	s.Lock()
	if s.Watch == nil {
		s.Watch = &WatchItem{}
	}
	s.Unlock()

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

func (s *Strolt) ping() {
	pingAt := time.Now()
	isError := false

	s.Lock()
	s.Watch.IsPingInProcess = true
	s.Watch.LatestPingAt = pingAt
	s.Unlock()

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
	}

	s.Lock()
	if isError {
		s.IsOnline = false
	} else {
		s.Watch.LatestSuccessPingAt = pingAt
		s.IsOnline = true
	}

	s.Watch.IsPingInProcess = false
	s.Unlock()
}

func (m *Manager) pingInstances() {
	for _, instance := range m.Strolt {
		isPingAllowed := instance.isPingAllowed()

		if isPingAllowed {
			go instance.ping()
		}
	}
}
