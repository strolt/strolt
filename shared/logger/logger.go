package logger

import (
	"fmt"

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
	case LogLevelTrace:
		l.logger.SetLevel(logrus.TraceLevel)
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
		l.logger.SetFormatter(&logrus.TextFormatter{
			DisableTimestamp:       true,
			DisableLevelTruncation: true,
		})
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
	newFields := Fields{}

	for key, value := range l.fields {
		newFields[key] = value
	}

	for key, value := range fields {
		newFields[key] = value
	}

	return &Logger{
		logger: l.logger,
		fields: newFields,
	}
}

func (l *Logger) WithField(field string, value interface{}) *Logger {
	fields := Fields{}
	fields[field] = value

	return l.WithFields(fields)
}

func (l *Logger) getLogger() *logrus.Entry {
	l.setLogLevel()
	l.setFormat()

	fields := Fields{}

	for field, value := range globalFields {
		fields[field] = value
	}

	for field, value := range l.fields {
		fields[field] = value
	}

	logger := l.logger.WithFields(logrus.Fields(fields))

	return logger
}

func (l *Logger) Info(arg interface{}) {
	l.getLogger().Info(arg)
}

func (l *Logger) Infof(format string, a ...any) {
	l.Info(fmt.Sprintf(format, a...))
}

func (l *Logger) Error(arg interface{}) {
	l.getLogger().Error(arg)
}

func (l *Logger) Errorf(format string, a ...any) {
	l.Error(fmt.Sprintf(format, a...))
}

func (l *Logger) Debug(arg interface{}) {
	l.getLogger().Debug(arg)
}

func (l *Logger) Debugf(format string, a ...any) {
	l.Debug(fmt.Sprintf(format, a...))
}

func (l *Logger) Warn(arg interface{}) {
	l.getLogger().Warn(arg)
}

func (l *Logger) Warnf(format string, a ...any) {
	l.Warn(fmt.Sprintf(format, a...))
}

func (l *Logger) Fatal(arg interface{}) {
	l.getLogger().Fatal(arg)
}

func (l *Logger) Fatalf(format string, a ...any) {
	l.Fatal(fmt.Sprintf(format, a...))
}

func (l *Logger) Trace(arg interface{}) {
	l.getLogger().Trace(arg)
}

func (l *Logger) Tracef(format string, a ...any) {
	l.Fatal(fmt.Sprintf(format, a...))
}
