package restic

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	taskcontext "github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"
)

func TestPruneStopsWithOperationContext(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("the fake restic executable uses a POSIX shell")
	}
	if _, err := exec.LookPath("sh"); err != nil {
		t.Skip("sh not available")
	}

	dataDir := t.TempDir()
	t.Cleanup(env.Scan)
	t.Setenv("STROLT_PATH_DATA", dataDir)
	env.Scan()

	readyFile := filepath.Join(dataDir, "prune-ready")
	binary := filepath.Join(dataDir, "restic-fake")
	script := "#!/bin/sh\nprintf '%s\\n' \"$$\" > \"$RESTIC_PRUNE_READY_FILE\"\nexec /bin/sleep 30\n"
	require.NoError(t, os.WriteFile(binary, []byte(script), 0o700))

	driver := &Restic{
		driverName: "prune-cancellation-test",
		taskName:   "prune-cancellation-test",
		logger:     logger.New(),
		config:     Config{BinPath: binary, Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: map[string]string{"RESTIC_PRUNE_READY_FILE": readyFile}},
		env:        Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""},
	}
	processContext, cancel := context.WithCancel(context.Background())
	defer cancel()
	operationContext := taskcontext.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: taskcontext.Operation{Time: taskcontext.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}
	operationContext.SetProcessContext(processContext)

	pruneDone := make(chan error, 1)
	go func() {
		_, err := driver.Prune(operationContext, false)
		pruneDone <- err
	}()

	require.Eventually(t, func() bool {
		_, err := os.Stat(readyFile)
		return err == nil
	}, 5*time.Second, 10*time.Millisecond)
	cancel()

	select {
	case err := <-pruneDone:
		require.ErrorIs(t, err, context.Canceled)
	case <-time.After(5 * time.Second):
		t.Fatal("restic prune did not stop after operation cancellation")
	}
}
