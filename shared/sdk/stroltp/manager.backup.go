package stroltp

import (
	"sync"
)

type BackupAllStatusItem struct {
	ProxyName    string `json:"proxyName,omitempty"`
	InstanceName string `json:"instanceName,omitempty"`
	ServiceName  string `json:"serviceName,omitempty"`
	TaskName     string `json:"taskName,omitempty"`
}

type ManagerhBackupAllResponse struct {
	ErrorStarted   []*BackupAllStatusItem `json:"errorStarted"`
	SuccessStarted []*BackupAllStatusItem `json:"successStarted"`
	*sync.Mutex
}

func ManagerBackupAll() *ManagerhBackupAllResponse {
	status := ManagerhBackupAllResponse{
		ErrorStarted:   []*BackupAllStatusItem{},
		SuccessStarted: []*BackupAllStatusItem{},
		Mutex:          &sync.Mutex{},
	}

	wg := sync.WaitGroup{}

	if manager != nil {
		for _, instance := range manager.Instances {
			wg.Add(1)

			go func(instance *Instance) {
				instance.RLock()

				result, err := instance.sdk.BackupAll()
				if err != nil {
					if len(instance.StroltInstances) > 0 {
						errorList := []*BackupAllStatusItem{}

						for _, stroltInstance := range instance.StroltInstances {
							if stroltInstance.Config == nil {
								continue
							}

							for serviceName, service := range stroltInstance.Config.Services {
								for taskName := range service {
									statusItem := BackupAllStatusItem{
										ProxyName:    instance.Name,
										InstanceName: stroltInstance.Name,
										ServiceName:  serviceName,
										TaskName:     taskName,
									}
									errorList = append(errorList, &statusItem)
								}
							}
						}

						status.Lock()
						status.ErrorStarted = append(status.ErrorStarted, errorList...)
						status.Unlock()
					}
				} else {
					status.Lock()

					for _, errorStarted := range result.Payload.ErrorStarted {
						item := BackupAllStatusItem{
							ProxyName:    instance.Name,
							InstanceName: errorStarted.InstanceName,
							ServiceName:  errorStarted.ServiceName,
							TaskName:     errorStarted.TaskName,
						}
						status.ErrorStarted = append(status.ErrorStarted, &item)
					}

					for _, successStarted := range result.Payload.SuccessStarted {
						item := BackupAllStatusItem{
							ProxyName:    instance.Name,
							InstanceName: successStarted.InstanceName,
							ServiceName:  successStarted.ServiceName,
							TaskName:     successStarted.TaskName,
						}
						status.SuccessStarted = append(status.SuccessStarted, &item)
					}

					status.Unlock()
				}

				instance.RUnlock()
				wg.Done()
			}(instance)
		}
	}

	wg.Wait()

	return &status
}
