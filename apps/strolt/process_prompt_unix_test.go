//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/charmbracelet/x/term"
	"github.com/creack/pty"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

func TestCLIPromptSignalCancellation(t *testing.T) {
	for _, scenario := range []struct {
		name    string
		args    []string
		partial string
		prompt  string
		ctrlC   bool
	}{
		{name: "service-selection", args: []string{"snapshots"}, prompt: "Select service"},
		{name: "confirmation", args: []string{"backup", "--service", "smoke", "--task", "task"}, prompt: "Are you sure?"},
		{name: "partial-confirmation", args: []string{"backup", "--service", "smoke", "--task", "task"}, partial: "y", prompt: "Are you sure?"},
		{name: "partial-escape", args: []string{"snapshots"}, partial: "\x1b[", prompt: "Select service"},
		{name: "ctrl-c", args: []string{"snapshots"}, partial: "\x03", prompt: "Select service", ctrlC: true},
		{name: "snapshot-restore", args: []string{"restore", "--service", "smoke", "--task", "task", "--destination", "storage"}, prompt: "Select snapshot"},
		{name: "snapshot-forget", args: []string{"forget", "--service", "smoke", "--task", "task", "--destination", "storage"}, prompt: "Select snapshot"},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			runCLIPromptSignalCase(t, scenario.args, scenario.partial, scenario.prompt, scenario.ctrlC)
		})
	}
}

func TestCLIPromptSnapshotInputOwnership(t *testing.T) {
	for _, ending := range []struct{ name, value string }{{"CR", "\r"}, {"LF", "\n"}, {"CRLF", "\r\n"}} {
		t.Run(ending.name, func(t *testing.T) { testCLIPromptSnapshotInputOwnership(t, ending.value) })
	}
}

func testCLIPromptSnapshotInputOwnership(t *testing.T, delimiter string) {
	t.Helper()
	for _, operation := range []string{"restore", "forget"} {
		t.Run(operation, func(t *testing.T) {
			root, configPath, marker, listingMarker := createCLIPromptResticFixture(t)
			inputPath := filepath.Join(root, "prompt-input")
			require.NoError(t, os.WriteFile(inputPath, []byte(delimiter+"n"+delimiter), 0o600))
			inputFile, err := os.Open(inputPath)
			require.NoError(t, err)
			t.Cleanup(func() { _ = inputFile.Close() })
			master, slave, err := pty.Open()
			require.NoError(t, err)
			t.Cleanup(func() { _ = master.Close(); _ = slave.Close() })
			output := &processLog{}
			readDone := make(chan struct{})
			go func() {
				defer close(readDone)
				buffer := make([]byte, 1024)
				for {
					n, readErr := master.Read(buffer)
					if n > 0 {
						_, _ = output.Write(buffer[:n])
					}
					if readErr != nil {
						return
					}
				}
			}()
			require.NoError(t, pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}))
			executable, err := os.Executable()
			require.NoError(t, err)
			args := []string{"-test.run=^TestStroltProcessHelper$", "--", "--config", configPath, operation, "--service", "smoke", "--task", "task", "--destination", "storage"}
			command := exec.Command(executable, args...)
			command.Env = append(os.Environ(), "STROLT_PROCESS_HELPER=1", "STROLT_PATH_DATA="+root, "STROLT_DISABLE_WATCH_FILES_CHANGED=true", "STROLT_PROMPT_MARKER="+marker, "STROLT_PROMPT_LIST_MARKER="+listingMarker)
			command.Stdin, command.Stdout, command.Stderr = inputFile, slave, slave
			require.NoError(t, command.Start())
			finished := make(chan error, 1)
			go func() { finished <- command.Wait() }()
			select {
			case err = <-finished:
			case <-time.After(10 * time.Second):
				_ = command.Process.Kill()
				_ = command.Wait()
				t.Fatal("snapshot selection did not consume the preloaded response")
			}
			require.NoError(t, slave.Close())
			select {
			case <-readDone:
			case <-time.After(2 * time.Second):
				t.Fatal("PTY output reader did not finish")
			}
			require.NoError(t, err, output.output())
			require.FileExists(t, listingMarker, "snapshot listing did not run")
			require.NoFileExists(t, marker, "restore or forget ran after declined confirmation")
		})
	}
}

func createCLIPromptResticFixture(t *testing.T) (string, string, string, string) {
	t.Helper()
	root := t.TempDir()
	source := filepath.Join(root, "source")
	require.NoError(t, os.Mkdir(source, 0o700))
	binary := filepath.Join(root, "restic")
	marker := filepath.Join(root, "operation-marker")
	listingMarker := filepath.Join(root, "snapshot-list-marker")
	script := `#!/bin/sh
if [ "$1" = version ]; then printf 'restic 0.19.1\n'; exit 0; fi
for argument in "$@"; do
  if [ "$argument" = snapshots ]; then printf '%s\n' listed >> "$STROLT_PROMPT_LIST_MARKER"; printf '%s\n' '[{"time":"2026-01-02T03:04:05Z","id":"0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef","short_id":"01234567","tags":["fixture"]}]'; exit 0; fi
done
printf '%s\n' "$*" >> "$STROLT_PROMPT_MARKER"
exit 0
`
	require.NoError(t, os.WriteFile(binary, []byte(script), 0o700))
	configuration := fmt.Sprintf(`services:
  smoke:
    task:
      source:
        driver: local
        config:
          path: %q
      destinations:
        storage:
          driver: restic
          config:
            binPath: %q
            env_extra:
              STROLT_PROMPT_MARKER: %q
              STROLT_PROMPT_LIST_MARKER: %q
          env:
            RESTIC_REPOSITORY: unused
`, source, binary, marker, listingMarker)
	configPath := filepath.Join(root, "config.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(configuration), 0o600))
	return root, configPath, marker, listingMarker
}

func runCLIPromptSignalCase(t *testing.T, args []string, partial, promptLabel string, ctrlC bool) {
	t.Helper()
	root, configPath, marker, listingMarker := createCLIPromptResticFixture(t)
	master, slave, err := pty.Open()
	require.NoError(t, err)
	defer func() { _ = master.Close() }()
	defer func() { _ = slave.Close() }()
	require.NoError(t, pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80}))
	before, err := term.GetState(slave.Fd())
	require.NoError(t, err)
	flags, err := unix.FcntlInt(slave.Fd(), unix.F_GETFL, 0)
	require.NoError(t, err)
	executable, err := os.Executable()
	require.NoError(t, err)
	commandArgs := append([]string{"-test.run=^TestStroltProcessHelper$", "--", "--config", configPath}, args...)
	command := exec.Command(executable, commandArgs...)
	command.Env = append(os.Environ(), "STROLT_PROCESS_HELPER=1", "STROLT_PATH_DATA="+root, "STROLT_DISABLE_WATCH_FILES_CHANGED=true", "STROLT_PROMPT_MARKER="+marker, "STROLT_PROMPT_LIST_MARKER="+listingMarker)
	command.Stdin, command.Stdout, command.Stderr = slave, slave, slave
	require.NoError(t, command.Start())
	defer func() {
		if command.ProcessState == nil {
			_ = command.Process.Kill()
			_ = command.Wait()
		}
	}()
	output := &processLog{}
	readDone := make(chan struct{})
	go func() {
		defer close(readDone)
		buffer := make([]byte, 1024)
		for {
			n, readErr := master.Read(buffer)
			if n > 0 {
				_, _ = output.Write(buffer[:n])
			}
			if readErr != nil {
				return
			}
		}
	}()
	require.Eventually(t, func() bool { return strings.Contains(output.output(), promptLabel) }, 10*time.Second, 10*time.Millisecond, "prompt not rendered; output: %s", output.output())
	if partial != "" {
		_, err = io.WriteString(master, partial)
		require.NoError(t, err)
	}
	finished := make(chan error, 1)
	go func() { finished <- command.Wait() }()
	if !ctrlC {
		require.NoError(t, command.Process.Signal(syscall.SIGTERM))
	}
	select {
	case err := <-finished:
		require.Error(t, err, "prompt cancellation should exit nonzero")
	case <-time.After(5 * time.Second):
		if ctrlC {
			t.Fatal("CLI ignored typed Ctrl-C")
		}
		t.Fatal("CLI prompt ignored SIGTERM")
	}
	after, err := term.GetState(slave.Fd())
	require.NoError(t, err)
	require.Equal(t, before, after, "terminal attributes were not restored")
	afterFlags, err := unix.FcntlInt(slave.Fd(), unix.F_GETFL, 0)
	require.NoError(t, err)
	require.Equal(t, flags, afterFlags, "shared descriptor flags were not restored")
	require.NoError(t, slave.Close())
	select {
	case <-readDone:
	case <-time.After(2 * time.Second):
		t.Fatal("PTY output reader did not finish after slave closure")
	}
	outputText := output.output()
	require.Contains(t, outputText, "\x1b[?25h", "cursor visibility restoration was not rendered")
	if strings.Contains(strings.Join(args, " "), "restore") || strings.Contains(strings.Join(args, " "), "forget") {
		require.FileExists(t, listingMarker, "snapshot listing did not run")
		require.NoFileExists(t, marker, "operation ran after canceled snapshot selection")
	}
	if strings.Contains(strings.Join(args, " "), "backup") {
		_, err := os.Stat(marker)
		require.True(t, os.IsNotExist(err), "backup ran after canceled confirmation")
	}
}
