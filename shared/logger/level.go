package logger

type LogLevel string

const (
	LogLevelInfo  LogLevel = "INFO"
	LogLevelDebug LogLevel = "DEBUG"

	LogLevelDefault = LogLevelInfo
)

var logLevel = LogLevelDefault

func GetLogLevel() LogLevel {
	return logLevel
}

func SetLogLevel(level LogLevel) {
	logLevel = level
}
