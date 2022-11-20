package logger

import (
	"github.com/sirupsen/logrus"
)

type Fields map[string]interface{}

type Logger struct {
	fields Fields
	logger *logrus.Logger
}

var globalFields = Fields{}

func (l *Logger) setLogLevel() {
	switch logLevel {
	case LogLevelDebug:
		l.logger.SetLevel(logrus.DebugLevel)

	case LogLevelInfo:
	default:
		l.logger.SetLevel(logrus.InfoLevel)
	}
}

func (l *Logger) setFormat() {
	if logFormat == LogFormatJSON {
		l.logger.SetFormatter(&logrus.JSONFormatter{})
	}

	if logFormat == LogFormatText {
		l.logger.SetFormatter(&logrus.TextFormatter{})
	}
}

func SetGlobalField(filed string, value interface{}) {
	globalFields[filed] = value
}

func New() *Logger {
	return &Logger{
		logger: logrus.New(),
		fields: Fields{},
	}
}

func (l *Logger) WithFields(fields Fields) *Logger {
	newFields := l.fields

	for key, value := range fields {
		newFields[key] = value
	}

	return &Logger{
		logger: l.logger,
		fields: newFields,
	}
}

func (l *Logger) WithField(field string, value interface{}) *Logger {
	return l.WithFields(Fields{field: value})
}

func (l *Logger) getLogger() *logrus.Entry {
	l.setLogLevel()
	l.setFormat()

	fields := globalFields

	for field, value := range l.fields {
		fields[field] = value
	}

	logger := l.logger.WithFields(logrus.Fields(globalFields))

	return logger
}

func (l *Logger) Info(args ...interface{}) {
	l.getLogger().Info(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.getLogger().Error(args...)
}

func (l *Logger) Debug(args ...interface{}) {
	l.getLogger().Debug(args...)
}

func (l *Logger) Warn(args ...interface{}) {
	l.getLogger().Warn(args...)
}
