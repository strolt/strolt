//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/x/term"
	"github.com/creack/pty"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/unix"
)

const promptPTYChildEnv = "STROLT_TEST_PROMPT_PTY_CHILD"

const promptPTYCompleted = "STROLT_PROMPT_COMPLETED"

func TestPromptTerminalSelectionAndConsecutiveConfirmations(t *testing.T) {
	runPromptPTYChild(t, "selection")
}

func TestPromptTerminalFilterKeyDoesNotBlockSubmission(t *testing.T) {
	runPromptPTYChild(t, "filter-key")
}

func TestPromptTerminalLineFeedConfirmation(t *testing.T) {
	runPromptPTYChild(t, "lf-confirm")
}

func TestPromptTerminalBracketedPaste(t *testing.T) {
	for _, mode := range []string{"paste-confirm", "paste-confirm-yes", "paste-select", "paste-confirm-slow-end", "paste-confirm-yes-slow-start"} {
		t.Run(mode, func(t *testing.T) { runPromptPTYChild(t, mode) })
	}
}

func TestPromptTerminalModifiedEnter(t *testing.T) {
	for _, mode := range []string{"alt-confirm", "alt-select"} {
		t.Run(mode, func(t *testing.T) { runPromptPTYChild(t, mode) })
	}
}

func writePromptWithMarkerGap(t *testing.T, master *os.File, first, rest string) {
	t.Helper()
	_, err := io.WriteString(master, first)
	require.NoError(t, err)
	timer := time.NewTimer(100 * time.Millisecond)
	defer timer.Stop()
	<-timer.C
	_, err = io.WriteString(master, rest)
	require.NoError(t, err)
}

func TestPromptTerminalCancellationPreservesRestorationFailure(t *testing.T) {
	runPromptPTYChild(t, "restore-failure")
}

func TestPromptTerminalCancellationRestoresState(t *testing.T) {
	for _, mode := range []string{"cancel-idle", "cancel-partial-confirmation", "cancel-partial-escape", "cancel-partial-paste"} {
		t.Run(mode, func(t *testing.T) { runPromptPTYChild(t, mode) })
	}
}

func TestPromptTerminalEOFAndCtrlC(t *testing.T) {
	for _, mode := range []string{"eof", "ctrl-c", "ctrl-c-after-escape"} {
		t.Run(mode, func(t *testing.T) { runPromptPTYChild(t, mode) })
	}
}

func TestPromptTerminalChild(t *testing.T) {
	mode := os.Getenv(promptPTYChildEnv)
	if mode == "" {
		return
	}
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGTERM, syscall.SIGINT)
	defer stop()
	err := runPromptPTYChildMode(ctx, mode)
	_, markerErr := fmt.Fprintln(os.Stdout, promptPTYCompleted)
	require.NoError(t, markerErr)
	release := os.NewFile(3, "prompt child release")
	if release == nil {
		t.Fatal("prompt child release pipe is unavailable")
	}
	defer func() { require.NoError(t, release.Close()) }()
	if _, readErr := io.Copy(io.Discard, release); readErr != nil {
		t.Fatal(readErr)
	}
	if err != nil {
		t.Fatal(err)
	}
}

func runPromptPTYChild(t *testing.T, mode string) {
	t.Helper()
	master, slave, err := pty.Open()
	require.NoError(t, err)
	defer func() { _ = master.Close(); _ = slave.Close() }()
	require.NoError(t, pty.Setsize(slave, &pty.Winsize{Rows: 24, Cols: 80, X: 0, Y: 0}))
	before, err := term.GetState(slave.Fd())
	require.NoError(t, err)
	flags, err := unix.FcntlInt(slave.Fd(), unix.F_GETFL, 0)
	require.NoError(t, err)
	executable, err := os.Executable()
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	releaseReader, releaseWriter, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = releaseReader.Close(); _ = releaseWriter.Close() }()
	command := exec.CommandContext(ctx, executable, "-test.run=^TestPromptTerminalChild$")
	command.Env = append(os.Environ(), promptPTYChildEnv+"="+mode)
	command.Stdin, command.Stdout, command.Stderr = slave, slave, slave
	command.ExtraFiles = []*os.File{releaseReader}
	var attrs syscall.SysProcAttr
	attrs.Setsid = true
	attrs.Setctty = true
	command.SysProcAttr = &attrs
	require.NoError(t, command.Start())
	require.NoError(t, releaseReader.Close())
	recorder := newPTYOutputRecorder(master)
	done := make(chan struct{})
	var waitErr error
	go func() { waitErr = command.Wait(); close(done) }()
	t.Cleanup(func() { cancel(); _ = releaseWriter.Close(); <-done })
	label := "choose"
	if mode == "cancel-partial-confirmation" || mode == "lf-confirm" || mode == "alt-confirm" || strings.HasPrefix(mode, "paste-confirm") {
		label = "Are you sure?"
	}
	require.Eventually(t, func() bool { return strings.Contains(recorder.output(), label) }, 5*time.Second, 5*time.Millisecond, "prompt did not render: %s", recorder.output())
	switch mode {
	case "selection":
		_, err = io.WriteString(master, "\x1b[B\r")
		require.NoError(t, err)
		require.Eventually(t, func() bool { return strings.Contains(recorder.output(), "Are you sure?") }, 5*time.Second, 5*time.Millisecond, "%s", recorder.output())
		_, err = io.WriteString(master, "y\r")
		require.NoError(t, err)
		require.Eventually(t, func() bool { return strings.Count(recorder.output(), "Are you sure?") >= 2 }, 5*time.Second, 5*time.Millisecond, "%s", recorder.output())
		_, err = io.WriteString(master, "n\r")
	case "filter-key":
		_, err = io.WriteString(master, "/zzz\r")
	case "lf-confirm":
		_, err = io.WriteString(master, "y\n")
	case "paste-confirm", "paste-select":
		_, err = io.WriteString(master, "\x1b[200~y\n\x1b[201~")
	case "paste-confirm-yes":
		_, err = io.WriteString(master, "\x1b[200~y\x1b[201~")
	case "paste-confirm-slow-end":
		writePromptWithMarkerGap(t, master, "\x1b[200~y\n\x1b", "[201~")
	case "paste-confirm-yes-slow-start":
		writePromptWithMarkerGap(t, master, "\x1b", "[200~y\x1b[201~")
	case "alt-confirm":
		_, err = io.WriteString(master, "y\x1b\r")
	case "alt-select":
		_, err = io.WriteString(master, "\x1b\r")
	case "ctrl-c-after-escape":
		_, err = io.WriteString(master, "\x1b\x03")
	case "cancel-partial-paste":
		_, err = io.WriteString(master, "\x1b[200~y\n")
	case "restore-failure":
		_, err = io.WriteString(master, "\r")
	case "cancel-partial-confirmation":
		_, err = io.WriteString(master, "y")
	case "cancel-partial-escape":
		_, err = io.WriteString(master, "\x1b[")
	case "eof":
		_, err = master.Write([]byte{4})
	case "ctrl-c":
		_, err = master.Write([]byte{3})
	}
	require.NoError(t, err)
	if strings.HasPrefix(mode, "paste-") {
		require.Never(t, func() bool { return strings.Contains(recorder.output(), promptPTYCompleted) }, 250*time.Millisecond, 5*time.Millisecond, "paste submitted without explicit Enter: %s", recorder.output())
		_, err = io.WriteString(master, "\r")
		require.NoError(t, err)
	}
	if strings.HasPrefix(mode, "cancel-") {
		require.NoError(t, command.Process.Signal(syscall.SIGTERM))
	}
	require.Eventually(t, func() bool { return strings.Contains(recorder.output(), promptPTYCompleted) }, 8*time.Second, 5*time.Millisecond, "prompt child did not complete: %s", recorder.output())
	// Inspect terminal state before the child releases its controlling session.
	after, err := term.GetState(slave.Fd())
	require.NoError(t, err)
	require.Equal(t, before, after, "terminal state was not restored")
	afterFlags, err := unix.FcntlInt(slave.Fd(), unix.F_GETFL, 0)
	require.NoError(t, err)
	// O_NONBLOCK is the only status bit this code changes; macOS adds private bits.
	require.Equal(t, flags&unix.O_NONBLOCK, afterFlags&unix.O_NONBLOCK, "shared descriptor nonblocking flag was not restored")
	require.NoError(t, releaseWriter.Close())
	select {
	case <-done:
		require.NoError(t, waitErr, "child failed: %s", recorder.output())
	case <-time.After(8 * time.Second):
		t.Fatalf("prompt child did not exit: %s", recorder.output())
	}
	require.NoError(t, slave.Close())
	select {
	case <-recorder.done:
	case <-time.After(time.Second):
		t.Fatal("PTY output reader did not finish")
	}
	require.Contains(t, recorder.output(), "\x1b[?25h", "cursor was not restored")
}

func runPromptPTYChildMode(ctx context.Context, mode string) error {
	if strings.HasPrefix(mode, "paste-confirm") {
		confirmed, err := runConfirmPromptForTest(ctx)
		if err != nil {
			return err
		}
		if confirmed != strings.HasPrefix(mode, "paste-confirm-yes") {
			return errors.New("pasted confirmation was interpreted incorrectly")
		}
		return nil
	}
	if mode == "paste-select" || mode == "alt-select" {
		selected := ""
		if err := runSelectPrompt(ctx, "choose", []string{"one", "two"}, &selected, nil); err != nil {
			return err
		}
		if selected != "one" {
			return fmt.Errorf("unexpected pasted selection %q", selected)
		}
		return nil
	}
	if mode == "restore-failure" {
		return runPromptRestorationFailure(ctx)
	}
	if mode == "lf-confirm" || mode == "alt-confirm" {
		confirmed, err := runConfirmPromptForTest(ctx)
		if err != nil {
			return err
		}
		if !confirmed {
			return errors.New("LF-terminated confirmation was not accepted")
		}
		return nil
	}
	if mode == "filter-key" {
		selected := ""
		if err := runSelectPrompt(ctx, "choose", []string{"one", "two"}, &selected, nil); err != nil {
			return err
		}
		if selected != "one" {
			return fmt.Errorf("unexpected selection %q", selected)
		}
		return nil
	}
	if mode == "selection" {
		selected, confirmed, declined := "", false, true
		for i := range 3 {
			var runErr error
			switch i {
			case 0:
				runErr = runSelectPrompt(ctx, "choose", []string{"one", "two"}, &selected, nil)
			case 1:
				confirmed, runErr = runConfirmPromptForTest(ctx)
			case 2:
				declined, runErr = runConfirmPromptForTest(ctx)
			}
			if runErr != nil {
				return runErr
			}
		}
		if selected != "two" || !confirmed || declined {
			return errors.New("selection and confirmation answers mismatch")
		}
		return nil
	}
	var err error
	if mode == "cancel-partial-confirmation" {
		_, err = runConfirmPromptForTest(ctx)
	} else {
		selected := ""
		err = runSelectPrompt(ctx, "choose", []string{"one", "two"}, &selected, nil)
	}
	switch {
	case strings.HasPrefix(mode, "cancel-"):
		if !errors.Is(err, context.Canceled) {
			return fmt.Errorf("expected cancellation, got %w", err)
		}
	case strings.HasPrefix(mode, "ctrl-c"):
		if !errors.Is(err, huh.ErrUserAborted) {
			return fmt.Errorf("expected user abort, got %w", err)
		}
	default:
		if !errors.Is(err, io.EOF) {
			return fmt.Errorf("expected EOF, got %w", err)
		}
	}
	return nil
}

func runPromptRestorationFailure(ctx context.Context) (resultErr error) {
	state, err := term.GetState(os.Stdin.Fd())
	if err != nil {
		return fmt.Errorf("capture terminal state: %w", err)
	}
	cleanupFD, err := unix.FcntlInt(os.Stdin.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return fmt.Errorf("duplicate test terminal: %w", err)
	}
	defer func() {
		resultErr = errors.Join(resultErr, term.Restore(uintptr(cleanupFD), state), unix.Close(cleanupFD))
	}()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	err = runPromptAnswer(ctx, nil, func(formCtx context.Context, input io.Reader) error {
		selected := ""
		form := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title("choose").Options(huh.NewOptions("one")...).Value(&selected))).WithAccessible(false).WithKeyMap(promptKeyMap()).WithProgramOptions(tea.WithoutSignalHandler()).WithInput(input).WithOutput(os.Stderr)
		runErr := form.RunWithContext(formCtx)
		cancel()
		return errors.Join(runErr, os.Stdin.Close())
	})
	if !errors.Is(err, context.Canceled) || !errors.Is(err, unix.EBADF) {
		return fmt.Errorf("cancellation or native restoration failure was lost: %w", err)
	}
	return nil
}

func runConfirmPromptForTest(ctx context.Context) (bool, error) {
	var command cobra.Command
	command.SetContext(ctx)
	return NewPrompt(&command).AskIsConfirm() //nolint:contextcheck // Prompt reads the context stored on its Cobra command.
}

type ptyOutputRecorder struct {
	mu   sync.Mutex
	text strings.Builder
	done chan struct{}
}

func newPTYOutputRecorder(master *os.File) *ptyOutputRecorder {
	recorder := &ptyOutputRecorder{mu: sync.Mutex{}, text: strings.Builder{}, done: make(chan struct{})}
	go func() {
		defer close(recorder.done)
		buffer := make([]byte, 1024)
		for {
			n, err := master.Read(buffer)
			if n > 0 {
				recorder.mu.Lock()
				if recorder.text.Len() < 1<<20 {
					_, _ = recorder.text.Write(buffer[:n])
				}
				recorder.mu.Unlock()
			}
			if err != nil {
				return
			}
		}
	}()
	return recorder
}

func (recorder *ptyOutputRecorder) output() string {
	recorder.mu.Lock()
	defer recorder.mu.Unlock()
	return recorder.text.String()
}
