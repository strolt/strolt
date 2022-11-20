package e2e_test

import (
	"fmt"
	"path/filepath"
	"strings"

	_ "github.com/lib/pq"
)

type File struct {
	Path  string
	Value string
}

var (
	fsInputPath = "/e2e/input"
)

var (
	files = []File{
		{
			Path:  filepath.Join(fsInputPath, "0.txt"),
			Value: "0",
		},
		{
			Path:  filepath.Join(fsInputPath, "1", "1.txt"),
			Value: "1",
		},
		{
			Path:  filepath.Join(fsInputPath, "1", "2", "2.txt"),
			Value: "2",
		},
	}
)

type Fs struct{}

func fs() *Fs {
	return &Fs{}
}

func (fs *Fs) isFile(path string) (bool, string) {
	o, err := runDockerComposeBash(fmt.Sprintf("cat %s", path))
	return err == nil, strings.Join(strings.Split(string(o), "\n")[:1], "\n")
}

func (fs *Fs) scan() ([]File, error) {
	o, err := runDockerComposeBash(fmt.Sprintf("ls -R1 %s", fsInputPath))
	if err != nil {
		return files, err
	}

	lines := strings.Split(string(o), "\n\n")

	for _, line := range lines {
		l := strings.Split(line, "\n")
		path := strings.TrimSuffix(l[0], ":")

		for _, f := range l[1:] {
			_path := filepath.Join(path, f)

			isFile, fileValue := fs.isFile(_path)
			if isFile {
				files = append(files, File{
					Path:  filepath.Join(path, f),
					Value: fileValue,
				})
			}
		}
	}

	return files, err
}

func (fs *Fs) createData() error {
	for _, file := range files {
		if _, err := runDockerComposeBash(fmt.Sprintf("mkdir -p %s", filepath.Dir(file.Path))); err != nil {
			fmt.Println(err) //nolint:forbidigo
			return err
		}

		if _, err := runDockerComposeBash(fmt.Sprintf("echo \"%s\" > %s", file.Value, file.Path)); err != nil {
			return err
		}
	}

	return nil
}

func (fs *Fs) dropData() error {
	_, err := runDockerComposeBash(fmt.Sprintf("rm -rf %s/*", fsInputPath))
	if err != nil {
		return err
	}

	return nil
}

func (file File) exists() (bool, File) {
	for _, _file := range files {
		if file.Path == _file.Path {
			return true, _file
		}
	}

	return false, File{}
}

func (fs *Fs) checkValidData() error {
	scannedFiles, err := fs.scan()
	if err != nil {
		return err
	}

	for _, scannedFile := range scannedFiles {
		isExists, _file := scannedFile.exists()
		if !isExists {
			return fmt.Errorf("'%s' not exists in mock", scannedFile.Path)
		}

		if _file.Value != scannedFile.Value {
			return fmt.Errorf("'%s' different content", scannedFile.Path)
		}
	}

	return nil
}
