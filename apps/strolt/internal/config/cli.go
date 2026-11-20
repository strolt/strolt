package config

// CliConfig holds configuration overrides provided via CLI flags.
type CliConfig struct {
	Tags []string
}

var cliConfig = CliConfig{Tags: nil}

// SetCliConfig stores CLI-provided configuration overrides.
func SetCliConfig(config *CliConfig) {
	cliConfig = *config
}

func getCliConfig() Config {
	return Config{
		TimeZone:     "",
		timeLocation: nil,
		Services:     nil,
		Tags:         cliConfig.Tags,
		Secrets:      nil,
		Extends:      Extends{Secrets: nil, Configs: nil},
		Definitions:  Definitions{Destinations: nil, Notifications: nil},
		API:          API{Users: nil},
	}
}
