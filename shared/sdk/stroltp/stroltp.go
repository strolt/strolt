// Package stroltp provides an SDK and instance manager for the strolt proxy API.
package stroltp

import (
	"errors"
	"fmt"
	"net/url"

	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client"

	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client/info"
	managerc "github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client/manager"
)

// SDK is a client for the strolt proxy API.
type SDK struct {
	client   *stroltp_client.StroltProxyAPI
	authInfo runtime.ClientAuthInfoWriter
}

// New creates an SDK client for the given host using basic auth credentials.
//
// host may be a bare "host:port" (defaults to the http scheme, preserving
// backwards compatibility) or a full URL such as "https://host:port" to enable
// TLS. Without this, basic-auth credentials would always be sent in cleartext.
func New(host, username, password string) *SDK {
	cfg := stroltp_client.DefaultTransportConfig()

	if u, err := url.Parse(host); err == nil && u.Host != "" && (u.Scheme == "http" || u.Scheme == "https") {
		cfg = cfg.WithHost(u.Host).WithSchemes([]string{u.Scheme})
	} else {
		cfg = cfg.WithHost(host)
	}

	c := stroltp_client.NewHTTPClientWithConfig(nil, cfg)

	return &SDK{
		client:   c,
		authInfo: runtimeClient.BasicAuth(username, password),
	}
}

// GetInfo returns information about the strolt proxy instance.
func (sdk *SDK) GetInfo() (*info.GetInfoOK, error) {
	result, err := sdk.client.Info.GetInfo(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("get info: %w", err)
	}

	return result, nil
}

// GetInstances returns the list of strolt instances managed by the proxy.
func (sdk *SDK) GetInstances() (*managerc.GetInstancesOK, error) {
	result, err := sdk.client.Manager.GetInstances(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("get instances: %w", err)
	}

	return result, nil
}

// Backup starts a backup for the given instance, service and task.
func (sdk *SDK) Backup(instanceName, serviceName, taskName string) (*managerc.BackupOK, error) {
	params := managerc.NewBackupParams()
	params.InstanceName = instanceName
	params.ServiceName = serviceName
	params.TaskName = taskName

	result, err := sdk.client.Manager.Backup(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *managerc.BackupInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("backup: %w", err)
	}

	return result, nil
}

// BackupAll starts a backup for all instances managed by the proxy.
func (sdk *SDK) BackupAll() (*managerc.BackupAllOK, error) {
	result, err := sdk.client.Manager.BackupAll(nil, sdk.authInfo)
	if err != nil {
		return nil, fmt.Errorf("backup all: %w", err)
	}

	return result, nil
}

// GetSnapshots returns snapshots for the given instance, service, task and destination.
func (sdk *SDK) GetSnapshots(instanceName, serviceName, taskName, destinationName string) (*managerc.GetSnapshotsOK, error) {
	params := managerc.NewGetSnapshotsParams()
	params.InstanceName = instanceName
	params.ServiceName = serviceName
	params.TaskName = taskName
	params.DestinationName = destinationName

	result, err := sdk.client.Manager.GetSnapshots(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *managerc.BackupInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get snapshots: %w", err)
	}

	return result, nil
}

// GetStats returns statistics for the given instance, service, task and destination.
func (sdk *SDK) GetStats(instanceName, serviceName, taskName, destinationName string) (*managerc.GetStatsOK, error) {
	params := managerc.NewGetStatsParams()
	params.InstanceName = instanceName
	params.ServiceName = serviceName
	params.TaskName = taskName
	params.DestinationName = destinationName

	result, err := sdk.client.Manager.GetStats(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *managerc.GetStatsInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get stats: %w", err)
	}

	return result, nil
}

// GetSnapshotsForPrune returns the snapshots that would be removed by a prune operation.
func (sdk *SDK) GetSnapshotsForPrune(instanceName, serviceName, taskName, destinationName string) (*managerc.GetSnapshotsForPruneOK, error) {
	params := managerc.NewGetSnapshotsForPruneParams()
	params.InstanceName = instanceName
	params.ServiceName = serviceName
	params.TaskName = taskName
	params.DestinationName = destinationName

	result, err := sdk.client.Manager.GetSnapshotsForPrune(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *managerc.GetSnapshotsForPruneInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("get snapshots for prune: %w", err)
	}

	return result, nil
}

// Prune removes outdated snapshots for the given instance, service, task and destination.
func (sdk *SDK) Prune(instanceName, serviceName, taskName, destinationName string) (*managerc.PruneOK, error) {
	params := managerc.NewPruneParams()
	params.InstanceName = instanceName
	params.ServiceName = serviceName
	params.TaskName = taskName
	params.DestinationName = destinationName

	result, err := sdk.client.Manager.Prune(params, sdk.authInfo)
	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *managerc.PruneInternalServerError:
			return result, errors.New(errResponse.Payload.Error)
		}

		return result, fmt.Errorf("prune: %w", err)
	}

	return result, nil
}
