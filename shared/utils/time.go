package utils

import (
	"time"
	_ "time/tzdata"
)

func TimeGetDefaultTimeZone() string {
	loc, err := time.LoadLocation("Local")
	if err == nil {
		return loc.String()
	}

	return time.UTC.String()
}
