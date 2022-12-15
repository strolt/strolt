package strolt

import (
	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/info"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/operations"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/public"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client/services"
)

type Sdk struct {
	client   *client.StroltAPI
	authInfo runtime.ClientAuthInfoWriter
}

func New(host, username, password string) *Sdk {
	cfg := client.DefaultTransportConfig().WithHost(host)
	c := client.NewHTTPClientWithConfig(nil, cfg)

	return &Sdk{
		client:   c,
		authInfo: runtimeClient.BasicAuth(username, password),
	}
}

func (sdk *Sdk) GetConfig() (*operations.GetConfigOK, error) {
	return sdk.client.Operations.GetConfig(nil, sdk.authInfo)
}

func (sdk *Sdk) GetSnapshots(serviceName, taskName, destinationName string) (*services.GetSnapshotsOK, error) {
	params := services.NewGetSnapshotsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	return sdk.client.Services.GetSnapshots(params, sdk.authInfo)
}

func (sdk *Sdk) Backup(serviceName, taskName string) (*services.BackupOK, error) {
	params := services.NewBackupParams()
	params.ServiceName = serviceName
	params.TaskName = taskName

	return sdk.client.Services.Backup(params, sdk.authInfo)
}

func (sdk *Sdk) GetSnapshotsForPrune(serviceName, taskName, destinationName string) (*services.GetSnapshotsForPruneOK, error) {
	params := services.NewGetSnapshotsForPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	return sdk.client.Services.GetSnapshotsForPrune(params, sdk.authInfo)
}

func (sdk *Sdk) Prune(serviceName, taskName, destinationName string) (*services.PruneOK, error) {
	params := services.NewPruneParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	return sdk.client.Services.Prune(params, sdk.authInfo)
}

func (sdk *Sdk) Ping() (*public.PingOK, error) {
	return sdk.client.Public.Ping(nil)
}

func (sdk *Sdk) GetMetrics() (*operations.GetStroltMetricsOK, error) {
	return sdk.client.Operations.GetStroltMetrics(nil, sdk.authInfo)
}

func (sdk *Sdk) GetInfo() (*info.GetInfoOK, error) {
	return sdk.client.Info.GetInfo(nil, sdk.authInfo)
}

func (sdk *Sdk) GetStats(serviceName, taskName, destinationName string) (*services.GetStatsOK, error) {
	params := services.NewGetStatsParams()
	params.TaskName = taskName
	params.ServiceName = serviceName
	params.DestinationName = destinationName

	return sdk.client.Services.GetStats(params, sdk.authInfo)
}
