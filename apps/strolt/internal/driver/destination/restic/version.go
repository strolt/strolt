package restic

import (
	"errors"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *Restic) getBinVersion() (string, error) {
	cmd := exec.Command(i.getBin(), "version")

	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	str := string(output)
	arr := strings.Split(str, "\n")

	if len(arr) == 0 {
		return "", errors.New("error parse 'restic version' output")
	}

	outputList := strings.Split(arr[0], " ")

	if len(outputList) < 2 { //nolint:gomnd
		return "", errors.New("error parse 'restic version' output")
	}

	return outputList[1], nil
}

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
