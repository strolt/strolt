package restic

import (
	"slices"
	"testing"
)

func TestGetKeepFlagsEmptyPolicy(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

	if flags := i.getKeepFlags(); len(flags) != 0 {
		t.Fatalf("getKeepFlags() with empty policy = %v, want none", flags)
	}
}

func TestGetKeepFlagsSingleField(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		keep ResticConfigKeep
		want string
	}{
		{name: "last", keep: ResticConfigKeep{Last: 3, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, want: "--keep-last=3"},
		{name: "hourly", keep: ResticConfigKeep{Last: 0, Hourly: 2, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, want: "--keep-hourly=2"},
		{name: "daily", keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 7}, want: "--keep-daily=7"},
		{name: "weekly", keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 4, Monthly: 0, Yearly: 0, Daily: 0}, want: "--keep-weekly=4"},
		{name: "monthly", keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 12, Yearly: 0, Daily: 0}, want: "--keep-monthly=12"},
		{name: "yearly", keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 5, Daily: 0}, want: "--keep-yearly=5"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, Keep: tc.keep}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

			flags := i.getKeepFlags()
			if len(flags) != 1 || flags[0] != tc.want {
				t.Fatalf("getKeepFlags() = %v, want exactly [%s]", flags, tc.want)
			}
		})
	}
}

func TestGetKeepFlagsCombined(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, Keep: ResticConfigKeep{
		Last: 1, Hourly: 2, Weekly: 4, Monthly: 5, Yearly: 6, Daily: 3,
	}}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

	flags := i.getKeepFlags()

	want := []string{
		"--keep-last=1",
		"--keep-hourly=2",
		"--keep-daily=3",
		"--keep-weekly=4",
		"--keep-monthly=5",
		"--keep-yearly=6",
	}

	if len(flags) != len(want) {
		t.Fatalf("getKeepFlags() = %v, want all of %v", flags, want)
	}

	for _, flag := range want {
		if !slices.Contains(flags, flag) {
			t.Fatalf("getKeepFlags() = %v, missing %s", flags, flag)
		}
	}
}

func TestGetBackupFlagsWithoutTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}

	if flags := i.getBackupFlags(); len(flags) != 0 {
		t.Fatalf("getBackupFlags() with empty config = %v, want none", flags)
	}
}

func TestGetBackupFlagsTuning(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, ReadConcurrency: 8, NoScan: true, PackSize: 64}}

	want := []string{"--read-concurrency", "8", "--no-scan", "--pack-size", "64"}

	if flags := i.getBackupFlags(); !slices.Equal(flags, want) {
		t.Fatalf("getBackupFlags() = %v, want %v", flags, want)
	}
}

// Zero means "not configured": restic keeps its own defaults instead of being
// passed a nonsensical --read-concurrency 0.
func TestGetBackupFlagsIgnoresNonPositiveCounts(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", NoScan: false, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, ReadConcurrency: -1, PackSize: 0}}

	if flags := i.getBackupFlags(); len(flags) != 0 {
		t.Fatalf("getBackupFlags() = %v, want none", flags)
	}
}

func TestGetGlobalFlagsOptions(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, Options: []string{"rest.connections=10", "s3.connections=20"}}}

	want := []string{"-o", "rest.connections=10", "-o", "s3.connections=20", "--json"}

	if flags := i.getGlobalFlags(); !slices.Equal(flags, want) {
		t.Fatalf("getGlobalFlags() = %v, want %v", flags, want)
	}
}

func TestGetGlobalFlagsWithLockDropsNoLock(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, NoLock: true, NoCache: true}}

	if flags := i.getGlobalFlags(); !slices.Contains(flags, "--no-lock") {
		t.Fatalf("getGlobalFlags() = %v, want --no-lock", flags)
	}

	flags := i.getGlobalFlagsWithLock()

	if slices.Contains(flags, "--no-lock") {
		t.Fatalf("getGlobalFlagsWithLock() = %v, want no --no-lock", flags)
	}

	for _, flag := range []string{"--no-cache", "--json"} {
		if !slices.Contains(flags, flag) {
			t.Fatalf("getGlobalFlagsWithLock() = %v, missing %s", flags, flag)
		}
	}
}

func TestGetGlobalFlagsWithLockKeepsOriginal(t *testing.T) {
	t.Parallel()

	i := &Restic{driverName: "", taskName: "", logger: nil, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil, NoLock: true}}

	_ = i.getGlobalFlagsWithLock()

	if flags := i.getGlobalFlags(); !slices.Contains(flags, "--no-lock") {
		t.Fatalf("getGlobalFlags() after getGlobalFlagsWithLock() = %v, want --no-lock", flags)
	}
}
