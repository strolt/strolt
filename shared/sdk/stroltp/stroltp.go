package stroltp

import (
	"errors"

	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client"

	"github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client/info"
	managerc "github.com/strolt/strolt/shared/sdk/stroltp/generated/stroltp_client/manager"
)

type SDK struct {
	client   *stroltp_client.StroltProxyAPI
	authInfo runtime.ClientAuthInfoWriter
}

func New(host, username, password string) *SDK {
	cfg := stroltp_client.DefaultTransportConfig().WithHost(host)
	c := stroltp_client.NewHTTPClientWithConfig(nil, cfg)

	return &SDK{
		client:   c,
		authInfo: runtimeClient.BasicAuth(username, password),
	}
}

func (sdk *SDK) GetInfo() (*info.GetInfoOK, error) {
	return sdk.client.Info.GetInfo(nil, sdk.authInfo)
}

func (sdk *SDK) GetInstances() (*managerc.GetInstancesOK, error) {
	return sdk.client.Manager.GetInstances(nil, sdk.authInfo)
}

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
	}

	return result, err
}

func (sdk *SDK) BackupAll() (*managerc.BackupAllOK, error) {
	return sdk.client.Manager.BackupAll(nil, sdk.authInfo)
}

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
	}

	return result, err
}

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
	}

	return result, err
}

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
	}

	return result, err
}

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
	}

	return result, err
}
