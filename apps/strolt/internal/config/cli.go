package config

type CliConfig struct {
	Tags []string
}

var cliConfig = CliConfig{}

func SetCliConfig(config *CliConfig) {
	cliConfig = *config
}

func getCliConfig() Config {
	return Config{
		Tags: cliConfig.Tags,
	}
}
