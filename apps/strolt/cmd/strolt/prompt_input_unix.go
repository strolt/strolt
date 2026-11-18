//go:build linux || darwin || freebsd || openbsd || netbsd || dragonfly

package cmd

import (
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

type promptInput struct {
	fd            int
	pollFD        int32
	originalFlags int
	readMu        sync.Mutex
	mu            sync.Mutex
	stopped       bool
	closeStarted  bool
	closeDone     chan struct{}
	closeErr      error
}

func newPromptInput(input *os.File) (io.ReadCloser, error) {
	if input == nil {
		return nil, errors.New("prompt input is nil")
	}
	fd, err := unix.FcntlInt(input.Fd(), unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return nil, fmt.Errorf("duplicate prompt input: %w", err)
	}
	if fd < 0 || fd > math.MaxInt32 {
		_ = unix.Close(fd)
		return nil, errors.New("prompt input descriptor exceeds int32 range")
	}
	flags, err := unix.FcntlInt(uintptr(fd), unix.F_GETFL, 0)
	if err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("get prompt input flags: %w", err)
	}
	if _, err = unix.FcntlInt(uintptr(fd), unix.F_SETFL, flags|unix.O_NONBLOCK); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("set prompt input nonblocking: %w", err)
	}
	return &promptInput{fd: fd, pollFD: int32(fd), originalFlags: flags, closeDone: make(chan struct{})}, nil
}

func (input *promptInput) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	input.readMu.Lock()
	defer input.readMu.Unlock()
	for {
		if input.isStopped() {
			return 0, io.EOF
		}
		polls := []unix.PollFd{{Fd: input.pollFD, Events: unix.POLLIN | unix.POLLHUP | unix.POLLERR}}
		count, err := unix.Poll(polls, int(promptInputPollInterval.Milliseconds()))
		if err != nil {
			if input.isStopped() {
				return 0, io.EOF
			}
			if errors.Is(err, unix.EINTR) {
				continue
			}
			return 0, fmt.Errorf("poll prompt input: %w", err)
		}
		if count == 0 {
			continue
		}
		if input.isStopped() {
			return 0, io.EOF
		}
		if polls[0].Revents&unix.POLLNVAL != 0 {
			return 0, errors.New("poll prompt input: invalid descriptor")
		}
		n, readErr := unix.Read(input.fd, data[:1])
		if input.isStopped() {
			return 0, io.EOF
		}
		if readErr != nil {
			if errors.Is(readErr, unix.EINTR) || errors.Is(readErr, unix.EAGAIN) || errors.Is(readErr, unix.EWOULDBLOCK) {
				if polls[0].Revents&unix.POLLERR != 0 {
					return 0, fmt.Errorf("read prompt input after poll error: %w", readErr)
				}
				continue
			}
			return n, fmt.Errorf("read prompt input: %w", readErr)
		}
		if n == 0 {
			return 0, io.EOF
		}
		return n, nil
	}
}

func (input *promptInput) Close() error {
	input.mu.Lock()
	if input.closeStarted {
		done := input.closeDone
		input.mu.Unlock()
		<-done
		input.mu.Lock()
		err := input.closeErr
		input.mu.Unlock()
		return err
	}
	input.closeStarted = true
	input.stopped = true
	input.mu.Unlock()
	input.readMu.Lock()
	_, flagErr := unix.FcntlInt(uintptr(input.fd), unix.F_SETFL, input.originalFlags)
	closeErr := unix.Close(input.fd)
	input.readMu.Unlock()
	input.mu.Lock()
	input.closeErr = errors.Join(flagErr, closeErr)
	close(input.closeDone)
	err := input.closeErr
	input.mu.Unlock()
	return err
}

func (input *promptInput) isStopped() bool {
	input.mu.Lock()
	defer input.mu.Unlock()
	return input.stopped
}
