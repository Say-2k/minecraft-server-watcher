package config

import (
	"strconv"
)

type Config struct {
	BotToken  string
	ChatID    int64
	MessageID int
	Command   string
}

// LoadFromArgs expects: <prog> BOT_TOKEN CHAT_ID MESSAGE_ID COMMAND
func LoadFromArgs(args []string) Config {
	var cfg Config
	if len(args) > 4 {
		cfg.BotToken = args[1]
		if v, err := strconv.ParseInt("-100"+args[2], 10, 64); err == nil {
			cfg.ChatID = v
		}
		if v, err := strconv.Atoi(args[3]); err == nil {
			cfg.MessageID = v
		}
		cfg.Command = args[4]
	}
	return cfg
}
