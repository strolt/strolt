package shellargs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSplit(t *testing.T) {
	t.Run("splits space-separated options", func(t *testing.T) {
		args, err := Split("--no-owner --no-privileges")
		require.NoError(t, err)
		assert.Equal(t, []string{"--no-owner", "--no-privileges"}, args)
	})

	t.Run("keeps quoted value with spaces as one argument", func(t *testing.T) {
		args, err := Split(`--comment="two words" --verbose`)
		require.NoError(t, err)
		assert.Equal(t, []string{"--comment=two words", "--verbose"}, args)
	})

	t.Run("single quotes group a value and keep backslashes literal", func(t *testing.T) {
		args, err := Split(`--user='db admin' --table='a\'`)
		require.NoError(t, err)
		assert.Equal(t, []string{"--user=db admin", `--table=a\`}, args)
	})

	t.Run("surrounding quotes around a whole option", func(t *testing.T) {
		args, err := Split(`"--opt with space" plain`)
		require.NoError(t, err)
		assert.Equal(t, []string{"--opt with space", "plain"}, args)
	})

	t.Run("explicit empty quoted argument is preserved", func(t *testing.T) {
		args, err := Split(`--table '' post`)
		require.NoError(t, err)
		assert.Equal(t, []string{"--table", "", "post"}, args,
			"dropping an explicit empty argument would silently change the argv")
	})

	t.Run("a lone empty quoted argument yields one empty argument", func(t *testing.T) {
		args, err := Split(`""`)
		require.NoError(t, err)
		assert.Equal(t, []string{""}, args)
	})

	t.Run("backslash escapes a space outside quotes", func(t *testing.T) {
		args, err := Split(`--comment=a\ b`)
		require.NoError(t, err)
		assert.Equal(t, []string{"--comment=a b"}, args)
	})

	t.Run("backslash escapes a quote outside quotes instead of opening one", func(t *testing.T) {
		args, err := Split(`--opt=say \"hi\"`)
		require.NoError(t, err)
		// The space still separates arguments; the escaped quotes become
		// literal characters of the second argument, exactly like a shell.
		assert.Equal(t, []string{"--opt=say", `"hi"`}, args)
	})

	t.Run("escaped quote inside a single argument", func(t *testing.T) {
		args, err := Split(`--comment=a\"b`)
		require.NoError(t, err)
		assert.Equal(t, []string{`--comment=a"b`}, args)
	})

	t.Run("escaped double quote inside double quotes", func(t *testing.T) {
		args, err := Split(`--opt="say \"hi\""`)
		require.NoError(t, err)
		assert.Equal(t, []string{`--opt=say "hi"`}, args)
	})

	t.Run("escaped backslash collapses to one backslash", func(t *testing.T) {
		args, err := Split(`--path=C:\\tmp\\x`)
		require.NoError(t, err)
		assert.Equal(t, []string{`--path=C:\tmp\x`}, args)
	})

	t.Run("non-escaping backslash inside double quotes stays literal", func(t *testing.T) {
		args, err := Split(`--opt="a\b"`)
		require.NoError(t, err)
		assert.Equal(t, []string{`--opt=a\b`}, args)
	})

	t.Run("trailing backslash with nothing after it stays literal", func(t *testing.T) {
		args, err := Split(`abc\`)
		require.NoError(t, err)
		assert.Equal(t, []string{`abc\`}, args)
	})

	t.Run("error on unterminated double quote", func(t *testing.T) {
		_, err := Split(`--comment="two words`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unterminated double")
	})

	t.Run("error on unterminated single quote", func(t *testing.T) {
		_, err := Split(`--table='a`)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unterminated single")
	})

	t.Run("empty and whitespace-only strings produce no arguments", func(t *testing.T) {
		for _, s := range []string{"", "   ", "\t\n"} {
			args, err := Split(s)
			require.NoError(t, err)
			assert.Empty(t, args)
		}
	})
}

func TestMustSplit(t *testing.T) {
	assert.Equal(t, []string{"--a", "--b"}, MustSplit("--a --b"))

	assert.Panics(t, func() { MustSplit(`--x="unterminated`) },
		"invalid input must fail loudly, not produce a broken argv")
}
