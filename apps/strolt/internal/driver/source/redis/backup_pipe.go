package redis

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
)

// BackupPipe runs redis-cli --rdb - and returns its stdout as a backup
// stream. The RDB magic is validated before the stream is handed out, so a
// server that answered an error string over the protocol fails here instead
// of producing an invalid snapshot.
func (i *Redis) BackupPipe(ctx context.Context) (io.ReadCloser, string, func() error, error) {
	cmd := exec.Command(i.getBinRedisCli(), i.getBackupArgs("-")...) //nolint:gosec,noctx // arguments come from validated configuration; driver context carries no std context
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	i.logger.Debug(cmd.Path)

	stderr := bufcap.New(stderrTailCap)
	cmd.Stderr = stderr

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, "", nil, fmt.Errorf("get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, "", nil, fmt.Errorf("start redis-cli: %w", err)
	}

	reader := bufio.NewReader(pipe)
	magic := make([]byte, len(rdbMagic))
	if _, err := io.ReadFull(reader, magic); err != nil {
		tail := i.abortPipedBackup(pipe, cmd, stderr)

		return nil, "", nil, streamError(fmt.Errorf("invalid RDB stream: %w", err), tail)
	}

	if string(magic) != rdbMagic {
		tail := i.abortPipedBackup(pipe, cmd, stderr)

		return nil, "", nil, streamError(fmt.Errorf("invalid RDB stream: got %q", magic), tail)
	}

	return &prefixedReader{reader: reader, head: magic, closer: pipe}, i.getFileName(), waitWithStderr(cmd, stderr), nil
}

// abortPipedBackup tears down a started redis-cli whose stream turned out
// invalid. The read end is closed before Wait so a blocked writer sees
// EPIPE and exits instead of deadlocking the wait. The stderr tail is read
// only after Wait completed, and returned so the caller can name the actual
// failure (a connection error or NOAUTH abort produces no stdout at all).
func (i *Redis) abortPipedBackup(pipe io.ReadCloser, cmd *exec.Cmd, stderr *bufcap.Writer) string {
	_ = pipe.Close()
	_ = cmd.Process.Kill()
	_ = cmd.Wait()

	return stderr.Tail()
}

func streamError(err error, stderrTail string) error {
	if stderrTail != "" {
		return fmt.Errorf("%w: %s", err, stderrTail)
	}

	return err
}

// prefixedReader replays already-consumed bytes before reading through to
// the underlying reader, so the magic check does not eat the stream head.
type prefixedReader struct {
	reader *bufio.Reader
	head   []byte
	closer io.Closer
}

func (p *prefixedReader) Read(b []byte) (int, error) {
	if len(p.head) > 0 {
		n := copy(b, p.head)
		p.head = p.head[n:]

		return n, nil
	}

	return p.reader.Read(b) //nolint:wrapcheck // io.EOF must reach the copier unwrapped: io.Copy ends streams on the sentinel itself, and a wrapped EOF would read as a failure
}

func (p *prefixedReader) Close() error {
	if err := p.closer.Close(); err != nil {
		return fmt.Errorf("close rdb stream: %w", err)
	}

	return nil
}
