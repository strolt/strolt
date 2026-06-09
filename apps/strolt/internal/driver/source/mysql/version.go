package mysql

import (
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *MySQL) getMySQLVersion() string {
	cmd := exec.Command(i.getBinMySQL(), "--version") //nolint:gosec,noctx // binary path comes from validated configuration; no context available

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")

	return arr[2]
}

func (i *MySQL) getMySQLDumpVersion() string {
	cmd := exec.Command(i.getBinMySQLDump(), "--version") //nolint:gosec,noctx // binary path comes from validated configuration; no context available

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")

	return arr[2]
}

// BinaryVersion returns versions of the mysql and mysqldump binaries.
func (i *MySQL) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	mysqlVersion := i.getMySQLVersion()
	mysqlDumpVersion := i.getMySQLDumpVersion()

	return []interfaces.DriverBinaryVersion{
		{Name: "mysql", Version: mysqlVersion},
		{Name: "mysqldump", Version: mysqlDumpVersion},
	}, nil
}
