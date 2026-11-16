package pg

import (
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/shared/process"
)

func (i *PgDump) getPgDumpVersion() string {
	cmd := process.Command(i.getBinPgDump(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

func (i *PgDump) getPgRestoreVersion() string {
	cmd := process.Command(i.getBinPgRestore(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

func (i *PgDump) getPsqlVersion() string {
	cmd := process.Command(i.getBinPSQL(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(string(output), " ")

	return arr[len(arr)-1]
}

// BinaryVersion returns versions of the pg_dump, pg_restore and psql binaries.
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
