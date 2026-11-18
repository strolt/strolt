package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/metrics"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/process"
)

var backupMetricsOnce sync.Once

func initBackupMetrics() { backupMetricsOnce.Do(metrics.Init) }

func TestBackupManualPreservesSourceErrors(t *testing.T) {
	initBackupMetrics()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executables require a POSIX shell")
	}

	for _, scenario := range []string{"exit", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			ready := filepath.Join(root, "source-ready")
			destinationMarker := filepath.Join(root, "destination-invoked")
			sourceScript := "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'pg_dump (PostgreSQL) 18.7'; exit 0; fi\n"
			if scenario == "exit" {
				sourceScript += "echo 'source fixture failure' >&2\nexit 17\n"
			} else {
				sourceScript += "printf ready > " + backupShellQuote(ready) + "\nwhile :; do /bin/sleep 1; done\n"
			}
			source := writeBackupFixture(t, root, "pg_dump", sourceScript)
			restore := writeBackupFixture(t, root, "pg_restore", "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'pg_restore (PostgreSQL) 18.7'; exit 0; fi\nexit 0\n")
			psql := writeBackupFixture(t, root, "psql", "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'psql (PostgreSQL) 18.7'; exit 0; fi\nexit 0\n")
			destination := writeBackupFixture(t, root, "restic", "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'restic 0.19.1'; exit 0; fi\nprintf invoked > "+backupShellQuote(destinationMarker)+"\nexit 0\n")
			created := prepareBackupTask(t, root, source, restore, psql, destination, false)
			defer func() { require.NoError(t, created.Close()) }()
			processCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			created.Context.SetProcessContext(processCtx)
			result := make(chan error, 1)
			go func() { result <- created.Backup() }()
			if scenario == "cancel" {
				require.Eventually(t, func() bool { _, err := os.Stat(ready); return err == nil }, 5*time.Second, 10*time.Millisecond)
				cancel()
			}
			select {
			case err := <-result:
				require.Error(t, err)
				if scenario == "cancel" {
					require.ErrorIs(t, err, context.Canceled)
				} else {
					var exitErr *process.ExitError
					require.ErrorAs(t, err, &exitErr)
					require.Equal(t, 17, exitErr.ExitCode())
					require.ErrorContains(t, err, "source fixture failure")
				}
				require.Equal(t, "source: "+created.Context.Source.Error, created.Context.Operation.Error)
				require.NotEmpty(t, created.Context.Source.Error)
				require.NoFileExists(t, destinationMarker)
			case <-time.After(5 * time.Second):
				cancel()
				t.Fatal("source backup did not finish")
			}
		})
	}
}

func TestBackupManualPreservesDestinationErrors(t *testing.T) {
	initBackupMetrics()
	if runtime.GOOS == "windows" {
		t.Skip("fixture executables require a POSIX shell")
	}

	for _, scenario := range []string{"exit", "cancel"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()
			destinationReady := filepath.Join(root, "destination-ready")
			source := writeBackupFixture(t, root, "pg_dump", "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'pg_dump (PostgreSQL) 18.7'; exit 0; fi\nwhile [ $# -gt 0 ]; do case \"$1\" in --file=*) output=${1#--file=};; esac; shift; done\nprintf dump > \"$output\"\n")
			restore := writeBackupFixture(t, root, "pg_restore", "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'pg_restore (PostgreSQL) 18.7'; exit 0; fi\nexit 0\n")
			psql := writeBackupFixture(t, root, "psql", "#!/bin/sh\nif [ \"$1\" = --version ]; then echo 'psql (PostgreSQL) 18.7'; exit 0; fi\nexit 0\n")
			resticScript := "#!/bin/sh\nif [ \"$1\" = version ]; then echo 'restic 0.19.1'; exit 0; fi\nif [ \"$RESTIC_REPOSITORY\" != unused ]; then echo '{\"message_type\":\"summary\",\"snapshot_id\":\"fixture\"}'; exit 0; fi\nprintf ready > " + backupShellQuote(destinationReady) + "\n"
			if scenario == "exit" {
				resticScript += "echo 'destination fixture failure' >&2\nexit 17\n"
			} else {
				resticScript += "while :; do /bin/sleep 1; done\n"
			}
			destination := writeBackupFixture(t, root, "restic", resticScript)
			created := prepareBackupTask(t, root, source, restore, psql, destination, scenario == "exit")
			defer func() { require.NoError(t, created.Close()) }()
			processCtx, cancel := context.WithCancel(context.Background())
			defer cancel()
			created.Context.SetProcessContext(processCtx)
			result := make(chan error, 1)
			go func() { result <- created.Backup() }()
			if scenario == "cancel" {
				require.Eventually(t, func() bool { _, err := os.Stat(destinationReady); return err == nil }, 5*time.Second, 10*time.Millisecond)
				cancel()
			}
			select {
			case err := <-result:
				require.Error(t, err)
				if scenario == "cancel" {
					require.ErrorIs(t, err, context.Canceled)
				} else {
					var exitErr *process.ExitError
					require.ErrorAs(t, err, &exitErr)
					require.Equal(t, 17, exitErr.ExitCode())
					require.ErrorContains(t, err, "destination fixture failure")
				}
				require.NotEmpty(t, created.Context.Destination["storage"].Error)
				require.Equal(t, "destination: [storage]: "+created.Context.Destination["storage"].Error, created.Context.Operation.Error)
				require.Empty(t, created.Context.Destination["archive"].Error)
				require.Contains(t, err.Error(), created.Context.Operation.Error)
			case <-time.After(5 * time.Second):
				cancel()
				t.Fatal("destination backup did not finish")
			}
		})
	}
}

func prepareBackupTask(t *testing.T, root, source, restore, psql, destination string, includeArchive bool) *Task {
	t.Helper()
	emptyConfig := filepath.Join(root, "empty.yml")
	require.NoError(t, os.WriteFile(emptyConfig, []byte("services: {}\n"), 0o600))
	t.Cleanup(func() {
		if err := config.Load(context.Background(), emptyConfig); err != nil {
			t.Errorf("restore empty config: %v", err)
		}
	})
	t.Cleanup(env.Scan)
	dataRoot := filepath.Join(root, "data")
	t.Setenv("STROLT_PATH_DATA", dataRoot)
	env.Scan()
	archiveConfig := ""
	if includeArchive {
		archiveConfig = fmt.Sprintf("        archive:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: successful\n", destination)
	}
	configuration := fmt.Sprintf("services:\n  smoke:\n    task:\n      operationMode: copy\n      source:\n        driver: pg\n        config:\n          bin_path_pg_dump: %q\n          bin_path_pg_restore: %q\n          bin_path_psql: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\n%s", source, restore, psql, destination, archiveConfig)
	configPath := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configuration), 0o600))
	require.NoError(t, config.Load(context.Background(), configPath))
	created, err := New(context.Background(), "smoke", "task", sctxt.TManual, sctxt.OpTypeBackup)
	require.NoError(t, err)
	return created
}

func writeBackupFixture(t *testing.T, root, name, script string) string {
	t.Helper()
	path := filepath.Join(root, name)
	require.NoError(t, os.WriteFile(path, []byte(script), 0o700))
	return path
}

func backupShellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
