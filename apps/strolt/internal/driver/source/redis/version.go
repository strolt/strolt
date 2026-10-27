package redis

import (
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

// getRedisCliVersion parses the version from `redis-cli --version`, whose
// first line is "redis-cli 8.4.2": the second whitespace-separated field.
func (i *Redis) getRedisCliVersion() string {
	cmd := exec.Command(i.getBinRedisCli(), "--version") //nolint:gosec,noctx // binary path comes from validated configuration; no context available

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	str := string(output)
	arr := strings.Split(str, "\n")

	if len(arr) != 0 {
		fields := strings.Split(arr[0], " ")
		if len(fields) > 1 {
			return fields[1]
		}
	}

	return str
}

// BinaryVersion returns the version of the redis-cli binary.
func (i *Redis) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	return []interfaces.DriverBinaryVersion{
		{Name: "redis-cli", Version: i.getRedisCliVersion()},
	}, nil
}
