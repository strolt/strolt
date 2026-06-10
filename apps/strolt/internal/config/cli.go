package config

// CliConfig holds configuration overrides provided via CLI flags.
type CliConfig struct {
	Tags []string
}

var cliConfig = CliConfig{}

// SetCliConfig stores CLI-provided configuration overrides.
func SetCliConfig(config *CliConfig) {
	cliConfig = *config
}

func getCliConfig() Config {
	return Config{
		Tags: cliConfig.Tags,
	}
}
