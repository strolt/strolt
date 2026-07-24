package restic

import (
	"slices"
	"testing"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
)

func backupTestContext() context.Context {
	ctx := context.Context{WorkDir: "/work", Tags: []string{"trigger=MANUAL"}}
	ctx.Operation.Time.Start = time.Unix(1600000000, 0)

	return ctx
}

func backupTestVersions() []interfaces.DriverBinaryVersion {
	return []interfaces.DriverBinaryVersion{{Name: "restic", Version: "0.19.0"}}
}

// containsArgPair reports whether args contains flag directly followed by value.
func containsArgPair(args []string, flag string, value string) bool {
	for index, arg := range args {
		if arg == flag && index+1 < len(args) && args[index+1] == value {
			return true
		}
	}

	return false
}

// TestBackupArgsWithoutTuning pins the command built for a config that uses
// none of the tuning fields: it must stay exactly what strolt sent before they
// existed.
func TestBackupArgsWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{}
	ctx := backupTestContext()

	want := []string{
		"--json",
		"backup", "--host", "strolt_host",
		"--tag", "trigger=MANUAL",
		"--tag", "restic=0.19.0",
		"--tag", ldflags.GetBinaryName() + "=" + ldflags.GetVersion(),
		"--tag", "stroltStartedAt=1600000000",
		".",
	}

	if got := i.backupArgs(ctx, "", false, backupTestVersions()); !slices.Equal(got, want) {
		t.Fatalf("backupArgs() = %v, want %v", got, want)
	}
}

func TestBackupArgsPipeWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{}

	got := i.backupArgs(backupTestContext(), "dump.sql", true, backupTestVersions())

	want := []string{"--stdin", "--stdin-filename=dump.sql"}
	if !slices.Equal(got[len(got)-len(want):], want) {
		t.Fatalf("backupArgs() = %v, want it to end with %v", got, want)
	}
}

func TestBackupArgsTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{config: Config{
		Options:         []string{"s3.connections=20", "rest.connections=10"},
		ReadConcurrency: 8,
		NoScan:          true,
		PackSize:        64,
		ExtraBackupArgs: []string{"--exclude-caches", "--group-by=host"},
	}}

	got := i.backupArgs(backupTestContext(), "", false, backupTestVersions())

	for _, want := range [][]string{
		{"-o", "s3.connections=20"},
		{"-o", "rest.connections=10"},
		{"--read-concurrency", "8"},
		{"--pack-size", "64"},
		{"--host", "strolt_host"},
	} {
		if !containsArgPair(got, want[0], want[1]) {
			t.Fatalf("backupArgs() = %v, want the pair %v", got, want)
		}
	}

	if !slices.Contains(got, "--no-scan") || !slices.Contains(got, "--json") {
		t.Fatalf("backupArgs() = %v, want --no-scan and --json", got)
	}

	// The extended options are global flags and must precede the command.
	if slices.Index(got, "-o") > slices.Index(got, "backup") {
		t.Fatalf("backupArgs() = %v, want -o before the backup command", got)
	}

	wantTail := []string{"--exclude-caches", "--group-by=host"}
	if !slices.Equal(got[len(got)-len(wantTail):], wantTail) {
		t.Fatalf("backupArgs() = %v, want it to end with the extra args %v", got, wantTail)
	}
}
