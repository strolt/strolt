package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"charm.land/huh/v2"

	"github.com/charmbracelet/x/term"
)

const promptInputPollInterval = 25 * time.Millisecond

func withPromptInput(ctx context.Context, run func(io.ReadCloser) error) error {
	if ctx == nil {
		return errors.New("prompt context is nil")
	}
	if err := ctx.Err(); err != nil {
		return errors.Join(err, context.Cause(ctx))
	}
	input, err := newPromptInput(os.Stdin)
	if err != nil {
		return err
	}

	callbackDone := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		_ = input.Close()
		close(callbackDone)
	})
	runErr := run(input)
	if !stop() {
		<-callbackDone
	}
	closeErr := input.Close()
	if err := ctx.Err(); err != nil {
		return errors.Join(err, context.Cause(ctx), runErr, closeErr)
	}
	return errors.Join(runErr, closeErr)
}

func runPromptAnswer(ctx context.Context, answers *promptAnswerState, run func(context.Context, io.Reader) error) error {
	return withPromptInput(ctx, func(raw io.ReadCloser) error {
		var state *term.State
		var err error
		if term.IsTerminal(os.Stdin.Fd()) {
			state, err = term.MakeRaw(os.Stdin.Fd())
			if err != nil {
				return fmt.Errorf("enter prompt raw mode: %w", err)
			}
		}
		formCtx, cancel := context.WithCancelCause(ctx)
		defer cancel(nil)
		input := newPromptAnswerInput(raw, cancel, answers)
		err = run(formCtx, input)
		cause := context.Cause(formCtx)
		closeErr := raw.Close()
		var restoreErr error
		if state != nil {
			restoreErr = term.Restore(os.Stdin.Fd(), state)
		}
		return errors.Join(err, cause, closeErr, restoreErr)
	})
}

type promptAnswerState struct {
	mu     sync.Mutex
	skipLF bool
}

func (state *promptAnswerState) takeSkipLF() bool {
	state.mu.Lock()
	defer state.mu.Unlock()
	skip := state.skipLF
	state.skipLF = false
	return skip
}

func (state *promptAnswerState) setSkipLF(skip bool) {
	state.mu.Lock()
	state.skipLF = skip
	state.mu.Unlock()
}

type promptAnswerInput struct {
	input       io.ReadCloser
	cancel      context.CancelCauseFunc
	mu          sync.Mutex
	answerDone  bool
	answers     *promptAnswerState
	skipLF      bool
	frame       [6]byte
	sourceByte  [1]byte
	frameSize   int
	frameOffset int
	inPaste     bool
}

func newPromptAnswerInput(input io.ReadCloser, cancel context.CancelCauseFunc, answers *promptAnswerState) *promptAnswerInput {
	if answers == nil {
		answers = &promptAnswerState{mu: sync.Mutex{}, skipLF: false}
	}
	return &promptAnswerInput{
		mu: sync.Mutex{}, answerDone: false, frame: [6]byte{}, sourceByte: [1]byte{}, frameSize: 0, frameOffset: 0, inPaste: false,
		input:   input,
		cancel:  cancel,
		answers: answers,
		skipLF:  answers.takeSkipLF(),
	}
}

func (input *promptAnswerInput) Read(data []byte) (int, error) {
	if len(data) == 0 {
		return 0, nil
	}
	input.mu.Lock()
	defer input.mu.Unlock()
	if input.frameOffset < input.frameSize {
		return input.deliverFrame(data), nil
	}
	if input.answerDone {
		return 0, io.EOF
	}
	input.frameSize, input.frameOffset = 0, 0
	for {
		b, err := input.readSourceByte()
		if err != nil {
			return 0, input.fail(err)
		}
		if input.skipLF {
			input.skipLF = false
			if b == '\n' {
				continue
			}
		}
		input.frame[0], input.frameSize = b, 1
		if b == '\x1b' {
			if err := input.readPasteMarker(); err != nil {
				return 0, input.fail(err)
			}
		}
		if !input.inPaste {
			if err := input.classifyAnswer(); err != nil {
				return 0, input.fail(err)
			}
		}
		return input.deliverFrame(data), nil
	}
}

func (input *promptAnswerInput) readSourceByte() (byte, error) {
	for {
		n, err := input.input.Read(input.sourceByte[:])
		if err != nil {
			return 0, fmt.Errorf("read input byte: %w", err)
		}
		if n > 0 {
			return input.sourceByte[0], nil
		}
	}
}

func (input *promptAnswerInput) readPasteMarker() error {
	const prefix = "\x1b[20"
	for input.frameSize < len(prefix) {
		b, err := input.readSourceByte()
		if err != nil {
			return err
		}
		index := input.frameSize
		input.frame[index] = b
		input.frameSize++
		if b != prefix[index] {
			return nil
		}
	}
	b, err := input.readSourceByte()
	if err != nil {
		return err
	}
	input.frame[input.frameSize] = b
	input.frameSize++
	if b != '0' && b != '1' {
		return nil
	}
	b, err = input.readSourceByte()
	if err != nil {
		return err
	}
	input.frame[input.frameSize] = b
	input.frameSize++
	if b == '~' {
		input.inPaste = input.frame[4] == '0'
	}
	return nil
}

func (input *promptAnswerInput) classifyAnswer() error {
	for index := range input.frameSize {
		switch input.frame[index] {
		case 3:
			return huh.ErrUserAborted
		case 4:
			return io.EOF
		case '\r':
			input.answerDone = true
			input.answers.setSkipLF(true)
		case '\n':
			input.frame[index] = '\r'
			input.answerDone = true
			input.answers.setSkipLF(false)
		}
	}
	return nil
}

func (input *promptAnswerInput) deliverFrame(data []byte) int {
	n := copy(data, input.frame[input.frameOffset:input.frameSize])
	input.frameOffset += n
	return n
}

func (input *promptAnswerInput) fail(err error) error {
	input.answerDone = true
	input.frameSize, input.frameOffset = 0, 0
	input.cancel(err)
	if errors.Is(err, io.EOF) {
		return io.EOF
	}
	return fmt.Errorf("read prompt answer: %w", err)
}
