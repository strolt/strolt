package task

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shortWriter accepts only the first limit bytes of every Write and then
// reports io.ErrShortWrite, mimicking a destination that stops mid-stream.
type shortWriter struct {
	limit int
}

func (w *shortWriter) Write(p []byte) (int, error) {
	if len(p) <= w.limit {
		return len(p), nil
	}

	return w.limit, io.ErrShortWrite
}

func TestCountingWriter(t *testing.T) {
	t.Run("counts bytes and forwards them to the underlying writer", func(t *testing.T) {
		var buf bytes.Buffer

		cw := &countingWriter{w: &buf}

		payload := []byte("hello world")

		n, err := cw.Write(payload)
		require.NoError(t, err)
		assert.Equal(t, len(payload), n)

		n, err = cw.Write(payload)
		require.NoError(t, err)
		assert.Equal(t, len(payload), n)

		assert.Equal(t, uint64(2*len(payload)), cw.written)
		assert.Equal(t, "hello worldhello world", buf.String())
	})

	t.Run("counts only the bytes accepted on a short write", func(t *testing.T) {
		cw := &countingWriter{w: &shortWriter{limit: 3}}

		n, err := cw.Write([]byte("hello"))
		require.ErrorIs(t, err, io.ErrShortWrite)
		assert.Equal(t, 3, n)
		assert.Equal(t, uint64(3), cw.written)
	})

	t.Run("io.Copy through the writer records the full stream size", func(t *testing.T) {
		var buf bytes.Buffer

		cw := &countingWriter{w: &buf}

		payload := bytes.Repeat([]byte("x"), 4096)

		written, err := io.Copy(cw, bytes.NewReader(payload))
		require.NoError(t, err)
		assert.Equal(t, int64(len(payload)), written)
		assert.Equal(t, uint64(len(payload)), cw.written)
	})

	t.Run("propagates an underlying writer error", func(t *testing.T) {
		wantErr := errors.New("boom")

		cw := &countingWriter{w: writerFunc(func(p []byte) (int, error) {
			return 0, wantErr
		})}

		n, err := cw.Write([]byte("data"))
		require.ErrorIs(t, err, wantErr)
		assert.Zero(t, n)
		assert.Zero(t, cw.written)
	})
}

type writerFunc func(p []byte) (int, error)

func (f writerFunc) Write(p []byte) (int, error) { return f(p) }
