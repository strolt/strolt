package schedule

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/stretchr/testify/require"
	"github.com/strolt/strolt/shared/process"
)

const scheduledHelperEnv = "STROLT_SCHEDULE_HELPER"

func TestScheduledProcessHelper(t *testing.T) {
	if os.Getenv(scheduledHelperEnv) != "1" {
		return
	}
	_, _ = fmt.Fprintln(os.Stdout, "ready")
	select {}
}

func TestRunCronWaitsForCancelledScheduledProcess(t *testing.T) {
	processContext, cancel := context.WithCancel(context.Background())
	defer cancel()

	cronScheduler := cron.New()
	releaseOperation := make(chan struct{})
	var releaseOnce sync.Once

	var started atomic.Bool
	jobStarted := make(chan error, 1)
	processWaited := make(chan error, 1)
	_, err := cronScheduler.AddFunc("@every 10ms", func() {
		if !started.CompareAndSwap(false, true) {
			return
		}

		command := process.CommandContext(processContext, os.Args[0], "-test.run=^TestScheduledProcessHelper$")
		command.Env = append(os.Environ(), scheduledHelperEnv+"=1")
		command.GracePeriod = 100 * time.Millisecond
		stdout, err := command.StdoutPipe()
		if err != nil {
			jobStarted <- err
			return
		}
		if err := command.Start(); err != nil {
			jobStarted <- err
			return
		}
		if _, err := bufio.NewReader(stdout).ReadString('\n'); err != nil {
			jobStarted <- err
			_ = command.Wait()
			return
		}

		jobStarted <- nil
		processWaited <- command.Wait()
		<-releaseOperation
	})
	require.NoError(t, err)

	runDone := make(chan struct{})
	t.Cleanup(func() {
		cancel()
		releaseOnce.Do(func() { close(releaseOperation) })
		select {
		case <-runDone:
		case <-time.After(5 * time.Second):
			t.Error("scheduler goroutine did not stop during cleanup")
		}
	})
	go func() {
		runCron(processContext, cronScheduler)
		close(runDone)
	}()

	select {
	case err := <-jobStarted:
		require.NoError(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("scheduled child did not start")
	}

	cancel()
	select {
	case err := <-processWaited:
		require.Error(t, err)
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("scheduled child was not reaped after cancellation")
	}

	select {
	case <-runDone:
		t.Fatal("scheduler returned before the in-flight operation completed")
	case <-time.After(100 * time.Millisecond):
	}

	releaseOnce.Do(func() { close(releaseOperation) })
	select {
	case <-runDone:
	case <-time.After(5 * time.Second):
		t.Fatal("scheduler did not return after the in-flight operation completed")
	}
}
