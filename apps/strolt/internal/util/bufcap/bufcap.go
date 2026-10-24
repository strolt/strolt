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
// oldest data first. A zero or negative limit keeps nothing.
type Writer struct {
	mu    sync.Mutex
	limit int
	buf   bytes.Buffer
}

// New returns a Writer that retains the last limit bytes written to it.
func New(limit int) *Writer {
	if limit < 0 {
		limit = 0
	}

	return &Writer{limit: limit}
}

// Write stores p, evicting the oldest bytes once the capacity is exceeded.
func (w *Writer) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.limit == 0 {
		return len(p), nil
	}

	w.buf.Write(p)

	if excess := w.buf.Len() - w.limit; excess > 0 {
		w.buf.Next(excess)
	}

	return len(p), nil
}

// Tail returns the retained bytes with surrounding whitespace trimmed.
func (w *Writer) Tail() string {
	w.mu.Lock()
	defer w.mu.Unlock()

	return string(bytes.TrimSpace(w.buf.Bytes()))
}
