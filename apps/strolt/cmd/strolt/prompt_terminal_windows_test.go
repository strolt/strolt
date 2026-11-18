//go:build windows

package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
	"unsafe"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows"
)

const promptPipeBufferSize = 4096

func createNamedPromptPipeMode(t *testing.T, messageMode, overlapped bool) (*os.File, *os.File) {
	t.Helper()
	name := `\\.\pipe\strolt-prompt-` + strings.ReplaceAll(t.Name(), "/", "-") + fmt.Sprint(time.Now().UnixNano())
	pipeMode := uint32(windows.PIPE_TYPE_BYTE | windows.PIPE_READMODE_BYTE | windows.PIPE_WAIT)
	if messageMode {
		pipeMode = windows.PIPE_TYPE_MESSAGE | windows.PIPE_READMODE_MESSAGE | windows.PIPE_WAIT
	}
	server, err := windows.CreateNamedPipe(windows.StringToUTF16Ptr(name), windows.PIPE_ACCESS_OUTBOUND, pipeMode, 1, promptPipeBufferSize, promptPipeBufferSize, 0, nil)
	require.NoError(t, err)
	writer := os.NewFile(uintptr(server), "prompt-pipe-writer")
	t.Cleanup(func() { _ = writer.Close() })
	access := uint32(windows.GENERIC_READ)
	if messageMode {
		access |= windows.FILE_WRITE_ATTRIBUTES
	}
	var flags uint32
	if overlapped {
		flags = windows.FILE_FLAG_OVERLAPPED
	}
	client, err := windows.CreateFile(windows.StringToUTF16Ptr(name), access, 0, nil, windows.OPEN_EXISTING, flags, 0)
	require.NoError(t, err)
	reader := os.NewFile(uintptr(client), "prompt-pipe-reader")
	t.Cleanup(func() { _ = reader.Close() })
	if messageMode {
		mode := uint32(windows.PIPE_READMODE_MESSAGE)
		require.NoError(t, windows.SetNamedPipeHandleState(client, &mode, nil, nil))
	}
	return reader, writer
}

func writePromptPipeMessage(t *testing.T, writer *os.File, data []byte) {
	t.Helper()
	var n uint32
	err := windows.WriteFile(windows.Handle(writer.Fd()), data, &n, nil)
	require.NoError(t, err)
	require.Equal(t, uint32(len(data)), n)
}

func waitPromptResult[T any](t *testing.T, ch <-chan T, label string) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		var zero T
		t.Fatalf("timed out waiting for %s", label)
		return zero
	}
}

type promptReadResult struct {
	n    int
	err  error
	data string
}

func readAllPromptInput(t *testing.T, input io.ReadCloser) ([]byte, error) {
	t.Helper()
	type result struct {
		data []byte
		err  error
	}
	done := make(chan result, 1)
	joined := false
	t.Cleanup(func() {
		_ = input.Close()
		if !joined {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Error("finite input reader did not join")
			}
		}
	})
	go func() { data, err := io.ReadAll(input); done <- result{data, err} }()
	got := waitPromptResult(t, done, "finite input EOF")
	joined = true
	return got.data, got.err
}

func TestPromptWindowsSynchronousPipeAndDisk(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	t.Cleanup(func() { _ = r.Close(); _ = w.Close() })
	input, err := newPromptInput(r)
	require.NoError(t, err)
	t.Cleanup(func() { _ = input.Close() })
	require.True(t, input.(*promptInput).synchronous)
	writerDone := make(chan error, 1)
	go func() {
		_, e := w.WriteString("pipe")
		if e == nil {
			e = w.Close()
		}
		writerDone <- e
	}()
	got, err := readAllPromptInput(t, input)
	require.NoError(t, err)
	require.Equal(t, "pipe", string(got))
	require.NoError(t, waitPromptResult(t, writerDone, "pipe writer"))
	require.NoError(t, input.Close())
	require.NoError(t, r.Close())
	path := filepath.Join(t.TempDir(), "input")
	require.NoError(t, os.WriteFile(path, []byte("disk"), 0o600))
	file, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	disk, err := newPromptInput(file)
	require.NoError(t, err)
	t.Cleanup(func() { _ = disk.Close() })
	require.Equal(t, fileTypeDisk, disk.(*promptInput).kind)
	require.True(t, disk.(*promptInput).synchronous)
	b := make([]byte, 1)
	n, err := disk.Read(b)
	require.Equal(t, 1, n)
	require.NoError(t, err)
	require.Equal(t, "d", string(b))
	got, err = readAllPromptInput(t, disk)
	require.NoError(t, err)
	require.Equal(t, "isk", string(got))
	_, err = disk.Read(b)
	require.Same(t, io.EOF, err)
	require.NoError(t, disk.Close())
	require.NoError(t, file.Close())
}

func TestPromptWindowsSynchronousDiskEmptyAndBufferedEOF(t *testing.T) {
	path := filepath.Join(t.TempDir(), "empty")
	require.NoError(t, os.WriteFile(path, nil, 0o600))
	file, err := os.Open(path)
	require.NoError(t, err)
	t.Cleanup(func() { _ = file.Close() })
	input, err := newPromptInput(file)
	require.NoError(t, err)
	t.Cleanup(func() { _ = input.Close() })
	require.Equal(t, fileTypeDisk, input.(*promptInput).kind)
	require.True(t, input.(*promptInput).synchronous)
	var one [1]byte
	for range 2 {
		n, err := input.Read(one[:])
		require.Zero(t, n)
		require.Same(t, io.EOF, err)
	}
	data, err := readAllPromptInput(t, input)
	require.NoError(t, err)
	require.Empty(t, data)
	_, err = bufio.NewReader(input).ReadByte()
	require.Same(t, io.EOF, err)
}

func TestPromptWindowsMessagePipeReaderContract(t *testing.T) {
	for _, overlapped := range []bool{false, true} {
		t.Run(map[bool]string{false: "synchronous", true: "overlapped"}[overlapped], func(t *testing.T) {
			r, w := createNamedPromptPipeMode(t, true, overlapped)
			input, err := newPromptInput(r)
			require.NoError(t, err)
			writePromptPipeMessage(t, w, []byte("y\r"))
			var first [1]byte
			n, readErr := input.Read(first[:])
			require.Equal(t, 1, n)
			require.NoError(t, readErr)
			require.Equal(t, byte('y'), first[0])
			reader := bufio.NewReader(input)
			line, err := reader.ReadString('\r')
			require.NoError(t, err)
			require.Equal(t, "\r", line)
			require.NoError(t, w.Close())
			_, err = reader.ReadByte()
			require.Same(t, io.EOF, err)
			require.NoError(t, input.Close())
		})
	}
}

func TestPromptWindowsOverlappedEmptyMessages(t *testing.T) {
	for _, overlapped := range []bool{false, true} {
		t.Run(fmt.Sprintf("overlapped-%t", overlapped), func(t *testing.T) { testPromptWindowsEmptyMessages(t, overlapped) })
	}
}

func testPromptWindowsEmptyMessages(t *testing.T, overlapped bool) {
	for _, count := range []int{1, 128} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			r, w := createNamedPromptPipeMode(t, true, overlapped)
			input, err := newPromptInput(r)
			require.NoError(t, err)
			backend := input.(*promptInput)
			require.Equal(t, !overlapped, backend.synchronous)
			writerDone := make(chan error, 1)
			setupDone := make(chan promptReadResult, 1)
			setupStarted, writerStarted := false, false
			var readerDone, readByteDone chan promptReadResult
			t.Cleanup(func() {
				_ = input.Close()
				_ = w.Close()
				if setupStarted {
					select {
					case <-setupDone:
					case <-time.After(5 * time.Second):
						t.Errorf("isolated native read did not stop during cleanup")
					}
				}
				for _, item := range []struct {
					started bool
					channel <-chan promptReadResult
					name    string
				}{{readerDone != nil, readerDone, "reader goroutine"}, {readByteDone != nil, readByteDone, "buffered EOF reader"}} {
					if item.started {
						select {
						case <-item.channel:
						case <-time.After(5 * time.Second):
							t.Errorf("%s did not stop during cleanup", item.name)
						}
					}
				}
				if writerStarted {
					select {
					case <-writerDone:
					case <-time.After(5 * time.Second):
						t.Errorf("writer goroutine did not stop during cleanup")
					}
				}
			})
			writeEmpty := func() {
				t.Helper()
				var written uint32
				require.NoError(t, windows.WriteFile(windows.Handle(w.Fd()), nil, &written, nil))
				require.Zero(t, written)
			}
			writeEmpty()
			setupStarted = true
			go func() {
				backend.readMu.Lock()
				var n int
				var readErr error
				if overlapped {
					n, readErr = backend.readOverlapped(make([]byte, 1))
				} else {
					n, readErr = backend.readSynchronous(make([]byte, 1))
				}
				backend.readMu.Unlock()
				setupDone <- promptReadResult{n: n, err: readErr}
			}()
			setup := waitPromptResult(t, setupDone, "isolated native empty read")
			setupStarted = false
			require.NoError(t, setup.err)
			require.Zero(t, setup.n)
			readerDone = make(chan promptReadResult, 1)
			buffered := bufio.NewReader(input)
			go func() {
				line, readErr := buffered.ReadString('\r')
				readerDone <- promptReadResult{n: len(line), err: readErr, data: line}
			}()
			writerStarted = true
			go func() {
				for range count {
					var written uint32
					if writeErr := windows.WriteFile(windows.Handle(w.Fd()), nil, &written, nil); writeErr != nil {
						writerDone <- writeErr
						return
					}
					if written != 0 {
						writerDone <- fmt.Errorf("empty write delivered %d bytes", written)
						return
					}
				}
				var written uint32
				writerDone <- windows.WriteFile(windows.Handle(w.Fd()), []byte("y\r"), &written, nil)
			}()
			got := waitPromptResult(t, readerDone, "buffered ReadString")
			readerDone = nil
			require.NoError(t, got.err)
			require.Equal(t, "y\r", got.data)
			writeErr := waitPromptResult(t, writerDone, "empty-message writer")
			writerStarted = false
			require.NoError(t, writeErr)
			require.NoError(t, w.Close())
			readByteDone = make(chan promptReadResult, 1)
			go func() {
				b, readErr := buffered.ReadByte()
				readByteDone <- promptReadResult{n: int(b), err: readErr}
			}()
			byteResult := waitPromptResult(t, readByteDone, "buffered EOF")
			readByteDone = nil
			require.Same(t, io.EOF, byteResult.err)
			require.NoError(t, input.Close())
		})
	}
}

func TestPromptWindowsOverlappedEmptyMessageCancellation(t *testing.T) {
	for _, overlapped := range []bool{false, true} {
		t.Run(fmt.Sprintf("overlapped-%t", overlapped), func(t *testing.T) { testPromptWindowsEmptyMessageCancellation(t, overlapped) })
	}
}

func testPromptWindowsEmptyMessageCancellation(t *testing.T, overlapped bool) {
	r, w := createNamedPromptPipeMode(t, true, overlapped)
	input, err := newPromptInput(r)
	require.NoError(t, err)
	backend := input.(*promptInput)
	require.Equal(t, !overlapped, backend.synchronous)
	setupDone := make(chan promptReadResult, 1)
	readDone := make(chan promptReadResult, 1)
	closeDone := make(chan error, 1)
	setupStarted, readStarted, closeStarted := false, false, false
	t.Cleanup(func() {
		_ = input.Close()
		_ = w.Close()
		for _, item := range []struct {
			started bool
			channel <-chan promptReadResult
			name    string
		}{{setupStarted, setupDone, "isolated native read"}, {readStarted, readDone, "cancelled reader"}} {
			if item.started {
				select {
				case <-item.channel:
				case <-time.After(5 * time.Second):
					t.Errorf("%s did not stop during cleanup", item.name)
				}
			}
		}
		if closeStarted {
			select {
			case <-closeDone:
			case <-time.After(5 * time.Second):
				t.Errorf("input Close did not stop during cleanup")
			}
		}
	})
	var written uint32
	require.NoError(t, windows.WriteFile(windows.Handle(w.Fd()), nil, &written, nil))
	require.Zero(t, written)
	setupStarted = true
	go func() {
		var n int
		var readErr error
		backend.readMu.Lock()
		if overlapped {
			n, readErr = backend.readOverlapped(make([]byte, 1))
		} else {
			n, readErr = backend.readSynchronous(make([]byte, 1))
		}
		backend.readMu.Unlock()
		setupDone <- promptReadResult{n: n, err: readErr}
	}()
	setup := waitPromptResult(t, setupDone, "isolated native empty read")
	setupStarted = false
	require.NoError(t, setup.err)
	require.Zero(t, setup.n)
	readStarted = true
	go func() { b := make([]byte, 1); n, e := input.Read(b); readDone <- promptReadResult{n: n, err: e} }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		owned := !backend.readMu.TryLock()
		if !owned {
			backend.readMu.Unlock()
		}
		if owned {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("Read did not acquire its read session")
		}
		time.Sleep(time.Millisecond)
	}
	closeStarted = true
	go func() { closeDone <- input.Close() }()
	closeErr := waitPromptResult(t, closeDone, "input Close")
	closeStarted = false
	require.NoError(t, closeErr)
	got := waitPromptResult(t, readDone, "cancelled Read")
	readStarted = false
	require.Zero(t, got.n)
	require.Same(t, io.EOF, got.err)
	require.NoError(t, w.Close())
}

func TestPromptWindowsSynchronousReadCancellation(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	input, err := newPromptInput(r)
	require.NoError(t, err)
	readDone := make(chan promptReadResult, 1)
	go func() { b := make([]byte, 1); n, e := input.Read(b); readDone <- promptReadResult{n: n, err: e} }()
	deadline := time.Now().Add(5 * time.Second)
	for {
		p := input.(*promptInput)
		p.mu.Lock()
		active := p.active
		p.mu.Unlock()
		if active {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("synchronous read did not become active")
		}
		time.Sleep(time.Millisecond)
	}
	closeDone := make(chan error, 1)
	go func() { closeDone <- input.Close() }()
	require.NoError(t, waitPromptResult(t, closeDone, "synchronous Close"))
	got := waitPromptResult(t, readDone, "cancelled synchronous Read")
	require.Zero(t, got.n)
	require.Same(t, io.EOF, got.err)
	require.NoError(t, r.Close())
	require.NoError(t, w.Close())
}

func TestPromptWindowsNativeTerminalStateRestored(t *testing.T) {
	if os.Getenv("STROLT_TEST_WINDOWS_PROMPT_CHILD") == "child" {
		runPromptWindowsConsoleChild(t)
		return
	}
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestPromptWindowsNativeTerminalStateRestored$")
	command.Env = append(os.Environ(), "STROLT_TEST_WINDOWS_PROMPT_CHILD=child", "STROLT_TEST_WINDOWS_PROMPT_CONSOLE=1")
	output, err := command.CombinedOutput()
	require.NoError(t, err, string(output))
}

func runPromptWindowsConsoleChild(t *testing.T) {
	free := windows.NewLazySystemDLL("kernel32.dll").NewProc("FreeConsole")
	_, _, _ = free.Call()
	defer func() { _, _, _ = free.Call() }()
	alloc := windows.NewLazySystemDLL("kernel32.dll").NewProc("AllocConsole")
	r, _, callErr := alloc.Call()
	if r == 0 {
		t.Fatalf("AllocConsole: %v", callErr)
	}
	openConsole := func(name string, access uint32) windows.Handle {
		h, err := windows.CreateFile(windows.StringToUTF16Ptr(name), access, windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE, nil, windows.OPEN_EXISTING, 0, 0)
		require.NoError(t, err)
		return h
	}
	in := openConsole("CONIN$", windows.GENERIC_READ|windows.GENERIC_WRITE)
	out := openConsole("CONOUT$", windows.GENERIC_READ|windows.GENERIC_WRITE)
	inFile, outFile := os.NewFile(uintptr(in), "CONIN$"), os.NewFile(uintptr(out), "CONOUT$")
	t.Cleanup(func() { _ = inFile.Close(); _ = outFile.Close() })
	require.NoError(t, windows.SetStdHandle(windows.STD_INPUT_HANDLE, in))
	require.NoError(t, windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, out))
	require.NoError(t, windows.SetStdHandle(windows.STD_ERROR_HANDLE, out))
	originalIn, originalOut, originalErr := os.Stdin, os.Stdout, os.Stderr
	os.Stdin, os.Stdout, os.Stderr = inFile, outFile, outFile
	defer func() {
		os.Stdin, os.Stdout, os.Stderr = originalIn, originalOut, originalErr
		_ = windows.SetStdHandle(windows.STD_INPUT_HANDLE, windows.Handle(originalIn.Fd()))
		_ = windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, windows.Handle(originalOut.Fd()))
		_ = windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(originalErr.Fd()))
	}()
	var mode uint32
	require.NoError(t, windows.GetConsoleMode(in, &mode))
	type cursorInfo struct {
		Size    uint32
		Visible int32
	}
	getCursor := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleCursorInfo")
	cursor := func() cursorInfo {
		var info cursorInfo
		result, _, e := getCursor.Call(uintptr(out), uintptr(unsafe.Pointer(&info)))
		require.NotZero(t, result, "GetConsoleCursorInfo: %v", e)
		return info
	}
	initialCursor := cursor()
	input, err := newPromptInput(os.Stdin)
	require.NoError(t, err)
	defer func() { _ = input.Close() }()
	writeRecords := func(records ...inputRecord) {
		t.Helper()
		proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("WriteConsoleInputW")
		var written uint32
		result, _, e := proc.Call(uintptr(in), uintptr(unsafe.Pointer(&records[0])), uintptr(len(records)), uintptr(unsafe.Pointer(&written)))
		runtime.KeepAlive(records)
		require.NotZero(t, result, "WriteConsoleInputW: %v", e)
		require.Equal(t, uint32(len(records)), written)
	}
	keyWithRepeat := func(vk, char uint16, down int32, eventType uint16, repeat uint16) inputRecord {
		var record inputRecord
		record.EventType = eventType
		value := keyEventRecord{KeyDown: down, RepeatCount: repeat, VirtualKeyCode: vk, UnicodeChar: char}
		*(*keyEventRecord)(unsafe.Pointer(&record.Event[0])) = value
		return record
	}
	key := func(vk, char uint16, down int32, eventType uint16) inputRecord {
		return keyWithRepeat(vk, char, down, eventType, 1)
	}

	writeRecords(keyWithRepeat(0x25, 0, 1, 1, 3))
	expectedRepeat := "\x1b[D\x1b[D\x1b[D"
	repeatDone := make(chan promptReadResult, 1)
	repeatJoined := false
	t.Cleanup(func() {
		_ = input.Close()
		if !repeatJoined {
			_ = waitPromptResult(t, repeatDone, "repeated console read cleanup")
		}
	})
	go func() {
		var repeated []byte
		var readErr error
		for len(repeated) < len(expectedRepeat) {
			var part [2]byte
			n, err := input.Read(part[:])
			repeated = append(repeated, part[:n]...)
			if err != nil {
				readErr = err
				break
			}
		}
		repeatDone <- promptReadResult{n: len(repeated), err: readErr, data: string(repeated)}
	}()
	repeated := waitPromptResult(t, repeatDone, "repeated console keys")
	repeatJoined = true
	require.NoError(t, repeated.err)
	require.Equal(t, expectedRepeat, repeated.data)
	writeRecords(keyWithRepeat(0, 'y', 1, 1, 2), key(0x0d, '\r', 1, 1))
	confirmed, confirmErr := runConfirmForWindowsTest(t.Context())
	require.NoError(t, confirmErr)
	require.False(t, confirmed)

	writeRecords(key(0x25, 0, 1, 1))
	got := make([]byte, 1)
	n, err := input.Read(got)
	require.NoError(t, err)
	require.Equal(t, 1, n)
	require.Equal(t, "\x1b", string(got))
	remaining := make([]byte, 2)
	n, err = input.Read(remaining)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, "[D", string(remaining))
	writeRecords(key(0, 0xd83d, 1, 1), key(0, 0xde00, 1, 1))
	short := make([]byte, 2)
	n, err = input.Read(short)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, []byte{0xf0, 0x9f}, short)
	remaining = make([]byte, 2)
	n, err = input.Read(remaining)
	require.NoError(t, err)
	require.Equal(t, 2, n)
	require.Equal(t, []byte{0x98, 0x80}, remaining)
	writeRecords(key(0, 'x', 1, 2))
	readDone := make(chan promptReadResult, 1)
	go func() { b := make([]byte, 1); n, e := input.Read(b); readDone <- promptReadResult{n: n, err: e} }()
	closeDone := make(chan error, 1)
	go func() { closeDone <- input.Close() }()
	require.NoError(t, waitPromptResult(t, closeDone, "console input Close"))
	readResult := waitPromptResult(t, readDone, "console cancellation")
	require.Zero(t, readResult.n)
	require.Same(t, io.EOF, readResult.err)
	writeRecords(key(0x28, 0, 1, 1), key(0x0d, '\r', 1, 1))
	selected := ""
	formDone := make(chan error, 1)
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	go func() { formDone <- runSelectPrompt(ctx, "choose", []string{"first", "second"}, &selected, nil) }()
	select {
	case err := <-formDone:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("production prompt did not finish")
	}
	require.Equal(t, "second", selected)
	var restored uint32
	require.NoError(t, windows.GetConsoleMode(in, &restored))
	require.Equal(t, mode, restored)
	require.Equal(t, initialCursor, cursor())
}

func TestPromptWindowsContextCancellation(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer reader.Close()
	defer writer.Close()
	original := os.Stdin
	os.Stdin = reader
	defer func() { os.Stdin = original }()
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() { selected := ""; done <- runSelectPrompt(ctx, "choose", []string{"one", "two"}, &selected, nil) }()
	cancel()
	select {
	case err := <-done:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("prompt input did not cancel")
	}
}

func TestPromptWindowsEmptyInputIsNotConfirmation(t *testing.T) {
	reader, writer, err := os.Pipe()
	require.NoError(t, err)
	defer reader.Close()
	defer writer.Close()
	original := os.Stdin
	os.Stdin = reader
	defer func() { os.Stdin = original }()
	require.NoError(t, writer.Close())
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	confirmed, err := runConfirmForWindowsTest(ctx)
	require.False(t, confirmed)
	require.Error(t, err, "empty input must not be interpreted as a negative confirmation")
}

func runConfirmForWindowsTest(ctx context.Context) (bool, error) {
	command := &cobra.Command{}
	command.SetContext(ctx)
	return NewPrompt(command).AskIsConfirm()
}
