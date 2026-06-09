package logger

// LogLevel is the verbosity level of the logger.
type LogLevel string

// Supported log levels and the default level.
const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelTrace LogLevel = "TRACE"

	LogLevelDefault = LogLevelInfo
)

var logLevel = LogLevelDefault

// GetLogLevel returns the current log level.
func GetLogLevel() LogLevel {
	return logLevel
}

// SetLogLevel sets the current log level.
func SetLogLevel(level LogLevel) {
	logLevel = level
}
