package e2e_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"

	tcexec "github.com/testcontainers/testcontainers-go/exec"
)

type Snapshot struct {
	ID      string `json:"id"`
	ShortID string `json:"shortId"`
	Date    string `json:"date"`
}

// execInStrolt runs a shell command inside the long-running strolt container.
// Reusing one container instead of spawning `docker run` per CLI call keeps
// the strolt working directory stable and avoids host docker CLI dependency.
func execInStrolt(command string) ([]byte, error) {
	if containerManager == nil || containerManager.GetStroltContainer() == nil {
		return nil, errors.New("strolt container not initialized")
	}

	exitCode, reader, err := containerManager.GetStroltContainer().Exec(
		ctx,
		[]string{"/bin/sh", "-c", command},
		tcexec.Multiplexed(),
	)
	if err != nil {
		return nil, err
	}

	output, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	fmt.Println(string(output)) //nolint:forbidigo

	if exitCode != 0 {
		return output, fmt.Errorf("command %q exited with code %d", command, exitCode)
	}

	return output, nil
}

func strolt(args ...string) error {
	_, err := stroltWithResponse(args...)

	return err
}

func stroltWithResponse(args ...string) ([]byte, error) {
	return execInStrolt("/strolt/bin/strolt " + strings.Join(args, " "))
}

func stroltGetSnapshotList(serviceName string, taskName string, destination string) ([]Snapshot, error) {
	output, err := stroltWithResponse("snapshots", "--service", serviceName, "--task", taskName, "--destination", destination, "--json")
	if err != nil {
		return nil, err
	}

	// The CLI mixes log lines with the JSON payload; the snapshot list is the
	// last line that parses as a JSON array.
	for _, line := range slices.Backward(strings.Split(string(output), "\n")) {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "[") {
			continue
		}

		var snapshots []Snapshot
		if err := json.Unmarshal([]byte(trimmed), &snapshots); err == nil {
			return snapshots, nil
		}
	}

	return nil, fmt.Errorf("no snapshot JSON array found in output: %s", output)
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
