package local

import (
	"os"
	"path"
	"testing"

	"github.com/strolt/strolt/shared/logger"
)

func newDestination(t *testing.T) (*Local, string) {
	t.Helper()

	dir := t.TempDir()

	return &Local{config: Config{Path: dir}, logger: logger.New()}, dir
}

func TestForgetRemovesSnapshot(t *testing.T) {
	t.Parallel()

	destination, dir := newDestination(t)

	for _, snapshotID := range []string{"snapshot-a", "snapshot-b"} {
		if err := os.MkdirAll(path.Join(dir, snapshotID), 0o750); err != nil {
			t.Fatalf("create snapshot %s: %v", snapshotID, err)
		}
	}

	if err := destination.Forget("snapshot-a"); err != nil {
		t.Fatalf("Forget() = %v, want nil", err)
	}

	if _, err := os.Stat(path.Join(dir, "snapshot-a")); !os.IsNotExist(err) {
		t.Fatalf("snapshot-a still exists after Forget(): %v", err)
	}

	if _, err := os.Stat(path.Join(dir, "snapshot-b")); err != nil {
		t.Fatalf("snapshot-b removed by Forget(): %v", err)
	}
}

func TestForgetUnknownSnapshot(t *testing.T) {
	t.Parallel()

	destination, dir := newDestination(t)

	if err := os.MkdirAll(path.Join(dir, "snapshot-a"), 0o750); err != nil {
		t.Fatalf("create snapshot: %v", err)
	}

	if err := destination.Forget("snapshot-unknown"); err == nil {
		t.Fatal("Forget() with unknown snapshot = nil, want error")
	}

	if _, err := os.Stat(path.Join(dir, "snapshot-a")); err != nil {
		t.Fatalf("snapshot-a removed by a failed Forget(): %v", err)
	}
}

// The snapshot ID is joined into a filesystem path, so a relative path must not
// escape the destination directory.
func TestForgetRejectsPathTraversal(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	destinationPath := path.Join(root, "destination")
	if err := os.MkdirAll(destinationPath, 0o750); err != nil {
		t.Fatalf("create destination: %v", err)
	}

	outsidePath := path.Join(root, "outside")
	if err := os.MkdirAll(outsidePath, 0o750); err != nil {
		t.Fatalf("create outside directory: %v", err)
	}

	destination := &Local{config: Config{Path: destinationPath}, logger: logger.New()}

	if err := destination.Forget("../outside"); err == nil {
		t.Fatal("Forget() with traversing id = nil, want error")
	}

	if _, err := os.Stat(outsidePath); err != nil {
		t.Fatalf("directory outside the destination was removed: %v", err)
	}
}

func TestUnlockIsNoop(t *testing.T) {
	t.Parallel()

	destination, _ := newDestination(t)

	if err := destination.Unlock(true); err != nil {
		t.Fatalf("Unlock() = %v, want nil", err)
	}
}
