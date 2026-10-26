package slack

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/logger"
)

func testContext() context.Context {
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
		config  map[string]any
		wantErr bool
	}{
		{name: "all fields missing", config: map[string]any{}, wantErr: true},
		{name: "teamId missing", config: map[string]any{"botId": "b", "hookId": "h"}, wantErr: true},
		{name: "botId missing", config: map[string]any{"teamId": "t", "hookId": "h"}, wantErr: true},
		{name: "hookId missing", config: map[string]any{"teamId": "t", "botId": "b"}, wantErr: true},
		{name: "valid", config: map[string]any{"teamId": "t", "botId": "b", "hookId": "h"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := New(Params{})
			s.SetLogger(logger.New())

			err := s.SetConfig(tt.config)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error for config %v, got nil", tt.config)
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for config %v: %v", tt.config, err)
			}
		})
	}
}

func TestSendPostsBlocksToWebhookPath(t *testing.T) {
	var gotPath string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		gotBody = b
	}))
	defer server.Close()

	s := New(Params{apiPrefix: server.URL})
	s.SetLogger(logger.New())

	if err := s.SetConfig(map[string]any{"teamId": "team", "botId": "bot", "hookId": "hook"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	s.Send(testContext())

	if want := "/team/bot/hook"; gotPath != want {
		t.Fatalf("expected webhook path %q, got %q", want, gotPath)
	}

	var payload struct {
		Blocks []map[string]any `json:"blocks"`
	}

	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("decode slack payload: %v", err)
	}

	if len(payload.Blocks) == 0 {
		t.Fatal("expected at least one slack block")
	}

	raw, _ := json.Marshal(payload.Blocks)

	if !strings.Contains(string(raw), "svc") || !strings.Contains(string(raw), "task") {
		t.Fatalf("slack payload missing service/task names: %s", raw)
	}
}

func TestSendTransportErrorDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close() // intentionally unreachable

	s := New(Params{apiPrefix: server.URL})
	s.SetLogger(logger.New())

	if err := s.SetConfig(map[string]any{"teamId": "t", "botId": "b", "hookId": "h"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	s.Send(testContext())
}
