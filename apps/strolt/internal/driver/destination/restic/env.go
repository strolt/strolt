package restic

import (
	"errors"
	"fmt"
	"maps"
	"slices"

	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
)

// Env describes the environment variables passed to the restic binary.
//
//nolint:revive // field names mirror the environment variable names expected by restic
type Env struct {
	// RESTIC_REPOSITORY_FILE  string `yaml:"RESTIC_REPOSITORY_FILE"`  // Name of file containing the repository location (replaces --repository-file)
	RESTIC_REPOSITORY string `yaml:"RESTIC_REPOSITORY"` // Location of repository (replaces -r)
	// RESTIC_PASSWORD_FILE    string `yaml:"RESTIC_PASSWORD_FILE"`    // Location of password file (replaces --password-file)
	RESTIC_PASSWORD string `yaml:"RESTIC_PASSWORD"` // The actual password for the repository
	// RESTIC_PASSWORD_COMMAND string `yaml:"RESTIC_PASSWORD_COMMAND"` // Command printing the password for the repository to stdout
	// RESTIC_KEY_HINT     string `yaml:"RESTIC_KEY_HINT"`     // ID of key to try decrypting first, before other keys
	// RESTIC_CACHE_DIR    string `yaml:"RESTIC_CACHE_DIR"`    // Location of the cache directory
	// RESTIC_PROGRESS_FPS string `yaml:"RESTIC_PROGRESS_FPS"` // Frames per second by which the progress bar is updated

	// TMPDIR string `yaml:"TMPDIR"` // Location for temporary files

	AWS_ACCESS_KEY_ID           string `yaml:"AWS_ACCESS_KEY_ID"`           // Amazon S3 access key ID
	AWS_SECRET_ACCESS_KEY       string `yaml:"AWS_SECRET_ACCESS_KEY"`       // Amazon S3 secret access key
	AWS_DEFAULT_REGION          string `yaml:"AWS_DEFAULT_REGION"`          //  Amazon S3 default region
	AWS_PROFILE                 string `yaml:"AWS_PROFILE"`                 // Amazon credentials profile (alternative to specifying key and region)
	AWS_SHARED_CREDENTIALS_FILE string `yaml:"AWS_SHARED_CREDENTIALS_FILE"` // Location of the AWS CLI shared credentials file (default: ~/.aws/credentials)

	ST_AUTH string `yaml:"ST_AUTH"` // Auth URL for keystone v1 authentication
	ST_USER string `yaml:"ST_USER"` // Username for keystone v1 authentication
	ST_KEY  string `yaml:"ST_KEY"`  // Password for keystone v1 authentication

	OS_AUTH_URL    string `yaml:"OS_AUTH_URL"`    // Auth URL for keystone authentication
	OS_REGION_NAME string `yaml:"OS_REGION_NAME"` // Region name for keystone authentication
	OS_USERNAME    string `yaml:"OS_USERNAME"`    // Username for keystone authentication
	OS_USER_ID     string `yaml:"OS_USER_ID"`     // User ID for keystone v3 authentication
	OS_PASSWORD    string `yaml:"OS_PASSWORD"`    // Password for keystone authentication
	OS_TENANT_ID   string `yaml:"OS_TENANT_ID"`   // Tenant ID for keystone v2 authentication
	OS_TENANT_NAME string `yaml:"OS_TENANT_NAME"` // Tenant name for keystone v2 authentication

	OS_USER_DOMAIN_NAME    string `yaml:"OS_USER_DOMAIN_NAME"`    // User domain name for keystone authentication
	OS_USER_DOMAIN_ID      string `yaml:"OS_USER_DOMAIN_ID"`      // User domain ID for keystone v3 authentication
	OS_PROJECT_NAME        string `yaml:"OS_PROJECT_NAME"`        // Project name for keystone authentication
	OS_PROJECT_DOMAIN_NAME string `yaml:"OS_PROJECT_DOMAIN_NAME"` // Project domain name for keystone authentication
	OS_PROJECT_DOMAIN_ID   string `yaml:"OS_PROJECT_DOMAIN_ID"`   // Project domain ID for keystone v3 authentication
	OS_TRUST_ID            string `yaml:"OS_TRUST_ID"`            // Trust ID for keystone v3 authentication

	OS_APPLICATION_CREDENTIAL_ID     string `yaml:"OS_APPLICATION_CREDENTIAL_ID"`     // Application Credential ID (keystone v3)
	OS_APPLICATION_CREDENTIAL_NAME   string `yaml:"OS_APPLICATION_CREDENTIAL_NAME"`   // Application Credential Name (keystone v3)
	OS_APPLICATION_CREDENTIAL_SECRET string `yaml:"OS_APPLICATION_CREDENTIAL_SECRET"` // Application Credential Secret (keystone v3)

	OS_STORAGE_URL string `yaml:"OS_STORAGE_URL"` // Storage URL for token authentication
	OS_AUTH_TOKEN  string `yaml:"OS_AUTH_TOKEN"`  // Auth token for token authentication

	B2_ACCOUNT_ID  string `yaml:"B2_ACCOUNT_ID"`  // Account ID or applicationKeyId for Backblaze B2
	B2_ACCOUNT_KEY string `yaml:"B2_ACCOUNT_KEY"` // Account Key or applicationKey for Backblaze B2

	AZURE_ACCOUNT_NAME string `yaml:"AZURE_ACCOUNT_NAME"` // Account name for Azure
	AZURE_ACCOUNT_KEY  string `yaml:"AZURE_ACCOUNT_KEY"`  // Account key for Azure

	GOOGLE_PROJECT_ID              string `yaml:"GOOGLE_PROJECT_ID"`              // Project ID for Google Cloud Storage
	GOOGLE_APPLICATION_CREDENTIALS string `yaml:"GOOGLE_APPLICATION_CREDENTIALS"` // Application Credentials for Google Cloud Storage (e.g. $HOME/.config/gs-secret-restic-key.json)

	RCLONE_BWLIMIT string `yaml:"RCLONE_BWLIMIT"` // rclone bandwidth limit
}

func (i *Restic) validateEnv() error {
	if i.env.RESTIC_REPOSITORY == "" {
		return errors.New("env RESTIC_REPOSITORY is empty")
	}

	return nil
}

func (i *Restic) getEnv() ([]string, error) {
	env := []string{}
	env = append(env, "RESTIC_PROGRESS_FPS=1")

	{
		d := dir.New()
		d.SetTaskName(i.taskName)
		d.SetDriverName(i.driverName)
		d.SetName("RESTIC_CACHE_DIR")

		path, err := d.CreateAsPersist()
		if err != nil {
			return nil, fmt.Errorf("create restic cache dir: %w", err)
		}

		env = append(env, "RESTIC_CACHE_DIR="+path)
	}

	return append(env, i.getEnvPairs()...), nil
}

// getEnvPairs returns the configured environment variables in NAME=value form.
func (i *Restic) getEnvPairs() []string {
	env := []string{}

	pairs := []struct {
		name  string
		value string
	}{
		{"RESTIC_REPOSITORY", i.env.RESTIC_REPOSITORY},
		{"RESTIC_COMPRESSION", i.config.Compression},
		{"RESTIC_PASSWORD", i.env.RESTIC_PASSWORD},
		{"AWS_ACCESS_KEY_ID", i.env.AWS_ACCESS_KEY_ID},
		{"AWS_SECRET_ACCESS_KEY", i.env.AWS_SECRET_ACCESS_KEY},
		{"AWS_DEFAULT_REGION", i.env.AWS_DEFAULT_REGION},
		{"AWS_PROFILE", i.env.AWS_PROFILE},
		{"AWS_SHARED_CREDENTIALS_FILE", i.env.AWS_SHARED_CREDENTIALS_FILE},
		{"ST_AUTH", i.env.ST_AUTH},
		{"ST_USER", i.env.ST_USER},
		{"ST_KEY", i.env.ST_KEY},
		{"OS_AUTH_URL", i.env.OS_AUTH_URL},
		{"OS_REGION_NAME", i.env.OS_REGION_NAME},
		{"OS_USERNAME", i.env.OS_USERNAME},
		{"OS_USER_ID", i.env.OS_USER_ID},
		{"OS_PASSWORD", i.env.OS_PASSWORD},
		{"OS_TENANT_ID", i.env.OS_TENANT_ID},
		{"OS_TENANT_NAME", i.env.OS_TENANT_NAME},
		{"OS_USER_DOMAIN_NAME", i.env.OS_USER_DOMAIN_NAME},
		{"OS_USER_DOMAIN_ID", i.env.OS_USER_DOMAIN_ID},
		{"OS_PROJECT_NAME", i.env.OS_PROJECT_NAME},
		{"OS_PROJECT_DOMAIN_NAME", i.env.OS_PROJECT_DOMAIN_NAME},
		{"OS_PROJECT_DOMAIN_ID", i.env.OS_PROJECT_DOMAIN_ID},
		{"OS_TRUST_ID", i.env.OS_TRUST_ID},
		{"OS_APPLICATION_CREDENTIAL_ID", i.env.OS_APPLICATION_CREDENTIAL_ID},
		{"OS_APPLICATION_CREDENTIAL_NAME", i.env.OS_APPLICATION_CREDENTIAL_NAME},
		{"OS_APPLICATION_CREDENTIAL_SECRET", i.env.OS_APPLICATION_CREDENTIAL_SECRET},
		{"OS_STORAGE_URL", i.env.OS_STORAGE_URL},
		{"OS_AUTH_TOKEN", i.env.OS_AUTH_TOKEN},
		{"B2_ACCOUNT_ID", i.env.B2_ACCOUNT_ID},
		{"B2_ACCOUNT_KEY", i.env.B2_ACCOUNT_KEY},
		{"AZURE_ACCOUNT_NAME", i.env.AZURE_ACCOUNT_NAME},
		{"AZURE_ACCOUNT_KEY", i.env.AZURE_ACCOUNT_KEY},
		{"GOOGLE_PROJECT_ID", i.env.GOOGLE_PROJECT_ID},
		{"GOOGLE_APPLICATION_CREDENTIALS", i.env.GOOGLE_APPLICATION_CREDENTIALS},
		{"RCLONE_BWLIMIT", i.env.RCLONE_BWLIMIT},
	}

	for _, pair := range pairs {
		if pair.value != "" {
			env = append(env, pair.name+"="+pair.value)
		}
	}

	// Appended last, in a stable order: os/exec keeps the last occurrence of a
	// duplicated variable, so env_extra overrides the typed fields above.
	for _, name := range slices.Sorted(maps.Keys(i.config.ExtraEnv)) {
		env = append(env, name+"="+i.config.ExtraEnv[name])
	}

	return env
}
