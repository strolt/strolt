package mysql

import (
	"slices"
	"testing"
)

func TestSetConfigTLSValidation(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name    string
		tls     string
		wantErr bool
	}{
		{name: "empty", tls: "", wantErr: false},
		{name: "disabled", tls: TLSModeDisabled, wantErr: false},
		{name: "skip-verify", tls: TLSModeSkipVerify, wantErr: false},
		{name: "unknown", tls: "bogus", wantErr: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &MySQL{}

			err := i.SetConfig(map[string]any{"tls": tc.tls})
			if (err != nil) != tc.wantErr {
				t.Fatalf("SetConfig(tls=%q) error = %v, wantErr %v", tc.tls, err, tc.wantErr)
			}
		})
	}
}

func TestGetCommonArgsTLS(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		tls      string
		wantFlag string
	}{
		{name: "disabled", tls: TLSModeDisabled, wantFlag: "--skip-ssl"},
		{name: "skip-verify", tls: TLSModeSkipVerify, wantFlag: "--skip-ssl-verify-server-cert"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &MySQL{}
			if err := i.SetConfig(map[string]any{"host": "db", "tls": tc.tls}); err != nil {
				t.Fatalf("SetConfig: %v", err)
			}

			args := i.getCommonArgs()
			if !slices.Contains(args, tc.wantFlag) {
				t.Fatalf("getCommonArgs() = %v, want flag %q", args, tc.wantFlag)
			}
		})
	}
}

func TestGetCommonArgsNoTLSFlagByDefault(t *testing.T) {
	t.Parallel()

	i := &MySQL{}
	if err := i.SetConfig(map[string]any{"host": "db"}); err != nil {
		t.Fatalf("SetConfig: %v", err)
	}

	for _, arg := range i.getCommonArgs() {
		if arg == "--skip-ssl" || arg == "--skip-ssl-verify-server-cert" {
			t.Fatalf("getCommonArgs() unexpectedly contains %q", arg)
		}
	}
}
