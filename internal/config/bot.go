package config

import (
	"log"
	"os"
	"strconv"
	"strings"
)

type BotConfig struct {
	BotToken         string
	ChatID           int64
	MessageID        int
	AdminIDs         []string
	SendFirstMessage bool
	BotPort          string
}

// LoadBotConfigFromEnv загружает конфигурацию бота из переменных окружения и возвращает указатель на структуру BotConfig.
func LoadBotConfigFromEnv() *BotConfig {
	var cfg BotConfig

	botToken, ok := os.LookupEnv("BOT_TOKEN")
	if !ok {
		log.Printf("Ошибка загрузки BOT_TOKEN из переменных окружения")
	}
	cfg.BotToken = botToken

	if chatId, ok := os.LookupEnv("CHAT_ID"); ok {
		if v, err := strconv.ParseInt("-100"+chatId, 10, 64); err == nil {
			cfg.ChatID = v
		} else {
			log.Printf("Ошибка загрузки CHAT_ID из переменных окружения")
		}
	}

	if messageId, ok := os.LookupEnv("MESSAGE_ID"); ok {
		if v, err := strconv.Atoi(messageId); err == nil {
			cfg.MessageID = v
		} else {
			log.Printf("Ошибка загрузки MESSAGE_ID из переменных окружения")
		}
	}

	if adminIds, ok := os.LookupEnv("ADMIN_IDS"); ok {
		massAdminIDs := strings.Split(adminIds, ",")
		cfg.AdminIDs = make([]string, len(massAdminIDs))
		for _, idStr := range massAdminIDs {
			cfg.AdminIDs = append(cfg.AdminIDs, idStr)
		}
	}

	if os.Getenv("SEND_FIRST_MESSAGE") == "true" {
		cfg.SendFirstMessage = true
	}

	if botPort, ok := os.LookupEnv("BOT_PORT"); ok {
		cfg.BotPort = botPort
	} else {
		log.Printf("Ошибка загрузки BOT_PORT из переменных окружения, используется значение по умолчанию: 50051")
		cfg.BotPort = "50051"
	}

	return &cfg
}
