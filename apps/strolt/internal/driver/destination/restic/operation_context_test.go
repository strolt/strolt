package restic

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	taskcontext "github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/process"
)

func TestRepositoryOperationsCancelAfterOutput(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("fixture executable requires a POSIX shell")
	}
	dataDir := t.TempDir()
	t.Cleanup(env.Scan)
	t.Setenv("STROLT_PATH_DATA", dataDir)
	env.Scan()

	operations := map[string]func(context.Context, *Restic) error{
		"snapshots": func(ctx context.Context, driver *Restic) error {
			_, err := driver.Snapshots(ctx)
			return err
		},
		"stats": func(ctx context.Context, driver *Restic) error {
			_, err := driver.Stats(ctx)
			return err
		},
		"forget": func(ctx context.Context, driver *Restic) error {
			return driver.Forget(ctx, "snapshot")
		},
		"unlock": func(ctx context.Context, driver *Restic) error {
			return driver.Unlock(ctx, false)
		},
		"init": func(ctx context.Context, driver *Restic) error {
			return driver.Init(ctx)
		},
		"prune": func(ctx context.Context, driver *Restic) error {
			operation := taskcontext.Context{}
			operation.SetProcessContext(ctx)
			_, err := driver.Prune(operation, false) //nolint:contextcheck // The operation retains ctx for its subprocess.
			return err
		},
		"version": func(ctx context.Context, driver *Restic) error {
			_, err := driver.BinaryVersion(ctx)
			return err
		},
		"config-validation": func(ctx context.Context, driver *Restic) error {
			return driver.SetConfig(ctx, map[string]any{"binPath": driver.config.BinPath})
		},
	}

	for name, run := range operations {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			ready := filepath.Join(root, "ready")
			release := filepath.Join(root, "release")
			t.Setenv("STROLT_TEST_READY", ready)
			t.Setenv("STROLT_TEST_RELEASE", release)
			binary := filepath.Join(root, "restic")
			script := "#!/bin/sh\nprintf 'progress before cancellation\\n'\nprintf ready > \"$STROLT_TEST_READY\"\nwhile [ ! -f \"$STROLT_TEST_RELEASE\" ]; do /bin/sleep 0.01; done\n"
			require.NoError(t, os.WriteFile(binary, []byte(script), 0o700))
			driver := &Restic{
				taskName: name, driverName: name, logger: logger.New(),
				config: Config{BinPath: binary, ExtraEnv: map[string]string{
					"STROLT_TEST_READY": ready, "STROLT_TEST_RELEASE": release,
				}},
			}
			ctx, cancel := context.WithCancel(t.Context())
			finished := make(chan struct{})
			var operationErr error
			t.Cleanup(func() {
				cancel()
				if err := os.WriteFile(release, nil, 0o600); err != nil {
					t.Error(err)
				}
				select {
				case <-finished:
				case <-time.After(5 * time.Second):
					t.Error("fixture operation did not exit during cleanup")
				}
			})
			go func() {
				operationErr = run(ctx, driver)
				close(finished)
			}()
			require.Eventually(t, func() bool {
				_, err := os.Stat(ready)
				return err == nil
			}, 5*time.Second, 10*time.Millisecond)
			cancel()
			select {
			case <-finished:
			case <-time.After(5 * time.Second):
				t.Fatal("operation ignored cancellation")
			}
			require.ErrorIs(t, operationErr, context.Canceled)
			var exitErr *process.ExitError
			require.ErrorAs(t, operationErr, &exitErr)
		})
	}
}
