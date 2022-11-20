package mysql

import (
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
)

func (i *MySQL) getMySQLVersion() string {
	cmd := exec.Command(i.getBinMySQL(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")

	return arr[2]
}

func (i *MySQL) getMySQLDumpVersion() string {
	cmd := exec.Command(i.getBinMySQLDump(), "--version")

	output, err := cmd.Output()
	if err != nil {
		return err.Error()
	}

	arr := strings.Split(strings.ReplaceAll(string(output), "  ", " "), " ")

	return arr[2]
}

func (i *MySQL) BinaryVersion() ([]interfaces.DriverBinaryVersion, error) {
	mysqlVersion := i.getMySQLVersion()
	mysqlDumpVersion := i.getMySQLDumpVersion()

	return []interfaces.DriverBinaryVersion{
		{Name: "mysql", Version: mysqlVersion},
		{Name: "mysqldump", Version: mysqlDumpVersion},
	}, nil
}
