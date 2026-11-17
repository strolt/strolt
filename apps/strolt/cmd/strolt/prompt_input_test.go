package cmd

import (
	"context"
	"io"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestPromptInputPreCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	called := false
	err := withPromptInput(ctx, func(io.ReadCloser) error { called = true; return nil })
	require.ErrorIs(t, err, context.Canceled)
	require.False(t, called)
}

func TestPromptInputOwnedSession(t *testing.T) {
	file, err := os.CreateTemp(t.TempDir(), "prompt-input")
	require.NoError(t, err)
	_, err = file.WriteString("ab")
	require.NoError(t, err)
	_, err = file.Seek(0, io.SeekStart)
	require.NoError(t, err)
	input, err := newPromptInput(file)
	require.NoError(t, err)
	buffer := make([]byte, 1)
	_, err = io.ReadFull(input, buffer)
	require.NoError(t, err)
	require.Equal(t, "a", string(buffer))
	require.NoError(t, input.Close())
	require.NoError(t, input.Close())
	_, err = input.Read(buffer)
	require.ErrorIs(t, err, io.EOF)
	_, err = file.Read(buffer)
	require.NoError(t, err)
	require.Equal(t, "b", string(buffer))
}

func TestPromptInputPipeSession(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()
	input, err := newPromptInput(reader)
	require.NoError(t, err)
	defer func() { _ = input.Close() }()
	_, err = writer.WriteString("one")
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	got, err := io.ReadAll(input)
	require.NoError(t, err)
	require.Equal(t, "one", string(got))
}

func TestPromptInputIdleCancellation(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("console cancellation is exercised by the PTY harness")
	}
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer func() { _ = reader.Close() }()
	defer func() { _ = writer.Close() }()
	original := os.Stdin
	os.Stdin = reader
	defer func() { os.Stdin = original }()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() {
		done <- withPromptInput(ctx, func(input io.ReadCloser) error {
			_, err := input.Read(make([]byte, 1))
			return err
		})
	}()
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("idle input read did not cancel")
	}
}

func TestPromptSelectionRejectsEmptyChoices(t *testing.T) {
	ctx, cancel := context.WithTimeout(t.Context(), time.Second)
	defer cancel()
	selected := "unchanged"
	require.ErrorIs(t, runSelectPrompt(ctx, "choose", nil, &selected, nil), errNoPromptChoices)
	require.Equal(t, "unchanged", selected)
	index := -1
	require.ErrorIs(t, runSelectIndexPrompt(ctx, "snapshot", nil, &index, nil), errNoPromptChoices)
	require.Equal(t, -1, index)
}
