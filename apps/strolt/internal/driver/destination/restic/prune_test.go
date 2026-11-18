package restic

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	taskcontext "github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"
)

func TestPruneStopsWithOperationContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake restic executable uses a POSIX shell")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	dataDir := t.TempDir()
	t.Cleanup(env.Scan)
	t.Setenv("STROLT_PATH_DATA", dataDir)
	env.Scan()

	readyFile := filepath.Join(dataDir, "prune-ready")
	binary := filepath.Join(dataDir, "restic-fake")
	script := "#!/bin/sh\nprintf '%s\\n' \"$$\" > \"$RESTIC_PRUNE_READY_FILE\"\nexec /bin/sleep 30\n"
	require.NoError(t, os.WriteFile(binary, []byte(script), 0o700))

	driver := &Restic{
		driverName: "prune-cancellation-test",
		taskName:   "prune-cancellation-test",
		logger:     logger.New(),
		config: Config{
			BinPath: binary,
			ExtraEnv: map[string]string{
				"RESTIC_PRUNE_READY_FILE": readyFile,
			},
		},
	}
	processContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	operationContext := taskcontext.Context{}
	operationContext.SetProcessContext(processContext)

	pruneDone := make(chan error, 1)
	go func() {
		_, err := driver.Prune(operationContext, false)
		pruneDone <- err
	}()

	require.Eventually(t, func() bool {
		_, err := os.Stat(readyFile)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	cancel()

	select {
	case err := <-pruneDone:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("restic prune did not stop after operation cancellation")
	}
}
