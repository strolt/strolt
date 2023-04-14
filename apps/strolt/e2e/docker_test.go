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

func runDockerCompose(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", append([]string{"compose"}, args...)...)
	fmt.Println(cmd.String()) //nolint:forbidigo

	return cmd.CombinedOutput()
}

func runDockerComposeBash(command string) ([]byte, error) {
	cmd := exec.Command("docker", "exec", "strolt", "/bin/sh", "-c", command)
	fmt.Println(cmd.String()) //nolint:forbidigo

	return cmd.Output()
}

func dockerComposeUp(services ...string) error {
	_, err := runDockerCompose(append([]string{"up", "-d"}, services...)...)
	return err
}

func dockerComposeUpStrolt() error {
	_, err := runDockerCompose("run", "-d", "--name", "strolt", "--entrypoint", "/bin/sh -c", "strolt", "sleep 99999")
	return err
}

func dockerComposeDown() error {
	if _, err := runDockerCompose("kill"); err != nil {
		return err
	}

	_, err := runDockerCompose("down", "-v")

	return err
}

func strolt(args ...string) error {
	o, err := stroltWithResponse(args...)
	log.Println(string(o))

	return err
}

func stroltWithResponse(args ...string) ([]byte, error) {
	cmd := exec.Command("docker", "exec", "strolt", "/bin/sh", "-c", fmt.Sprintf("/strolt/bin/strolt %s", strings.Join(args, " ")))
	fmt.Println(cmd.String()) //nolint:forbidigo

	return cmd.CombinedOutput()
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

	lastItem := lineList[len(lineList)-2]

	var snapshots []Snapshot

	err = json.Unmarshal([]byte(lastItem), &snapshots)
	if err != nil {
		return nil, err
	}

	return snapshots, nil
}

//nolint:unparam
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
