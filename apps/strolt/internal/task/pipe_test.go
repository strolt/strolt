package task

import (
	"context"
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
	"github.com/strolt/strolt/shared/process"
)

func TestCopyPipe_writerDiesEarly(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd, cancelReader := newPipeCommand(t, "yes 0123456789")
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd, cancelWriter := newPipeCommand(t, "head -c 4096 >/dev/null; exit 3")
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.Error(t, err, "an aborted peer must fail the transfer")
		var joined interface{ Unwrap() []error }
		require.ErrorAs(t, err, &joined)
		var writerExit *process.ExitError
		for _, cause := range joined.Unwrap() {
			if errors.As(cause, &writerExit) && writerExit.ExitCode() == 3 {
				return
			}
		}
		t.Fatalf("writer exit status 3 missing from %v", err)
	case <-time.After(30 * time.Second):
		stopPipeCommands(t, done, cancelReader, cancelWriter, "writer exit")
	}
}

func TestCopyPipe_streamsUntilEof(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd, cancelReader := newPipeCommand(t, "printf 'payload-1234567890'")
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd, cancelWriter := newPipeCommand(t, "cat")
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(30 * time.Second):
		stopPipeCommands(t, done, cancelReader, cancelWriter, "success path")
	}
}

func TestCopyPipe_readerDiesEarly(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd, cancelReader := newPipeCommand(t, "printf 'partial'; exit 4")
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd, cancelWriter := newPipeCommand(t, "cat")
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.Error(t, err, "a crashed producer must fail the transfer")
		assert.Contains(t, err.Error(), "exit status 4")
	case <-time.After(30 * time.Second):
		stopPipeCommands(t, done, cancelReader, cancelWriter, "reader exit")
	}
}

func TestCopyPipe_zeroByteStream(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}

	readerCmd, cancelReader := newPipeCommand(t, "true")
	reader, err := readerCmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, readerCmd.Start())

	writerCmd, cancelWriter := newPipeCommand(t, "cat")
	writer, err := writerCmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, writerCmd.Start())

	done := make(chan error, 1)
	go func() { done <- copyPipe(reader, writer, readerCmd.Wait, writerCmd.Wait) }()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(30 * time.Second):
		stopPipeCommands(t, done, cancelReader, cancelWriter, "empty stream")
	}
}

func newPipeCommand(t *testing.T, script string) (*process.Cmd, context.CancelFunc) {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	cmd := process.CommandContext(ctx, "sh", "-c", script)
	cmd.GracePeriod = 100 * time.Millisecond
	t.Cleanup(cancel)
	return cmd, cancel
}

func stopPipeCommands(t *testing.T, done <-chan error, cancelReader, cancelWriter context.CancelFunc, operation string) {
	t.Helper()
	cancelReader()
	cancelWriter()
	select {
	case err := <-done:
		t.Fatalf("copyPipe deadlocked: timed out on %s; cleanup returned %v", operation, err)
	case <-time.After(5 * time.Second):
		t.Fatalf("copyPipe did not stop after timeout on %s", operation)
	}
}

type sizedTestReader struct {
	io.Reader

	size uint64
}

func (r *sizedTestReader) Size() (uint64, bool) { return r.size, true }

func TestRestoreProgress_countsBytesAndReportsTotal(t *testing.T) {
	var mu sync.Mutex
	var lines []string
	logf := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, args...))
	}
	p := newRestoreProgress(&sizedTestReader{Reader: strings.NewReader("irrelevant"), size: 11}, logf)
	require.True(t, p.hasTotal)
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
	logf := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, args...))
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
	assert.Contains(t, lines[0], "restore stream aborted after 3 B of data")
}

func TestRestoreProgress_withoutTotalOmitsPercentage(t *testing.T) {
	var mu sync.Mutex
	var lines []string
	logf := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, args...))
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
	logf := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, sprintf(format, args...))
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
	require.GreaterOrEqual(t, len(lines), 2)
	assert.Contains(t, lines[0], "restore in progress: 3 B")
}

func TestRestoreProgress_finishWaitsForTicker(t *testing.T) {
	var mu sync.Mutex
	var order []string
	logf := func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		order = append(order, sprintf(format, args...))
	}
	p := newRestoreProgress(strings.NewReader("irrelevant"), logf)
	p.interval = time.Millisecond
	w := &progressWriter{p: p, underlying: nopWriteCloser{}}
	_, err := io.Copy(w, strings.NewReader("x"))
	require.NoError(t, err)
	p.Start()
	time.Sleep(10 * time.Millisecond)
	p.Finish(time.Now(), nil)
	mu.Lock()
	defer mu.Unlock()
	require.NotEmpty(t, order)
	assert.Contains(t, order[len(order)-1], "restore stream complete")
}

func TestProgressWriter_closeDelegatesToWrappedCloser(t *testing.T) {
	if !shAvailable(t) {
		t.Skip("sh not available")
	}
	cmd, _ := newPipeCommand(t, "cat")
	writer, err := cmd.StdinPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())
	w := &progressWriter{p: newRestoreProgress(strings.NewReader(""), func(string, ...any) {}), underlying: writer}
	require.NoError(t, w.Close())
	_, err = w.Write([]byte("x"))
	require.Error(t, err)
	require.NoError(t, cmd.Wait())
}

type nopWriteCloser struct{}

func (nopWriteCloser) Write(b []byte) (int, error) { return len(b), nil }
func (nopWriteCloser) Close() error                { return nil }

type closingWriteCloser struct{ w io.Writer }

func (c *closingWriteCloser) Write(b []byte) (int, error) { return c.w.Write(b) }
func (c *closingWriteCloser) Close() error                { return nil }

func shAvailable(t *testing.T) bool {
	t.Helper()
	path, err := exec.LookPath("sh")
	return err == nil && path != ""
}

func sprintf(format string, args ...any) string {
	return fmt.Sprintf(format, args...) //nolint:forbidigo // test-local formatter over the injectable logger sink
}
