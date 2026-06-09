// Package utils provides small helper functions shared across strolt services.
package utils

import (
	"time"
	_ "time/tzdata" // embed timezone database so LoadLocation works without OS zoneinfo
)

// TimeGetDefaultTimeZone returns the name of the local timezone, falling back to UTC.
func TimeGetDefaultTimeZone() string {
	loc, err := time.LoadLocation("Local")
	if err == nil {
		return loc.String()
	}

	return time.UTC.String()
}
