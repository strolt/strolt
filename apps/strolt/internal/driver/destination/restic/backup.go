package restic

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
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
	tags, err := i.BinaryVersion()
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	var args []string
	args = append(args, i.getGlobalFlags()...)
	args = append(args, "backup")

	{
		for _, tag := range ctx.Tags {
			args = append(args, "--tag", tag)
		}

		for _, tag := range tags {
			args = append(args, "--tag", fmt.Sprintf("%s=%s", tag.Name, tag.Version))
		}

		args = append(args, "--tag", fmt.Sprintf("%s=%s", ldflags.GetBinaryName(), ldflags.GetVersion()))
	}

	args = append(args, i.getBackupFlags()...)
	args = append(args, ".")

	cmd := exec.Command(i.getBin(), args...)
	cmd.Dir = ctx.WorkDir

	env, err := i.getEnv()
	if err != nil {
		return sctxt.BackupOutput{}, err
	}

	cmd.Env = env

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
