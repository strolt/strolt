package logger

type LogFormat string

const (
	LogFormatJSON LogFormat = "JSON"
	LogFormatText LogFormat = "TEXT"

	logFormatDefault = LogFormatText
)

var logFormat = logFormatDefault

func GetAvailableLogFormat() []LogFormat {
	return []LogFormat{LogFormatJSON, LogFormatText}
}

func IsLogFormatAvailable(logFormat LogFormat) bool {
	for _, ll := range GetAvailableLogFormat() {
		if logFormat == ll {
			return true
		}
	}

	return false
}

func GetLogFormat() LogFormat {
	return logFormat
}

func SetLogFormat(format LogFormat) {
	if IsLogFormatAvailable(format) {
		logFormat = format
	} else {
		logFormat = logFormatDefault
	}
}
