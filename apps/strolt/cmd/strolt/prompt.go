package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"
)

type Prompt struct {
	ServiceName     string
	TaskName        string
	DestinationName string
	SnapshotName    string

	cmd *cobra.Command
}

func NewPrompt(cmd *cobra.Command) *Prompt {
	return &Prompt{
		cmd: cmd,
	}
}

func (p *Prompt) ScanServiceName() error {
	c := config.Get()
	serviceName, err := p.cmd.Flags().GetString("service")

	if serviceName == "" || err != nil {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "select service",
			Items:        config.GetServiceNameList(),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return err
		}

		serviceName = result
	}

	p.ServiceName = serviceName

	if _, ok := c.Services[serviceName]; !ok {
		return fmt.Errorf("does not exists service - '%s'", serviceName)
	}

	return nil
}

func (p *Prompt) ScanTaskName() error {
	if p.ServiceName == "" {
		return fmt.Errorf("serviceName is empty")
	}

	c := config.Get()
	taskName, _ := p.cmd.Flags().GetString("task")

	if taskName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "select task",
			Items:        config.GetTaskNameList(p.ServiceName),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return err
		}

		taskName = result
	}

	p.TaskName = taskName

	if _, ok := c.Services[p.ServiceName][p.TaskName]; !ok {
		return fmt.Errorf("does not exists task - '%s'", p.TaskName)
	}

	return nil
}

func (p *Prompt) ScanDestinationName() error {
	if p.ServiceName == "" {
		return fmt.Errorf("serviceName is empty")
	}

	if p.TaskName == "" {
		return fmt.Errorf("taskName is empty")
	}

	destinationName, _ := p.cmd.Flags().GetString("destination")
	if destinationName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "Select destination",
			Items:        config.GetDestinationNameList(p.ServiceName, p.TaskName),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return err
		}

		destinationName = result
	}

	p.DestinationName = destinationName

	t, err := task.New(p.ServiceName, p.TaskName, sctxt.TManual, sctxt.OpTypeSnapshots)
	if err != nil {
		return err
	}
	defer t.Close()

	if !t.IsAvailableDestinationName(p.DestinationName) {
		return fmt.Errorf("does not exists destination - %s", p.DestinationName)
	}

	return nil
}

func (p *Prompt) ScanSnapshotName() error {
	if p.ServiceName == "" {
		return fmt.Errorf("serviceName is empty")
	}

	if p.TaskName == "" {
		return fmt.Errorf("taskName is empty")
	}

	if p.DestinationName == "" {
		return fmt.Errorf("destinationName is empty")
	}

	t, err := task.New(p.ServiceName, p.TaskName, sctxt.TManual, sctxt.OpTypeRestore)
	if err != nil {
		return err
	}
	defer t.Close()

	snapshotList, err := t.GetSnapshotList(p.DestinationName)
	if err != nil {
		return err
	}

	snapshotName, err := p.cmd.Flags().GetString("snapshot")
	if err != nil {
		return err
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
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "select snapshot",
			Items:        items,
		}

		i, _, err := prompt.Run()
		if err != nil {
			return err
		}

		snapshotName = snapshotList[i].ID
	}

	if !snapshotList.IsAvailable(snapshotName) {
		return fmt.Errorf("snapshot not exists")
	}

	p.SnapshotName = snapshotName

	return nil
}

func (p *Prompt) AskIsConfirm() bool {
	prompt := promptui.Prompt{
		HideEntered: true,
		Label:       "Are you shure",
		IsConfirm:   true,
	}

	_, err := prompt.Run()

	return err == nil
}
