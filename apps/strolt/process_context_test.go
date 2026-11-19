package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestStroltProcessHelper(t *testing.T) {
	if os.Getenv("STROLT_PROCESS_HELPER") != "1" {
		return
	}
	for index, argument := range os.Args {
		if argument == "--" {
			os.Args = append([]string{os.Args[0]}, os.Args[index+1:]...)
			main()
			os.Exit(0)
		}
	}
	t.Fatal("missing helper command arguments")
}

type processLog struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (log *processLog) Write(data []byte) (int, error) {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.buffer.Write(data)
}

func (log *processLog) output() string {
	log.mu.Lock()
	defer log.mu.Unlock()
	return log.buffer.String()
}

type runningStrolt struct {
	command  *exec.Cmd
	finished chan struct{}
	waitErr  error
	log      processLog
	ready    string
	release  string
}

func startStroltProcess(t *testing.T, port string, arguments ...string) *runningStrolt {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("signal fixture requires a POSIX shell")
	}
	root := t.TempDir()
	ready, release := filepath.Join(root, "ready"), filepath.Join(root, "release")
	binary := filepath.Join(root, "restic")
	script := "#!/bin/sh\nif [ \"$1\" = version ]; then printf 'restic 0.19.1\\n'; exit 0; fi\nprintf 'operation progress\\n'\nprintf '%s\\n' \"$$\" > \"$STROLT_TEST_READY\"\nwhile [ ! -f \"$STROLT_TEST_RELEASE\" ]; do /bin/sleep 0.01; done\n"
	require.NoError(t, os.WriteFile(binary, []byte(script), 0o700))
	source := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(source, 0o700))
	configuration := fmt.Sprintf("services:\n  smoke:\n    task:\n      source:\n        driver: local\n        config:\n          path: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n            env_extra:\n              STROLT_TEST_READY: %q\n              STROLT_TEST_RELEASE: %q\n          env:\n            RESTIC_REPOSITORY: unused\napi:\n  users:\n    test: test\n", source, binary, ready, release)
	return startStroltProcessFromConfig(t, port, root, configuration, arguments...)
}

func startStroltProcessFromConfig(t *testing.T, port, root, configuration string, arguments ...string) *runningStrolt {
	t.Helper()
	ready, release := filepath.Join(root, "ready"), filepath.Join(root, "release")
	configPath := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configuration), 0o600))
	executable, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	args := append([]string{"-test.run=^TestStroltProcessHelper$", "--", "--config", configPath}, arguments...)
	command := exec.CommandContext(ctx, executable, args...)
	command.Env = append(os.Environ(), "STROLT_PROCESS_HELPER=1", "STROLT_PATH_DATA="+root,
		"STROLT_HOST=127.0.0.1", "STROLT_PORT="+port, "STROLT_DISABLE_WATCH_FILES_CHANGED=true")
	process := &runningStrolt{command: command, finished: make(chan struct{}), waitErr: nil, log: processLog{mu: sync.Mutex{}, buffer: bytes.Buffer{}}, ready: ready, release: release}
	command.Stdout, command.Stderr = &process.log, &process.log
	require.NoError(t, command.Start())
	go func() {
		process.waitErr = command.Wait()
		close(process.finished)
	}()
	t.Cleanup(func() {
		_ = os.WriteFile(release, nil, 0o600)
		_ = command.Process.Signal(syscall.SIGTERM)
		select {
		case <-process.finished:
		case <-time.After(5 * time.Second):
			cancel()
			<-process.finished
		}
		cancel()
	})
	return process
}

func (process *runningStrolt) waitForOperation(t *testing.T) int {
	t.Helper()
	var pid int
	require.Eventually(t, func() bool {
		data, err := os.ReadFile(process.ready)
		if err != nil {
			return false
		}
		pid, err = strconv.Atoi(string(bytes.TrimSpace(data)))
		return err == nil
	}, 5*time.Second, 10*time.Millisecond, "operation did not start: %s", process.log.output())
	return pid
}

func assertProcessGone(t *testing.T, pid int) {
	t.Helper()
	require.Eventually(t, func() bool {
		process, err := os.FindProcess(pid)
		if err != nil {
			return errors.Is(err, syscall.ESRCH)
		}
		defer func() { _ = process.Release() }()
		err = process.Signal(syscall.Signal(0))
		return errors.Is(err, os.ErrProcessDone) || errors.Is(err, syscall.ESRCH)
	}, time.Second, 10*time.Millisecond, "subprocess %d survived cancellation", pid)
}

func TestAPIStatsDoesNotProbeSourceBinaries(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture requires POSIX shell scripts and signals")
	}
	root := t.TempDir()
	markers := map[string]string{}
	for _, name := range []string{"pg_dump", "pg_restore", "psql"} {
		marker := filepath.Join(root, name+".marker")
		script := fmt.Sprintf("#!/bin/sh\nprintf 'invoked\\n' > %q\nexit 17\n", marker)
		require.NoError(t, os.WriteFile(filepath.Join(root, name), []byte(script), 0o700))
		markers[name] = marker
	}
	destinationMarker := filepath.Join(root, "destination-stats.marker")
	restic := filepath.Join(root, "restic")
	resticScript := fmt.Sprintf("#!/bin/sh\nif [ \"$1\" = version ]; then printf 'restic 0.19.1\\n'; exit 0; fi\nif [ \"$1\" = --json ] && [ \"$2\" = stats ]; then printf 'called\\n' > %q; printf '%%s\\n' '{\"total_size\":8,\"total_file_count\":2,\"snapshots_count\":2}'; exit 0; fi\nprintf 'unexpected restic invocation\\n' >&2\nexit 17\n", destinationMarker)
	require.NoError(t, os.WriteFile(restic, []byte(resticScript), 0o700))
	configuration := fmt.Sprintf("services:\n  smoke:\n    task:\n      source:\n        driver: pg\n        config:\n          bin_path_pg_dump: %q\n          bin_path_pg_restore: %q\n          bin_path_psql: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\napi:\n  users:\n    test: test\n", filepath.Join(root, "pg_dump"), filepath.Join(root, "pg_restore"), filepath.Join(root, "psql"), restic)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())
	process := startStroltProcessFromConfig(t, port, root, configuration, "start")
	client := &http.Client{Timeout: 10 * time.Second, Transport: nil, CheckRedirect: nil, Jar: nil}
	t.Cleanup(client.CloseIdleConnections)
	baseURL := "http://" + net.JoinHostPort("127.0.0.1", port)
	require.Eventually(t, func() bool {
		response, err := client.Get(baseURL + "/api/v1/ping")
		if err != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond, process.log.output())
	request, err := http.NewRequest(http.MethodGet, baseURL+"/api/v1/services/smoke/tasks/task/destinations/storage/stats", nil)
	require.NoError(t, err)
	request.SetBasicAuth("test", "test")
	response, err := client.Do(request)
	require.NoError(t, err)
	body, readErr := io.ReadAll(response.Body)
	require.NoError(t, readErr)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusOK, response.StatusCode, string(body))
	var result struct {
		Data struct {
			TotalSize      uint64 `json:"totalSize"`
			TotalFileCount uint64 `json:"totalFileCount"`
			SnapshotsCount int    `json:"snapshotsCount"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(body, &result))
	require.Equal(t, uint64(8), result.Data.TotalSize)
	require.Equal(t, uint64(2), result.Data.TotalFileCount)
	require.Equal(t, 2, result.Data.SnapshotsCount)
	_, err = os.Stat(destinationMarker)
	require.NoError(t, err)
	for _, marker := range markers {
		_, err = os.Stat(marker)
		require.ErrorIs(t, err, os.ErrNotExist, "source binary marker exists: %s", marker)
	}
	require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
	select {
	case <-process.finished:
		require.NoError(t, process.waitErr, process.log.output())
	case <-time.After(5 * time.Second):
		t.Fatal("daemon did not exit cleanly")
	}
}

func TestCLIRepositoryOperationInterrupt(t *testing.T) {
	for _, command := range []string{"snapshots", "stats", "init"} {
		t.Run(command, func(t *testing.T) {
			arguments := []string{command}
			if command != "init" {
				arguments = append(arguments, "--service", "smoke", "--task", "task", "--destination", "storage")
			}
			process := startStroltProcess(t, "0", arguments...)
			pid := process.waitForOperation(t)
			require.NoError(t, process.command.Process.Signal(os.Interrupt))
			select {
			case <-process.finished:
			case <-time.After(5 * time.Second):
				t.Fatalf("CLI ignored interrupt: %s", process.log.output())
			}
			assertProcessGone(t, pid)
		})
	}
}

func TestDaemonShutdownCancelsRepositoryRequest(t *testing.T) {
	for _, operation := range []struct {
		name   string
		method string
		path   string
	}{
		{"snapshots", http.MethodGet, "/snapshots"},
		{"stats", http.MethodGet, "/stats"},
		{"prune-preview", http.MethodGet, "/snapshots/prune"},
		{"prune", http.MethodPost, "/prune"},
	} {
		t.Run(operation.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("signal fixture requires a POSIX shell")
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			_, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			require.NoError(t, listener.Close())
			process := startStroltProcess(t, port, "start")
			client := &http.Client{Timeout: 10 * time.Second, Transport: nil, CheckRedirect: nil, Jar: nil}
			t.Cleanup(client.CloseIdleConnections)
			baseURL := "http://" + net.JoinHostPort("127.0.0.1", port)
			require.Eventually(t, func() bool {
				response, err := client.Get(baseURL + "/api/v1/ping")
				if err != nil {
					return false
				}
				_ = response.Body.Close()
				return response.StatusCode == http.StatusOK
			}, 5*time.Second, 10*time.Millisecond)
			request, err := http.NewRequest(operation.method, baseURL+"/api/v1/services/smoke/tasks/task/destinations/storage"+operation.path, nil)
			require.NoError(t, err)
			request.SetBasicAuth("test", "test")
			requestFinished := make(chan struct{})
			go func() {
				defer close(requestFinished)
				response, err := client.Do(request)
				if err == nil {
					_, _ = io.Copy(io.Discard, response.Body)
					_ = response.Body.Close()
				}
			}()
			pid := process.waitForOperation(t)
			require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
			select {
			case <-process.finished:
				require.NoError(t, process.waitErr, process.log.output())
			case <-time.After(5 * time.Second):
				t.Fatalf("daemon shutdown blocked on %s: %s", operation, process.log.output())
			}
			assertProcessGone(t, pid)
			select {
			case <-requestFinished:
			case <-time.After(5 * time.Second):
				t.Fatal("request did not finish after daemon shutdown")
			}
		})
	}
}

func TestClientCancellationStopsRepositoryRequest(t *testing.T) {
	for _, operation := range []struct {
		name   string
		method string
		path   string
	}{
		{"snapshots", http.MethodGet, "/snapshots"},
		{"stats", http.MethodGet, "/stats"},
		{"prune-preview", http.MethodGet, "/snapshots/prune"},
		{"prune", http.MethodPost, "/prune"},
	} {
		t.Run(operation.name, func(t *testing.T) {
			if runtime.GOOS == "windows" {
				t.Skip("signal fixture requires a POSIX shell")
			}
			listener, err := net.Listen("tcp", "127.0.0.1:0")
			require.NoError(t, err)
			_, port, err := net.SplitHostPort(listener.Addr().String())
			require.NoError(t, err)
			require.NoError(t, listener.Close())
			process := startStroltProcess(t, port, "start")
			client := &http.Client{Timeout: 10 * time.Second, Transport: nil, CheckRedirect: nil, Jar: nil}
			t.Cleanup(client.CloseIdleConnections)
			baseURL := "http://" + net.JoinHostPort("127.0.0.1", port)
			require.Eventually(t, func() bool {
				response, err := client.Get(baseURL + "/api/v1/ping")
				if err != nil {
					return false
				}
				_ = response.Body.Close()
				return response.StatusCode == http.StatusOK
			}, 5*time.Second, 10*time.Millisecond)

			requestContext, cancelRequest := context.WithCancel(context.Background())
			defer cancelRequest()
			path := "/api/v1/services/smoke/tasks/task/destinations/storage" + operation.path
			request, err := http.NewRequestWithContext(requestContext, operation.method, baseURL+path, nil)
			require.NoError(t, err)
			request.SetBasicAuth("test", "test")
			requestFinished := make(chan struct{})
			go func() {
				defer close(requestFinished)
				response, err := client.Do(request)
				if err == nil {
					_, _ = io.Copy(io.Discard, response.Body)
					_ = response.Body.Close()
				}
			}()
			pid := process.waitForOperation(t)
			cancelRequest()
			select {
			case <-requestFinished:
			case <-time.After(5 * time.Second):
				t.Fatal("HTTP request did not return after cancellation")
			}
			assertProcessGone(t, pid)
			response, err := client.Get(baseURL + "/api/v1/ping")
			require.NoError(t, err)
			require.Equal(t, http.StatusOK, response.StatusCode)
			require.NoError(t, response.Body.Close())
			require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
			select {
			case <-process.finished:
				require.NoError(t, process.waitErr, process.log.output())
			case <-time.After(5 * time.Second):
				t.Fatal("daemon shutdown blocked after canceled request")
			}
		})
	}
}

func TestBackgroundBackupRetainsDaemonLifetime(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("signal fixture requires a POSIX shell")
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	_, port, err := net.SplitHostPort(listener.Addr().String())
	require.NoError(t, err)
	require.NoError(t, listener.Close())
	process := startStroltProcess(t, port, "start")
	client := &http.Client{Timeout: 10 * time.Second, Transport: nil, CheckRedirect: nil, Jar: nil}
	t.Cleanup(client.CloseIdleConnections)
	baseURL := "http://" + net.JoinHostPort("127.0.0.1", port)
	require.Eventually(t, func() bool {
		response, err := client.Get(baseURL + "/api/v1/ping")
		if err != nil {
			return false
		}
		_ = response.Body.Close()
		return response.StatusCode == http.StatusOK
	}, 5*time.Second, 10*time.Millisecond)

	requestContext, cancelRequest := context.WithCancel(context.Background())
	request, err := http.NewRequestWithContext(requestContext, http.MethodPost,
		baseURL+"/api/v1/services/smoke/tasks/task/backup", nil)
	require.NoError(t, err)
	request.SetBasicAuth("test", "test")
	response, err := client.Do(request)
	require.NoError(t, err)
	_, err = io.Copy(io.Discard, response.Body)
	require.NoError(t, err)
	require.NoError(t, response.Body.Close())
	require.Equal(t, http.StatusOK, response.StatusCode)
	cancelRequest()
	pid := process.waitForOperation(t)
	assertProcessAlive(t, pid)
	ping, err := client.Get(baseURL + "/api/v1/ping")
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, ping.StatusCode)
	require.NoError(t, ping.Body.Close())

	require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
	select {
	case <-process.finished:
		require.NoError(t, process.waitErr, process.log.output())
	case <-time.After(5 * time.Second):
		t.Fatal("daemon shutdown did not drain background backup")
	}
	assertProcessGone(t, pid)
}

func assertProcessAlive(t *testing.T, pid int) {
	t.Helper()
	process, err := os.FindProcess(pid)
	require.NoError(t, err)
	defer func() { _ = process.Release() }()
	require.NoError(t, process.Signal(syscall.Signal(0)))
}
