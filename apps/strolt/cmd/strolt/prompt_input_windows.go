//go:build windows

package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"sync"
	"syscall"
	"time"
	"unicode/utf16"
	"unsafe"

	"golang.org/x/sys/windows"
)

const (
	fileModeInformationClass = 16
	fileTypeDisk             = windows.FILE_TYPE_DISK
	fileTypeChar             = windows.FILE_TYPE_CHAR
	fileTypePipe             = windows.FILE_TYPE_PIPE
	waitObject0              = windows.WAIT_OBJECT_0
	waitTimeout              = uint32(windows.WAIT_TIMEOUT)
)

var (
	promptKernel32      = windows.NewLazySystemDLL("kernel32.dll")
	readConsoleInputW   = promptKernel32.NewProc("ReadConsoleInputW")
	cancelSynchronousIo = promptKernel32.NewProc("CancelSynchronousIo")
)

type promptInput struct {
	readMu          sync.Mutex
	mu              sync.Mutex
	stopped         bool
	closeDone       chan struct{}
	closeErr        error
	handle          windows.Handle
	cancel          windows.Handle
	kind            uint32
	synchronous     bool
	pending         []byte
	repeatBytes     []byte
	repeatRemaining uint16
	activeThread    windows.Handle
	active          bool
	fileOffset      int64
	highSurrogate   uint16
}

type inputRecord struct {
	EventType uint16
	Padding   uint16
	Event     [16]byte
}
type keyEventRecord struct {
	KeyDown         int32
	RepeatCount     uint16
	VirtualKeyCode  uint16
	VirtualScanCode uint16
	UnicodeChar     uint16
	ControlKeyState uint32
}

func newPromptInput(input *os.File) (io.ReadCloser, error) {
	if input == nil {
		return nil, errors.New("prompt input is nil")
	}
	var handle windows.Handle
	if err := windows.DuplicateHandle(windows.CurrentProcess(), windows.Handle(input.Fd()), windows.CurrentProcess(), &handle, 0, false, windows.DUPLICATE_SAME_ACCESS); err != nil {
		return nil, fmt.Errorf("duplicate prompt input: %w", err)
	}
	var cancel windows.Handle
	fail := func(err error) (io.ReadCloser, error) {
		if cancel != 0 {
			err = errors.Join(err, windows.CloseHandle(cancel))
		}
		return nil, errors.Join(err, windows.CloseHandle(handle))
	}
	var err error
	cancel, err = windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return fail(fmt.Errorf("create prompt cancellation event: %w", err))
	}
	kind, err := windows.GetFileType(handle)
	if err != nil {
		return fail(fmt.Errorf("query prompt input type: %w", err))
	}
	var consoleMode uint32
	if windows.GetConsoleMode(handle, &consoleMode) == nil {
		if kind != fileTypeChar {
			return fail(errors.New("unsupported prompt input handle"))
		}
		return &promptInput{handle: handle, cancel: cancel, kind: kind, closeDone: make(chan struct{})}, nil
	}
	if kind != fileTypePipe && kind != fileTypeDisk {
		return fail(errors.New("unsupported prompt input handle"))
	}
	var iosb windows.IO_STATUS_BLOCK
	var mode uint32
	if err := windows.NtQueryInformationFile(handle, &iosb, (*byte)(unsafe.Pointer(&mode)), uint32(unsafe.Sizeof(mode)), fileModeInformationClass); err != nil {
		return fail(fmt.Errorf("query prompt input mode: %w", err))
	}
	p := &promptInput{handle: handle, cancel: cancel, kind: kind, synchronous: mode&(windows.FILE_SYNCHRONOUS_IO_ALERT|windows.FILE_SYNCHRONOUS_IO_NONALERT) != 0, closeDone: make(chan struct{})}
	if kind == fileTypeDisk {
		p.fileOffset, err = currentFileOffset(handle)
		if err != nil {
			return fail(fmt.Errorf("query prompt input offset: %w", err))
		}
	}
	return p, nil
}

func currentFileOffset(handle windows.Handle) (int64, error) {
	var high int32
	low, err := windows.SetFilePointer(handle, 0, &high, windows.FILE_CURRENT)
	return int64(high)<<32 | int64(low), err
}

func setFileOffset(handle windows.Handle, offset int64) error {
	high := int32(offset >> 32)
	_, err := windows.SetFilePointer(handle, int32(offset), &high, windows.FILE_BEGIN)
	return err
}

func (p *promptInput) isStopped() bool { p.mu.Lock(); defer p.mu.Unlock(); return p.stopped }
func (p *promptInput) Read(b []byte) (int, error) {
	if len(b) == 0 {
		return 0, nil
	}
	p.readMu.Lock()
	defer p.readMu.Unlock()
	if p.isStopped() {
		return 0, io.EOF
	}
	if len(p.pending) > 0 {
		return p.drain(b), nil
	}
	if p.kind == fileTypeChar {
		return p.readConsole(b)
	}
	for {
		if p.isStopped() {
			return 0, io.EOF
		}
		var n int
		var err error
		if p.synchronous {
			n, err = p.readSynchronous(b)
		} else {
			n, err = p.readOverlapped(b)
		}
		if p.kind == fileTypePipe && n == 0 && err == nil {
			continue
		}
		return n, err
	}
}

func (p *promptInput) drain(b []byte) int {
	n := copy(b, p.pending)
	p.pending = p.pending[n:]
	if len(p.pending) == 0 && p.repeatRemaining > 0 {
		p.pending = p.repeatBytes
		p.repeatRemaining--
	}
	return n
}

func (p *promptInput) beginRepeat(count uint16) {
	p.repeatBytes = p.pending
	p.repeatRemaining = 0
	if len(p.pending) > 0 && count > 1 {
		p.repeatRemaining = count - 1
	}
}

func (p *promptInput) readConsole(b []byte) (int, error) {
	for {
		if p.isStopped() {
			return 0, io.EOF
		}
		event, err := windows.WaitForMultipleObjects([]windows.Handle{p.cancel, p.handle}, false, windows.INFINITE)
		if err != nil {
			return 0, fmt.Errorf("wait for console input: %w", err)
		}
		if event == waitObject0 {
			return 0, io.EOF
		}
		if event != waitObject0+1 {
			return 0, fmt.Errorf("unexpected console wait result %d", event)
		}
		if p.isStopped() {
			return 0, io.EOF
		}
		var record inputRecord
		var count uint32
		result, _, callErr := readConsoleInputW.Call(uintptr(p.handle), uintptr(unsafe.Pointer(&record)), 1, uintptr(unsafe.Pointer(&count)))
		runtime.KeepAlive(&record)
		runtime.KeepAlive(&count)
		if result == 0 {
			if callErr == syscall.Errno(0) {
				callErr = syscall.EINVAL
			}
			return 0, fmt.Errorf("read console input: %w", callErr)
		}
		if count == 0 {
			continue
		}
		p.translate(record)
		if len(p.pending) > 0 {
			return p.drain(b), nil
		}
	}
}

func (p *promptInput) translate(record inputRecord) {
	if record.EventType != 1 {
		return
	}
	key := *(*keyEventRecord)(unsafe.Pointer(&record.Event[0]))
	if key.KeyDown == 0 {
		return
	}
	if key.UnicodeChar == 0 {
		switch key.VirtualKeyCode {
		case 0x25:
			p.pending = []byte("\x1b[D")
		case 0x26:
			p.pending = []byte("\x1b[A")
		case 0x27:
			p.pending = []byte("\x1b[C")
		case 0x28:
			p.pending = []byte("\x1b[B")
		case 0x2e:
			p.pending = []byte("\x1b[3~")
		case 0x24:
			p.pending = []byte("\x1b[H")
		case 0x23:
			p.pending = []byte("\x1b[F")
		}
		p.beginRepeat(key.RepeatCount)
		return
	}
	c := rune(key.UnicodeChar)
	if key.UnicodeChar >= 0xd800 && key.UnicodeChar <= 0xdbff {
		p.highSurrogate = key.UnicodeChar
		return
	}
	if key.UnicodeChar >= 0xdc00 && key.UnicodeChar <= 0xdfff && p.highSurrogate != 0 {
		c = utf16.DecodeRune(rune(p.highSurrogate), c)
	}
	p.highSurrogate = 0
	if key.ControlKeyState&0x0003 != 0 {
		p.pending = append([]byte{27}, []byte(string(c))...)
	} else {
		p.pending = []byte(string(c))
	}
	p.beginRepeat(key.RepeatCount)
}

func (p *promptInput) readSynchronous(b []byte) (int, error) {
	runtime.LockOSThread()
	thread, err := windows.OpenThread(windows.THREAD_TERMINATE, false, windows.GetCurrentThreadId())
	if err != nil {
		runtime.UnlockOSThread()
		return 0, fmt.Errorf("open input read thread: %w", err)
	}
	p.mu.Lock()
	if p.stopped {
		p.mu.Unlock()
		_ = windows.CloseHandle(thread)
		runtime.UnlockOSThread()
		return 0, io.EOF
	}
	p.activeThread, p.active = thread, true
	p.mu.Unlock()
	var n uint32
	var readErr error
	if p.isStopped() {
		readErr = io.EOF
	} else {
		readErr = windows.ReadFile(p.handle, b[:1], &n, nil)
	}
	if n > 0 && errors.Is(readErr, windows.ERROR_MORE_DATA) {
		readErr = nil
	}
	if n == 0 && readErr == windows.ERROR_OPERATION_ABORTED && p.isStopped() {
		readErr = io.EOF
	}
	if readErr == windows.ERROR_BROKEN_PIPE || readErr == windows.ERROR_HANDLE_EOF {
		readErr = io.EOF
	}
	if n == 0 && readErr == nil && p.kind == fileTypeDisk {
		readErr = io.EOF
	}
	if n > 0 && p.kind == fileTypeDisk {
		p.fileOffset += int64(n)
		if offsetErr := setFileOffset(p.handle, p.fileOffset); offsetErr != nil {
			readErr = errors.Join(readErr, offsetErr)
		}
	}
	p.mu.Lock()
	p.active, p.activeThread = false, 0
	closeErr := windows.CloseHandle(thread)
	p.mu.Unlock()
	runtime.UnlockOSThread()
	if closeErr != nil {
		readErr = errors.Join(readErr, closeErr)
	}
	return int(n), readErr
}

func (p *promptInput) readOverlapped(b []byte) (count int, resultErr error) {
	event, err := windows.CreateEvent(nil, 1, 0, nil)
	if err != nil {
		return 0, fmt.Errorf("create input completion event: %w", err)
	}
	defer func() {
		if closeErr := windows.CloseHandle(event); closeErr != nil {
			resultErr = errors.Join(resultErr, closeErr)
		}
	}()
	var ov windows.Overlapped
	ov.HEvent = event
	if p.kind == fileTypeDisk {
		ov.Offset = uint32(p.fileOffset)
		ov.OffsetHigh = uint32(uint64(p.fileOffset) >> 32)
	}
	var one [1]byte
	var n uint32
	err = windows.ReadFile(p.handle, one[:], &n, &ov)
	if err != nil && err != windows.ERROR_IO_PENDING && err != windows.ERROR_MORE_DATA {
		if err == windows.ERROR_BROKEN_PIPE || err == windows.ERROR_HANDLE_EOF {
			return 0, io.EOF
		}
		return 0, err
	}
	for {
		ready, waitErr := windows.WaitForMultipleObjects([]windows.Handle{p.cancel, event}, false, windows.INFINITE)
		if waitErr != nil || (ready != waitObject0 && ready != waitObject0+1) {
			if waitErr == nil {
				waitErr = fmt.Errorf("unexpected input wait result %d", ready)
			}
			cancelErr := windows.CancelIoEx(p.handle, &ov)
			if cancelErr != nil && cancelErr != windows.ERROR_NOT_FOUND {
				waitErr = errors.Join(waitErr, cancelErr)
			}
			return 0, errors.Join(waitErr, windows.GetOverlappedResult(p.handle, &ov, &n, true))
		}
		if ready == waitObject0 {
			cancelErr := windows.CancelIoEx(p.handle, &ov)
			resultErr = windows.GetOverlappedResult(p.handle, &ov, &n, true)
			if cancelErr != nil && cancelErr != windows.ERROR_NOT_FOUND {
				return 0, errors.Join(cancelErr, resultErr)
			}
			if n == 0 {
				if resultErr != nil && resultErr != windows.ERROR_OPERATION_ABORTED {
					return 0, resultErr
				}
				return 0, io.EOF
			}
			break
		}
		resultErr = windows.GetOverlappedResult(p.handle, &ov, &n, true)
		break
	}
	if n > 0 && resultErr == windows.ERROR_MORE_DATA {
		resultErr = nil
	}
	if n == 0 {
		if resultErr != nil {
			if resultErr == windows.ERROR_BROKEN_PIPE || resultErr == windows.ERROR_HANDLE_EOF {
				return 0, io.EOF
			}
			return 0, resultErr
		}
		if p.kind == fileTypePipe {
			return 0, nil
		}
		return 0, io.EOF
	}
	if p.kind == fileTypeDisk {
		p.fileOffset += int64(n)
		if offsetErr := setFileOffset(p.handle, p.fileOffset); offsetErr != nil {
			resultErr = errors.Join(resultErr, offsetErr)
		}
	}
	b[0] = one[0]
	return int(n), resultErr
}

func (p *promptInput) Close() error {
	p.mu.Lock()
	if p.stopped {
		done := p.closeDone
		p.mu.Unlock()
		<-done
		p.mu.Lock()
		err := p.closeErr
		p.mu.Unlock()
		return err
	}
	p.stopped = true
	eventErr := windows.SetEvent(p.cancel)
	p.mu.Unlock()
	var cancelErr error
	ticker := time.NewTicker(promptInputPollInterval)
	defer ticker.Stop()
	for {
		p.mu.Lock()
		if !p.active {
			p.mu.Unlock()
			break
		}
		result, _, callErr := cancelSynchronousIo.Call(uintptr(p.activeThread))
		if result == 0 && callErr != windows.ERROR_NOT_FOUND && cancelErr == nil {
			if callErr == syscall.Errno(0) {
				callErr = syscall.EINVAL
			}
			cancelErr = fmt.Errorf("cancel synchronous input: %w", callErr)
		}
		p.mu.Unlock()
		<-ticker.C
	}
	p.readMu.Lock()
	defer p.readMu.Unlock()
	closeErr := errors.Join(eventErr, cancelErr, windows.CloseHandle(p.cancel), windows.CloseHandle(p.handle))
	p.mu.Lock()
	p.closeErr = closeErr
	close(p.closeDone)
	p.mu.Unlock()
	return closeErr
}
