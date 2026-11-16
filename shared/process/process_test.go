package process

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestProcessHelper(t *testing.T) {
	args := os.Args
	for len(args) > 0 && args[0] != "--" {
		args = args[1:]
	}
	if len(args) < 2 {
		return
	}
	args = args[1:]
	switch args[0] {
	case "exit":
		_, _ = fmt.Fprintln(os.Stderr, "command diagnostic")
		code, _ := strconv.Atoi(args[1])
		os.Exit(code)
	case "environment":
		dir, err := os.Getwd()
		if err != nil {
			os.Exit(98)
		}
		_, _ = fmt.Fprint(os.Stdout, os.Getenv("STROLT_PROCESS_VALUE")+"|"+dir)
	case "cat":
		if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
			os.Exit(98)
		}
	case "wait":
		_, _ = fmt.Fprintln(os.Stdout, "ready")
		time.Sleep(30 * time.Second)
	case "spam":
		data := bytes.Repeat([]byte("stream"), 1024)
		for {
			if _, err := os.Stdout.Write(data); err != nil {
				os.Exit(3)
			}
		}
	case "consume-once":
		_, _ = io.CopyN(io.Discard, os.Stdin, 1024)
		os.Exit(7)
	}
	os.Exit(0)
}

func helperCommand(t *testing.T, mode string, args ...string) *Cmd {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	t.Cleanup(cancel)
	command := CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestProcessHelper$", "--", mode}, args...)...)
	command.GracePeriod = 100 * time.Millisecond
	return command
}

func TestOutputFailurePreservesStatusAndDiagnostic(t *testing.T) {
	_, err := helperCommand(t, "exit", "7").Output()
	var status *ExitError
	if !errors.As(err, &status) || status.ExitCode() != 7 {
		t.Fatalf("command error = %v, want exit code 7", err)
	}
	if !strings.Contains(err.Error(), "command diagnostic") {
		t.Fatalf("missing stderr diagnostic: %v", err)
	}
	if err := helperCommand(t, "exit", "0").Run(); err != nil {
		t.Fatal(err)
	}
}

func TestCombinedOutputRetainsFailureOutput(t *testing.T) {
	output, err := helperCommand(t, "exit", "9").CombinedOutput()
	var status *ExitError
	if !errors.As(err, &status) || status.ExitCode() != 9 || string(output) != "command diagnostic\n" {
		t.Fatalf("output = %q, error = %v", output, err)
	}
}

func TestEnvironmentAndWorkingDirectory(t *testing.T) {
	t.Setenv("STROLT_PROCESS_VALUE", "inherited")
	for _, tc := range []struct {
		name string
		env  []string
		want string
	}{
		{name: "inherit", want: "inherited"},
		{name: "empty", env: []string{}, want: ""},
		{name: "explicit", env: []string{"STROLT_PROCESS_VALUE=first", "STROLT_PROCESS_VALUE=last"}, want: "last"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			command := helperCommand(t, "environment")
			command.Env = tc.env
			command.Dir = t.TempDir()
			output, err := command.Output()
			if err != nil {
				t.Fatal(err)
			}
			if string(output) != tc.want+"|"+command.Dir {
				t.Fatalf("environment/cwd = %q, want %q", output, tc.want+"|"+command.Dir)
			}
		})
	}
}

func TestPipesStreamAllBytesBeforeWait(t *testing.T) {
	command := helperCommand(t, "cat")
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	payload := bytes.Repeat([]byte("stream-payload"), 32768)
	written := make(chan error, 1)
	go func() {
		_, err := stdin.Write(payload)
		written <- errors.Join(err, stdin.Close())
	}()
	output, readErr := io.ReadAll(stdout)
	writeErr := <-written
	waitErr := command.Wait()
	if err := errors.Join(readErr, writeErr, waitErr); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(output, payload) {
		t.Fatal("stream payload was truncated or corrupted")
	}
}

func TestEarlyConsumerExitUnblocksProducer(t *testing.T) {
	producer := helperCommand(t, "spam")
	reader, err := producer.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	consumer := helperCommand(t, "consume-once")
	writer, err := consumer.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := producer.Start(); err != nil {
		t.Fatal(err)
	}
	if err := consumer.Start(); err != nil {
		_ = reader.Close()
		_ = producer.Wait()
		t.Fatal(err)
	}
	_, copyErr := io.Copy(writer, reader)
	_ = reader.Close()
	_ = writer.Close()
	producerErr, consumerErr := producer.Wait(), consumer.Wait()
	var status *ExitError
	if copyErr == nil || producerErr == nil || !errors.As(consumerErr, &status) || status.ExitCode() != 7 {
		t.Fatalf("copy = %v, producer = %v, consumer = %v", copyErr, producerErr, consumerErr)
	}
	if errors.Is(producerErr, context.DeadlineExceeded) || errors.Is(consumerErr, context.DeadlineExceeded) {
		t.Fatal("pipe failure only finished after the timeout")
	}
}

func TestCancellationPreservesCause(t *testing.T) {
	ctx, cancel := context.WithCancelCause(t.Context())
	defer cancel(nil)
	command := helperCommand(t, "wait")
	command.ctx = ctx
	reader, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	if _, err := bufio.NewReader(reader).ReadString('\n'); err != nil {
		t.Fatal(err)
	}
	cause := errors.New("operation stopped")
	cancel(cause)
	if err := command.Wait(); !errors.Is(err, context.Canceled) || !errors.Is(err, cause) {
		t.Fatalf("cancellation identity lost: %v", err)
	}
}

func TestFailedStartClosesCreatedPipes(t *testing.T) {
	command := Command(filepath.Join(t.TempDir(), "missing"))
	stdin, err := command.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := command.Start(); err == nil {
		t.Fatal("missing executable started")
	}
	if _, err := stdin.Write([]byte("late")); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("stdin remains open: %v", err)
	}
	if _, err := stdout.Read(make([]byte, 1)); !errors.Is(err, os.ErrClosed) {
		t.Fatalf("stdout remains open: %v", err)
	}
}
