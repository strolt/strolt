package e2e_test

import (
	"fmt"
	"path/filepath"
	"strings"
)

type File struct {
	Path  string
	Value string
}

var fsInputPath = "/e2e/input"

// files is the immutable fixture set; scan() must never modify it.
var files = []File{
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

type Fs struct{}

func fs() *Fs {
	return &Fs{}
}

func (fs *Fs) isFile(path string) (bool, string) {
	o, err := execInStrolt("cat " + path)
	if err != nil {
		return false, ""
	}

	value, _, _ := strings.Cut(string(o), "\n")

	return true, value
}

// scan returns the files currently present under fsInputPath.
func (fs *Fs) scan() ([]File, error) {
	o, err := execInStrolt("ls -R1 " + fsInputPath)
	if err != nil {
		return nil, err
	}

	var scanned []File

	for block := range strings.SplitSeq(string(o), "\n\n") {
		lines := strings.Split(strings.TrimSpace(block), "\n")
		if len(lines) == 0 || lines[0] == "" {
			continue
		}

		dir := strings.TrimSuffix(lines[0], ":")

		for _, entry := range lines[1:] {
			if entry == "" {
				continue
			}

			path := filepath.Join(dir, entry)

			isFile, fileValue := fs.isFile(path)
			if isFile {
				scanned = append(scanned, File{
					Path:  path,
					Value: fileValue,
				})
			}
		}
	}

	return scanned, nil
}

func (fs *Fs) createData() error {
	for _, file := range files {
		if _, err := execInStrolt("mkdir -p " + filepath.Dir(file.Path)); err != nil {
			return err
		}

		if _, err := execInStrolt(fmt.Sprintf("echo \"%s\" > %s", file.Value, file.Path)); err != nil {
			return err
		}
	}

	return nil
}

func (fs *Fs) dropData() error {
	_, err := execInStrolt(fmt.Sprintf("rm -rf %s/*", fsInputPath))
	if err != nil {
		return err
	}

	return nil
}

// checkValidData verifies the input directory matches the fixture set
// exactly: every fixture file exists with the right content and nothing
// extra is present.
func (fs *Fs) checkValidData() error {
	scannedFiles, err := fs.scan()
	if err != nil {
		return err
	}

	if len(scannedFiles) != len(files) {
		return fmt.Errorf("expected %d files in %s, found %d: %+v", len(files), fsInputPath, len(scannedFiles), scannedFiles)
	}

	for _, want := range files {
		found := false

		for _, got := range scannedFiles {
			if got.Path != want.Path {
				continue
			}

			if got.Value != want.Value {
				return fmt.Errorf("'%s' different content: want %q, got %q", want.Path, want.Value, got.Value)
			}

			found = true

			break
		}

		if !found {
			return fmt.Errorf("'%s' not found in %s", want.Path, fsInputPath)
		}
	}

	return nil
}
