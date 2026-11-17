//go:build !windows && !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly

package cmd

import (
	"errors"
	"io"
	"os"
)

func newPromptInput(_ *os.File) (io.ReadCloser, error) {
	return nil, errors.New("interactive input cancellation is unsupported on this platform")
}
