package restic

import (
	"slices"
	"testing"
	"time"

	"github.com/strolt/strolt/apps/strolt/internal/context"
	"github.com/strolt/strolt/apps/strolt/internal/driver/interfaces"
	"github.com/strolt/strolt/apps/strolt/internal/ldflags"
)

func backupTestContext() context.Context {
	ctx := context.Context{Trigger: "", ServiceName: "", TaskName: "", OpertationType: "", Event: "", Operation: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Source: context.Operation{Time: context.Time{Start: time.Time{}, Stop: time.Time{}}, Error: ""}, Destination: nil, WorkDir: "/work", IsWorkDirTmp: false, Tags: []string{"trigger=MANUAL"}, SourceLocalPath: ""}
	ctx.Operation.Time.Start = time.Unix(1600000000, 0)

	return ctx
}

func backupTestVersions() []interfaces.DriverBinaryVersion {
	return []interfaces.DriverBinaryVersion{{Name: "restic", Version: "0.19.0"}}
}

// containsArgPair reports whether args contains flag directly followed by value.
func containsArgPair(args []string, flag string, value string) bool {
	for index, arg := range args {
		if arg == flag && index+1 < len(args) && args[index+1] == value {
			return true
		}
	}

	return false
}

// TestBackupArgsWithoutTuning pins the command built for a config that uses
// none of the tuning fields: it must stay exactly what strolt sent before they
// existed.
func TestBackupArgsWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}
	ctx := backupTestContext()

	want := []string{
		"--json",
		"backup", "--host", "strolt_host",
		"--tag", "trigger=MANUAL",
		"--tag", "restic=0.19.0",
		"--tag", ldflags.GetBinaryName() + "=" + ldflags.GetVersion(),
		"--tag", "stroltStartedAt=1600000000",
		".",
	}

	if got := i.backupArgs(ctx, "", false, backupTestVersions()); !slices.Equal(got, want) {
		t.Fatalf("backupArgs() = %v, want %v", got, want)
	}
}

func TestBackupArgsPipeWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

	got := i.backupArgs(backupTestContext(), "dump.sql", true, backupTestVersions())

	want := []string{"--stdin", "--stdin-filename=dump.sql"}
	if !slices.Equal(got[len(got)-len(want):], want) {
		t.Fatalf("backupArgs() = %v, want it to end with %v", got, want)
	}
}

func TestBackupArgsTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: []string{"s3.connections=20", "rest.connections=10"}, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 8, NoScan: true, PackSize: 64, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: []string{"--exclude-caches", "--group-by=host"}, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

	got := i.backupArgs(backupTestContext(), "", false, backupTestVersions())

	for _, want := range [][]string{
		{"-o", "s3.connections=20"},
		{"-o", "rest.connections=10"},
		{"--read-concurrency", "8"},
		{"--pack-size", "64"},
		{"--host", "strolt_host"},
	} {
		if !containsArgPair(got, want[0], want[1]) {
			t.Fatalf("backupArgs() = %v, want the pair %v", got, want)
		}
	}

	if !slices.Contains(got, "--no-scan") || !slices.Contains(got, "--json") {
		t.Fatalf("backupArgs() = %v, want --no-scan and --json", got)
	}

	// The extended options are global flags and must precede the command.
	if slices.Index(got, "-o") > slices.Index(got, "backup") {
		t.Fatalf("backupArgs() = %v, want -o before the backup command", got)
	}

	wantTail := []string{"--exclude-caches", "--group-by=host"}
	if !slices.Equal(got[len(got)-len(wantTail):], wantTail) {
		t.Fatalf("backupArgs() = %v, want it to end with the extra args %v", got, wantTail)
	}
}
