package e2e_test

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"maps"
	"path"
	"slices"
	"strings"
)

const fsInputPath = "/e2e/input"

// Text fixtures: relative path -> content (a trailing newline is added on
// write). Names exercise unicode, emoji and spaces; the empty file checks
// zero-length handling.
var textFiles = map[string]string{
	"0.txt":                              "0",
	"1/1.txt":                            "1",
	"1/2/2.txt":                          "2",
	"unicode/файл-🚀.txt":                 "привет мир",
	"dir with space/file with space.txt": "space",
	"empty/empty.txt":                    "",
}

// Binary fixtures created inside the container; their content is reproduced
// in Go by expectedManifest, so corruption is detected end to end.
const (
	zerosPath = "bin/zeros.bin"
	zerosSize = 65536
	seqPath   = "bin/seq.txt"
	seqCount  = 10000
)

type Fs struct{}

func fs() *Fs {
	return &Fs{}
}

func (fs *Fs) createData() error {
	for rel, content := range textFiles {
		abs := fsInputPath + "/" + rel

		cmd := fmt.Sprintf("mkdir -p '%s' && ", path.Dir(abs))
		if content == "" {
			cmd += fmt.Sprintf(": > '%s'", abs)
		} else {
			cmd += fmt.Sprintf("printf '%%s\\n' '%s' > '%s'", content, abs)
		}

		if _, err := execInStrolt(cmd); err != nil {
			return err
		}
	}

	cmd := fmt.Sprintf("mkdir -p '%s/bin' && head -c %d /dev/zero > '%s/%s' && seq 0 %d > '%s/%s'",
		fsInputPath, zerosSize, fsInputPath, zerosPath, seqCount-1, fsInputPath, seqPath)
	if _, err := execInStrolt(cmd); err != nil {
		return err
	}

	return nil
}

func (fs *Fs) dropData() error {
	if _, err := execInStrolt(fmt.Sprintf("rm -rf %s/* 2>/dev/null; true", fsInputPath)); err != nil {
		return err
	}

	return nil
}

// hashAll returns the sha256 manifest (relative path -> hash) of every file
// currently present under fsInputPath, computed inside the container.
func (fs *Fs) hashAll() (map[string]string, error) {
	o, err := execInStrolt(fmt.Sprintf("cd '%s' && find . -type f -exec sha256sum {} +", fsInputPath))
	if err != nil {
		return nil, err
	}

	manifest := map[string]string{}

	for line := range strings.SplitSeq(string(o), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		hash, p, ok := strings.Cut(line, "  ")
		if !ok {
			return nil, fmt.Errorf("unexpected sha256sum output line: %q", line)
		}

		manifest[strings.TrimPrefix(p, "./")] = hash
	}

	return manifest, nil
}

// expectedManifest reproduces the fixture content in Go, so the comparison
// does not depend on any state captured inside the container.
func expectedManifest() map[string]string {
	manifest := map[string]string{}

	for rel, content := range textFiles {
		if content == "" {
			manifest[rel] = sha256Hex(nil)
		} else {
			manifest[rel] = sha256Hex([]byte(content + "\n"))
		}
	}

	manifest[zerosPath] = sha256Hex(make([]byte, zerosSize))

	var seq strings.Builder
	for i := range seqCount {
		fmt.Fprintf(&seq, "%d\n", i)
	}

	manifest[seqPath] = sha256Hex([]byte(seq.String()))

	return manifest
}

func sha256Hex(data []byte) string {
	sum := sha256.Sum256(data)

	return hex.EncodeToString(sum[:])
}

// checkValidData verifies the input directory matches the fixture set
// exactly: every fixture file exists with the right content and nothing
// extra is present.
func (fs *Fs) checkValidData() error {
	got, err := fs.hashAll()
	if err != nil {
		return err
	}

	want := expectedManifest()

	if !maps.Equal(got, want) {
		return fmt.Errorf("input directory differs from fixtures:\nwant files: %v\ngot files: %v",
			slices.Sorted(maps.Keys(want)), slices.Sorted(maps.Keys(got)))
	}

	return nil
}
