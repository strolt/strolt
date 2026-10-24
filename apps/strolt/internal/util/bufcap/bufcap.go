// Package bufcap provides a concurrency-safe bounded byte buffer that keeps
// only the most recent writes. It is used to capture subprocess stderr on
// long-running pipe streams where the total output is unbounded but only the
// tail explains a failure.
package bufcap

import (
	"bytes"
	"sync"
)

// Writer keeps at most cap bytes of everything written to it, discarding the
// oldest data first. A zero or negative cap keeps nothing.
type Writer struct {
	mu  sync.Mutex
	cap int
	buf bytes.Buffer
}

// New returns a Writer that retains the last cap bytes written to it.
func New(cap int) *Writer {
	if cap < 0 {
		cap = 0
	}

	return &Writer{cap: cap}
}

// Write stores p, evicting the oldest bytes once the capacity is exceeded.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.cap == 0 {
		return len(p), nil
	}

	w.buf.Write(p)

	if excess := w.buf.Len() - w.cap; excess > 0 {
		w.buf.Next(excess)
	}

	return len(p), nil //nolint:wrapcheck // io.Writer contract, no error path exists
}

// Tail returns the retained bytes with surrounding whitespace trimmed.
func (w *Writer) Tail() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return string(bytes.TrimSpace(w.buf.Bytes()))
}
