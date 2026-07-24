package restic

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

const messageTypeSummary = "summary"

type resticBackupOutput struct {
	MessageType         string  `json:"message_type"` // "summary"
	FilesNew            uint    `json:"files_new"`
	FilesChanged        uint    `json:"files_changed"`
	FilesUnmodified     uint    `json:"files_unmodified"`
	DirsNew             uint    `json:"dirs_new"`
	DirsChanged         uint    `json:"dirs_changed"`
	DirsUnmodified      uint    `json:"dirs_unmodified"`
	DataBlobs           int     `json:"data_blobs"`
	TreeBlobs           int     `json:"tree_blobs"`
	DataAdded           uint64  `json:"data_added"`
	TotalFilesProcessed uint    `json:"total_files_processed"`
	TotalBytesProcessed uint64  `json:"total_bytes_processed"`
	TotalDuration       float64 `json:"total_duration"` // in seconds
	SnapshotID          string  `json:"snapshot_id"`
	DryRun              bool    `json:"dry_run,omitempty"`
}

func (o *resticBackupOutput) toBackupOutput() sctxt.BackupOutput {
	return sctxt.BackupOutput{
		FilesNew:            o.FilesNew,
		FilesChanged:        o.FilesChanged,
		FilesUnmodified:     o.FilesUnmodified,
		DirsNew:             o.DirsNew,
		DirsChanged:         o.DirsChanged,
		DirsUnmodified:      o.DirsUnmodified,
		TotalFilesProcessed: o.TotalFilesProcessed,
		TotalBytesProcessed: o.TotalBytesProcessed,
		SnapshotID:          o.SnapshotID,
	}
}

// parseBackupSummary extracts restic's summary object from its --json backup
// output. restic prints the summary as the last JSON line, after any status
// updates, so the scan runs from the end and skips blank trailing lines.
func parseBackupSummary(output []byte) (sctxt.BackupOutput, error) {
	lines := strings.Split(string(output), "\n")

	for _, line := range slices.Backward(lines) {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var summary resticBackupOutput
		if err := json.Unmarshal([]byte(line), &summary); err != nil {
			return sctxt.BackupOutput{}, fmt.Errorf("unmarshal backup output: %w", err)
		}

		// Skip non-summary progress lines (e.g. "status"); accept the summary
		// message, or a bare object without a message_type for compatibility.
		if summary.MessageType != "" && summary.MessageType != messageTypeSummary {
			continue
		}

		return summary.toBackupOutput(), nil
	}

	return sctxt.BackupOutput{}, errors.New("restic backup summary not found")
}

// Backup runs a restic backup of the working directory and returns its summary.
func (i *Restic) Backup(ctx context.Context) (sctxt.BackupOutput, error) {
	cmd, err := i.backupCmd(ctx, "", false)
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	i.logger.Debug(cmd.String())

	output, err := startCmd(cmd)
	if err != nil {
		i.logger.Error(err)
		return sctxt.BackupOutput{}, err
	}

	i.logger.Debug(string(output))

	backupOutput, err := parseBackupSummary(output)
	if err != nil {
		i.logger.Error(err)
		return sctxt.BackupOutput{}, err
	}

	return backupOutput, nil
}

// BackupPipe returns a writer that streams data into a restic backup via stdin.
// The returned wait func collects restic's exit status and parses the --json
// summary it prints to stdout, so a piped backup reports the same statistics
// (snapshot_id, file counts, processed size) as a manual backup.
func (i *Restic) BackupPipe(ctx context.Context, filename string) (io.WriteCloser, func() (sctxt.BackupOutput, error), error) {
	cmd, err := i.backupCmd(ctx, filename, true)
	if err != nil {
		return nil, nil, err
	}

	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, fmt.Errorf("open stdin pipe: %w", err)
	}

	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Start(); err != nil {
		return nil, nil, fmt.Errorf("start restic backup: %w", err)
	}

	wait := func() (sctxt.BackupOutput, error) {
		if err := cmd.Wait(); err != nil {
			return sctxt.BackupOutput{}, fmt.Errorf("wait restic backup: %w", err)
		}

		output, err := parseBackupSummary(stdout.Bytes())
		if err != nil {
			i.logger.Error(err)
			return sctxt.BackupOutput{}, err
		}

		return output, nil
	}

	return writer, wait, nil
}

// IsSupportedBackupPipe reports whether piped backups are supported.
func (i *Restic) IsSupportedBackupPipe(ctx context.Context) bool {
	return true
}
