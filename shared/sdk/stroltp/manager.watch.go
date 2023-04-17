package stroltp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/sdk/common"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_models"
)

func (m *Manager) Watch(ctx context.Context, cancel func()) {
	log := logger.New()
	ticker := time.NewTicker(5 * time.Second) //nolint:gomnd
	quit := make(chan struct{})

	go func() {
		<-ctx.Done()

		log.Debug("stop stroltp manager...")

		close(quit)

		log.Warn("stroltp manager stopped")
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

		if s.Info == nil || s.Info.UpdatedAt == "" {
			go s.updateStroltInstances(time.Now().Unix())
		} else {
			rInfoUpdatedAt, err := time.Parse(time.RFC3339, result.Payload.UpdatedAt)

			if err != nil {
				s.log.Errorf("parse time result.Payload.UpdatedAt %s", result.Payload.UpdatedAt)
			} else if rInfoUpdatedAt.Unix() > s.StroltInstancesUpdatedAt {
				go s.updateStroltInstances(rInfoUpdatedAt.Unix())
			}
		}

		s.Info = result.Payload
		s.Unlock()
	}
}

func convertInstances(src []*stroltp_models.ManagerPreparedInstance) []*common.ManagerPreparedInstance {
	data, err := json.Marshal(src)
	if err != nil {
		return []*common.ManagerPreparedInstance{}
	}

	var list []*common.ManagerPreparedInstance

	if err := json.Unmarshal(data, &list); err != nil {
		return []*common.ManagerPreparedInstance{}
	}

	return list
}

func (s *Instance) updateStroltInstances(infoTime int64) {
	result, err := s.sdk.GetInstances()
	if err != nil {
		s.log.Debugf("error get config %s", err)
		return
	}

	convertedInstanceList := convertInstances(result.Payload)

	s.Lock()

	s.StroltInstances = convertedInstanceList

	s.StroltInstancesUpdatedAt = infoTime

	s.Unlock()
}
