package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/ovineko/procguard"
	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/process"
)

func TestMain(m *testing.M) {
	name := strings.TrimSuffix(filepath.Base(os.Args[0]), ".exe")
	if after, ok := strings.CutPrefix(name, "version-helper-"); ok {
		versionHelper(after)
		return
	}
	if os.Getenv("STROLT_VERSION_CHILD") == "1" {
		os.Exit(runVersionChild())
	}
	os.Exit(m.Run())
}

func versionHelper(binary string) {
	root := os.Getenv("STROLT_VERSION_FIXTURE_ROOT")
	if root == "" {
		root = filepath.Dir(os.Args[0])
	}
	failPath := filepath.Join(root, "fail-binary")
	if data, err := os.ReadFile(failPath); err == nil && strings.TrimSpace(string(data)) == binary {
		fmt.Fprintln(os.Stderr, "version diagnostic")
		os.Exit(17)
	}
	if strings.HasPrefix(binary, "restic") {
		armPath := filepath.Join(root, "factory-arm")
		readyPath := filepath.Join(root, "factory-ready")
		_ = os.WriteFile(filepath.Join(root, "factory-process-seen"), nil, 0o600)
		if _, err := os.Stat(armPath); err == nil {
			if _, err := os.Stat(readyPath); err == nil {
				_ = os.WriteFile(filepath.Join(root, "factory-forbidden"), nil, 0o600)
				os.Exit(17)
			}
			fmt.Fprintln(os.Stderr, "factory probe diagnostic")
			_ = os.WriteFile(readyPath, nil, 0o600)
			deadline := time.Now().Add(10 * time.Second)
			for time.Now().Before(deadline) {
				if _, err := os.Stat(filepath.Join(root, "factory-release")); err == nil {
					os.Exit(18)
				}
				time.Sleep(10 * time.Millisecond)
			}
			os.Exit(18)
		}
		countPath := filepath.Join(root, "restic-count")
		file, _ := os.OpenFile(countPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
		if file != nil {
			_, _ = file.WriteString("probe\n")
			_ = file.Close()
		}
		version := "restic 0.19.1"
		if binary == "restic-other" {
			version = "restic 0.19.2"
		}
		if _, err := fmt.Fprintln(os.Stdout, version); err != nil {
			os.Exit(2)
		}
		return
	}
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		if binary == "pg_dump" {
			if readyPath := os.Getenv("STROLT_VERSION_PROBE_READY"); readyPath != "" {
				_ = os.WriteFile(readyPath, nil, 0o600)
				timer := time.NewTimer(30 * time.Second)
				<-timer.C
				os.Exit(18)
			}
		}
		versions := map[string]string{
			"pg_dump":    "pg_dump (PostgreSQL) 18.7",
			"pg_restore": "pg_restore (PostgreSQL) 18.7",
			"psql":       "psql (PostgreSQL) 18.7",
		}
		if _, err := fmt.Fprintln(os.Stdout, versions[binary]); err != nil {
			os.Exit(2)
		}
		return
	}
	os.Exit(0)
}

func runVersionChild() int {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if cancelPath := os.Getenv("STROLT_VERSION_CANCEL"); cancelPath != "" {
		done := make(chan struct{})
		go func() {
			defer close(done)
			ticker := time.NewTicker(10 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					if _, err := os.Stat(cancelPath); err == nil {
						cancel()
						return
					}
				}
			}
		}()
		defer func() { cancel(); <-done }()
	}
	args := os.Args
	for i, arg := range args {
		if arg != "--" {
			continue
		}
		os.Args = append([]string{args[0]}, args[i+1:]...)
		if err := ExecuteContext(ctx); err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 2
		}
		return 0
	}
	return 2
}

func TestVersionCommandReportsProbeDiagnosticsAndContinues(t *testing.T) {
	root := t.TempDir()
	probePaths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic", "restic-other")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, probePaths, sourcePath, true)
	require.NoError(t, os.WriteFile(filepath.Join(root, "fail-binary"), []byte("pg_dump"), 0o600))

	output, err := runVersionProcess(t, root, configPath, "")
	require.Error(t, err)
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.NotZero(t, exitError.ExitCode())
	require.Contains(t, output, "pg_restore=18.7")
	require.Contains(t, output, "psql=18.7")
	require.Contains(t, output, "version diagnostic")
	require.Contains(t, output, "restic=0.19.1")
	require.Contains(t, output, "restic=0.19.2")
}

func TestVersionCommandLateSourceFailureKeepsEarlierVersions(t *testing.T) {
	root := t.TempDir()
	probePaths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, probePaths, sourcePath, false)
	require.NoError(t, os.WriteFile(filepath.Join(root, "fail-binary"), []byte("psql"), 0o600))

	output, err := runVersionProcess(t, root, configPath, "")
	require.Error(t, err)
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.NotZero(t, exitError.ExitCode())
	require.Contains(t, output, "pg_dump=18.7")
	require.Contains(t, output, "pg_restore=18.7")
	require.Contains(t, output, "version diagnostic")
	require.NotContains(t, output, "psql=18.7")
	require.Contains(t, output, "restic=0.19.1")
}

func TestVersionCommandExplicitConfigSucceeds(t *testing.T) {
	root := t.TempDir()
	paths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, paths, sourcePath, false)
	output, err := runVersionProcess(t, root, configPath, "")
	require.NoError(t, err, output)
	for _, binary := range []string{"pg_dump=18.7", "pg_restore=18.7", "psql=18.7", "restic=0.19.1"} {
		require.Contains(t, output, binary)
	}
}

func TestVersionCommandDestinationFailureKeepsOtherVersions(t *testing.T) {
	root := t.TempDir()
	paths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic", "restic-other")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, paths, sourcePath, true)
	require.NoError(t, os.WriteFile(filepath.Join(root, "fail-binary"), []byte("restic"), 0o600))
	output, err := runVersionProcess(t, root, configPath, "")
	require.Error(t, err)
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.NotZero(t, exitError.ExitCode())
	require.Contains(t, output, "pg_dump=18.7")
	require.Contains(t, output, "restic=0.19.2")
	require.Contains(t, output, "version diagnostic")
}

func TestVersionCommandDefaultSkipsConfigAndExplicitErrors(t *testing.T) {
	root := t.TempDir()
	invalidDefault := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(invalidDefault, []byte("not: [valid"), 0o600))
	output, err := runVersionArgs(t, root)
	require.NoError(t, err, output)
	require.Contains(t, output, "strolt ")

	for name, data := range map[string][]byte{
		"missing": nil,
		"invalid": []byte("not: [valid"),
	} {
		t.Run(name, func(t *testing.T) {
			configPath := filepath.Join(root, name+".yml")
			if data != nil {
				require.NoError(t, os.WriteFile(configPath, data, 0o600))
			}
			output, err := runVersionArgs(t, root, "--config", configPath)
			require.Error(t, err)
			var exitError *exec.ExitError
			require.ErrorAs(t, err, &exitError)
			require.NotZero(t, exitError.ExitCode())
			_ = output
		})
	}
	_, err = runVersionArgs(t, root, "--not-a-flag")
	require.Error(t, err)
	var exitError *exec.ExitError
	require.ErrorAs(t, err, &exitError)
	require.NotZero(t, exitError.ExitCode())
}

func runVersionArgs(t *testing.T, root string, args ...string) (string, error) {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	commandArgs := append([]string{"-test.run=^TestVersionProcessEntry$", "--", "version"}, args...)
	command := exec.Command(executable, commandArgs...)
	command.Env = append(os.Environ(), "STROLT_VERSION_CHILD=1", "STROLT_PATH_DATA="+filepath.Join(root, "data"))
	command.Dir = root
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestVersionCommandStopsOnProbeCancellation(t *testing.T) {
	root := t.TempDir()
	probePaths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, probePaths, sourcePath, false)
	probeReady := filepath.Join(root, "probe-ready")
	cancelPath := filepath.Join(root, "cancel")
	result := make(chan struct {
		output string
		err    error
	}, 1)
	go func() {
		output, err := runVersionProcess(t, root, configPath, probeReady, cancelPath)
		result <- struct {
			output string
			err    error
		}{output, err}
	}()
	require.Eventually(t, func() bool {
		_, err := os.Stat(probeReady)
		return err == nil
	}, 10*time.Second, 10*time.Millisecond)
	require.NoError(t, os.WriteFile(cancelPath, nil, 0o600))
	select {
	case outcome := <-result:
		require.Error(t, outcome.err)
		require.NotContains(t, outcome.output, "restic=")
	case <-time.After(5 * time.Second):
		t.Fatal("version command continued after probe cancellation")
	}
	require.NoFileExists(t, filepath.Join(root, "restic-count"), "destination probe ran after source cancellation")
}

func TestVersionFactoryCancellationPreservesError(t *testing.T) {
	root := t.TempDir()
	paths := installVersionHelpers(t, root, "pg_dump", "pg_restore", "psql", "restic")
	sourcePath := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(sourcePath, 0o700))
	configPath := writeVersionConfig(t, root, paths, sourcePath, false)
	configData, err := os.ReadFile(configPath)
	require.NoError(t, err)
	needle := fmt.Sprintf("binPath: %q\n          env:", paths["restic"])
	replacement := fmt.Sprintf("binPath: %q\n            env_extra:\n              STROLT_VERSION_FIXTURE_ROOT: %q\n          env:", paths["restic"], root)
	require.Contains(t, string(configData), needle)
	require.NoError(t, os.WriteFile(configPath, []byte(strings.Replace(string(configData), needle, replacement, 1)), 0o600))
	executable, err := os.Executable()
	require.NoError(t, err)
	command := exec.Command(executable, "-test.run=^TestVersionFactoryCancellationHelper$", "--", "factory-cancellation", configPath, root)
	command.Env = append(os.Environ(), "STROLT_PATH_DATA="+filepath.Join(root, "data"), "STROLT_VERSION_FIXTURE_ROOT="+root)
	output, err := command.CombinedOutput()
	if err != nil {
		entries, _ := os.ReadDir(root)
		t.Fatalf("factory child failed: %v; entries: %v; output: %s", err, entries, output)
	}
	require.Contains(t, string(output), "pg_dump=18.7")
	require.NoFileExists(t, filepath.Join(root, "factory-forbidden"))
}

func TestVersionFactoryCancellationHelper(t *testing.T) {
	args := os.Args
	for i := range args {
		if args[i] != "factory-cancellation" || i+2 >= len(args) {
			continue
		}
		configPath, root := args[i+1], args[i+2]
		env.Scan()
		require.NoError(t, config.Load(context.Background(), configPath))
		readyPath := filepath.Join(root, "factory-ready")
		require.NoError(t, os.WriteFile(filepath.Join(root, "factory-arm"), nil, 0o600))
		ctx, cancel := context.WithCancelCause(context.Background())
		result := make(chan error, 1)
		go func() { result <- printDriverVersions(ctx) }()
		joined := false
		t.Cleanup(func() {
			cancel(errors.New("factory cancellation cause"))
			_ = os.WriteFile(filepath.Join(root, "factory-release"), nil, 0o600)
			if joined {
				return
			}
			select {
			case collectorErr := <-result:
				if collectorErr != nil && !errors.Is(collectorErr, context.Canceled) {
					t.Errorf("version collector cleanup returned unexpected error: %v", collectorErr)
				}
			case <-time.After(5 * time.Second):
				t.Error("version collector did not stop during cleanup")
			}
		})
		ready := false
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			if _, err := os.Stat(readyPath); err == nil {
				ready = true
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		if !ready {
			entries, _ := os.ReadDir(root)
			t.Errorf("factory readiness missing; fixture entries: %v", entries)
			cancel(errors.New("factory cancellation cause"))
			_ = os.WriteFile(filepath.Join(root, "factory-release"), nil, 0o600)
			return
		}
		cause := errors.New("factory cancellation cause")
		cancel(cause)
		require.NoError(t, os.WriteFile(filepath.Join(root, "factory-release"), nil, 0o600))
		var err error
		select {
		case err = <-result:
		case <-time.After(5 * time.Second):
			t.Fatal("version collector did not stop after cancellation")
		}
		joined = true
		require.ErrorIs(t, err, context.Canceled)
		require.ErrorIs(t, err, cause)
		var exitErr *process.ExitError
		require.ErrorAs(t, err, &exitErr)
		require.Equal(t, procguard.CleanupConfirmed, exitErr.Exit.Cleanup)
		require.NotErrorIs(t, err, procguard.ErrCleanupUnconfirmed)
		require.ErrorContains(t, err, "factory probe diagnostic")
		require.ErrorContains(t, err, fmt.Sprintf("service %s task %s destination %s", "smoke", "task", "storage"))
		require.NoFileExists(t, filepath.Join(root, "factory-forbidden"))
		return
	}
}

func installVersionHelpers(t *testing.T, root string, names ...string) map[string]string {
	t.Helper()
	executable, err := os.Executable()
	require.NoError(t, err)
	paths := make(map[string]string, len(names))
	for _, name := range names {
		path := filepath.Join(root, "version-helper-"+name)
		if runtime.GOOS == "windows" {
			path += ".exe"
		}
		require.NoError(t, os.Link(executable, path))
		paths[name] = path
	}
	return paths
}

func writeVersionConfig(t *testing.T, root string, paths map[string]string, sourcePath string, addSecondTask bool) string {
	t.Helper()
	services := fmt.Sprintf("  smoke:\n    task:\n      source:\n        driver: pg\n        config:\n          bin_path_pg_dump: %q\n          bin_path_pg_restore: %q\n          bin_path_psql: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\n", paths["pg_dump"], paths["pg_restore"], paths["psql"], paths["restic"])
	if addSecondTask {
		services += fmt.Sprintf("  other:\n    task:\n      source:\n        driver: local\n        config:\n          path: %q\n      destinations:\n        storage:\n          driver: restic\n          config:\n            binPath: %q\n          env:\n            RESTIC_REPOSITORY: unused\n", sourcePath, paths["restic-other"])
	}
	configPath := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte("services:\n"+services), 0o600))
	return configPath
}

func firstCancelPath(paths []string) string {
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

func runVersionProcess(t *testing.T, root, configPath, probeReady string, cancelPath ...string) (string, error) {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	args := []string{"-test.run=^TestVersionProcessEntry$", "--", "version", "--config", configPath}
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = root
	command.Env = append(os.Environ(),
		"STROLT_VERSION_CHILD=1",
		"STROLT_PATH_DATA="+filepath.Join(root, "data"),
		"STROLT_VERSION_PROBE_READY="+probeReady,
		"STROLT_VERSION_FIXTURE_ROOT="+root,
		"STROLT_VERSION_CANCEL="+firstCancelPath(cancelPath),
	)
	output, err := command.CombinedOutput()
	return string(output), err
}

func TestVersionProcessEntry(t *testing.T) {
	if os.Getenv("STROLT_VERSION_CHILD") != "1" {
		t.Skip("child-only test entrypoint")
	}
}
