// Package strolt provides an SDK and instance manager for the strolt API.
package strolt

import (
	"errors"
	"fmt"

	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_client"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_client/info"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_client/operations"
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_client/services"
)

// SDK is a client for the strolt API.
type SDK struct {
	client   *strolt_client.StroltAPI
	authInfo runtime.ClientAuthInfoWriter
}

// New creates an SDK client for the given host with basic auth credentials.
func New(host, username, password string) *SDK {
	cfg := strolt_client.DefaultTransportConfig().WithHost(host)
	c := strolt_client.NewHTTPClientWithConfig(nil, cfg)

	return &SDK{
		client:   c,
		authInfo: runtimeClient.BasicAuth(username, password),
	}
}

// GetConfig fetches the instance configuration.
func (sdk *SDK) GetConfig() (*operations.GetConfigOK, error) {
	result, err := sdk.client.Operations.GetConfig(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("get config: %w", err)
	}

	return result, nil
}

// GetSnapshots fetches snapshots for the given service, task and destination.
func (sdk *SDK) GetSnapshots(serviceName, taskName, destinationName string) (*services.GetSnapshotsOK, error) {
	params := services.NewGetSnapshotsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetSnapshots(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:errorlint
		case *services.GetSnapshotsBadRequest:
			return result, errors.New(errResponse.Payload.Error)
		case *services.GetSnapshotsForPruneInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get snapshots: %w", err)
	}

	return result, nil
}

// Backup starts a backup for the given service and task.
func (sdk *SDK) Backup(serviceName, taskName string) (*services.BackupOK, error) {
	params := services.NewBackupParams()
	params.ServiceName = serviceName
	params.TaskName = taskName

	result, err := sdk.client.Services.Backup(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.BackupInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("backup: %w", err)
	}

	return result, nil
}

// GetSnapshotsForPrune fetches snapshots that would be removed by a prune.
func (sdk *SDK) GetSnapshotsForPrune(serviceName, taskName, destinationName string) (*services.GetSnapshotsForPruneOK, error) {
	params := services.NewGetSnapshotsForPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetSnapshotsForPrune(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.GetSnapshotsForPruneInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get snapshots for prune: %w", err)
	}

	return result, nil
}

// Prune removes outdated snapshots for the given service, task and destination.
func (sdk *SDK) Prune(serviceName, taskName, destinationName string) (*services.PruneOK, error) {
	params := services.NewPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.Prune(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.PruneInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("prune: %w", err)
	}

	return result, nil
}

// GetInfo fetches general information about the instance.
func (sdk *SDK) GetInfo() (*info.GetInfoOK, error) {
	result, err := sdk.client.Info.GetInfo(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("get info: %w", err)
	}

	return result, nil
}

// GetStats fetches statistics for the given service, task and destination.
func (sdk *SDK) GetStats(serviceName, taskName, destinationName string) (*services.GetStatsOK, error) {
	params := services.NewGetStatsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetStats(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.GetStatsInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get stats: %w", err)
	}

	return result, nil
}

// GetStatus fetches the task manager status of the instance.
func (sdk *SDK) GetStatus() (*services.GetStatusOK, error) {
	result, err := sdk.client.Services.GetStatus(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("get status: %w", err)
	}

	return result, nil
}
