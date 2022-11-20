package pg

import (
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *PgDump) getPgDumpVersion() string {
	cmd := exec.Command(i.getBinPgDump(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

func (i *PgDump) getPgRestoreVersion() string {
	cmd := exec.Command(i.getBinPgRestore(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

func (i *PgDump) getPsqlVersion() string {
	cmd := exec.Command(i.getBinPSQL(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

func (i *PgDump) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	pgDumpVersion := i.getPgDumpVersion()
	pgRestoreVersion := i.getPgRestoreVersion()
	pgPsqlVersion := i.getPsqlVersion()

	return []interfaces.DriverBinaryVersion{
		{Name: "pg_dump", Version: pgDumpVersion},
		{Name: "pg_restore", Version: pgRestoreVersion},
		{Name: "psql", Version: pgPsqlVersion},
	}, nil
}
