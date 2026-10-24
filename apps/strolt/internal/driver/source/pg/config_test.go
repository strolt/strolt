package pg

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetArgs_paramsAreSplitIntoArgvElements pins the regression where
// backup_params/restore_params/common_params were passed to exec as ONE argv
// element: a config like "--no-owner --no-privileges" then made the pg tool
// abort with an invalid-option error before touching any data — a restore
// into an empty database that never worked.
func TestGetArgs_paramsAreSplitIntoArgvElements(t *testing.T) {
	d := New()
	d.config.Format = FormatCustom
	d.config.Database = "db"
	d.config.CommonParams = "--common-a --common-b"
	d.config.BackupParams = "--backup-a --backup-b"
	d.config.RestoreParams = "--restore-a --restore-b"

	assert.Equal(t,
		[]string{"--backup-a", "--backup-b", "--common-a", "--common-b", "--no-password", "--dbname=db", "--format=c"},
		d.getBackupArgs(),
		"backup params must become separate argv elements",
	)

	assert.Equal(t,
		[]string{"--restore-a", "--restore-b", "--common-a", "--common-b", "--no-password", "--dbname=db", "--exit-on-error"},
		d.getRestoreArgs(),
		"restore params must become separate argv elements",
	)
}

// TestSetConfig_rejectsBrokenParamsQuoting pins the fail-fast contract: a
// params string that cannot be tokenized must be rejected when the config is
// loaded, not discovered as an invalid-option crash at restore time.
func TestSetConfig_rejectsBrokenParamsQuoting(t *testing.T) {
	d := New()

	err := d.SetConfig(map[string]any{
		"database":       "db",
		"format":         "c",
		"restore_params": `--comment="unterminated`,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore_params")
	assert.Contains(t, err.Error(), "unterminated")
}

func TestSetConfig_acceptsQuotedValues(t *testing.T) {
	d := New()

	require.NoError(t, d.SetConfig(map[string]any{
		"database":       "db",
		"format":         "c",
		"restore_params": `--no-owner --comment="two words"`,
	}))

	args := d.getRestoreArgs()
	assert.Equal(t, []string{"--no-owner", "--comment=two words"}, args[0:2],
		"a quoted value with spaces must stay one argv element",
	)
}

// TestRealBinary_acceptsSplitRestoreParams runs the real pg_restore with the
// argv strolt builds for a multi-option restore_params config, next to the
// old single-element argv. pg tools special-case --version at argv[1] only,
// so /dev/null is used as the archive: with accepted options the failure is
// about the file's content, never about unrecognized options.
func TestRealBinary_acceptsSplitRestoreParams(t *testing.T) {
	bin, err := exec.LookPath("pg_restore")
	if err != nil {
		t.Skip("pg_restore not available")
	}

	d := New()
	d.config.RestoreParams = "--no-owner --no-privileges"

	// The pre-fix argv: params as one element.
	oldCmd := exec.Command(bin, d.config.RestoreParams, "--no-password", "--exit-on-error", "/dev/null") //nolint:gosec // fixed system binary, args from the test
	oldOut, err := oldCmd.CombinedOutput()
	require.Error(t, err, "the old argv must fail")
	assert.Contains(t, string(oldOut), "unrecognized option",
		"the pre-fix single-element argv is rejected by the real pg_restore")

	// The argv strolt builds now: params split into elements.
	newCmd := exec.Command(bin, append(d.getRestoreArgs(), "/dev/null")...) //nolint:gosec // fixed system binary, args from the test
	newOut, err := newCmd.CombinedOutput()
	require.Error(t, err, "reading /dev/null as an archive must fail")
	assert.NotContains(t, string(newOut), "unrecognized option",
		"the split params must be accepted by the real pg_restore: %s", newOut)
}
