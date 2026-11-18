package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestDaemonStartupRetainsExistingTemporaryFiles(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("daemon fixture requires POSIX executables and signals")
	}
	root := t.TempDir()
	retained := filepath.Join(root, "strolt-data", "tmp", "i-smoke", "t-task", "n-temp_work_dir", "work-existing", "sentinel")
	require.NoError(t, os.MkdirAll(filepath.Dir(retained), 0o700))
	require.NoError(t, os.WriteFile(retained, []byte("retained operation data"), 0o600))
	source := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(source, 0o700))
	binary := filepath.Join(root, "restic")
	require.NoError(t, os.WriteFile(binary, []byte("#!/bin/sh\nif [ \"$1\" = version ]; then printf 'restic 0.19.1\\n'; exit 0; fi\nexit 17\n"), 0o700))
	configuration := fmt.Sprintf("services:\n  smoke:\n    task:\n      source:\n        driver: local\n        config:\n          path: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\napi:\n  users:\n    test: test\n", source, binary)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())
	process := startStroltProcessFromConfig(t, port, root, configuration, "start")
	client := &http.Client{Timeout: time.Second}
	t.Cleanup(client.CloseIdleConnections)
	url := "http://" + net.JoinHostPort("127.0.0.1", port) + "/api/v1/ping"
	require.Eventually(t, func() bool {
		response, err := client.Get(url)
		if err != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond, "daemon did not start: %s", process.log.output())
	data, err := os.ReadFile(retained)
	require.NoError(t, err)
	require.Equal(t, "retained operation data", string(data))
	require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
	select {
	case <-process.finished:
		require.NoError(t, process.waitErr, process.log.output())
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not stop")
	}
	data, err = os.ReadFile(retained)
	require.NoError(t, err)
	require.Equal(t, "retained operation data", string(data))
}
