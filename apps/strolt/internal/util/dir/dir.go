package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"
)

type Directory struct {
	serviceName string
	taskName    string
	driverName  string
	name        string
	isTemp      bool
}

const prefix = "strolt-data"

func New() *Directory {
	return &Directory{}
}

func (d *Directory) SetServiceName(serviceName string) {
	d.serviceName = serviceName
}

func (d *Directory) SetTaskName(taskName string) {
	d.taskName = taskName
}

func (d *Directory) SetDriverName(driverName string) {
	d.driverName = driverName
}

func (d *Directory) SetName(name string) {
	d.name = name
}

func getBasePath(isTemp bool) (string, error) {
	path := ""

	if env.PathData() != "" {
		p, err := filepath.Abs(env.PathData())
		if err != nil {
			return "", err
		}

		path = p
	} else {
		p, err := filepath.Abs(filepath.Dir(os.Args[0]))
		if err != nil {
			return "", err
		}

		path = p
	}

	parts := []string{path, prefix}

	if isTemp {
		parts = append(parts, "tmp")
	}

	return filepath.Join(parts...), nil
}

func (d *Directory) path() (string, error) {
	path, err := getBasePath(d.isTemp)
	if err != nil {
		return "", err
	}

	parts := []string{path}

	if d.serviceName != "" {
		parts = append(parts, fmt.Sprintf("i-%s", d.serviceName))
	}

	if d.taskName != "" {
		parts = append(parts, fmt.Sprintf("t-%s", d.taskName))
	}

	if d.driverName != "" {
		parts = append(parts, fmt.Sprintf("d-%s", d.driverName))
	}

	if d.name != "" {
		parts = append(parts, fmt.Sprintf("n-%s", d.name))
	}

	return filepath.Join(parts...), nil
}

func (d *Directory) create() (string, error) {
	log := logger.New()

	path, err := d.path()
	if err != nil {
		return "", err
	}

	isExists, err := Exists(path)
	if err != nil {
		return "", err
	}

	if isExists {
		log.Debug(fmt.Sprintf("error create directory '%s' - already exists", path))
	} else {
		if err := os.MkdirAll(path, 0o700); err != nil {
			return "", err
		}

		log.Debug(fmt.Sprintf("created directory '%s'", path))
	}

	return path, nil
}

func (d *Directory) CreateAsTmp() (string, error) {
	d.isTemp = true
	return d.create()
}

func (d *Directory) CreateAsPersist() (string, error) {
	return d.create()
}

func Remove(path string) error {
	return os.RemoveAll(path)
}

func RemoveTempDirectories() error {
	path, err := getBasePath(true)
	if err != nil {
		return err
	}

	isExists, err := Exists(path)
	if err != nil {
		return err
	}

	if isExists {
		return os.RemoveAll(path)
	}

	return nil
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}

	if os.IsNotExist(err) {
		return false, nil
	}

	return false, err
}

func IsNotADirectoryError(err error) bool {
	return strings.HasSuffix(err.Error(), "not a directory")
}

func Exists(path string) (bool, error) {
	isExists, err := exists(path)
	if err != nil {
		if IsNotADirectoryError(err) {
			return false, nil
		}

		return false, err
	}

	if !isExists {
		return false, nil
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		return false, err
	}

	return fileInfo.IsDir(), nil
}
