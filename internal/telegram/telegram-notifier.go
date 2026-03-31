package telegram

import (
	"errors"
	"log"
	"minecraft-server-watcher/v2/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	SERV_RUNNING = "Сервер запущен 🟢"
	SERV_STOP    = "Сервер остановлен 🔴"
	SERV_START   = "Сервер запускается 🟡"
)

type TelegramNotifier struct {
	bot    *tgbotapi.BotAPI
	config *config.BotConfig
}

func NewTelegramNotifier(config *config.BotConfig) (*TelegramNotifier, error) {
	if config.BotToken == "" {
		return nil, errors.New("Telegram bot token is empty")
	}
	bot, err := tgbotapi.NewBotAPI(config.BotToken)
	if err != nil {
		return nil, err
	}
	return &TelegramNotifier{bot: bot, config: config}, nil
}

func (t *TelegramNotifier) OnStart() {
	if t == nil || t.bot == nil {
		return
	}
	log.Println("Сообщение через 5 минут будет обновлено на", SERV_RUNNING)
	msg := tgbotapi.NewEditMessageText(t.config.ChatID, t.config.MessageID, SERV_START)
	t.sendEdit(msg)
}

func (t *TelegramNotifier) OnRunning() {
	if t == nil || t.bot == nil {
		return
	}
	msg := tgbotapi.NewEditMessageText(t.config.ChatID, t.config.MessageID, SERV_RUNNING)
	t.sendEdit(msg)
}

func (t *TelegramNotifier) OnStop() {
	if t == nil || t.bot == nil {
		return
	}
	msg := tgbotapi.NewEditMessageText(t.config.ChatID, t.config.MessageID, SERV_STOP)
	t.sendEdit(msg)
}

func (t *TelegramNotifier) sendEdit(msg tgbotapi.EditMessageTextConfig) {
	_, err := t.bot.Send(msg)
	if err != nil {
		log.Printf("Сообщение не обновлено: %v", err)
		return
	}
	log.Println("Сообщение обновлено на", msg.Text)
}
