package restic

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
)

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

	outputs := strings.Split(string(output), "\n")

	var backupOutput resticBackupOutput
	if err := json.Unmarshal([]byte(outputs[len(outputs)-2]), &backupOutput); err != nil {
		i.logger.Error(err)
		return sctxt.BackupOutput{}, err
	}

	return sctxt.BackupOutput{
		FilesNew:            backupOutput.FilesNew,
		FilesChanged:        backupOutput.FilesChanged,
		FilesUnmodified:     backupOutput.FilesUnmodified,
		DirsNew:             backupOutput.DirsNew,
		DirsChanged:         backupOutput.DirsChanged,
		DirsUnmodified:      backupOutput.DirsUnmodified,
		TotalFilesProcessed: backupOutput.TotalFilesProcessed,
		TotalBytesProcessed: backupOutput.TotalBytesProcessed,
		SnapshotID:          backupOutput.SnapshotID,
	}, nil
}

func (i *Restic) BackupPipe(ctx context.Context, filename string) (io.WriteCloser, func() error, error) {
	cmd, err := i.backupCmd(ctx, filename, true)
	if err != nil {
		return nil, nil, err
	}

	writer, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, err
	}

	if err := cmd.Start(); err != nil {
		return nil, nil, err
	}

	return writer, cmd.Wait, nil
}

func (i *Restic) IsSupportedBackupPipe(ctx context.Context) bool {
	return true
}
