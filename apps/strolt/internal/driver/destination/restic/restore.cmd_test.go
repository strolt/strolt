package restic

import (
	"slices"
	"testing"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/shared/logger"
)

func TestRestoreArgsWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: logger.New(), config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}
	ctx := context.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "/work", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}

	want := []string{"--json", "restore", "snapshot-id", "--target", "/work"}

	if got := i.restoreArgs(ctx, "snapshot-id", "", false); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}

func TestRestoreArgsCustomTarget(t *testing.T) {
	t.Parallel()

	i := &Restic{
		driverName: "", taskName: "", logger: logger.New(),
		config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "/restore/staging", ExtraBackupArgs: nil, ExtraRestoreArgs: []string{"--verify"}, ExtraEnv: nil},
		env:    Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""},
	}
	ctx := context.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "/work", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}

	want := []string{"--json", "restore", "snapshot-id", "--target", "/restore/staging", "--verify"}

	if got := i.restoreArgs(ctx, "snapshot-id", "", false); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}

// The dump used by piped restores writes to stdout, so neither the restore
// target nor the restore-only extra args apply to it.
func TestRestoreArgsPipeIgnoresRestoreOptions(t *testing.T) {
	t.Parallel()

	i := &Restic{
		driverName: "", taskName: "", logger: logger.New(),
		config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "/restore/staging", ExtraBackupArgs: nil, ExtraRestoreArgs: []string{"--verify"}, ExtraEnv: nil},
		env:    Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""},
	}
	ctx := context.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "/work", IsWorkDirTmp: false, Tags: nil, SourceLocalPath: ""}

	want := []string{"--json", "dump", "snapshot-id", "/dump.sql"}

	if got := i.restoreArgs(ctx, "snapshot-id", "/dump.sql", true); !slices.Equal(got, want) {
		t.Fatalf("restoreArgs() = %v, want %v", got, want)
	}
}
