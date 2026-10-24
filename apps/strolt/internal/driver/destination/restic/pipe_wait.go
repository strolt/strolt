package restic

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/util/bufcap"
)

// stderrTailCap bounds how much subprocess stderr is retained on pipe
// streams. restic writes its progress and fatal errors to stderr; a bounded
// tail explains a failed dump/backup without unbounded memory.
const stderrTailCap = 8 * 1024

// waitWithStderr returns a wait function that appends the captured stderr
// tail to a non-zero exit, so a piped operation failure names the actual
// restic error instead of a bare exit status.
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
