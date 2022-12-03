package strolt

import (
	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client"
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
	params := &services.GetSnapshotsParams{
		TaskName:        taskName,
		ServiceName:     serviceName,
		DestinationName: destinationName,
	}

	return sdk.client.Services.GetSnapshots(params, sdk.authInfo)
}

func (sdk *Sdk) GetSnapshotsForPrune(serviceName, taskName, destinationName string) (*services.GetSnapshotsForPruneOK, error) {
	params := &services.GetSnapshotsForPruneParams{
		TaskName:        taskName,
		ServiceName:     serviceName,
		DestinationName: destinationName,
	}

	return sdk.client.Services.GetSnapshotsForPrune(params, sdk.authInfo)
}

func (sdk *Sdk) Prune(serviceName, taskName, destinationName string) (*services.PruneOK, error) {
	params := &services.PruneParams{
		TaskName:        taskName,
		ServiceName:     serviceName,
		DestinationName: destinationName,
	}

	return sdk.client.Services.Prune(params, sdk.authInfo)
}

func (sdk *Sdk) Ping() (*public.PingOK, error) {
	return sdk.client.Public.Ping(nil)
}
