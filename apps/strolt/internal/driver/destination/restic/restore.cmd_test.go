package restic

import (
	"slices"
	"testing"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/shared/logger"
)

func TestRestoreArgsWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{logger: logger.New()}
	ctx := context.Context{WorkDir: "/work"}

	want := []string{"--json", "restore", "snapshot-id", "--target", "/work"}

	if got := i.restoreArgs(ctx, "snapshot-id", "", false); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}

func TestRestoreArgsCustomTarget(t *testing.T) {
	t.Parallel()

	i := &Restic{
		logger: logger.New(),
		config: Config{RestoreTarget: "/restore/staging", ExtraRestoreArgs: []string{"--verify"}},
	}
	ctx := context.Context{WorkDir: "/work"}

	want := []string{"--json", "restore", "snapshot-id", "--target", "/restore/staging", "--verify"}

	if got := i.restoreArgs(ctx, "snapshot-id", "", false); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}

// The dump used by piped restores writes to stdout, so neither the restore
// target nor the restore-only extra args apply to it.
func TestRestoreArgsPipeIgnoresRestoreOptions(t *testing.T) {
	t.Parallel()

	i := &Restic{
		logger: logger.New(),
		config: Config{RestoreTarget: "/restore/staging", ExtraRestoreArgs: []string{"--verify"}},
	}
	ctx := context.Context{WorkDir: "/work"}

	want := []string{"--json", "dump", "snapshot-id", "/dump.sql"}

	if got := i.restoreArgs(ctx, "snapshot-id", "/dump.sql", true); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}
