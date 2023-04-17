package common

import (
	"github.com/strolt/strolt/shared/sdk/strolt/generated/strolt_models"
)

type ManagerPreparedInstance struct {
	ProxyName  *string                      `json:"proxyName,omitempty"`
	Name       string                       `json:"name"`
	Config     *strolt_models.Config        `json:"config"`
	TaskStatus *strolt_models.ManagerStatus `json:"taskStatus"`
	IsOnline   bool                         `json:"isOnline"`
} // @name ManagerPreparedInstance
