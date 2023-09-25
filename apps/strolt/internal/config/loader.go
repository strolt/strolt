package config

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/apps/strolt/internal/sctxt"
	"github.com/strolt/strolt/shared/utils"

	_ "time/tzdata"

	"gopkg.in/yaml.v3"
)

var (
	loadedAt time.Time
)

func GetLoadedAt() time.Time {
	return loadedAt
}

type FileInfo struct {
	Config Config

	ConfigPathname              string
	ExtendedConfigPathnameList  []string
	ExtendedSecretsPathnameList []string

	ExtendedFileInfoList []FileInfo
	ExtendedSecretsList  []Secrets
}

type loaded struct {
	Config   Config
	FileList []string
}

func scan(pathname string) (*FileInfo, error) {
	configPathname, err := filepath.Abs(pathname)
	if err != nil {
		return nil, err
	}

	fi := FileInfo{
		ConfigPathname: configPathname,
	}

	configData, err := os.ReadFile(pathname)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(configData, &fi.Config); err != nil {
		return nil, err
	}

	for _, pathname := range fi.Config.Extends.Configs {
		extendedConfigPathname, err := filepath.Abs(path.Join(path.Dir(configPathname), pathname))
		if err != nil {
			return nil, err
		}

		if configPathname == extendedConfigPathname {
			return nil, fmt.Errorf("recursive config load '%s'", extendedConfigPathname)
		}

		fi.ExtendedConfigPathnameList = append(fi.ExtendedConfigPathnameList, extendedConfigPathname)

		extendedConfigFileInfo, err := scan(extendedConfigPathname)
		if err != nil {
			return nil, err
		}

		fi.ExtendedFileInfoList = append(fi.ExtendedFileInfoList, *extendedConfigFileInfo)
	}

	for _, pathname := range fi.Config.Extends.Secrets {
		extendedSecretsPathname, err := filepath.Abs(path.Join(path.Dir(configPathname), pathname))
		if err != nil {
			return nil, err
		}

		fi.ExtendedSecretsPathnameList = append(fi.ExtendedSecretsPathnameList, extendedSecretsPathname)

		extendedSecrets, err := loadSecrets(extendedSecretsPathname)
		if err != nil {
			return nil, err
		}

		fi.ExtendedSecretsList = append(fi.ExtendedSecretsList, *extendedSecrets)
	}

	fi.Config.Extends = Extends{}

	return &fi, nil
}

func (fi *FileInfo) fileList(isMain bool) []string {
	fl := []string{}

	if isMain {
		fl = append(fl, fi.ConfigPathname)
	}

	fl = append(fl, fi.ExtendedConfigPathnameList...)
	fl = append(fl, fi.ExtendedSecretsPathnameList...)

	for _, extendedFi := range fi.ExtendedFileInfoList {
		fl = append(fl, extendedFi.fileList(false)...)
	}

	return fl
}

func loadSecrets(pathname string) (*Secrets, error) {
	s := Secrets{}

	secretsData, err := os.ReadFile(pathname)
	if err != nil {
		return nil, err
	}

	err = yaml.Unmarshal(secretsData, &s)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (c *Config) setDefaults() error {
	{
		if c.TimeZone == "" {
			c.TimeZone = utils.TimeGetDefaultTimeZone()
		}

		timeLocation, err := time.LoadLocation(c.TimeZone)
		if err != nil {
			return err
		}

		c.timeLocation = timeLocation
	}

	for notificationName, notification := range c.Definitions.Notifications {
		if len(notification.Events) == 0 {
			notification.Events = []sctxt.EventType{sctxt.EvOperationStop, sctxt.EvOperationError}
		}

		c.Definitions.Notifications[notificationName] = notification
	}

	for serviceName, service := range c.Services {
		for taskName, task := range service {
			if task.OperationMode == "" {
				task.OperationMode = OperationModePreferPipe
				c.Services[serviceName][taskName] = task
			}
		}
	}

	return nil
}

func load(pathname string) (loaded, error) {
	fi, err := scan(pathname)
	if err != nil {
		return loaded{}, err
	}

	fi.ExtendedFileInfoList = append(fi.ExtendedFileInfoList,
		FileInfo{
			Config: getCliConfig(),
		},
		FileInfo{
			Config: Config{
				Tags: env.GlobalTags(),
			},
		})

	c, err := fi.merge()
	if err != nil {
		return loaded{}, err
	}

	if err := c.mergeDestinationExtends(); err != nil {
		return loaded{}, err
	}

	if err := c.replaceSecrets(); err != nil {
		return loaded{}, err
	}

	c.Secrets = map[string]string{}
	c.Definitions = Definitions{
		Notifications: c.Definitions.Notifications,
	}

	if err := c.setDefaults(); err != nil {
		return loaded{}, err
	}

	return loaded{
		Config:   *c,
		FileList: fi.fileList(true),
	}, nil
}

func Load(pathname string) error {
	l, err := load(pathname)
	if err != nil {
		return err
	}

	if err := l.Config.validate(); err != nil {
		return errors.Wrapf(err, "validate")
	}

	config = l.Config
	fileList = l.FileList

	// {
	// 	//TODO: remove this
	// 	data, _ := yaml.Marshal(config)
	// 	os.WriteFile("result.config.yml", data, 0o700)
	// }

	loadedAt = time.Now()

	return nil
}
