// Package dir manages the strolt data and temporary directories.
package dir

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/strolt/strolt/apps/strolt/internal/env"
	"github.com/strolt/strolt/shared/logger"
)

// Directory builds paths for strolt data directories and creates them.
type Directory struct {
	serviceName string
	taskName    string
	driverName  string
	name        string
	isTemp      bool
}

const prefix = "strolt-data"

// New creates an empty Directory builder.
func New() *Directory {
	return &Directory{}
}

// SetServiceName sets the service name part of the directory path.
func (d *Directory) SetServiceName(serviceName string) {
	d.serviceName = serviceName
}

// SetTaskName sets the task name part of the directory path.
func (d *Directory) SetTaskName(taskName string) {
	d.taskName = taskName
}

// SetDriverName sets the driver name part of the directory path.
func (d *Directory) SetDriverName(driverName string) {
	d.driverName = driverName
}

// SetName sets the name part of the directory path.
func (d *Directory) SetName(name string) {
	d.name = name
}

func getBasePath(isTemp bool) (string, error) {
	base := env.PathData()
	if base == "" {
		base = filepath.Dir(os.Args[0])
	}

	path, err := filepath.Abs(base)
	if err != nil {
		return "", fmt.Errorf("resolve absolute path: %w", err)
	}

	parts := []string{path, prefix}

	if isTemp {
		parts = append(parts, "tmp")
	}

	return filepath.Join(parts...), nil
}

// CreateAsTmp creates the directory under the temporary base path.
func (d *Directory) CreateAsTmp() (string, error) {
	d.isTemp = true
	return d.create()
}

// CreateAsPersist creates the directory under the persistent base path.
func (d *Directory) CreateAsPersist() (string, error) {
	return d.create()
}

func (d *Directory) path() (string, error) {
	path, err := getBasePath(d.isTemp)
	if err != nil {
		return "", err
	}

	parts := []string{path}

	if d.serviceName != "" {
		parts = append(parts, "i-"+d.serviceName)
	}

	if d.taskName != "" {
		parts = append(parts, "t-"+d.taskName)
	}

	if d.driverName != "" {
		parts = append(parts, "d-"+d.driverName)
	}

	if d.name != "" {
		parts = append(parts, "n-"+d.name)
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
			return "", fmt.Errorf("create directory: %w", err)
		}

		log.Debug(fmt.Sprintf("created directory '%s'", path))
	}

	return path, nil
}

// Remove deletes the directory at the given path with all its contents.
func Remove(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove directory: %w", err)
	}

	return nil
}

// RemoveTempDirectories deletes the temporary base directory if it exists.
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
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("remove temp directory: %w", err)
		}
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

	return false, fmt.Errorf("stat: %w", err)
}

// IsNotADirectoryError reports whether the error means the path is not a directory.
func IsNotADirectoryError(err error) bool {
	return strings.HasSuffix(err.Error(), "not a directory")
}

// Exists reports whether the given path exists and is a directory.
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
		return false, fmt.Errorf("stat: %w", err)
	}

	return fileInfo.IsDir(), nil
}
