package logger

import (
	"fmt"
	"maps"

	"github.com/sirupsen/logrus"
)

// Fields is a map of structured log fields.
type Fields map[string]any

// Logger is a structured logger that wraps logrus with field support.
type Logger struct {
	fields Fields
	logger *logrus.Logger
}

var globalFields = Fields{}

// SetGlobalField sets a field that is included in every log entry.
func SetGlobalField(filed string, value any) {
	globalFields[filed] = value
}

// New creates a new Logger with empty fields.
func New() *Logger {
	return &Logger{
		logger: logrus.New(),
		fields: Fields{},
	}
}

// WithFields returns a new Logger with the given fields merged in.
func (l *Logger) WithFields(fields Fields) *Logger {
	newFields := Fields{}

	maps.Copy(newFields, l.fields)

	maps.Copy(newFields, fields)

	return &Logger{
		logger: l.logger,
		fields: newFields,
	}
}

// WithField returns a new Logger with the given field merged in.
func (l *Logger) WithField(field string, value any) *Logger {
	fields := Fields{}
	fields[field] = value

	return l.WithFields(fields)
}

// Info logs a message at info level.
func (l *Logger) Info(arg any) {
	l.getLogger().Info(arg)
}

// Infof logs a formatted message at info level.
func (l *Logger) Infof(format string, a ...any) {
	l.Info(fmt.Sprintf(format, a...))
}

// Error logs a message at error level.
func (l *Logger) Error(arg any) {
	l.getLogger().Error(arg)
}

// Errorf logs a formatted message at error level.
func (l *Logger) Errorf(format string, a ...any) {
	l.Error(fmt.Sprintf(format, a...))
}

// Debug logs a message at debug level.
func (l *Logger) Debug(arg any) {
	l.getLogger().Debug(arg)
}

// Debugf logs a formatted message at debug level.
func (l *Logger) Debugf(format string, a ...any) {
	l.Debug(fmt.Sprintf(format, a...))
}

// Warn logs a message at warn level.
func (l *Logger) Warn(arg any) {
	l.getLogger().Warn(arg)
}

// Warnf logs a formatted message at warn level.
func (l *Logger) Warnf(format string, a ...any) {
	l.Warn(fmt.Sprintf(format, a...))
}

// Fatal logs a message at fatal level and exits.
func (l *Logger) Fatal(arg any) {
	l.getLogger().Fatal(arg)
}

// Fatalf logs a formatted message at fatal level and exits.
func (l *Logger) Fatalf(format string, a ...any) {
	l.Fatal(fmt.Sprintf(format, a...))
}

// Trace logs a message at trace level.
func (l *Logger) Trace(arg any) {
	l.getLogger().Trace(arg)
}

// Tracef logs a formatted message at trace level.
func (l *Logger) Tracef(format string, a ...any) {
	l.Trace(fmt.Sprintf(format, a...))
}

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

func (l *Logger) getLogger() *logrus.Entry {
	l.setLogLevel()
	l.setFormat()

	fields := Fields{}

	maps.Copy(fields, globalFields)

	maps.Copy(fields, l.fields)

	logger := l.logger.WithFields(logrus.Fields(fields))

	return logger
}
