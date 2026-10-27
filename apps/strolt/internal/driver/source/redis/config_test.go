package redis

import (
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetConfig_requiresHostOrURI(t *testing.T) {
	r := New()

	err := r.SetConfig(map[string]any{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "host or uri is empty")
}

func TestSetConfig_acceptsHostOnly(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{"host": "redis"}))
}

func TestSetConfig_acceptsURIOnly(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{"uri": "redis://redis:6379"}))
}

// TestSetConfig_rejectsBrokenParamsQuoting pins the fail-fast contract: a
// params string that cannot be tokenized must be rejected when the config is
// loaded, not discovered as an invalid-option crash at backup time.
func TestSetConfig_rejectsBrokenParamsQuoting(t *testing.T) {
	for _, field := range []string{"backup_params", "common_params"} {
		t.Run(field, func(t *testing.T) {
			r := New()

			err := r.SetConfig(map[string]any{
				"host": "redis",
				field:  `--bigoptions='unterminated`,
			})
			require.Error(t, err)
			assert.Contains(t, err.Error(), field)
			assert.Contains(t, err.Error(), "unterminated")
		})
	}
}

func TestSetConfig_acceptsQuotedValues(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{
		"host":          "redis",
		"common_params": `--bigoptions='two words'`,
	}))

	args := r.getCommonArgs()
	assert.Equal(t, "--bigoptions=two words", args[len(args)-1], "a quoted value with spaces must stay one argv element")
}

// TestGetArgs_paramsAreSplitIntoArgvElements pins the regression where params
// were passed to exec as ONE argv element: a config like "--tls --verbose"
// then made redis-cli abort with an invalid-option error before touching any
// data.
func TestGetArgs_paramsAreSplitIntoArgvElements(t *testing.T) {
	r := New()
	r.config.CommonParams = "--common-a --common-b"
	r.config.BackupParams = "--backup-a --backup-b"
	r.config.Host = "redis"

	args := r.getBackupArgs("-")
	assert.Equal(t, "--backup-a", args[0], "backup params must become separate argv elements")
	assert.Equal(t, "--backup-b", args[1], "backup params must become separate argv elements")
	assert.Contains(t, args, "--common-a")
	assert.Contains(t, args, "--common-b")
}

func TestGetBackupArgs_uriMode(t *testing.T) {
	r := New()
	r.config.URI = "redis://redis:6379"

	args := r.getBackupArgs("-")
	assert.Equal(t, []string{"-u", "redis://redis:6379"}, args[:2])
	assert.NotContains(t, args, "-h", "uri mode must not pass a host")
}

func TestGetBackupArgs_hostMode(t *testing.T) {
	r := New()
	r.config.Host = "redis"
	r.config.Port = 6380

	args := r.getBackupArgs("-")
	assert.Equal(t, []string{"-h", "redis", "-p", "6380"}, args[:4])
}

func TestGetBackupArgs_noDefaultPortArg(t *testing.T) {
	r := New()
	r.config.Host = "redis"

	assert.NotContains(t, r.getBackupArgs("-"), "-p", "redis-cli defaults to 6379 itself")
}

// TestGetBackupArgs_pipeShape pins the two contracts the pipe task depends
// on: stdout streaming via "--rdb -" as the last option, and a filename with
// no ./ prefix (restic --stdin-filename must be a plain name).
func TestGetBackupArgs_pipeShape(t *testing.T) {
	r := New()
	r.config.Host = "redis"

	args := r.getBackupArgs("-")
	assert.Equal(t, []string{"--rdb", "-"}, args[len(args)-2:])

	assert.False(t, strings.HasPrefix(r.getFileName(), "./"))
	assert.Equal(t, "strolt_driver_redis.rdb", r.getFileName())
}

// TestGetBackupArgs_passwordNeverInArgv pins the secret-handling contract:
// the password must reach redis-cli via the REDISCLI_AUTH environment
// variable only, never as an argv element visible in `ps`.
func TestGetBackupArgs_passwordNeverInArgv(t *testing.T) {
	r := New()
	r.config.Host = "redis"
	r.config.Username = "default"
	r.config.Password = "s3cret"

	args := r.getBackupArgs("-")
	assert.Contains(t, args, "--user")
	assert.NotContains(t, args, "s3cret")
	assert.NotContains(t, args, "-a")

	assert.Contains(t, r.getEnv(), "REDISCLI_AUTH=s3cret")
}

// TestSetConfig_stripsURIUserinfo pins the secret-handling contract for
// connection strings: URI userinfo must never reach argv (it is visible in
// `ps`, and redis-cli treats any userinfo — even colon-less — as AUTH,
// which overrides REDISCLI_AUTH). Password goes to env, username to --user.
func TestSetConfig_stripsURIUserinfo(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{
		"uri": "redis://default:s3cret-p%40ss@redis:6379",
	}))

	args := r.getBackupArgs("-")
	assert.Contains(t, args, "redis://redis:6379", "argv must carry the credential-free URI")
	assert.NotContains(t, args, "s3cret-p@ss")
	assert.Contains(t, args, "--user", "the URI username must move to the --user flag")

	env := r.getEnv()
	assert.Contains(t, env, "REDISCLI_AUTH=s3cret-p@ss", "the URI password must reach redis-cli via env, decoded")
}

// TestSetConfig_colonlessUserinfoIsPassword pins redis-cli URI semantics:
// userinfo without a colon is the AUTH password, not a username. Moving it
// to REDISCLI_AUTH must preserve that meaning.
func TestSetConfig_colonlessUserinfoIsPassword(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{
		"uri": "redis://sekret@redis:6379",
	}))

	args := r.getBackupArgs("-")
	assert.Contains(t, args, "redis://redis:6379")
	assert.NotContains(t, args, "sekret")
	assert.NotContains(t, args, "--user", "no colon means no username in redis-cli semantics")

	assert.Contains(t, r.getEnv(), "REDISCLI_AUTH=sekret")
}

func TestSetConfig_passwordFieldWinsOverURI(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{
		"uri":      "redis://default:from-uri@redis:6379",
		"password": "explicit",
	}))

	assert.Contains(t, r.getEnv(), "REDISCLI_AUTH=explicit")
	assert.NotContains(t, r.getEnv(), "from-uri")
	assert.Contains(t, r.getBackupArgs("-"), "redis://redis:6379")
}

func TestSetConfig_uriWithoutUserinfoUntouched(t *testing.T) {
	r := New()

	require.NoError(t, r.SetConfig(map[string]any{
		"uri": "redis://redis:6379/0",
	}))

	assert.Contains(t, r.getBackupArgs("-"), "redis://redis:6379/0")
	assert.NotContains(t, r.getEnv(), "REDISCLI_AUTH=")
}

func TestGetEnv_plainPairs(t *testing.T) {
	r := New()
	r.env = map[string]string{"REDIS_URL": "redis://redis"}
	r.config.Password = "s3cret"

	env := r.getEnv()
	assert.Contains(t, env, "REDIS_URL=redis://redis", "env pairs must be plain KEY=VALUE, not shell-quoted")
	assert.Equal(t, "REDISCLI_AUTH=s3cret", env[len(env)-1], "REDISCLI_AUTH must be appended last so it wins collisions")
}

func TestGetEnv_noPasswordNoAuthVar(t *testing.T) {
	r := New()

	assert.NotContains(t, r.getEnv(), "REDISCLI_AUTH=")
}

func TestValidateRDBArtifact(t *testing.T) {
	t.Run("valid magic", func(t *testing.T) {
		path := t.TempDir() + "/dump.rdb"
		require.NoError(t, os.WriteFile(path, []byte("REDIS0011payload"), 0o644))

		assert.NoError(t, New().validateRDBArtifact(path))
		assert.FileExists(t, path, "a valid artifact must be kept")
	})

	t.Run("bad magic is removed", func(t *testing.T) {
		path := t.TempDir() + "/dump.rdb"
		require.NoError(t, os.WriteFile(path, []byte("WRONGpayload"), 0o644))

		err := New().validateRDBArtifact(path)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "bad magic")
		assert.NoFileExists(t, path, "an invalid artifact must not reach a destination")
	})

	t.Run("short file is removed", func(t *testing.T) {
		path := t.TempDir() + "/dump.rdb"
		require.NoError(t, os.WriteFile(path, []byte("RE"), 0o644))

		err := New().validateRDBArtifact(path)
		require.Error(t, err)
		assert.NoFileExists(t, path)
	})
}
