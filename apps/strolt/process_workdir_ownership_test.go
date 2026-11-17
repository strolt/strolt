package main

import (
	"fmt"
	"io"
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

func TestDaemonWorkdirOwnershipDuringRejectedRequests(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture requires POSIX shell scripts and signals")
	}

	root := t.TempDir()
	ready := filepath.Join(root, "backup-ready")
	release := filepath.Join(root, "backup-release")
	writeScript := func(name, script string) string {
		t.Helper()
		path := filepath.Join(root, name)
		require.NoError(t, os.WriteFile(path, []byte(script), 0o700))
		return path
	}
	pgRestore := writeScript("pg_restore", "#!/bin/sh\nprintf 'pg_restore (PostgreSQL) 17.0\\n'\n")
	psql := writeScript("psql", "#!/bin/sh\nprintf 'psql (PostgreSQL) 17.0\\n'\n")
	slowDump := writeScript("pg_dump_slow", fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = '--version' ]; then printf 'pg_dump (PostgreSQL) 17.0\\n'; exit 0; fi\ntouch sentinel\nprintf 'ready\\n' > %q.tmp && mv %q.tmp %q\nwhile [ ! -f %q ]; do /bin/sleep 0.01; done\n", ready, ready, ready, release))
	failingDump := writeScript("pg_dump_failure", fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = '--version' ]; then printf 'probe failed\\n' > %q.tmp && mv %q.tmp %q; exit 17; fi\nexit 17\n", filepath.Join(root, "version-failure"), filepath.Join(root, "version-failure"), filepath.Join(root, "version-failure")))
	restic := writeScript("restic", "#!/bin/sh\nif [ \"$1\" = version ]; then printf 'restic 0.19.1\\n'; exit 0; fi\nexit 17\n")
	configuration := fmt.Sprintf(`services:
  smoke:
    task:
      source:
        driver: pg
        config:
          bin_path_pg_dump: %q
          bin_path_pg_restore: %q
          bin_path_psql: %q
      destinations:
        storage:
          driver: restic
          config:
            binPath: %q
          env:
            RESTIC_REPOSITORY: unused
  probeFailure:
    task:
      source:
        driver: pg
        config:
          bin_path_pg_dump: %q
          bin_path_pg_restore: %q
          bin_path_psql: %q
      destinations:
        storage:
          driver: restic
          config:
            binPath: %q
          env:
            RESTIC_REPOSITORY: unused
api:
  users:
    test: test
`, slowDump, pgRestore, psql, restic, failingDump, pgRestore, psql, restic)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())
	process := startStroltProcessFromConfig(t, port, root, configuration, "start")
	client := &http.Client{Timeout: 10 * time.Second}
	t.Cleanup(client.CloseIdleConnections)
	baseURL := "http://" + net.JoinHostPort("127.0.0.1", port)
	readyForRequests := false
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		response, requestErr := client.Get(baseURL + "/api/v1/ping")
		if requestErr == nil {
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				readyForRequests = true
				break
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
	require.True(t, readyForRequests, "daemon did not become ready: %s", process.log.output())

	requestBackup := func(service, task string) *http.Response {
		t.Helper()
		request, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/services/"+service+"/tasks/"+task+"/backup", nil)
		require.NoError(t, err)
		request.SetBasicAuth("test", "test")
		response, err := client.Do(request)
		require.NoError(t, err)
		return response
	}
	response := requestBackup("smoke", "task")
	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusOK, response.StatusCode, string(body))
	require.Eventually(t, func() bool { return fileExists(ready) }, 5*time.Second, 10*time.Millisecond, process.log.output())
	workParent := filepath.Join(root, "strolt-data", "tmp", "i-smoke", "t-task", "n-temp_work_dir")
	entries, err := os.ReadDir(workParent)
	require.NoError(t, err)
	require.Len(t, entries, 1)
	activeWorkDir := filepath.Join(workParent, entries[0].Name())
	require.FileExists(t, filepath.Join(activeWorkDir, "sentinel"))

	for range 3 {
		response = requestBackup("smoke", "task")
		body, err = io.ReadAll(response.Body)
		require.NoError(t, err)
		require.NoError(t, response.Body.Close())
		require.Equal(t, http.StatusBadRequest, response.StatusCode, string(body))
		require.FileExists(t, filepath.Join(activeWorkDir, "sentinel"))
	}
	response = requestBackup("probeFailure", "task")
	body, err = io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusInternalServerError, response.StatusCode, string(body))
	require.Eventually(t, func() bool { return fileExists(filepath.Join(root, "version-failure")) }, time.Second, 10*time.Millisecond)
	require.FileExists(t, filepath.Join(activeWorkDir, "sentinel"))

	require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
	select {
	case <-process.finished:
		require.NoError(t, process.waitErr, process.log.output())
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not stop active backup")
	}
	require.NoDirExists(t, activeWorkDir)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
