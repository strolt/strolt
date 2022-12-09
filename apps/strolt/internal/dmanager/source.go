package dmanager

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/local"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/mongodb"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/mysql"
	"github.com/strolt/strolt/apps/strolt/internal/driver/source/pg"
	"github.com/strolt/strolt/shared/logger"
)

type Source string

const (
	DriverSourceLocal   Source = "local"
	DriverSourceMysql   Source = "mysql"
	DriverSourcePg      Source = "pg"
	DriverSourceMongodb Source = "mongodb"
)

func GetAvailableDriverSource() []Source {
	return []Source{
		DriverSourceLocal,
		DriverSourceMysql,
		DriverSourcePg,
		DriverSourceMongodb,
	}
}

func IsAvailableDriverSource(driver Source) bool {
	for _, d := range GetAvailableDriverSource() {
		if driver == d {
			return true
		}
	}

	return false
}

func GetSourceDriver(driver Source, serviceName string, taskName string, driverConfig interface{}, driverEnv interface{}) (interfaces.DriverSourceInterface, error) {
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
		return nil, err
	}

	if err := d.SetEnv(driverEnv); err != nil {
		return nil, err
	}

	return d, nil
}
