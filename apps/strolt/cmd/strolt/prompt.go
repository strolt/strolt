package cmd

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/spf13/cobra"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

// Prompt collects service, task, destination and snapshot names from flags or interactive prompts.
type Prompt struct {
	ServiceName     string
	TaskName        string
	DestinationName string
	SnapshotName    string

	cmd         *cobra.Command
	answerState promptAnswerState
}

// NewPrompt creates a Prompt bound to the given command.
func NewPrompt(cmd *cobra.Command) *Prompt {
	return &Prompt{
		cmd: cmd,
	}
}

// ScanServiceName resolves the service name from the flag or an interactive prompt.
func (p *Prompt) ScanServiceName() error {
	c := config.Get()
	serviceName, err := p.cmd.Flags().GetString("service")

	if serviceName == "" || err != nil {
		var selected string
		err := runSelectPrompt(p.cmd.Context(), "Select service", config.GetServiceNameList(), &selected, &p.answerState)
		if err != nil {
			return fmt.Errorf("select service prompt: %w", err)
		}
		serviceName = selected
	}

	p.ServiceName = serviceName

	if _, ok := c.Services[serviceName]; !ok {
		return fmt.Errorf("does not exists service - '%s'", serviceName)
	}

	return nil
}

// ScanTaskName resolves the task name from the flag or an interactive prompt.
func (p *Prompt) ScanTaskName() error {
	if p.ServiceName == "" {
		return errors.New("serviceName is empty")
	}

	c := config.Get()
	taskName, _ := p.cmd.Flags().GetString("task")

	if taskName == "" {
		var selected string
		err := runSelectPrompt(p.cmd.Context(), "Select task", config.GetTaskNameList(p.ServiceName), &selected, &p.answerState)
		if err != nil {
			return fmt.Errorf("select task prompt: %w", err)
		}
		taskName = selected
	}

	p.TaskName = taskName

	if _, ok := c.Services[p.ServiceName][p.TaskName]; !ok {
		return fmt.Errorf("does not exists task - '%s'", p.TaskName)
	}

	return nil
}

// ScanDestinationName resolves the destination name from the flag or an interactive prompt.
func (p *Prompt) ScanDestinationName() error {
	if p.ServiceName == "" {
		return errors.New("serviceName is empty")
	}

	if p.TaskName == "" {
		return errors.New("taskName is empty")
	}

	destinationName, _ := p.cmd.Flags().GetString("destination")
	if destinationName == "" {
		var selected string
		err := runSelectPrompt(p.cmd.Context(), "Select destination", config.GetDestinationNameList(p.ServiceName, p.TaskName), &selected, &p.answerState)
		if err != nil {
			return fmt.Errorf("select destination prompt: %w", err)
		}
		destinationName = selected
	}

	p.DestinationName = destinationName

	t, err := task.New(p.cmd.Context(), p.ServiceName, p.TaskName, sctxt.TManual, sctxt.OpTypeSnapshots)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	defer func() { _ = t.Close() }()

	if !t.IsAvailableDestinationName(p.DestinationName) {
		return fmt.Errorf("does not exists destination - %s", p.DestinationName)
	}

	return nil
}

// ScanSnapshotName resolves the snapshot name from the flag or an interactive prompt.
func (p *Prompt) ScanSnapshotName() error {
	if p.ServiceName == "" {
		return errors.New("serviceName is empty")
	}

	if p.TaskName == "" {
		return errors.New("taskName is empty")
	}

	if p.DestinationName == "" {
		return errors.New("destinationName is empty")
	}

	t, err := task.New(p.cmd.Context(), p.ServiceName, p.TaskName, sctxt.TManual, sctxt.OpTypeRestore)
	if err != nil {
		return fmt.Errorf("create task: %w", err)
	}
	defer func() { _ = t.Close() }()

	snapshotList, err := t.GetSnapshotList(p.DestinationName)
	if err != nil {
		return fmt.Errorf("get snapshot list: %w", err)
	}

	snapshotName, err := p.cmd.Flags().GetString("snapshot")
	if err != nil {
		return fmt.Errorf("get snapshot flag: %w", err)
	}

	items := make([]string, len(snapshotList))

	for i, snapshot := range snapshotList {
		snapshotID := snapshot.ShortID
		if snapshotID == "" {
			snapshotID = snapshot.ID
		}

		items[i] = fmt.Sprintf("%s (%s) [%s]", snapshotID, snapshot.Time.Format(time.RFC3339), strings.Join(snapshot.Tags, ", "))
	}

	if snapshotName == "" {
		var selectedIndex int
		err := runSelectIndexPrompt(p.cmd.Context(), "Select snapshot", items, &selectedIndex, &p.answerState)
		if err != nil {
			return fmt.Errorf("select snapshot prompt: %w", err)
		}
		snapshotName = snapshotList[selectedIndex].ID
	}

	if !snapshotList.IsAvailable(snapshotName) {
		return errors.New("snapshot not exists")
	}

	p.SnapshotName = snapshotName
	return nil
}

// AskIsConfirm asks the user for confirmation and reports whether it was given.
func (p *Prompt) AskIsConfirm() (bool, error) {
	confirmed := false
	err := runPromptAnswer(p.cmd.Context(), &p.answerState, func(formCtx context.Context, input io.Reader) error {
		answer := ""
		form := huh.NewForm(huh.NewGroup(huh.NewInput().Title("Are you sure? [y/N]").Value(&answer))).WithAccessible(false).WithKeyMap(promptKeyMap()).WithProgramOptions(tea.WithoutSignalHandler()).WithInput(input).WithOutput(os.Stderr)
		runErr := form.RunWithContext(formCtx)
		if runErr == nil {
			confirmed = strings.EqualFold(answer, "y")
		}
		if runErr != nil {
			return fmt.Errorf("run confirmation prompt: %w", runErr)
		}
		return nil
	})
	if err != nil {
		return false, err
	}
	return confirmed, err
}

var errNoPromptChoices = errors.New("no choices available")

func promptKeyMap() *huh.KeyMap {
	keys := huh.NewDefaultKeyMap()
	keys.Select.Filter.SetKeys()
	keys.Select.Filter.SetEnabled(false)
	keys.Select.Submit.SetKeys("enter", "alt+enter")
	keys.Input.Submit.SetKeys("enter", "alt+enter")
	return keys
}

func runSelectPrompt(ctx context.Context, title string, options []string, selected *string, answers *promptAnswerState) error {
	if len(options) == 0 {
		return fmt.Errorf("%s: %w", title, errNoPromptChoices)
	}
	return runPromptAnswer(ctx, answers, func(formCtx context.Context, input io.Reader) error {
		form := huh.NewForm(huh.NewGroup(huh.NewSelect[string]().Title(title).Options(huh.NewOptions(options...)...).Value(selected))).WithAccessible(false).WithKeyMap(promptKeyMap()).WithProgramOptions(tea.WithoutSignalHandler()).WithInput(input).WithOutput(os.Stderr)
		if err := form.RunWithContext(formCtx); err != nil {
			return fmt.Errorf("run selection prompt: %w", err)
		}
		return nil
	})
}

func runSelectIndexPrompt(ctx context.Context, title string, options []string, selected *int, answers *promptAnswerState) error {
	if len(options) == 0 {
		return fmt.Errorf("%s: %w", title, errNoPromptChoices)
	}
	return runPromptAnswer(ctx, answers, func(formCtx context.Context, input io.Reader) error {
		optionsTyped := make([]huh.Option[int], len(options))
		for i, option := range options {
			optionsTyped[i] = huh.NewOption(option, i)
		}
		form := huh.NewForm(huh.NewGroup(huh.NewSelect[int]().Title(title).Options(optionsTyped...).Value(selected))).WithAccessible(false).WithKeyMap(promptKeyMap()).WithProgramOptions(tea.WithoutSignalHandler()).WithInput(input).WithOutput(os.Stderr)
		if err := form.RunWithContext(formCtx); err != nil {
			return fmt.Errorf("run snapshot selection: %w", err)
		}
		return nil
	})
}
