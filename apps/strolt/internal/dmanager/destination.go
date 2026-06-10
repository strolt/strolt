// Package dmanager resolves source, destination, and notification drivers.
package dmanager

import (
	"fmt"
	"slices"

	"github.com/strolt/strolt/apps/strolt/internal/driver/destination/local"
	"github.com/strolt/strolt/apps/strolt/internal/driver/destination/restic"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/logger"
)

// Destination is the name of a destination driver.
type Destination string

// Supported destination drivers.
const (
	DriverDestinationLocal  Destination = "local"
	DriverDestinationRestic Destination = "restic"
)

// GetAvailableDriverDestination returns the supported destination drivers.
func GetAvailableDriverDestination() []Destination {
	return []Destination{
		DriverDestinationLocal,
		DriverDestinationRestic,
	}
}

// IsAvailableDriverDestination reports whether the destination driver is supported.
func IsAvailableDriverDestination(driver Destination) bool {
	return slices.Contains(GetAvailableDriverDestination(), driver)
}

// GetDestinationDriver creates and configures the requested destination driver.
func GetDestinationDriver(destinationName string, driver Destination, serviceName string, taskName string, driverConfig any, driverEnv any) (interfaces.DriverDestinationInterface, error) {
	destinationDrivers := map[Destination]interfaces.DriverDestinationInterface{
		DriverDestinationLocal:  local.New(),
		DriverDestinationRestic: restic.New(),
	}

	d, ok := destinationDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("destination driver '%s' does not exists", driver)
	}

	d.SetTaskName(taskName)
	d.SetDriverName(destinationName)

	loggerFields := logger.Fields{
		"driverType":  "destination",
		"serviceName": serviceName,
		"taskName":    taskName,
		"driver":      driver,
	}
	d.SetLogger(logger.New().WithFields(loggerFields))

	if err := d.SetConfig(driverConfig); err != nil {
		return nil, fmt.Errorf("set destination driver config: %w", err)
	}

	if err := d.SetEnv(driverEnv); err != nil {
		return nil, fmt.Errorf("set destination driver env: %w", err)
	}

	return d, nil
}
