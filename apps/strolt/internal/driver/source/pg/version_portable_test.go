package pg

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/shared/process"
)

func TestPortablePostgresMalformedProbeOutput(t *testing.T) {
	for _, mode := range []string{"empty", "truncated", "stderr-only"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			first := installProbe(t, root, "pg_dump", mode)
			second := installProbe(t, root, "pg_restore", "forbidden")
			third := installProbe(t, root, "psql", "forbidden")
			versions, err := (&PgDump{logger: nil, env: nil, config: PgDumpConfig{Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: first, BinPathPgRestore: second, BinPathPsql: third}}).BinaryVersion(context.Background())
			require.Nil(t, versions)
			require.Error(t, err)
			require.NoFileExists(t, filepath.Join(root, "invoked"))
		})
	}
}

func TestPortablePostgresProbeFailureOrdering(t *testing.T) {
	binaries := []string{"pg_dump", "pg_restore", "psql"}
	for failed, name := range binaries {
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			paths := make([]string, len(binaries))
			for index, binary := range binaries {
				mode := "success"
				if index == failed {
					mode = "fail"
				} else if index > failed {
					mode = "forbidden"
				}
				paths[index] = installProbe(t, root, binary, mode)
			}
			driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: paths[0], BinPathPgRestore: paths[1], BinPathPsql: paths[2]}}
			versions, err := driver.BinaryVersion(context.Background())
			require.Nil(t, versions)
			var exitError *process.ExitError
			require.ErrorAs(t, err, &exitError)
			require.Equal(t, 17, exitError.ExitCode())
			require.ErrorContains(t, err, "probe diagnostic")
			require.NoFileExists(t, filepath.Join(root, "invoked"))
		})
	}
}

func TestPortablePostgresProbeCancellation(t *testing.T) {
	root := t.TempDir()
	first := installProbe(t, root, "pg_dump", "gate")
	second := installProbe(t, root, "pg_restore", "forbidden")
	driver := &PgDump{logger: nil, env: nil, config: PgDumpConfig{BinPathPsql: "", Host: "", Port: 0, Database: "", Username: "", Password: "", Format: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathPgDump: first, BinPathPgRestore: second}}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() {
		versions, err := driver.BinaryVersion(ctx)
		if versions != nil {
			result <- errors.New("probe returned partial versions")
			return
		}
		result <- err
	}()
	require.Eventually(t, func() bool {
		_, err := os.Stat(filepath.Join(root, "ready"))
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	cancel()
	select {
	case err := <-result:
		require.ErrorIs(t, err, context.Canceled)
		var exitError *process.ExitError
		require.ErrorAs(t, err, &exitError)
	case <-time.After(5 * time.Second):
		t.Fatal("running probe did not stop after cancellation")
	}
	require.NoFileExists(t, filepath.Join(root, "invoked"))
}
