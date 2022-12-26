package strolt

import (
	"fmt"

	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/info"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/operations"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/services"
)

type SDK struct {
	client   *client.StroltAPI
	authInfo runtime.ClientAuthInfoWriter
}

func New(host, username, password string) *SDK {
	cfg := client.DefaultTransportConfig().WithHost(host)
	c := client.NewHTTPClientWithConfig(nil, cfg)

	return &SDK{
		client:   c,
		authInfo: runtimeClient.BasicAuth(username, password),
	}
}

func (sdk *SDK) GetConfig() (*operations.GetConfigOK, error) {
	return sdk.client.Operations.GetConfig(nil, sdk.authInfo)
}

func (sdk *SDK) GetSnapshots(serviceName, taskName, destinationName string) (*services.GetSnapshotsOK, error) {
	params := services.NewGetSnapshotsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetSnapshots(params, sdk.authInfo)

	if err != nil {
		switch errResponse := err.(type) { //nolint:errorlint
		case *services.GetSnapshotsBadRequest:
			return result, fmt.Errorf(errResponse.Payload.Error)
		case *services.GetSnapshotsForPruneInternalServerError:
			return result, fmt.Errorf(errResponse.Payload.Error)
		}
	}

	return result, err
}

func (sdk *SDK) Backup(serviceName, taskName string) (*services.BackupOK, error) {
	params := services.NewBackupParams()
	params.ServiceName = serviceName
	params.TaskName = taskName

	result, err := sdk.client.Services.Backup(params, sdk.authInfo)

	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.BackupInternalServerError:
			return result, fmt.Errorf(errResponse.Payload.Error)
		}
	}

	return result, err
}

func (sdk *SDK) GetSnapshotsForPrune(serviceName, taskName, destinationName string) (*services.GetSnapshotsForPruneOK, error) {
	params := services.NewGetSnapshotsForPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetSnapshotsForPrune(params, sdk.authInfo)

	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.GetSnapshotsForPruneInternalServerError:
			return result, fmt.Errorf(errResponse.Payload.Error)
		}
	}

	return result, err
}

func (sdk *SDK) Prune(serviceName, taskName, destinationName string) (*services.PruneOK, error) {
	params := services.NewPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.Prune(params, sdk.authInfo)

	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.PruneInternalServerError:
			return result, fmt.Errorf(errResponse.Payload.Error)
		}
	}

	return result, err
}

func (sdk *SDK) GetMetrics() (*operations.GetStroltMetricsOK, error) {
	return sdk.client.Operations.GetStroltMetrics(nil, sdk.authInfo)
}

func (sdk *SDK) GetInfo() (*info.GetInfoOK, error) {
	return sdk.client.Info.GetInfo(nil, sdk.authInfo)
}

func (sdk *SDK) GetStats(serviceName, taskName, destinationName string) (*services.GetStatsOK, error) {
	params := services.NewGetStatsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	result, err := sdk.client.Services.GetStats(params, sdk.authInfo)

	if err != nil {
		switch errResponse := err.(type) { //nolint:gocritic,errorlint
		case *services.GetStatsInternalServerError:
			return result, fmt.Errorf(errResponse.Payload.Error)
		}
	}

	return result, err
}

func (sdk *SDK) GetStatus() (*services.GetStatusOK, error) {
	return sdk.client.Services.GetStatus(nil, sdk.authInfo)
}
