// Package restic implements a destination driver backed by the restic backup tool.
package restic

import (
	gocontext "context"
	"fmt"
	"strings"

	"github.com/strolt/strolt/shared/logger"
	"github.com/strolt/strolt/shared/process"

	"gopkg.in/yaml.v3"
)

// Restic is a destination driver that stores snapshots in a restic repository.
type Restic struct {
	driverName string
	taskName   string
	logger     *logger.Logger
	config     Config
	env        Env
}

// New creates a new Restic driver instance.
func New() *Restic {
	return &Restic{driverName: "", taskName: "", logger: nil, config: Config{BinPath: "", Cacert: "", CleanupCache: false, LimitDownload: 0, LimitUpload: 0, NoCache: false, NoLock: false, TLSClientCert: "", Options: nil, ExcludePattern: nil, ExcludeFiles: nil, ExcludeLargerThan: "", Host: "", IExclude: "", IExcludeFile: "", IgnoreCTime: false, IgnoreINode: false, OneFileSystem: false, Tags: nil, WithATime: false, Compression: "", ReadConcurrency: 0, NoScan: false, PackSize: 0, Keep: ResticConfigKeep{Last: 0, Hourly: 0, Weekly: 0, Monthly: 0, Yearly: 0, Daily: 0}, RestoreTarget: "", ExtraBackupArgs: nil, ExtraRestoreArgs: nil, ExtraEnv: nil}, env: Env{RESTIC_REPOSITORY: "", RESTIC_PASSWORD: "", AWS_ACCESS_KEY_ID: "", AWS_SECRET_ACCESS_KEY: "", AWS_DEFAULT_REGION: "", AWS_PROFILE: "", AWS_SHARED_CREDENTIALS_FILE: "", ST_AUTH: "", ST_USER: "", ST_KEY: "", OS_AUTH_URL: "", OS_REGION_NAME: "", OS_USERNAME: "", OS_USER_ID: "", OS_PASSWORD: "", OS_TENANT_ID: "", OS_TENANT_NAME: "", OS_USER_DOMAIN_NAME: "", OS_USER_DOMAIN_ID: "", OS_PROJECT_NAME: "", OS_PROJECT_DOMAIN_NAME: "", OS_PROJECT_DOMAIN_ID: "", OS_TRUST_ID: "", OS_APPLICATION_CREDENTIAL_ID: "", OS_APPLICATION_CREDENTIAL_NAME: "", OS_APPLICATION_CREDENTIAL_SECRET: "", OS_STORAGE_URL: "", OS_AUTH_TOKEN: "", B2_ACCOUNT_ID: "", B2_ACCOUNT_KEY: "", AZURE_ACCOUNT_NAME: "", AZURE_ACCOUNT_KEY: "", GOOGLE_PROJECT_ID: "", GOOGLE_APPLICATION_CREDENTIALS: "", RCLONE_BWLIMIT: ""}}
}

// SetTaskName sets the task name for the driver.
func (i *Restic) SetTaskName(taskName string) {
	i.taskName = taskName
}

// SetDriverName sets the driver name.
func (i *Restic) SetDriverName(driverName string) {
	i.driverName = driverName
}

// SetConfig parses and validates the driver configuration.
func (i *Restic) SetConfig(processContext gocontext.Context, config any) error {
	data, err := yaml.Marshal(config)
	if err != nil {
		return fmt.Errorf("marshal config: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.config); err != nil {
		return fmt.Errorf("unmarshal config: %w", err)
	}

	return i.validateConfig(processContext)
}

// SetEnv parses and validates the driver environment variables.
func (i *Restic) SetEnv(env any) error {
	data, err := yaml.Marshal(env)
	if err != nil {
		return fmt.Errorf("marshal env: %w", err)
	}

	if err := yaml.Unmarshal(data, &i.env); err != nil {
		return fmt.Errorf("unmarshal env: %w", err)
	}

	return i.validateEnv()
}

// SetLogger sets the logger for the driver.
func (i *Restic) SetLogger(logger *logger.Logger) {
	i.logger = logger
}

func startCmd(cmd *process.Cmd) ([]byte, error) {
	output, err := cmd.CombinedOutput()
	if err != nil {
		if diagnostic := strings.TrimSpace(string(output)); diagnostic != "" {
			return nil, fmt.Errorf("run restic command: %w: %s", err, diagnostic)
		}

		return nil, fmt.Errorf("run restic command: %w", err)
	}

	return output, nil
}
