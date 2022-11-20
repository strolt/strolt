package cmd

import (
	"fmt"

	"github.com/strolt/strolt/apps/strolt/internal/config"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/apps/strolt/internal/task"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func getServiceName(cmd *cobra.Command) (string, error) {
	c := config.Get()
	serviceName, _ := cmd.Flags().GetString("service")

	if serviceName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "Select service",
			Items:        config.GetServiceNameList(),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return "", err
		}

		serviceName = result
	}

	if _, ok := c.Services[serviceName]; !ok {
		fmt.Println("Please set --service string") //nolint:forbidigo
		return "", fmt.Errorf("does not exists service - %s", serviceName)
	}

	return serviceName, nil
}

func getTaskName(cmd *cobra.Command, serviceName string) (string, error) {
	c := config.Get()
	taskName, _ := cmd.Flags().GetString("task")

	if taskName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "Select task",
			Items:        config.GetTaskNameList(serviceName),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return "", err
		}

		taskName = result
	}

	if _, ok := c.Services[serviceName][taskName]; !ok {
		fmt.Println("Please set -t, --task string") //nolint:forbidigo
		return "", fmt.Errorf("does not exists task - %s", taskName)
	}

	return taskName, nil
}

func getDestinationName(cmd *cobra.Command, serviceName string, taskName string) (string, error) {
	destinationName, _ := cmd.Flags().GetString("destination")
	if destinationName == "" {
		prompt := promptui.Select{
			HideSelected: true,
			Label:        "Select destination",
			Items:        config.GetDestinationNameList(serviceName, taskName),
		}

		_, result, err := prompt.Run()
		if err != nil {
			return "", err
		}

		destinationName = result
	}

	t, err := task.New(serviceName, taskName, sctxt.TManual, sctxt.OpTypeSnapshots)
	if err != nil {
		return "", err
	}
	defer t.Close()

	if !t.IsAvailableDestinationName(destinationName) {
		fmt.Println("Please set -d, --destination string") //nolint:forbidigo
		return "", fmt.Errorf("does not exists destination - %s", destinationName)
	}

	return destinationName, nil
}

func isConfirm() bool {
	prompt := promptui.Prompt{
		HideEntered: true,
		Label:       "Are you shure",
		IsConfirm:   true,
	}

	_, err := prompt.Run()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err) //nolint:forbidigo
		return false
	}

	return true
}
