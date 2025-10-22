package e2e_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type Snapshot struct {
	ID      string `json:"id"`
	ShortID string `json:"shortId"`
	Date    string `json:"date"`
}

func runDockerExec(containerName string, command string) ([]byte, error) {
	cmd := exec.Command("docker", "exec", containerName, "/bin/sh", "-c", command)

	output, err := cmd.Output()

	fmt.Println(string(output)) //nolint:forbidigo

	return output, err
}

func runDockerComposeBash(command string) ([]byte, error) {
	if containerManager == nil || containerManager.GetStroltContainer() == nil {
		return nil, errors.New("strolt container not initialized")
	}

	containerName, err := GetContainerName(ctx, containerManager.GetStroltContainer())
	if err != nil {
		return nil, err
	}

	return runDockerExec(containerName, command)
}

func strolt(args ...string) error {
	o, err := stroltWithResponse(args...)
	log.Println(string(o))

	return err
}

func stroltWithResponse(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", "run", "--rm", "--network", "strolt",
		"-v", "./strolt.yml:/strolt/config.yml:ro",
		"-v", "./.strolt:/strolt/.strolt:ro",
		"-v", "./.temp/input:/e2e/input",
		"--entrypoint", "/bin/sh",
		"strolt/strolt:development",
		"-c", "/strolt/bin/strolt "+strings.Join(args, " "))

	output, err := cmd.CombinedOutput()

	fmt.Println(string(output)) //nolint:forbidigo

	return output, err
}

func stroltGetSnapshotList(serviceName string, taskName string, destination string) ([]Snapshot, error) {
	output, err := stroltWithResponse("snapshots", "--service", serviceName, "--task", taskName, "--destination", destination, "--json")
	if err != nil {
		return nil, err
	}

	lineList := strings.Split(string(output), "\n")

	if len(lineList) == 0 {
		return nil, errors.New("snapshots not exists")
	}

	// Find the last non-empty line that contains JSON array
	var lastItem string

	for i := len(lineList) - 1; i >= 0; i-- {
		trimmed := strings.TrimSpace(lineList[i])
		if trimmed != "" && strings.HasPrefix(trimmed, "[") {
			lastItem = trimmed
			break
		}
	}

	if lastItem == "" {
		return nil, errors.New("no valid JSON array found in output")
	}

	var snapshots []Snapshot

	err = json.Unmarshal([]byte(lastItem), &snapshots)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal snapshots: %w, data: %s", err, lastItem)
	}

	return snapshots, nil
}

func stroltGetLatestSnapshotID(serviceName string, taskName string, destination string) (string, error) {
	snapshots, err := stroltGetSnapshotList(serviceName, taskName, destination)
	if err != nil {
		return "", err
	}

	if len(snapshots) == 0 {
		return "", errors.New("snapshots not exists")
	}

	return snapshots[0].ID, nil
}
