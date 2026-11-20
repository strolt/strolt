package mongodb

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

func TestPortableMongoMalformedProbeOutput(t *testing.T) {
	for _, mode := range []string{"empty", "truncated", "stderr-only"} {
		t.Run(mode, func(t *testing.T) {
			root := t.TempDir()
			first := installProbe(t, root, "mongodump", mode)
			second := installProbe(t, root, "mongorestore", "forbidden")
			versions, err := (&MongoDB{logger: nil, env: nil, config: Config{URI: "", Host: "", Port: 0, Database: "", Username: "", Password: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathMongoDump: first, BinPathMongoRestore: second}}).BinaryVersion(context.Background())
			require.Nil(t, versions)
			require.Error(t, err)
			require.NoFileExists(t, filepath.Join(root, "invoked"))
		})
	}
}

func TestPortableMongoProbeFailureOrdering(t *testing.T) {
	binaries := []string{"mongodump", "mongorestore"}
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
			driver := &MongoDB{logger: nil, env: nil, config: Config{URI: "", Host: "", Port: 0, Database: "", Username: "", Password: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathMongoDump: paths[0], BinPathMongoRestore: paths[1]}}
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

func TestPortableMongoProbeCancellation(t *testing.T) {
	root := t.TempDir()
	first := installProbe(t, root, "mongodump", "gate")
	second := installProbe(t, root, "mongorestore", "forbidden")
	driver := &MongoDB{logger: nil, env: nil, config: Config{URI: "", Host: "", Port: 0, Database: "", Username: "", Password: "", BackupParams: "", RestoreParams: "", CommonParams: "", BinPathMongoDump: first, BinPathMongoRestore: second}}
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
