package config

import (
	"strconv"
	"strings"
)

type Config struct {
	BotToken  string
	ChatID    int64
	MessageID int
	Command   string
	AdminIDs  []int64
}

// LoadFromArgs expects: <prog> BOT_TOKEN CHAT_ID MESSAGE_ID COMMAND
func LoadFromArgs(args []string) Config {
	var cfg Config
	if len(args) > 5 {
		cfg.BotToken = args[1]
		if v, err := strconv.ParseInt("-100"+args[2], 10, 64); err == nil {
			cfg.ChatID = v
		}
		if v, err := strconv.Atoi(args[3]); err == nil {
			cfg.MessageID = v
		}
		cfg.Command = args[4]
		strAdminIDs := strings.Split(args[5], ",")
		cfg.AdminIDs = make([]int64, len(strAdminIDs))
		for _, idStr := range strAdminIDs {
			if v, err := strconv.ParseInt(idStr, 10, 64); err == nil {
				cfg.AdminIDs = append(cfg.AdminIDs, v)
			}
		}
	}
	return cfg
}
