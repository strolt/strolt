package mongodb

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetArgs_paramsAreSplitIntoArgvElements pins the regression where
// backup_params/restore_params/common_params were passed to exec as ONE argv
// element: a config like "--quiet --verbose" then made the mongo tool abort
// with an invalid-option error before touching any data.
func TestGetArgs_paramsAreSplitIntoArgvElements(t *testing.T) {
	m := New()
	m.config.CommonParams = "--common-a --common-b"
	m.config.BackupParams = "--backup-a --backup-b"
	m.config.RestoreParams = "--restore-a --restore-b"

	backup := m.getBackupArgs()
	assert.Equal(t, "--backup-a", backup[0], "backup params must become separate argv elements")
	assert.Equal(t, "--backup-b", backup[1], "backup params must become separate argv elements")

	restore := m.getRestoreArgs()
	assert.Equal(t, "--restore-a", restore[0], "restore params must become separate argv elements")
	assert.Equal(t, "--restore-b", restore[1], "restore params must become separate argv elements")
}

// TestSetConfig_rejectsBrokenParamsQuoting pins the fail-fast contract: a
// params string that cannot be tokenized must be rejected when the config is
// loaded, not discovered as an invalid-option crash at restore time.
func TestSetConfig_rejectsBrokenParamsQuoting(t *testing.T) {
	m := New()

	err := m.SetConfig(map[string]any{
		"database":       "db",
		"restore_params": `--comment='unterminated`,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "restore_params")
	assert.Contains(t, err.Error(), "unterminated")
}

func TestSetConfig_acceptsQuotedValues(t *testing.T) {
	m := New()

	require.NoError(t, m.SetConfig(map[string]any{
		"database":       "db",
		"restore_params": `--quiet --nsInclude='my db'`,
	}))

	args := m.getRestoreArgs()
	assert.Equal(t, "--quiet", args[0])
	assert.Equal(t, "--nsInclude=my db", args[1], "a quoted value with spaces must stay one argv element")
}
