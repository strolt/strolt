// Package shared provides common helpers used by strolt services.
package shared

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"syscall"

	"github.com/strolt/strolt/shared/process"
)

// RestartSelf replaces the current process with a fresh instance of its own binary.
func RestartSelf() error {
	self, err := os.Executable()
	if err != nil {
		return fmt.Errorf("get executable path: %w", err)
	}

	args := os.Args
	env := os.Environ()

	// Windows does not support exec syscall.
	if runtime.GOOS == "windows" {
		cmd := process.CommandContext(context.Background(), self, args[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Stdin = os.Stdin
		cmd.Env = env

		if err := cmd.Run(); err != nil {
			return fmt.Errorf("restart process: %w", err)
		}

		os.Exit(0)
	}

	if err := syscall.Exec(self, args, env); err != nil { //nolint:gosec // self is the path of the current executable
		return fmt.Errorf("exec %q: %w", self, err)
	}

	return nil
}
