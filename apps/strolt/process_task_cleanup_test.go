package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCLIDatabaseTaskFailureCleansWorkDirectory(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("process fixture requires POSIX shell scripts and signals")
	}

	for _, operation := range []string{"backup", "restore"} {
		for _, outcome := range []string{"failure", "cancellation"} {
			t.Run(operation+"/"+outcome, func(t *testing.T) {
				root := t.TempDir()
				marker := filepath.Join(root, "operation-started")
				pgDump := writeTaskCleanupBinary(t, root, "pg_dump", "pg_dump", marker, outcome)
				pgRestore := writeTaskCleanupBinary(t, root, "pg_restore", "pg_restore", marker, outcome)
				psql := writeTaskCleanupBinary(t, root, "psql", "psql", marker, outcome)
				restic := writeTaskCleanupBinary(t, root, "restic", "restic", marker, outcome)
				configuration := fmt.Sprintf(`services:
  smoke:
    task:
      operationMode: copy
      source:
        driver: pg
        config:
          bin_path_pg_dump: %q
          bin_path_pg_restore: %q
          bin_path_psql: %q
          format: c
      destinations:
        storage:
          driver: restic
          config:
            binPath: %q
          env:
            RESTIC_REPOSITORY: unused
`, pgDump, pgRestore, psql, restic)

				arguments := []string{operation, "--service", "smoke", "--task", "task", "--y"}
				if operation == "restore" {
					arguments = append(arguments, "--destination", "storage", "--snapshot", "snapshot")
				}
				process := startStroltProcessFromConfig(t, "0", root, configuration, arguments...)
				pid := waitForTaskCleanupOperation(t, process, marker)
				if outcome == "cancellation" {
					require.NoError(t, process.command.Process.Signal(syscall.SIGTERM))
				}
				select {
				case <-process.finished:
				case <-time.After(5 * time.Second):
					t.Fatalf("%s %s did not finish: %s", operation, outcome, process.log.output())
				}
				var exitErr *exec.ExitError
				require.ErrorAs(t, process.waitErr, &exitErr)
				require.NotZero(t, exitErr.ExitCode())
				require.FileExists(t, marker)
				tempParent := filepath.Join(root, "strolt-data", "tmp", "i-smoke", "t-task", "n-temp_work_dir")
				entries, err := os.ReadDir(tempParent)
				require.NoError(t, err)
				require.Empty(t, entries, "failed or canceled CLI operation leaked its owned work directory")
				assertProcessGone(t, pid)
			})
		}
	}
}

func waitForTaskCleanupOperation(t *testing.T, process *runningStrolt, marker string) int {
	t.Helper()
	deadline := time.NewTimer(5 * time.Second)
	defer deadline.Stop()
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()
	for {
		if data, err := os.ReadFile(marker); err == nil {
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			require.NoError(t, err)
			return pid
		}
		select {
		case <-process.finished:
			data, err := os.ReadFile(marker)
			if err != nil {
				t.Fatalf("CLI exited before starting its operation (err=%v): %s", process.waitErr, process.log.output())
			}
			pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
			require.NoError(t, err)
			return pid
		case <-deadline.C:
			t.Fatalf("CLI did not start its operation: %s", process.log.output())
		case <-ticker.C:
		}
	}
}

func writeTaskCleanupBinary(t *testing.T, root, name, binary, marker, outcome string) string {
	t.Helper()
	path := filepath.Join(root, name)
	versionArgument := "version"
	if binary != "restic" {
		versionArgument = "--version"
	}
	version := binary + " (PostgreSQL) 17.0\\n"
	if binary == "restic" {
		version = "restic 0.19.1\\n"
	}
	script := "#!/bin/sh\nif [ \"$1\" = " + shellQuote(versionArgument) + " ]; then printf " + shellQuote(version) + "; exit 0; fi\n"
	if binary == "restic" {
		script += "if [ \"$1\" = '--json' ] && [ \"$2\" = 'snapshots' ]; then printf '[{\"id\":\"snapshot\",\"short_id\":\"snapshot\",\"time\":\"2024-01-01T00:00:00Z\"}]\\n'; exit 0; fi\n"
	}
	script += "printf '%s\\n' \"$$\" > " + shellQuote(marker+".tmp") + " && mv " + shellQuote(marker+".tmp") + " " + shellQuote(marker) + "\n"
	if outcome == "cancellation" {
		script += "exec /bin/sleep 30\n"
	} else {
		script += "exit 17\n"
	}
	require.NoError(t, os.WriteFile(path, []byte(script), 0o700))
	return path
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
