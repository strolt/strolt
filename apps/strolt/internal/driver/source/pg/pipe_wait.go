package pg

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
)

// stderrTailCap bounds how much subprocess stderr is retained on pipe
// streams. psql/pg_restore print the failing statement's error to stderr; a
// bounded tail is enough to explain an abort without risking unbounded
// memory on chatty runs.
const stderrTailCap = 8 * 1024

// waitWithStderr returns a wait function that appends the captured stderr
// tail to a non-zero exit, so a piped restore failure names the actual SQL
// error instead of a bare "exit status 1".
func waitWithStderr(cmd *exec.Cmd, stderr *bufcap.Writer) func() error {
	return func() error {
		err := cmd.Wait()
		if err == nil {
			return nil
		}

		if tail := strings.TrimSpace(stderr.Tail()); tail != "" {
			return fmt.Errorf("%w: %s", err, tail)
		}

		return err //nolint:wrapcheck // the exec error is the failure; stderr added above when available
	}
}
