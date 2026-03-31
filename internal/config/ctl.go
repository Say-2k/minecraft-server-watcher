package config

type CtlConfig struct {
	Command string
}

// LoadCtlConfigFromArgs expects: <prog> COMMAND
func LoadCtlConfigFromArgs(args []string) *CtlConfig {
	var cfg CtlConfig
	if len(args) > 1 {
		cfg.Command = args[1]
	}
	return &cfg
}
