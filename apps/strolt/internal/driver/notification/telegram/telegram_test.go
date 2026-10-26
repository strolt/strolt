package telegram

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
		{name: "both missing", config: map[string]any{}, wantErr: true},
		{name: "token missing", config: map[string]any{"chatId": "123"}, wantErr: true},
		{name: "chatId missing", config: map[string]any{"token": "tok"}, wantErr: true},
		{name: "valid", config: map[string]any{"token": "tok", "chatId": "123"}, wantErr: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := New(Params{})
			g.SetLogger(logger.New())

			err := g.SetConfig(tt.config)

			if tt.wantErr && err == nil {
				t.Fatalf("expected error for config %v, got nil", tt.config)
			}

			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error for config %v: %v", tt.config, err)
			}
		})
	}
}

func TestSendNumericChatID(t *testing.T) {
	assertSend(t, "123456", "chat_id=123456", "")
}

func TestSendChannelChatIDGetsAtPrefix(t *testing.T) {
	assertSend(t, "mychannel", "chat_id=@mychannel", "")
}

func assertSend(t *testing.T, chatID string, wantQuery string, _ string) {
	t.Helper()

	var gotPath string
	var gotRawQuery string
	var gotBody []byte

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotRawQuery = r.URL.RawQuery

		b, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("read body: %v", err)
		}

		gotBody = b
	}))
	defer server.Close()

	g := New(Params{apiPrefix: server.URL + "/bot"})
	g.SetLogger(logger.New())

	if err := g.SetConfig(map[string]any{"token": "tok", "chatId": chatID}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	g.Send(testContext())

	if want := "/bot" + "tok" + "/sendMessage"; gotPath != want {
		t.Fatalf("expected path %q, got %q", want, gotPath)
	}

	if !strings.Contains(gotRawQuery, wantQuery) {
		t.Fatalf("expected query to contain %q, got %q", wantQuery, gotRawQuery)
	}

	if !strings.Contains(gotRawQuery, "no_webpage=true") {
		t.Fatalf("expected no_webpage=true in query, got %q", gotRawQuery)
	}

	var payload struct {
		Text      string `json:"text"`
		ParseMode string `json:"parse_mode"`
	}

	if err := json.Unmarshal(gotBody, &payload); err != nil {
		t.Fatalf("decode telegram payload: %v", err)
	}

	if payload.ParseMode != "HTML" {
		t.Fatalf("expected parse_mode HTML, got %q", payload.ParseMode)
	}

	if !strings.Contains(payload.Text, "svc") || !strings.Contains(payload.Text, "task") {
		t.Fatalf("telegram text missing service/task names: %q", payload.Text)
	}
}

func TestSendTransportErrorDoesNotPanic(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	server.Close() // intentionally unreachable

	g := New(Params{apiPrefix: server.URL + "/bot"})
	g.SetLogger(logger.New())

	if err := g.SetConfig(map[string]any{"token": "tok", "chatId": "123"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	g.Send(testContext())
}
