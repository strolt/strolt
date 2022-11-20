package dmanager

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/driver/destination/local"
	"github.com/strolt/strolt/apps/strolt/internal/driver/destination/restic"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/logger"
)

type Destination string

const (
	DriverDestinationLocal  Destination = "local"
	DriverDestinationRestic Destination = "restic"
)

func GetAvailableDriverDestination() []Destination {
	return []Destination{
		DriverDestinationLocal,
		DriverDestinationRestic,
	}
}

func IsAvailableDriverDestination(driver Destination) bool {
	for _, d := range GetAvailableDriverDestination() {
		if driver == d {
			return true
		}
	}

	return false
}

func GetDestinationDriver(destinationName string, driver Destination, serviceName string, taskName string, driverConfig interface{}, driverEnv interface{}) (interfaces.DriverDestinationInterface, error) {
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
		return nil, err
	}

	if err := d.SetEnv(driverEnv); err != nil {
		return nil, err
	}

	return d, nil
}
