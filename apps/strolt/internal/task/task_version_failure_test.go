package task

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/process"
)

func TestMain(m *testing.M) {
	if strings.HasPrefix(strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe"), "task-version-probe") {
		if len(os.Args) > 1 && os.Args[1] == "--version" {
			fmt.Fprintln(os.Stderr, "source probe diagnostic")
			os.Exit(17)
		}
		if len(os.Args) > 1 && os.Args[1] != "version" {
			_ = os.WriteFile(filepath.Join(filepath.Dir(os.Args[0]), "backup-invoked"), nil, 0o600)
		}
		os.Exit(0)
	}
	os.Exit(m.Run())
}

func TestNewClosesWorkdirWhenSourceVersionProbeFails(t *testing.T) {
	root := t.TempDir()
	emptyConfigPath := filepath.Join(root, "empty.yml")
	require.NoError(t, os.WriteFile(emptyConfigPath, []byte("services: {}\n"), 0o600))
	t.Cleanup(func() {
		if err := config.Load(context.Background(), emptyConfigPath); err != nil {
			t.Errorf("restore empty config: %v", err)
		}
	})
	t.Cleanup(env.Scan)
	dataRoot := filepath.Join(root, "data")
	workRoot := filepath.Join(dataRoot, "strolt-data", "tmp")
	require.NoError(t, os.MkdirAll(workRoot, 0o700))
	t.Setenv("STROLT_PATH_DATA", dataRoot)
	env.Scan()

	executable, err := os.Executable()
	require.NoError(t, err)
	probe := filepath.Join(root, "task-version-probe")
	if runtime.GOOS == "windows" {
		probe += ".exe"
	}
	require.NoError(t, os.Link(executable, probe))
	configuration := fmt.Sprintf("services:\n  smoke:\n    task:\n      source:\n        driver: pg\n        config:\n          bin_path_pg_dump: %q\n          bin_path_pg_restore: %q\n          bin_path_psql: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\n", probe, probe, probe, probe)
	configPath := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configuration), 0o600))
	require.NoError(t, config.Load(context.Background(), configPath))
	taskDir := filepath.Join(workRoot, "i-smoke", "t-task", "n-temp_work_dir")
	require.NoError(t, os.MkdirAll(taskDir, 0o700))

	created, err := New(context.Background(), "smoke", "task", sctxt.TManual, sctxt.OpTypeBackup)
	require.Empty(t, created.Context.WorkDir)
	require.Error(t, err)
	require.ErrorContains(t, err, "source probe diagnostic")
	var exitError *process.ExitError
	require.ErrorAs(t, err, &exitError)
	require.Equal(t, 17, exitError.ExitCode())
	after, err := os.ReadDir(taskDir)
	require.NoError(t, err)
	require.Empty(t, after, "failed admission leaked an owned temporary work directory")
	require.NoFileExists(t, filepath.Join(root, "backup-invoked"))
}
