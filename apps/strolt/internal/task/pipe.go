package task

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/dustin/go-humanize"
)

// copyPipe streams a snapshot from the destination reader into the source
// writer and collects both exit statuses without deadlocking.
//
// When the copy fails mid-stream, the surviving process is blocked writing
// into a pipe nobody drains anymore (or waiting for an EOF that io.Copy will
// never deliver). Waiting on such a process blocks forever with zero CPU —
// observed in production as a restore that hung after pg_restore aborted
// early while restic dump kept streaming from the repository. Closing our end
// of the pipe first turns the blocked write into EPIPE/SIGPIPE, the process
// exits, and the wait returns with a real error instead of hanging.
func copyPipe(reader io.ReadCloser, writer io.WriteCloser, waitReader, waitWriter func() error) error {
	_, copyErr := io.Copy(writer, reader)

	if copyErr != nil {
		_ = reader.Close()
	}

	errs := []error{copyErr}

	if err := waitReader(); err != nil {
		errs = append(errs, fmt.Errorf("destination process: %w", err))
	}

	// EOF must reach the reading process before its exit status is collected.
	errs = append(errs, writer.Close())

	if err := waitWriter(); err != nil {
		errs = append(errs, fmt.Errorf("source process: %w", err))
	}

	return errors.Join(errs...)
}

// sizeAwareReader is implemented by destination readers that know the exact
// size of the stream they deliver (restic ls node size, a stat'ed file for
// the local driver). Drivers that cannot know it simply skip the interface.
type sizeAwareReader interface {
	Size() (uint64, bool)
}

const restoreProgressInterval = 10 * time.Second

// restoreProgress reports how much data a piped restore has moved. Restore is
// usually run manually against a live database, so a long restore must be
// distinguishable from a hang: the ticker keeps logging even when the stream
// stalls (speed drops to 0), and the final line states exactly how many bytes
// the restore process actually accepted.
type restoreProgress struct {
	logf     func(format string, a ...any)
	total    uint64
	hasTotal bool
	interval time.Duration

	// written counts bytes accepted by the restore process, not bytes read
	// from the destination: on an aborted copy the last read chunk can be
	// rejected with EPIPE, and reporting it would overstate the transfer.
	written atomic.Int64
	done    chan struct{}
	stopped chan struct{}
}

func newRestoreProgress(reader io.Reader, logf func(format string, a ...any)) *restoreProgress {
	p := &restoreProgress{
		logf:     logf,
		interval: restoreProgressInterval,
		done:     make(chan struct{}),
		stopped:  make(chan struct{}),
	}

	if sizer, ok := reader.(sizeAwareReader); ok {
		if total, ok := sizer.Size(); ok {
			p.total = total
			p.hasTotal = true
		}
	}

	return p
}

// writer wraps the restore process stdin so only accepted bytes are counted.
type progressWriter struct {
	p          *restoreProgress
	underlying io.WriteCloser
}

// Write records the bytes the restore process accepted.
func (w *progressWriter) Write(b []byte) (int, error) {
	n, err := w.underlying.Write(b)
	w.p.written.Add(int64(n))
	return n, err //nolint:wrapcheck // the stream error belongs to the pipe, not the reporter
}

// Close delegates to the wrapped stdin so copyPipe can deliver EOF.
func (w *progressWriter) Close() error {
	return w.underlying.Close() //nolint:wrapcheck // passthrough of the stream close
}

// Start begins the periodic progress ticker. It stops when Finish is called.
func (p *restoreProgress) Start() {
	go func() {
		defer close(p.stopped)

		ticker := time.NewTicker(p.interval)
		defer ticker.Stop()

		var lastCount int64
		lastTime := time.Now()

		for {
			select {
			case <-p.done:
				return
			case now := <-ticker.C:
				written := p.written.Load()
				speed := float64(written-lastCount) / now.Sub(lastTime).Seconds()
				lastCount, lastTime = written, now

				if p.hasTotal {
					p.logf("restore in progress: %s / %s (%.1f%%), %s",
						humanize.IBytes(uint64(max(written, 0))), humanize.IBytes(p.total),
						100*float64(written)/float64(p.total), iBytesPerSec(speed))
				} else {
					p.logf("restore in progress: %s, %s",
						humanize.IBytes(uint64(max(written, 0))), iBytesPerSec(speed))
				}
			}
		}
	}()
}

// Finish stops the ticker and logs the final line. It reports the amount of
// data the restore process accepted even when the operation failed, because
// that is what tells a truncated restore apart from a complete one.
func (p *restoreProgress) Finish(startedAt time.Time, err error) {
	close(p.done)
	<-p.stopped

	written := p.written.Load()
	elapsed := time.Since(startedAt)

	avg := 0.0
	if elapsed > 0 {
		avg = float64(written) / elapsed.Seconds()
	}

	if err != nil {
		p.logf("restore stream aborted after %s of data (%.1fs, average %s): %v",
			humanize.IBytes(uint64(max(written, 0))), elapsed.Seconds(), iBytesPerSec(avg), err)
		return
	}

	p.logf("restore stream complete: %s in %.1fs (average %s)",
		humanize.IBytes(uint64(max(written, 0))), elapsed.Seconds(), iBytesPerSec(avg))
}

func iBytesPerSec(bps float64) string {
	unit, divisor := "B", float64(humanize.Byte)

	switch {
	case bps >= humanize.GiByte:
		unit, divisor = "GiB", humanize.GiByte
	case bps >= humanize.MiByte:
		unit, divisor = "MiB", humanize.MiByte
	case bps >= humanize.KiByte:
		unit, divisor = "KiB", humanize.KiByte
	}

	return strconv.FormatFloat(bps/divisor, 'f', 1, 64) + " " + unit + "/s"
}
