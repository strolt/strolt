package dmanager

import (
	"fmt"
	"slices"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/local"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/mongodb"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/mysql"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/pg"
	"github.com/strolt/strolt/shared/logger"
)

// Source is the name of a source driver.
type Source string

// Supported source drivers.
const (
	DriverSourceLocal   Source = "local"
	DriverSourceMysql   Source = "mysql"
	DriverSourcePg      Source = "pg"
	DriverSourceMongodb Source = "mongodb"
)

// GetAvailableDriverSource returns the supported source drivers.
func GetAvailableDriverSource() []Source {
	return []Source{
		DriverSourceLocal,
		DriverSourceMysql,
		DriverSourcePg,
		DriverSourceMongodb,
	}
}

// IsAvailableDriverSource reports whether the source driver is supported.
func IsAvailableDriverSource(driver Source) bool {
	return slices.Contains(GetAvailableDriverSource(), driver)
}

// GetSourceDriver creates and configures the requested source driver.
func GetSourceDriver(driver Source, serviceName string, taskName string, driverConfig any, driverEnv any) (interfaces.DriverSourceInterface, error) {
	sourceDrivers := map[Source]interfaces.DriverSourceInterface{
		DriverSourceLocal:   local.New(),
		DriverSourcePg:      pg.New(),
		DriverSourceMongodb: mongodb.New(),
		DriverSourceMysql:   mysql.New(),
	}

	d, ok := sourceDrivers[driver]
	if !ok {
		return nil, fmt.Errorf("source driver '%s' does not exists", driver)
	}

	loggerFields := logger.Fields{
		"driverType":  "source",
		"serviceName": serviceName,
		"taskName":    taskName,
		"driver":      driver,
	}
	d.SetLogger(logger.New().WithFields(loggerFields))

	if err := d.SetConfig(driverConfig); err != nil {
		return nil, fmt.Errorf("set source driver config: %w", err)
	}

	if err := d.SetEnv(driverEnv); err != nil {
		return nil, fmt.Errorf("set source driver env: %w", err)
	}

	return d, nil
}
