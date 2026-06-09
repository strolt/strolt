// Package template renders notification messages for task operations.
package template

import (
	"fmt"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/constants"
	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/dustin/go-humanize"
)

// Template holds the rendered parts of a notification message.
type Template struct {
	Header            string
	Body              string
	CopyrightMarkdown string
	CopyrightHTML     string
}

// Emoji defines the emoji set used by a notification driver.
type Emoji struct {
	TriggerHook     string
	TriggerSchedule string
	TriggerManual   string
	Error           string
}

var emojies = map[string]Emoji{
	"slack": {
		TriggerHook:     ":hook:",
		TriggerSchedule: ":timer_clock:",
		TriggerManual:   ":bust_in_silhouette:",
		Error:           ":red_circle:",
	},
	"telegram": {
		TriggerHook:     "🪝",
		TriggerSchedule: "⏰",
		TriggerManual:   "👤",
		Error:           "🔴",
	},
}

func getTriggerEmoji(driver string, trigger sctxt.TriggerType) string {
	emoji, ok := emojies[driver]
	if !ok {
		return ""
	}

	switch trigger {
	case sctxt.TApi:
		return emoji.TriggerHook + " "

	case sctxt.TSchedule:
		return emoji.TriggerSchedule + " "

	case sctxt.TManual:
		return emoji.TriggerManual + " "
	}

	return ""
}

func getErrorEmoji(driver string, ctx context.Context) string { //nolint:revive // ctx is the package-local operation context, not a stdlib context.Context
	if ctx.Operation.Error == "" {
		return ""
	}

	emoji, ok := emojies[driver]
	if !ok {
		return ""
	}

	return " " + emoji.Error
}

// commaUint formats an unsigned counter with thousand separators.
func commaUint(v uint) string {
	return humanize.Comma(int64(v)) //nolint:gosec // backup counters never exceed int64
}

// New builds a notification Template for the given driver from the operation context.
func New(driver string, ctx context.Context) Template { //nolint:revive // ctx is the package-local operation context, not a stdlib context.Context
	t := Template{}

	t.CopyrightMarkdown = fmt.Sprintf("<%s|strolt>", constants.RepoURL)
	t.CopyrightHTML = fmt.Sprintf("<a href=%q>strolt</a>", constants.RepoURL)
	t.Header = fmt.Sprintf("%s[%s] [%s] - %s%s", getTriggerEmoji(driver, ctx.Trigger), ctx.ServiceName, ctx.TaskName, ctx.OpertationType, getErrorEmoji(driver, ctx))

	t.Body += fmt.Sprintf("Event: %s", ctx.Event)
	t.Body += "\nStart: " + ctx.Operation.Time.Start.Format(time.RFC3339)

	if !ctx.Operation.Time.Stop.IsZero() {
		t.Body += "\nStop: " + ctx.Operation.Time.Stop.Format(time.RFC3339)
		t.Body += fmt.Sprintf("\nDuration: %s", ctx.Operation.Time.Stop.Sub(ctx.Operation.Time.Start))
	}

	if ctx.Operation.Error != "" {
		t.Body += "\n\nError: " + ctx.Operation.Error
	}

	for destinationName, destination := range ctx.Destination {
		t.Body += fmt.Sprintf("\n\n[destination] %s:", destinationName)

		if destination.BackupOutput.SnapshotID != "" {
			t.Body += "\n    snapshot_id: " + destination.BackupOutput.SnapshotID
		}

		if destination.BackupOutput.FilesNew != 0 {
			t.Body += "\n    files_new: " + commaUint(destination.BackupOutput.FilesNew)
		}

		if destination.BackupOutput.FilesChanged != 0 {
			t.Body += "\n    files_changed: " + commaUint(destination.BackupOutput.FilesChanged)
		}

		if destination.BackupOutput.FilesUnmodified != 0 {
			t.Body += "\n    files_unmodified: " + commaUint(destination.BackupOutput.FilesUnmodified)
		}

		if destination.BackupOutput.DirsNew != 0 {
			t.Body += "\n    dirs_new: " + commaUint(destination.BackupOutput.DirsNew)
		}

		if destination.BackupOutput.DirsChanged != 0 {
			t.Body += "\n    dirs_changed: " + commaUint(destination.BackupOutput.DirsChanged)
		}

		if destination.BackupOutput.DirsUnmodified != 0 {
			t.Body += "\n    dirs_unmodified: " + commaUint(destination.BackupOutput.DirsUnmodified)
		}

		if destination.BackupOutput.TotalFilesProcessed != 0 {
			t.Body += "\n    total_files_processed: " + commaUint(destination.BackupOutput.TotalFilesProcessed)
		}

		if destination.BackupOutput.TotalBytesProcessed != 0 {
			t.Body += "\n    total_size_processed: " + humanize.Bytes(destination.BackupOutput.TotalBytesProcessed)
		}
	}

	return t
}
