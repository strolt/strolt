package stroltp

import (
	"github.com/go-openapi/runtime"
	runtimeClient "github.com/go-openapi/runtime/client"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/stroltp/generated/client"

	// "github.com/strolt/strolt/apps/stroltm/internal/sdk/stroltp/generated/client/info"
	// "github.com/strolt/strolt/apps/stroltm/internal/sdk/stroltp/generated/client/operations"
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/stroltp/generated/client/services"
)

type SDK struct {
	client   *client.StroltProxyAPI
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

func (sdk *SDK) GetList() (*services.GetListOK, error) {
	return sdk.client.Services.GetList(nil, sdk.authInfo)
}
