package task

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCopyPipe_writerDiesEarly must not hang: the process behind the reader
// keeps producing data while the process behind the writer has already
// exited. Waiting for the producer without closing the pipe first used to
// block forever with zero CPU (the production deadlock behind hung restores).
func TestCopyPipe_writerDiesEarly(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	// The reader side produces an endless stream, far more than any pipe
	// buffer holds, exactly like `restic dump` streaming a large dump.
	readerCmd := exec.Command("sh", "-c", "yes 0123456789") //nolint:gosec // fixed test command
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	// The writer side consumes a little and then aborts, exactly like
	// pg_restore exiting on the first SQL error mid-restore.
	writerCmd := exec.Command("sh", "-c", "head -c 4096 >/dev/null; exit 3") //nolint:gosec // fixed test command
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	type result struct {
		err error
	}

	done := make(chan result, 1)

	go func() {
		done <- result{err: copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait)}
	}()

	select {
	case res := <-done:
		require.Error(t, res.err, "an aborted peer must fail the transfer")

		writerStatusReported := false

		var joined interface{ Unwrap() []error }
		if errors.As(res.err, &joined) {
			for _, err := range joined.Unwrap() {
				var exitErr *exec.ExitError
				if errors.As(err, &exitErr) && exitErr.ExitCode() == 3 {
					writerStatusReported = true
				}
			}
		}

		assert.True(t, writerStatusReported, "the aborted writer exit status (3) must be part of the error: %v", res.err)

	case <-time.After(30 * time.Second):
		_ = readerCmd.Process.Kill()
		_ = writerCmd.Process.Kill()
		t.Fatal("copyPipe deadlocked: the reader process was never unblocked after the writer died")
	}
}

// TestCopyPipe_streamsUntilEof verifies the success path: all data crosses
// the pipe, both waits return nil and the transfer reports no error.
func TestCopyPipe_streamsUntilEof(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd := exec.Command("sh", "-c", "printf 'payload-1234567890'") //nolint:gosec // fixed test command
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd := exec.Command("sh", "-c", "cat") //nolint:gosec // fixed test command
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)

	go func() {
		done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait)
	}()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(30 * time.Second):
		_ = readerCmd.Process.Kill()
		_ = writerCmd.Process.Kill()
		t.Fatal("copyPipe deadlocked on the success path")
	}
}

// TestCopyPipe_readerDiesEarly covers the producer crashing mid-stream: the
// copy sees EOF (or a read error), the producer's non-zero exit must surface
// in the joined error, and the consumer must still receive EOF and be reaped.
func TestCopyPipe_readerDiesEarly(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd := exec.Command("sh", "-c", "printf 'partial'; exit 4") //nolint:gosec // fixed test command
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd := exec.Command("sh", "-c", "cat") //nolint:gosec // fixed test command
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.Error(t, err, "a crashed producer must fail the transfer")
		assert.Contains(t, err.Error(), "exit status 4", "the producer exit status must be reported: %v", err)
	case <-time.After(30 * time.Second):
		_ = readerCmd.Process.Kill()
		_ = writerCmd.Process.Kill()
		t.Fatal("copyPipe deadlocked after the reader died")
	}
}

// TestCopyPipe_zeroByteStream covers an empty dump: no data crosses the pipe,
// both processes exit cleanly and the transfer must succeed without hanging.
func TestCopyPipe_zeroByteStream(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd := exec.Command("sh", "-c", "true") //nolint:gosec // fixed test command
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd := exec.Command("sh", "-c", "cat") //nolint:gosec // fixed test command
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(30 * time.Second):
		_ = readerCmd.Process.Kill()
		_ = writerCmd.Process.Kill()
		t.Fatal("copyPipe deadlocked on an empty stream")
	}
}

// sizedTestReader is a sizeAwareReader stub for the progress reporter tests.
type sizedTestReader struct {
	io.Reader
	size uint64
}

func (r *sizedTestReader) Size() (uint64, bool) { return r.size, true }

func TestRestoreProgress_countsBytesAndReportsTotal(t *testing.T) {
	var mu sync.Mutex
	var lines []string

	logf := func(format string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, a...))
	}

	p := newRestoreProgress(&sizedTestReader{Reader: strings.NewReader("irrelevant"), size: 11}, logf)
	require.True(t, p.hasTotal, "the reporter must pick the size up from a sizeAwareReader")

	p.Start()

	w := &progressWriter{p: p, underlying: nopWriteCloser{}}
	_, err := io.Copy(w, strings.NewReader("hello world"))
	require.NoError(t, err)

	p.Finish(time.Now().Add(-2*time.Second), nil)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "restore stream complete: 11 B")
	assert.Contains(t, lines[0], "5.5 B/s")
}

func TestRestoreProgress_countsOnlyAcceptedBytes(t *testing.T) {
	var mu sync.Mutex
	var lines []string

	logf := func(format string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, a...))
	}

	p := newRestoreProgress(strings.NewReader("irrelevant"), logf)

	w := &progressWriter{p: p, underlying: &closingWriteCloser{w: &shortWriter{limit: 3}}}

	_, err := io.Copy(w, strings.NewReader("hello"))
	require.ErrorIs(t, err, io.ErrShortWrite)

	p.Start()
	p.Finish(time.Now(), err)

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "restore stream aborted after 3 B of data",
		"only bytes the restore process accepted may be reported")
}

func TestRestoreProgress_withoutTotalOmitsPercentage(t *testing.T) {
	var mu sync.Mutex
	var lines []string

	logf := func(format string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, a...))
	}

	p := newRestoreProgress(strings.NewReader("irrelevant"), logf)
	require.False(t, p.hasTotal)

	p.Start()

	w := &progressWriter{p: p, underlying: nopWriteCloser{}}
	_, err := io.Copy(w, strings.NewReader("data"))
	require.NoError(t, err)

	p.Finish(time.Now().Add(-time.Second), errors.New("boom"))

	mu.Lock()
	defer mu.Unlock()

	require.Len(t, lines, 1)
	assert.Contains(t, lines[0], "restore stream aborted after 4 B of data")
	assert.Contains(t, lines[0], "boom")
}

func TestRestoreProgress_periodicLinesReportSpeed(t *testing.T) {
	var mu sync.Mutex
	var lines []string

	logf := func(format string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, a...))
	}

	p := newRestoreProgress(strings.NewReader("irrelevant"), logf)
	p.interval = 20 * time.Millisecond

	w := &progressWriter{p: p, underlying: nopWriteCloser{}}
	_, err := io.Copy(w, strings.NewReader("abc"))
	require.NoError(t, err)

	p.Start()
	time.Sleep(80 * time.Millisecond)
	p.Finish(time.Now(), nil)

	mu.Lock()
	defer mu.Unlock()

	require.GreaterOrEqual(t, len(lines), 2, "a periodic line must be logged while the reporter runs")
	assert.Contains(t, lines[0], "restore in progress: 3 B")
}

func TestRestoreProgress_finishWaitsForTicker(t *testing.T) {
	var mu sync.Mutex
	var order []string

	logf := func(format string, a ...any) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, sprintf(format, a...))
	}

	p := newRestoreProgress(strings.NewReader("irrelevant"), logf)
	p.interval = time.Millisecond

	w := &progressWriter{p: p, underlying: nopWriteCloser{}}
	_, err := io.Copy(w, strings.NewReader("x"))
	require.NoError(t, err)

	p.Start()
	time.Sleep(10 * time.Millisecond)
	p.Finish(time.Now(), nil)

	// Finish must wait for the ticker goroutine: no periodic line may land
	// after the final line.
	mu.Lock()
	defer mu.Unlock()

	require.NotEmpty(t, order)
	assert.Contains(t, order[len(order)-1], "restore stream complete",
		"no periodic line may be logged after the final line")
}

func TestProgressWriter_closeDelegatesToWrappedCloser(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	cmd := exec.Command("sh", "-c", "cat") //nolint:gosec // fixed test command
	writer, err := cmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())
	defer func() { _ = cmd.Wait() }()

	w := &progressWriter{p: newRestoreProgress(strings.NewReader(""), func(string, ...any) {}), underlying: writer}
	require.NoError(t, w.Close())

	_, err = w.Write([]byte("x"))
	require.Error(t, err, "the wrapped stdin pipe must be closed after Close")
}

type nopWriteCloser struct{}

func (c nopWriteCloser) Write(b []byte) (int, error) { return len(b), nil }

func (c nopWriteCloser) Close() error { return nil }

type closingWriteCloser struct{ w io.Writer }

func (c *closingWriteCloser) Write(b []byte) (int, error) { return c.w.Write(b) }

func (c *closingWriteCloser) Close() error { return nil }

func shAvailable(t *testing.T) bool {
	t.Helper()

	path, err := exec.LookPath("sh")
	if err != nil || path == "" {
		return false
	}

	return true
}

func sprintf(format string, a ...any) string {
	return fmt.Sprintf(format, a...) //nolint:forbidigo // test-local formatter over the injectable logger sink
}
