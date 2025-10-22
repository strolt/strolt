package restic

import (
	"errors"

	"github.com/strolt/strolt/apps/strolt/internal/util/dir"
)

type Env struct {
	// RESTIC_REPOSITORY_FILE  string `yaml:"RESTIC_REPOSITORY_FILE"`  // Name of file containing the repository location (replaces --repository-file)
	RESTIC_REPOSITORY string `yaml:"RESTIC_REPOSITORY"` //nolint:stylecheck // Location of repository (replaces -r)
	// RESTIC_PASSWORD_FILE    string `yaml:"RESTIC_PASSWORD_FILE"`    // Location of password file (replaces --password-file)
	RESTIC_PASSWORD string `yaml:"RESTIC_PASSWORD"` //nolint:stylecheck // The actual password for the repository
	// RESTIC_PASSWORD_COMMAND string `yaml:"RESTIC_PASSWORD_COMMAND"` // Command printing the password for the repository to stdout
	// RESTIC_KEY_HINT     string `yaml:"RESTIC_KEY_HINT"`     // ID of key to try decrypting first, before other keys
	// RESTIC_CACHE_DIR    string `yaml:"RESTIC_CACHE_DIR"`    // Location of the cache directory
	// RESTIC_PROGRESS_FPS string `yaml:"RESTIC_PROGRESS_FPS"` // Frames per second by which the progress bar is updated

	// TMPDIR string `yaml:"TMPDIR"` // Location for temporary files

	AWS_ACCESS_KEY_ID           string `yaml:"AWS_ACCESS_KEY_ID"`           //nolint:stylecheck // Amazon S3 access key ID
	AWS_SECRET_ACCESS_KEY       string `yaml:"AWS_SECRET_ACCESS_KEY"`       //nolint:stylecheck // Amazon S3 secret access key
	AWS_DEFAULT_REGION          string `yaml:"AWS_DEFAULT_REGION"`          //nolint:stylecheck //  Amazon S3 default region
	AWS_PROFILE                 string `yaml:"AWS_PROFILE"`                 //nolint:stylecheck // Amazon credentials profile (alternative to specifying key and region)
	AWS_SHARED_CREDENTIALS_FILE string `yaml:"AWS_SHARED_CREDENTIALS_FILE"` //nolint:stylecheck // Location of the AWS CLI shared credentials file (default: ~/.aws/credentials)

	ST_AUTH string `yaml:"ST_AUTH"` //nolint:stylecheck // Auth URL for keystone v1 authentication
	ST_USER string `yaml:"ST_USER"` //nolint:stylecheck // Username for keystone v1 authentication
	ST_KEY  string `yaml:"ST_KEY"`  //nolint:stylecheck // Password for keystone v1 authentication

	OS_AUTH_URL    string `yaml:"OS_AUTH_URL"`    //nolint:stylecheck // Auth URL for keystone authentication
	OS_REGION_NAME string `yaml:"OS_REGION_NAME"` //nolint:stylecheck // Region name for keystone authentication
	OS_USERNAME    string `yaml:"OS_USERNAME"`    //nolint:stylecheck // Username for keystone authentication
	OS_USER_ID     string `yaml:"OS_USER_ID"`     //nolint:stylecheck // User ID for keystone v3 authentication
	OS_PASSWORD    string `yaml:"OS_PASSWORD"`    //nolint:stylecheck // Password for keystone authentication
	OS_TENANT_ID   string `yaml:"OS_TENANT_ID"`   //nolint:stylecheck // Tenant ID for keystone v2 authentication
	OS_TENANT_NAME string `yaml:"OS_TENANT_NAME"` //nolint:stylecheck // Tenant name for keystone v2 authentication

	OS_USER_DOMAIN_NAME    string `yaml:"OS_USER_DOMAIN_NAME"`    //nolint:stylecheck // User domain name for keystone authentication
	OS_USER_DOMAIN_ID      string `yaml:"OS_USER_DOMAIN_ID"`      //nolint:stylecheck // User domain ID for keystone v3 authentication
	OS_PROJECT_NAME        string `yaml:"OS_PROJECT_NAME"`        //nolint:stylecheck // Project name for keystone authentication
	OS_PROJECT_DOMAIN_NAME string `yaml:"OS_PROJECT_DOMAIN_NAME"` //nolint:stylecheck // Project domain name for keystone authentication
	OS_PROJECT_DOMAIN_ID   string `yaml:"OS_PROJECT_DOMAIN_ID"`   //nolint:stylecheck // Project domain ID for keystone v3 authentication
	OS_TRUST_ID            string `yaml:"OS_TRUST_ID"`            //nolint:stylecheck // Trust ID for keystone v3 authentication

	OS_APPLICATION_CREDENTIAL_ID     string `yaml:"OS_APPLICATION_CREDENTIAL_ID"`     //nolint:stylecheck // Application Credential ID (keystone v3)
	OS_APPLICATION_CREDENTIAL_NAME   string `yaml:"OS_APPLICATION_CREDENTIAL_NAME"`   //nolint:stylecheck // Application Credential Name (keystone v3)
	OS_APPLICATION_CREDENTIAL_SECRET string `yaml:"OS_APPLICATION_CREDENTIAL_SECRET"` //nolint:stylecheck // Application Credential Secret (keystone v3)

	OS_STORAGE_URL string `yaml:"OS_STORAGE_URL"` //nolint:stylecheck // Storage URL for token authentication
	OS_AUTH_TOKEN  string `yaml:"OS_AUTH_TOKEN"`  //nolint:stylecheck // Auth token for token authentication

	B2_ACCOUNT_ID  string `yaml:"B2_ACCOUNT_ID"`  //nolint:stylecheck // Account ID or applicationKeyId for Backblaze B2
	B2_ACCOUNT_KEY string `yaml:"B2_ACCOUNT_KEY"` //nolint:stylecheck // Account Key or applicationKey for Backblaze B2

	AZURE_ACCOUNT_NAME string `yaml:"AZURE_ACCOUNT_NAME"` //nolint:stylecheck // Account name for Azure
	AZURE_ACCOUNT_KEY  string `yaml:"AZURE_ACCOUNT_KEY"`  //nolint:stylecheck // Account key for Azure

	GOOGLE_PROJECT_ID              string `yaml:"GOOGLE_PROJECT_ID"`              //nolint:stylecheck // Project ID for Google Cloud Storage
	GOOGLE_APPLICATION_CREDENTIALS string `yaml:"GOOGLE_APPLICATION_CREDENTIALS"` //nolint:stylecheck // Application Credentials for Google Cloud Storage (e.g. $HOME/.config/gs-secret-restic-key.json)

	RCLONE_BWLIMIT string `yaml:"RCLONE_BWLIMIT"` //nolint:stylecheck // rclone bandwidth limit
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
			return nil, err
		}

		env = append(env, "RESTIC_CACHE_DIR="+path)
	}

	if i.env.RESTIC_REPOSITORY != "" {
		env = append(env, "RESTIC_REPOSITORY="+i.env.RESTIC_REPOSITORY)
	}

	if i.config.Compression != "" {
		env = append(env, "RESTIC_COMPRESSION="+i.config.Compression)
	}

	if i.env.RESTIC_PASSWORD != "" {
		env = append(env, "RESTIC_PASSWORD="+i.env.RESTIC_PASSWORD)
	}

	if i.env.AWS_ACCESS_KEY_ID != "" {
		env = append(env, "AWS_ACCESS_KEY_ID="+i.env.AWS_ACCESS_KEY_ID)
	}

	if i.env.AWS_SECRET_ACCESS_KEY != "" {
		env = append(env, "AWS_SECRET_ACCESS_KEY="+i.env.AWS_SECRET_ACCESS_KEY)
	}

	if i.env.AWS_DEFAULT_REGION != "" {
		env = append(env, "AWS_DEFAULT_REGION="+i.env.AWS_DEFAULT_REGION)
	}

	if i.env.AWS_PROFILE != "" {
		env = append(env, "AWS_PROFILE="+i.env.AWS_PROFILE)
	}

	if i.env.AWS_SHARED_CREDENTIALS_FILE != "" {
		env = append(env, "AWS_SHARED_CREDENTIALS_FILE="+i.env.AWS_SHARED_CREDENTIALS_FILE)
	}

	if i.env.ST_AUTH != "" {
		env = append(env, "ST_AUTH="+i.env.ST_AUTH)
	}

	if i.env.ST_USER != "" {
		env = append(env, "ST_USER="+i.env.ST_USER)
	}

	if i.env.ST_KEY != "" {
		env = append(env, "ST_KEY="+i.env.ST_KEY)
	}

	if i.env.OS_AUTH_URL != "" {
		env = append(env, "OS_AUTH_URL="+i.env.OS_AUTH_URL)
	}

	if i.env.OS_REGION_NAME != "" {
		env = append(env, "OS_REGION_NAME="+i.env.OS_REGION_NAME)
	}

	if i.env.OS_USERNAME != "" {
		env = append(env, "OS_USERNAME="+i.env.OS_USERNAME)
	}

	if i.env.OS_USER_ID != "" {
		env = append(env, "OS_USER_ID="+i.env.OS_USER_ID)
	}

	if i.env.OS_PASSWORD != "" {
		env = append(env, "OS_PASSWORD="+i.env.OS_PASSWORD)
	}

	if i.env.OS_TENANT_ID != "" {
		env = append(env, "OS_TENANT_ID="+i.env.OS_TENANT_ID)
	}

	if i.env.OS_TENANT_NAME != "" {
		env = append(env, "OS_TENANT_NAME="+i.env.OS_TENANT_NAME)
	}

	if i.env.OS_USER_DOMAIN_NAME != "" {
		env = append(env, "OS_USER_DOMAIN_NAME="+i.env.OS_USER_DOMAIN_NAME)
	}

	if i.env.OS_USER_DOMAIN_ID != "" {
		env = append(env, "OS_USER_DOMAIN_ID="+i.env.OS_USER_DOMAIN_ID)
	}

	if i.env.OS_PROJECT_NAME != "" {
		env = append(env, "OS_PROJECT_NAME="+i.env.OS_PROJECT_NAME)
	}

	if i.env.OS_PROJECT_DOMAIN_NAME != "" {
		env = append(env, "OS_PROJECT_DOMAIN_NAME="+i.env.OS_PROJECT_DOMAIN_NAME)
	}

	if i.env.OS_PROJECT_DOMAIN_ID != "" {
		env = append(env, "OS_PROJECT_DOMAIN_ID="+i.env.OS_PROJECT_DOMAIN_ID)
	}

	if i.env.OS_TRUST_ID != "" {
		env = append(env, "OS_TRUST_ID="+i.env.OS_TRUST_ID)
	}

	if i.env.OS_APPLICATION_CREDENTIAL_ID != "" {
		env = append(env, "OS_APPLICATION_CREDENTIAL_ID="+i.env.OS_APPLICATION_CREDENTIAL_ID)
	}

	if i.env.OS_APPLICATION_CREDENTIAL_NAME != "" {
		env = append(env, "OS_APPLICATION_CREDENTIAL_NAME="+i.env.OS_APPLICATION_CREDENTIAL_NAME)
	}

	if i.env.OS_APPLICATION_CREDENTIAL_SECRET != "" {
		env = append(env, "OS_APPLICATION_CREDENTIAL_SECRET="+i.env.OS_APPLICATION_CREDENTIAL_SECRET)
	}

	if i.env.OS_STORAGE_URL != "" {
		env = append(env, "OS_STORAGE_URL="+i.env.OS_STORAGE_URL)
	}

	if i.env.OS_AUTH_TOKEN != "" {
		env = append(env, "OS_AUTH_TOKEN="+i.env.OS_AUTH_TOKEN)
	}

	if i.env.B2_ACCOUNT_ID != "" {
		env = append(env, "B2_ACCOUNT_ID="+i.env.B2_ACCOUNT_ID)
	}

	if i.env.B2_ACCOUNT_KEY != "" {
		env = append(env, "B2_ACCOUNT_KEY="+i.env.B2_ACCOUNT_KEY)
	}

	if i.env.AZURE_ACCOUNT_NAME != "" {
		env = append(env, "AZURE_ACCOUNT_NAME="+i.env.AZURE_ACCOUNT_NAME)
	}

	if i.env.AZURE_ACCOUNT_KEY != "" {
		env = append(env, "AZURE_ACCOUNT_KEY="+i.env.AZURE_ACCOUNT_KEY)
	}

	if i.env.GOOGLE_PROJECT_ID != "" {
		env = append(env, "GOOGLE_PROJECT_ID="+i.env.GOOGLE_PROJECT_ID)
	}

	if i.env.GOOGLE_APPLICATION_CREDENTIALS != "" {
		env = append(env, "GOOGLE_APPLICATION_CREDENTIALS="+i.env.GOOGLE_APPLICATION_CREDENTIALS)
	}

	if i.env.RCLONE_BWLIMIT != "" {
		env = append(env, "RCLONE_BWLIMIT="+i.env.RCLONE_BWLIMIT)
	}

	return env, nil
}
