package pg

import (
	"fmt"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
	"github.com/strolt/strolt/shared/process"
)

// stderrTailCap bounds how much subprocess stderr is retained on pipe
// streams. psql/pg_restore print the failing statement's error to stderr; a
// bounded tail is enough to explain an abort without risking unbounded
// memory on chatty runs.
const stderrTailCap = 8 * 1024

// waitWithStderr appends the captured stderr tail to a non-zero exit.
func waitWithStderr(cmd *process.Cmd, stderr *bufcap.Writer) func() error {
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
