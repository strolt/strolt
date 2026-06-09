package restic

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

// BinaryVersion returns the version of the restic binary used by the driver.
func (i *Restic) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	resticVersion, err := i.getBinVersion()
	if err != nil {
		return nil, err
	}

	return []interfaces.DriverBinaryVersion{
		{
			Name:    "restic",
			Version: resticVersion,
		},
	}, nil
}

func (i *Restic) getBinVersion() (string, error) {
	cmd := exec.CommandContext(context.Background(), i.getBin(), "version") //nolint:gosec // restic binary path comes from validated config

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("restic version: %w", err)
	}

	str := string(output)
	arr := strings.Split(str, "\n")

	if len(arr) == 0 {
		return "", errors.New("error parse 'restic version' output")
	}

	outputList := strings.Split(arr[0], " ")

	if len(outputList) < 2 { //nolint:mnd
		return "", errors.New("error parse 'restic version' output")
	}

	return outputList[1], nil
}
