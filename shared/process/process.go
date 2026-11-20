// Package process applies strolt's command and output policy to procguard launches.
package process

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/ovineko/procguard"
)

// Cmd owns one launch. Configure it before Start, close stdin and drain stdout
// before Wait. Args excludes argv[0], matching procguard.Spec.
type Cmd struct {
	Path        string
	Args        []string
	Dir         string
	Env         []string
	Stdin       io.Reader
	Stdout      io.Writer
	Stderr      io.Writer
	GracePeriod time.Duration

	ctx     context.Context //nolint:containedctx // A command retains its operation context until completion.
	child   *procguard.Child
	started bool
	stdin   commandPipe
	stdout  commandPipe
}

type commandPipe struct {
	parent *os.File
	child  *os.File
}

// ExitError preserves the process status and cancellation or cleanup diagnostics.
type ExitError struct {
	Exit  procguard.Exit
	Cause error
}

func (e *ExitError) Error() string { return e.Cause.Error() }

// Unwrap preserves errors.Is support for cancellation and cleanup failures.
func (e *ExitError) Unwrap() error { return e.Cause }

// ExitCode is -1 for a process terminated by a signal.
func (e *ExitError) ExitCode() int { return e.Exit.Code }

// Command runs independently of operation cancellation, for standalone probes.
func Command(path string, args ...string) *Cmd {
	return CommandContext(context.Background(), path, args...)
}

// CommandContext binds a command's process unit to an operation lifetime.
func CommandContext(ctx context.Context, path string, args ...string) *Cmd {
	if ctx == nil {
		panic("process: nil context")
	}
	return &Cmd{
		Path:        path,
		Args:        args,
		Dir:         "",
		Env:         nil,
		Stdin:       nil,
		Stdout:      nil,
		Stderr:      nil,
		GracePeriod: 0,
		ctx:         ctx,
		child:       nil,
		started:     false,
		stdin:       commandPipe{parent: nil, child: nil},
		stdout:      commandPipe{parent: nil, child: nil},
	}
}

// String is a diagnostic command line, not shell-escaped executable input.
func (c *Cmd) String() string {
	var line strings.Builder
	line.WriteString(c.Path)
	for _, arg := range c.Args {
		line.WriteByte(' ')
		line.WriteString(arg)
	}
	return line.String()
}

// StdinPipe returns the write end; closing it delivers EOF to the command.
func (c *Cmd) StdinPipe() (io.WriteCloser, error) {
	if c.started || c.Stdin != nil || c.stdin.child != nil {
		return nil, errors.New("process: stdin already configured or command started")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create stdin pipe: %w", err)
	}
	c.stdin = commandPipe{parent: writer, child: reader}
	c.Stdin = reader
	return writer, nil
}

// StdoutPipe returns the read end; drain or close it before waiting for the command.
func (c *Cmd) StdoutPipe() (io.ReadCloser, error) {
	if c.started || c.Stdout != nil || c.stdout.child != nil {
		return nil, errors.New("process: stdout already configured or command started")
	}
	reader, writer, err := os.Pipe()
	if err != nil {
		return nil, fmt.Errorf("create stdout pipe: %w", err)
	}
	c.stdout = commandPipe{parent: reader, child: writer}
	c.Stdout = writer
	return reader, nil
}

// Start closes parent copies of the child's pipe ends; failure closes all owned ends.
func (c *Cmd) Start() error {
	if c.started {
		return errors.New("process: command already started")
	}
	c.started = true
	if (c.stdin.child != nil && c.Stdin != c.stdin.child) ||
		(c.stdout.child != nil && c.Stdout != c.stdout.child) {
		closeFiles(c.stdin.child, c.stdout.child, c.stdin.parent, c.stdout.parent)
		return errors.New("process: pipe stream replaced before Start")
	}
	child, err := procguard.Start(c.ctx, procguard.Spec{
		Path:            c.Path,
		Args:            c.Args,
		Dir:             c.Dir,
		Env:             c.Env,
		Stdin:           c.Stdin,
		Stdout:          c.Stdout,
		Stderr:          c.Stderr,
		ExtraFiles:      nil,
		GracePeriod:     c.GracePeriod,
		KillImmediately: false,
		Limits:          procguard.Limits{AddressSpace: 0, OOMScoreAdj: nil},
		AfterStart:      nil,
		Cgroup:          nil,
	})
	closeFiles(c.stdin.child, c.stdout.child)
	if err != nil {
		closeFiles(c.stdin.parent, c.stdout.parent)
		return fmt.Errorf("start command: %w", err)
	}
	c.child = child
	return nil
}

// Wait joins process-unit cleanup and output draining, treating failed status as an error.
func (c *Cmd) Wait() error {
	if c.child == nil {
		return errors.New("process: Wait called before successful Start")
	}
	exit := c.child.Wait()
	closeFiles(c.stdin.parent, c.stdout.parent)
	cause := exit.Err
	if exit.Signal != nil {
		cause = errors.Join(cause, fmt.Errorf("signal: %s", exit.Signal))
	} else if exit.Code != 0 {
		cause = errors.Join(cause, fmt.Errorf("exit status %d", exit.Code))
	}
	if exit.Requested {
		cause = errors.Join(cause, c.ctx.Err(), context.Cause(c.ctx))
	}
	if exit.Cleanup != procguard.CleanupConfirmed {
		cause = errors.Join(cause, procguard.ErrCleanupUnconfirmed)
	}
	if cause != nil {
		return &ExitError{Exit: exit, Cause: cause}
	}
	return nil
}

// Run joins the same lifecycle as Start followed by Wait.
func (c *Cmd) Run() error {
	if err := c.Start(); err != nil {
		return err
	}
	return c.Wait()
}

// Output captures stdout and includes captured stderr in failures when no Stderr is supplied.
func (c *Cmd) Output() ([]byte, error) {
	if c.Stdout != nil {
		return nil, errors.New("process: Stdout already set")
	}
	var output, stderr bytes.Buffer
	if c.Stderr == nil {
		c.Stderr = &stderr
	}
	c.Stdout = &output
	err := c.Run()
	if err != nil && stderr.Len() > 0 {
		err = fmt.Errorf("%w: %s", err, stderr.String())
	}
	return output.Bytes(), err
}

// CombinedOutput keeps both streams together for driver diagnostics.
func (c *Cmd) CombinedOutput() ([]byte, error) {
	if c.Stdout != nil || c.Stderr != nil {
		return nil, errors.New("process: output already set")
	}
	var output commandBuffer
	c.Stdout, c.Stderr = &output, &output
	err := c.Run()
	return output.buffer.Bytes(), err
}

func closeFiles(files ...*os.File) {
	for _, file := range files {
		if file != nil {
			_ = file.Close()
		}
	}
}

type commandBuffer struct {
	mu     sync.Mutex
	buffer bytes.Buffer
}

func (w *commandBuffer) Write(data []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.buffer.Write(data) //nolint:wrapcheck // io.Writer adapter preserves the buffer's error.
}
