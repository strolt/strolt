// Package logger provides a structured logging wrapper around logrus.
package logger

import "slices"

// LogFormat is the output format of the logger.
type LogFormat string

// Supported log formats.
const (
	LogFormatJSON LogFormat = "JSON"
	LogFormatText LogFormat = "TEXT"

	logFormatDefault = LogFormatText
)

var logFormat = logFormatDefault

// GetAvailableLogFormat returns the list of supported log formats.
func GetAvailableLogFormat() []LogFormat {
	return []LogFormat{LogFormatJSON, LogFormatText}
}

// IsLogFormatAvailable reports whether the given log format is supported.
func IsLogFormatAvailable(logFormat LogFormat) bool {
	return slices.Contains(GetAvailableLogFormat(), logFormat)
}

// GetLogFormat returns the current log format.
func GetLogFormat() LogFormat {
	return logFormat
}

// SetLogFormat sets the log format, falling back to the default if unsupported.
func SetLogFormat(format LogFormat) {
	if IsLogFormatAvailable(format) {
		logFormat = format
	} else {
		logFormat = logFormatDefault
	}
}
