package pg

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/context"
)

func (i *PgDump) Restore(ctx context.Context) error {
	filename, err := i.getFilenameFromBackup(ctx)
	if err != nil {
		return err
	}

	if i.isRestoreWithPSQL(filename) {
		return i.restoreWithPSQL(ctx, filename)
	}

	return i.restoreWithPgRestore(ctx, filename)
}

func (i *PgDump) isRestoreWithPSQL(filename string) bool {
	return strings.HasSuffix(filename, ".sql")
}

func (i *PgDump) getFilenameFromBackup(ctx context.Context) (string, error) {
	files, err := os.ReadDir(ctx.WorkDir)
	if err != nil {
		return "", err
	}

	for _, file := range files {
		if strings.HasPrefix(file.Name(), FileNamePrefix) {
			return file.Name(), nil
		}
	}

	return "", fmt.Errorf("not found dump")
}

func (i *PgDump) restoreWithPSQL(ctx context.Context, filename string) error {
	i.logger.Info("restore with psql")

	args := i.getCommonArgs()
	args = append(args, "--file="+filename)
	cmd := exec.Command(i.getBinPSQL(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	outputByte, err := cmd.CombinedOutput()
	outputString := string(outputByte)
	arr := strings.Split(outputString, "\n")

	if err != nil {
		i.logger.Error(outputString)

		if len(arr) > 0 {
			lastMessage := arr[len(arr)-1]
			if lastMessage == "" && len(arr) > 1 {
				lastMessage = arr[len(arr)-2]
			}

			return fmt.Errorf("%w (%v)", err, lastMessage)
		}

		return err
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

func (i *PgDump) restoreWithPgRestore(ctx context.Context, filename string) error {
	i.logger.Info("restore with pg_restore")

	args := i.getRestoreArgs()
	args = append(args, filename)
	cmd := exec.Command(i.getBinPgRestore(), args...)
	cmd.Dir = ctx.WorkDir
	cmd.Env = i.getEnv()

	outputByte, err := cmd.CombinedOutput()
	outputString := string(outputByte)
	arr := strings.Split(outputString, "\n")

	if err != nil {
		i.logger.Error(outputString)

		if len(arr) > 0 {
			lastMessage := arr[len(arr)-1]
			if lastMessage == "" && len(arr) > 1 {
				lastMessage = arr[len(arr)-2]
			}

			return fmt.Errorf("%w (%v)", err, lastMessage)
		}

		return err
	}

	if outputString != "" {
		i.logger.Info(outputString)
	}

	return nil
}

func (i *PgDump) RestorePipe(_ context.Context) error {
	return errors.New("not support pipe")
}

func (i *PgDump) IsSupportedRestorePipe(_ context.Context) bool {
	return false
}
