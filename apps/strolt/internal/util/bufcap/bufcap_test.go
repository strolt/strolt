package bufcap

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWriter_keepsLastBytes(t *testing.T) {
	w := New(8)

	for _, chunk := range []string{"aaaa", "bbbb", "cccc"} {
		n, err := w.Write([]byte(chunk))
		require.NoError(t, err)
		require.Equal(t, len(chunk), n)
	}

	assert.Equal(t, "bbbbcccc", w.Tail(), "the oldest bytes must be evicted first")
}

func TestWriter_trimsSurroundingWhitespace(t *testing.T) {
	w := New(64)
	_, err := w.Write([]byte("  fatal: boom\n\n"))
	require.NoError(t, err)

	assert.Equal(t, "fatal: boom", w.Tail())
}

func TestWriter_zeroCapKeepsNothing(t *testing.T) {
	w := New(0)

	n, err := w.Write([]byte("data"))
	require.NoError(t, err)
	require.Equal(t, 4, n)

	assert.Empty(t, w.Tail())
}

func TestWriter_concurrentWritesAreSafe(t *testing.T) {
	w := New(1024)

	done := make(chan struct{})

	for range 8 {
		go func() {
			defer func() { done <- struct{}{} }()

			for range 100 {
				_, _ = w.Write([]byte(strings.Repeat("x", 64)))
			}
		}()
	}

	for range 8 {
		<-done
	}

	assert.Len(t, w.Tail(), 1024)
}
