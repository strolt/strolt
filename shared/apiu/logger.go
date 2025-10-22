package apiu

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/strolt/strolt/shared/logger"
)

func Logger() func(next http.Handler) http.Handler {
	return middleware.RequestLogger(&StructuredLogger{Fields: logger.Fields{}})
}

type StructuredLogger struct {
	Fields logger.Fields
}

func (l *StructuredLogger) NewLogEntry(r *http.Request) middleware.LogEntry { //nolint:ireturn
	entry := &StructuredLoggerEntry{Fields: logger.Fields{}}

	username, _, ok := r.BasicAuth()
	if ok {
		entry.Fields["username"] = username
	}

	if reqID := middleware.GetReqID(r.Context()); reqID != "" {
		entry.Fields["req_id"] = reqID
	}

	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}

	entry.Fields["http_scheme"] = scheme
	entry.Fields["http_proto"] = r.Proto
	entry.Fields["http_method"] = r.Method
	entry.Fields["remote_addr"] = r.RemoteAddr
	entry.Fields["user_agent"] = r.UserAgent()
	entry.Fields["uri"] = fmt.Sprintf("%s://%s%s", scheme, r.Host, r.RequestURI)

	return entry
}

type StructuredLoggerEntry struct {
	Fields logger.Fields
}

func (l *StructuredLoggerEntry) Write(status, bytes int, header http.Header, elapsed time.Duration, extra interface{}) {
	logger.New().WithFields(l.Fields).WithFields(logger.Fields{
		"resp_status": status, "resp_bytes_length": bytes,
		"resp_elapsed_ms": float64(elapsed.Nanoseconds()) / 1000000.0, //nolint:mnd
	}).Info("api")
}

func (l *StructuredLoggerEntry) Panic(v interface{}, stack []byte) {
	logger.New().WithFields(l.Fields).WithFields(logger.Fields{
		"stack": string(stack),
		"panic": fmt.Sprintf("%+v", v),
	}).Error("api")
}
