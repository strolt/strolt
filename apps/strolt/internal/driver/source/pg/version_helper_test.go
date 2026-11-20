package pg

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	parts := strings.Split(name, "-")
	if len(parts) == 4 && parts[0] == "version" && parts[1] == "helper" {
		probeHelper(parts[2], parts[3])
		return
	}
	os.Exit(m.Run())
}

func probeHelper(binary, mode string) {
	switch mode {
	case "fail":
		fmt.Fprintln(os.Stderr, "probe diagnostic")
		os.Exit(17)
	case "forbidden":
		_ = os.WriteFile(filepath.Join(filepath.Dir(os.Args[0]), "invoked"), nil, 0o600)
	case "gate":
		_ = os.WriteFile(filepath.Join(filepath.Dir(os.Args[0]), "ready"), nil, 0o600)
		// Keep a runnable timer so the runtime does not report a deadlock before cancellation.
		for {
			time.Sleep(time.Hour)
		}
	case "empty":
	case "truncated":
		if _, err := fmt.Fprintln(os.Stdout, "pg_dump"); err != nil {
			os.Exit(2)
		}
	case "stderr-only":
		fmt.Fprintln(os.Stderr, "version diagnostic")
	default:
		versions := map[string]string{
			"pg_dump":    "pg_dump (PostgreSQL) 18.7",
			"pg_restore": "pg_restore (PostgreSQL) 18.7",
			"psql":       "psql (PostgreSQL) 18.7",
		}
		if _, err := fmt.Fprintln(os.Stdout, versions[binary]); err != nil {
			os.Exit(1)
		}
	}
}

func installProbe(t *testing.T, directory, binary, mode string) string {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	name := fmt.Sprintf("version-helper-%s-%s", binary, mode)
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	path := filepath.Join(directory, name)
	if err := os.Link(executable, path); err != nil {
		t.Fatal(err)
	}
	return path
}
