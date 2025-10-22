package restic

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type ResticConfigBackupFlags struct {
}

type ResticConfigGlobalFlags struct {
}

type ResticConfigKeep struct {
	Last    int `yaml:"last"`
	Hourly  int `yaml:"hourly"`
	Weekly  int `yaml:"weekly"`
	Monthly int `yaml:"monthly"`
	Yearly  int `yaml:"yearly"`
	Daily   int `yaml:"daily"`
}

type Config struct {
	BinPath string `yaml:"binPath"`

	// GlobalFlags
	Cacert        string `yaml:"cacert"`          // --cacert                     file to load root certificates from (default: use system certificates)
	CleanupCache  bool   `yaml:"cleanup-cache"`   // --cleanup-cache              auto remove old cache directories
	LimitDownload int    `yaml:"limit-download"`  // --limit-download int         limits downloads to a maximum rate in KiB/s. (default: unlimited)
	LimitUpload   int    `yaml:"limit-upload"`    // --limit-upload int           limits uploads to a maximum rate in KiB/s. (default: unlimited)
	NoCache       bool   `yaml:"no-cache"`        //  --no-cache                   do not use a local cache
	NoLock        bool   `yaml:"no-lock"`         //  --no-lock                    do not lock the repository, this allows some operations on read-only repositories
	TLSClientCert string `yaml:"tls-client-cert"` // --tls-client-cert file       path to a file containing PEM encoded TLS client certificate and private key

	// BackupFlags
	ExcludePattern    []string `yaml:"exclude"`             // --exclude pattern                        exclude a pattern (can be specified multiple times)
	ExcludeFiles      []string `yaml:"exclude-file"`        // --exclude-file file                      read exclude patterns from a file (can be specified multiple times)
	ExcludeLargerThan string   `yaml:"exclude-larger-than"` // --exclude-larger-than size               max size of the files to be backed up (allowed suffixes: k/K, m/M, g/G, t/T)
	Host              string   `yaml:"host"`                // --host hostname                          set the hostname for the snapshot manually. To prevent an expensive rescan use the "parent" flag
	IExclude          string   `yaml:"iexclude"`            // --iexclude pattern                       same as --exclude pattern but ignores the casing of filenames
	IExcludeFile      string   `yaml:"iexclude-file"`       // --iexclude-file file                     same as --exclude-file but ignores casing of filenames in patterns
	IgnoreCTime       bool     `yaml:"ignore-ctime"`        // --ignore-ctime                           ignore ctime changes when checking for modified files
	IgnoreINode       bool     `yaml:"ignore-inode"`        // --ignore-inode                           ignore inode number changes when checking for modified files_changed
	OneFileSystem     bool     `yaml:"one-file-system"`     // --one-file-system                        exclude other file systems, don't cross filesystem boundaries and subvolumes
	Tags              []string `yaml:"tags"`                // --tag tags                               add tags for the new snapshot in the format `tag[,tag,...]` (can be specified multiple times) (default [])
	WithATime         bool     `yaml:"with-atime"`          // --with-atime                             store the atime for all files and directories
	Compression       string   `yaml:"compression"`         //  --compression mode                      compression mode (only available for repository format version 2), one of (auto|off|max) (default auto)

	// Forget Flags
	Keep ResticConfigKeep `yaml:"keep"`
}

func (i *Restic) getBackupFlags() []string {
	var flags []string

	if len(i.config.ExcludePattern) != 0 {
		for _, pattern := range i.config.ExcludePattern {
			flags = append(flags, "--exclude", pattern)
		}
	}

	if len(i.config.ExcludeFiles) != 0 {
		for _, file := range i.config.ExcludeFiles {
			flags = append(flags, "--exclude-file", file)
		}
	}

	if i.config.ExcludeLargerThan != "" {
		flags = append(flags, "--exclude-larger-than", i.config.ExcludeLargerThan)
	}

	if i.config.Host != "" {
		flags = append(flags, "--host", i.config.Host)
	}

	if i.config.IExclude != "" {
		flags = append(flags, "--iexclude", i.config.IExclude)
	}

	if i.config.IgnoreCTime {
		flags = append(flags, "--ignore-ctime")
	}

	if i.config.IgnoreINode {
		flags = append(flags, "--ignore-inode")
	}

	if i.config.OneFileSystem {
		flags = append(flags, "--one-file-system")
	}

	if len(i.config.Tags) != 0 {
		flags = append(flags, "--tag", strings.Join(i.config.Tags, ","))
	}

	if i.config.WithATime {
		flags = append(flags, "--with-atime")
	}

	return flags
}

func (i *Restic) getGlobalFlags() []string {
	var flags []string

	if i.config.Cacert != "" {
		flags = append(flags, "--cacert", i.config.Cacert)
	}

	if i.config.CleanupCache {
		flags = append(flags, "--cleanup-cache")
	}

	if i.config.LimitDownload != 0 {
		flags = append(flags, "--limit-download", strconv.Itoa(i.config.LimitDownload))
	}

	if i.config.LimitUpload != 0 {
		flags = append(flags, "--limit-upload", strconv.Itoa(i.config.LimitUpload))
	}

	if i.config.NoCache {
		flags = append(flags, "--no-cache")
	}

	if i.config.NoLock {
		flags = append(flags, "--no-lock")
	}

	if i.config.TLSClientCert != "" {
		flags = append(flags, "--tls-client-cert", i.config.TLSClientCert)
	}

	flags = append(flags, "--json")

	return flags
}

func (i *Restic) getKeepFlags() []string {
	var flags []string

	if i.config.Keep.Daily != 0 {
		flags = append(flags, fmt.Sprintf("--keep-daily=%d", i.config.Keep.Daily))
	}

	if i.config.Keep.Hourly != 0 {
		flags = append(flags, fmt.Sprintf("--keep-hourly=%d", i.config.Keep.Hourly))
	}

	if i.config.Keep.Last != 0 {
		flags = append(flags, fmt.Sprintf("--keep-last=%d", i.config.Keep.Last))
	}

	if i.config.Keep.Monthly != 0 {
		flags = append(flags, fmt.Sprintf("--keep-monthly=%d", i.config.Keep.Monthly))
	}

	if i.config.Keep.Weekly != 0 {
		flags = append(flags, fmt.Sprintf("--keep-weekly=%d", i.config.Keep.Weekly))
	}

	if i.config.Keep.Yearly != 0 {
		flags = append(flags, fmt.Sprintf("--keep-yearly=%d", i.config.Keep.Yearly))
	}

	return flags
}

func (i *Restic) validateConfigBinPath() error {
	if i.config.BinPath != "" {
		_, err := os.Stat(i.config.BinPath)
		if os.IsNotExist(err) {
			return errors.New("bin not exists")
		}

		binPathAbs, err := filepath.Abs(i.config.BinPath)
		if err != nil {
			return err
		}

		i.config.BinPath = binPathAbs
	}

	if !i.isExistsBin() {
		return errors.New("bin not exists")
	}

	return nil
}

func (i *Restic) validateConfigGlobalFlags() error {
	if i.config.Cacert != "" {
		abs, err := filepath.Abs(i.config.Cacert)
		if err != nil {
			return err
		}

		i.config.Cacert = abs
	}

	if i.config.TLSClientCert != "" {
		abs, err := filepath.Abs(i.config.TLSClientCert)
		if err != nil {
			return err
		}

		i.config.TLSClientCert = abs
	}

	return nil
}

func (i *Restic) validateKeep() {
	if i.config.Keep.Daily == 0 &&
		i.config.Keep.Hourly == 0 &&
		i.config.Keep.Last == 0 &&
		i.config.Keep.Monthly == 0 &&
		i.config.Keep.Weekly == 0 &&
		i.config.Keep.Yearly == 0 {
		i.logger.Warn("does not defined section 'keep' in config")
	}
}

func (i *Restic) validateCompression() error {
	if i.config.Compression != "" {
		if i.config.Compression == "off" || i.config.Compression == "auto" || i.config.Compression == "max" {
			return nil
		}

		return fmt.Errorf("unknown compression type '%s', supported [off, auto (DEFAULT), max]", i.config.Compression)
	}

	return nil
}

func (i *Restic) validateConfig() error {
	if err := i.validateConfigBinPath(); err != nil {
		return err
	}

	if err := i.validateConfigGlobalFlags(); err != nil {
		return err
	}

	if err := i.validateCompression(); err != nil {
		return err
	}

	i.validateKeep()

	return nil
}

func (i *Restic) isExistsBin() bool {
	cmd := exec.Command(i.getBin(), "version")

	if _, err := cmd.Output(); err != nil {
		return false
	}

	return true
}

func (i *Restic) getBin() string {
	if i.config.BinPath != "" {
		return i.config.BinPath
	}

	return "restic"
}
