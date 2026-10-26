package webhook_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/notification/webhook"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

func newTestContext() context.Context {
	return context.Context{
		Trigger:        sctxt.TManual,
		ServiceName:    "svc",
		TaskName:       "task",
		OpertationType: sctxt.OpTypeBackup,
		Event:          sctxt.EvOperationStop,
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		url     string
		wantErr bool
	}{
		{name: "empty url", url: "", wantErr: true},
		{name: "unsupported scheme", url: "ftp://example.com/hook", wantErr: true},
		{name: "missing host", url: "http:///hook", wantErr: true},
		{name: "valid http", url: "http://example.com/hook", wantErr: false},
		{name: "valid https", url: "https://example.com/hook", wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := webhook.New()
			w.SetLogger(logger.New())

			err := w.SetConfig(map[string]any{"url": tt.url})

			if tt.wantErr && err == nil {
				t.Fatalf("expected error for url %q, got nil", tt.url)
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for url %q: %v", tt.url, err)
			}
		})
	}
}

func TestSendPostsContextAsJSON(t *testing.T) {
	var received map[string]any

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}

		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected application/json content type, got %q", ct)
		}

		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode body: %v", err)
		}
	}))
	defer server.Close()

	w := webhook.New()
	w.SetLogger(logger.New())

	if err := w.SetConfig(map[string]any{"url": server.URL}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	w.Send(newTestContext())

	// Decode into an independent structure so a change to the production JSON
	// tags cannot silently break webhook consumers.
	for key, want := range map[string]any{
		"serviceName":    "svc",
		"taskName":       "task",
		"opertationType": string(sctxt.OpTypeBackup),
		"event":          string(sctxt.EvOperationStop),
	} {
		if got := received[key]; got != want {
			t.Errorf("payload field %q = %v, want %v", key, got, want)
		}
	}
}

func TestSendServerSideErrorDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("boom"))
	}))
	defer server.Close()

	w := webhook.New()
	w.SetLogger(logger.New())

	if err := w.SetConfig(map[string]any{"url": server.URL}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	w.Send(newTestContext())
}

func TestSendTransportErrorDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close() // intentionally unreachable

	w := webhook.New()
	w.SetLogger(logger.New())

	if err := w.SetConfig(map[string]any{"url": server.URL}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	w.Send(newTestContext())
}
