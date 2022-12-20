package template

import (
	"fmt"
	"testing"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/constants"
	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	taskName := "task_name"
	startTime := time.Now()
	stopTime := startTime.Add(time.Minute)

	t.Run("copyright", func(t *testing.T) {
		t.Run("markdown", func(t *testing.T) {
			tmpl := New("", context.Context{})
			assert.Equal(t, fmt.Sprintf("<%s|strolt>", constants.RepoURL), tmpl.CopyrightMarkdown)
		})
	})

	t.Run("trigger", func(t *testing.T) {
		t.Run(string(sctxt.TApi), func(t *testing.T) {
			tmpl := New("slack", context.Context{
				Trigger:        sctxt.TApi,
				TaskName:       taskName,
				OpertationType: sctxt.OpTypeBackup,
			})

			header := fmt.Sprintf("%s [] [%s] - %s", emojies["slack"].TriggerHook, taskName, sctxt.OpTypeBackup)
			assert.Equal(t, header, tmpl.Header)
		})

		t.Run(string(sctxt.TSchedule), func(t *testing.T) {
			tmpl := New("slack", context.Context{
				Trigger:        sctxt.TSchedule,
				TaskName:       taskName,
				OpertationType: sctxt.OpTypeBackup,
			})

			header := fmt.Sprintf("%s [] [%s] - %s", emojies["slack"].TriggerSchedule, taskName, sctxt.OpTypeBackup)
			assert.Equal(t, header, tmpl.Header)
		})

		t.Run(string(sctxt.TManual), func(t *testing.T) {
			tmpl := New("slack", context.Context{
				Trigger:        sctxt.TManual,
				TaskName:       taskName,
				OpertationType: sctxt.OpTypeBackup,
			})

			header := fmt.Sprintf("%s [] [%s] - %s", emojies["slack"].TriggerManual, taskName, sctxt.OpTypeBackup)
			assert.Equal(t, header, tmpl.Header)
		})

		t.Run(string(sctxt.TSchedule)+"/error", func(t *testing.T) {
			tmpl := New("slack", context.Context{
				Trigger:        sctxt.TSchedule,
				TaskName:       taskName,
				OpertationType: sctxt.OpTypeBackup,
				Operation: context.Operation{
					Error: fmt.Errorf("error").Error(),
				},
			})

			header := fmt.Sprintf("%s [] [%s] - %s %s", emojies["slack"].TriggerSchedule, taskName, sctxt.OpTypeBackup, emojies["slack"].Error)
			assert.Equal(t, header, tmpl.Header)
		})
	})

	t.Run(string(sctxt.EvOperationStart), func(t *testing.T) {
		tmpl := New("", context.Context{
			TaskName:       taskName,
			OpertationType: sctxt.OpTypeBackup,
			Event:          sctxt.EvOperationStart,
			Operation: context.Operation{
				Time: context.Time{
					Start: startTime,
				},
			},
		})

		header := fmt.Sprintf("[] [%s] - %s", taskName, sctxt.OpTypeBackup)
		assert.Equal(t, header, tmpl.Header)

		body := fmt.Sprintf(`Event: %s`, sctxt.EvOperationStart)
		body += fmt.Sprintf("\nStart: %s", startTime.Format(time.RFC3339))
		assert.Equal(t, body, tmpl.Body)
	})

	t.Run(string(sctxt.EvOperationStop), func(t *testing.T) {
		tmpl := New("", context.Context{
			TaskName:       taskName,
			OpertationType: sctxt.OpTypeBackup,
			Event:          sctxt.EvOperationStart,
			Operation: context.Operation{
				Time: context.Time{
					Start: startTime,
					Stop:  stopTime,
				},
			},
		})

		header := fmt.Sprintf("[] [%s] - %s", taskName, sctxt.OpTypeBackup)
		assert.Equal(t, header, tmpl.Header)

		body := fmt.Sprintf(`Event: %s`, sctxt.EvOperationStart)
		body += fmt.Sprintf("\nStart: %s", startTime.Format(time.RFC3339))
		body += fmt.Sprintf("\nStop: %s", stopTime.Format(time.RFC3339))
		body += fmt.Sprintf("\nDuration: %s", stopTime.Sub(startTime))

		assert.Equal(t, body, tmpl.Body)
	})

	t.Run(string(sctxt.EvOperationError), func(t *testing.T) {
		err := fmt.Errorf("error")
		tmpl := New("", context.Context{
			TaskName:       taskName,
			OpertationType: sctxt.OpTypeBackup,
			Event:          sctxt.EvOperationError,
			Operation: context.Operation{
				Error: err.Error(),
				Time: context.Time{
					Start: startTime,
					Stop:  stopTime,
				},
			},
		})

		header := fmt.Sprintf("[] [%s] - %s", taskName, sctxt.OpTypeBackup)
		assert.Equal(t, header, tmpl.Header)

		body := fmt.Sprintf(`Event: %s`, sctxt.EvOperationError)
		body += fmt.Sprintf("\nStart: %s", startTime.Format(time.RFC3339))
		body += fmt.Sprintf("\nStop: %s", stopTime.Format(time.RFC3339))
		body += fmt.Sprintf("\nDuration: %s", stopTime.Sub(startTime))
		body += fmt.Sprintf("\n\nError: %s", err)

		assert.Equal(t, body, tmpl.Body)
	})
}
