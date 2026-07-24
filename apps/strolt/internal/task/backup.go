// Package task implements backup, restore, prune and related task operations.
package task

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/dmanager"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

// ErrNotSupportedPipeMode is returned when pipe mode is requested but not supported.
var ErrNotSupportedPipeMode = errors.New("source or one of destinations does not support pipe mode")

func (t *Task) backupSourceToWorkDir() error {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return err
	}

	if err := sourceDriver.Backup(t.Context); err != nil {
		return fmt.Errorf("source backup: %w", err)
	}

	return nil
}

func (t *Task) getSourceDriver() (interfaces.DriverSourceInterface, error) {
	sourceDriver, err := dmanager.GetSourceDriver(t.TaskConfig.Source.Driver, t.ServiceName, t.TaskName, t.TaskConfig.Source.Config, t.TaskConfig.Source.Env)
	if err != nil {
		return nil, fmt.Errorf("get source driver: %w", err)
	}

	return sourceDriver, nil
}

func (t *Task) isAvailableBackupPipe() (bool, error) {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return false, err
	}

	if !sourceDriver.IsSupportedBackupPipe(t.Context) {
		return false, nil
	}

	for destinationName := range t.TaskConfig.Destinations {
		destinationDriver, err := t.getDestinationDriver(destinationName)
		if err != nil {
			return false, err
		}

		if !destinationDriver.IsSupportedBackupPipe(t.Context) {
			return false, nil
		}
	}

	return true, nil
}

func (t *Task) backupManual() error {
	if err := t.managerStart(sctxt.OpTypeBackup); err != nil {
		return err
	}
	defer t.managerStop()

	var resultError error

	t.eventOperationStart()

	isSourceError := false
	isDestinationError := false

	{
		t.eventSourceStart()

		err := t.backupSourceToWorkDir()
		if err != nil {
			t.log.Error(err)
			t.eventSourceError(err)

			isSourceError = true
		} else {
			t.eventSourceStop()
		}
	}

	if !isSourceError {
		for destinationName := range t.TaskConfig.Destinations {
			t.eventDestinationStart(destinationName)

			backupOutput, err := t.backupWorkDirToDestination(destinationName)
			if err != nil {
				t.log.Error(err)
				t.eventDestinationError(destinationName, err)

				if !isDestinationError {
					isDestinationError = true
				}
			} else {
				t.eventDestinationStop(destinationName, backupOutput)
			}
		}
	}

	if isSourceError {
		resultError = fmt.Errorf("source: %s", t.Context.Source.Error)
		t.eventOperationError(resultError)
	}

	if isDestinationError {
		destinationErrors := []string{}

		for destinationName, operation := range t.Context.Destination {
			if operation.Error != "" {
				destinationErrors = append(destinationErrors, fmt.Sprintf("[%s]: %s", destinationName, operation.Error))
			}
		}

		resultError = fmt.Errorf("destination: %s", strings.Join(destinationErrors, ", "))
		t.eventOperationError(resultError)
	}

	if !isSourceError && !isDestinationError {
		t.eventOperationStop()
	}

	t.notifyWaitGroup().Wait()

	return resultError
}

func (t *Task) backupPipe() error {
	if err := t.managerStart(sctxt.OpTypeBackup); err != nil {
		return err
	}
	defer t.managerStop()

	t.eventOperationStart()

	err := t.runBackupPipe()

	if err != nil {
		t.eventOperationError(err)
	} else {
		t.eventOperationStop()
	}

	t.notifyWaitGroup().Wait()

	return err
}

// countingWriter wraps an io.Writer and records how many bytes were written to
// it. io.Copy into an io.MultiWriter writes serially in a single goroutine, so a
// plain counter is race-free without synchronization.
type countingWriter struct {
	w       io.Writer
	written uint64
}

func (c *countingWriter) Write(p []byte) (int, error) {
	n, err := c.w.Write(p)
	c.written += uint64(n) //nolint:gosec // Write never reports a negative count
	//nolint:wrapcheck // io.Writer contract: the underlying error must pass through unchanged (io.Copy inspects it)
	return n, err
}

// destinationPipe holds the per-destination pipe state while a piped backup is
// streaming. Kept in a slice so streaming and cleanup order stay stable and
// independent of map iteration order.
type destinationPipe struct {
	name     string
	counter  *countingWriter
	closer   io.Closer
	wait     func() (sctxt.BackupOutput, error)
	closeErr error
}

// runBackupPipe streams the source dump into every destination. The exit
// status of the source and destination processes must fail the backup: a
// crashed dump with a closed stream would otherwise turn into a
// "successful" empty snapshot.
//
// It also emits the per-destination start/stop/error events so the report
// includes a destination block, and it records the streamed byte count and the
// restic summary (snapshot_id, file counts, processed size) in each
// destination's BackupOutput.
func (t *Task) runBackupPipe() error {
	sourceDriver, err := t.getSourceDriver()
	if err != nil {
		return err
	}

	reader, filename, sourceWait, err := sourceDriver.BackupPipe(t.Context)
	if err != nil {
		return fmt.Errorf("source backup pipe: %w", err)
	}

	waitSource := sync.OnceValue(sourceWait)

	defer func() {
		_ = reader.Close()
		_ = waitSource()
	}()

	destinations := make([]destinationPipe, 0, len(t.TaskConfig.Destinations))

	defer func() {
		for idx := range destinations {
			_ = destinations[idx].closer.Close()
			_, _ = destinations[idx].wait()
		}
	}()

	for destinationName := range t.TaskConfig.Destinations {
		destinationDriver, err := t.getDestinationDriver(destinationName)
		if err != nil {
			return err
		}

		writer, destinationWait, err := destinationDriver.BackupPipe(t.Context, filename)
		if err != nil {
			return fmt.Errorf("destination backup pipe: %w", err)
		}

		destinations = append(destinations, destinationPipe{
			name:    destinationName,
			counter: &countingWriter{w: writer},
			closer:  writer,
			wait:    sync.OnceValues(destinationWait),
		})
	}

	// All pipes are established; announce every destination before streaming.
	writers := make([]io.Writer, 0, len(destinations))

	for idx := range destinations {
		t.eventDestinationStart(destinations[idx].name)
		writers = append(writers, destinations[idx].counter)
	}

	_, copyErr := io.Copy(io.MultiWriter(writers...), reader)

	errs := []error{copyErr}

	sourceErr := waitSource()
	if sourceErr != nil {
		errs = append(errs, fmt.Errorf("source process: %w", sourceErr))
	}

	// The source dump governs trust in the whole backup: a crashed dump yields a
	// truncated stream that restic would still turn into a snapshot, so a source
	// failure must fail every destination too.
	streamErr := copyErr
	if streamErr == nil {
		streamErr = sourceErr
	}

	// Close every writer first so the destinations see EOF and finalize in
	// parallel, then collect their exit status and summary.
	for idx := range destinations {
		closeErr := destinations[idx].closer.Close()
		destinations[idx].closeErr = closeErr

		if closeErr != nil {
			errs = append(errs, closeErr)
		}
	}

	for idx := range destinations {
		destination := &destinations[idx]

		output, waitErr := destination.wait()
		if waitErr != nil {
			errs = append(errs, fmt.Errorf("destination process: %w", waitErr))
		}

		switch {
		case destination.closeErr != nil || waitErr != nil:
			t.eventDestinationError(destination.name, errors.Join(destination.closeErr, waitErr))

		case streamErr != nil:
			t.eventDestinationError(destination.name, streamErr)

		default:
			output.TotalBytesStreamed = destination.counter.written
			t.eventDestinationStop(destination.name, output)
		}
	}

	return errors.Join(errs...)
}

// Backup runs the backup operation in pipe or manual mode depending on the task config.
func (t *Task) Backup() error {
	isAvailablePipe, err := t.isAvailableBackupPipe()
	if err != nil {
		return err
	}

	if t.TaskConfig.OperationMode == config.OperationModePipe {
		if !isAvailablePipe {
			return ErrNotSupportedPipeMode
		}

		return t.backupPipe()
	}

	if t.TaskConfig.OperationMode == config.OperationModePreferPipe && isAvailablePipe {
		return t.backupPipe()
	}

	return t.backupManual()
}

func (t *Task) backupWorkDirToDestination(destinationName string) (sctxt.BackupOutput, error) {
	destinationDriver, err := t.getDestinationDriver(destinationName)
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	output, err := destinationDriver.Backup(t.Context)
	if err != nil {
		return sctxt.BackupOutput{}, fmt.Errorf("destination backup: %w", err)
	}

	return output, nil
}

func (t *Task) getDestinationDriver(destinationName string) (interfaces.DriverDestinationInterface, error) {
	destination, ok := t.TaskConfig.Destinations[destinationName]
	if !ok {
		return nil, errors.New("destination not exits")
	}

	destinationDriver, err := dmanager.GetDestinationDriver(destinationName, destination.Driver, t.ServiceName, t.TaskName, destination.Config, destination.Env)
	if err != nil {
		return nil, fmt.Errorf("get destination driver: %w", err)
	}

	return destinationDriver, nil
}
