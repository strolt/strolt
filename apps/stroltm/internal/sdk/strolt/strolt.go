package strolt

import (
	"github.com/strolt/strolt/apps/stroltm/internal/sdk/strolt/generated/client"
)

type Sdk struct {
	client *client.StroltAPI
}

func New(host string) *Sdk {
	cfg := client.DefaultTransportConfig().WithHost(host)
	client := client.NewHTTPClientWithConfig(nil, cfg)

	return &Sdk{
		client,
	}
}

// func (sdk *Sdk) GetConfig() (*operations.GetAPIConfigOK, error) {
// 	params := operations.NewGetAPIConfigParams()
// 	return sdk.client.Operations.GetAPIConfig(params)
// }

// func (sdk *Sdk) GetPrune(serviceName, taskName, destinationName string) (*services.GetAPIServicesServiceNameTasksTaskNameDestinationsDestinationNamePruneOK, error) {
// 	params := services.NewGetAPIServicesServiceNameTasksTaskNameDestinationsDestinationNamePruneParams()
// 	params.SetServiceName(serviceName)
// 	params.SetTaskName(taskName)
// 	params.SetDestinationName(destinationName)

// 	return sdk.client.Services.GetAPIServicesServiceNameTasksTaskNameDestinationsDestinationNamePrune(params)
// }
