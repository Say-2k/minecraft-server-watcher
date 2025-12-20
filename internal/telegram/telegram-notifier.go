//go:build linux

package telegram

import (
	"context"
	"errors"
	"log"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	SERV_START    = "Сервер запущен 🟢"
	SERV_STOP     = "Сервер остановлен 🔴"
	SERV_STARTING = "Сервер запускается 🟡"
)

type TelegramNotifier struct {
	bot       *tgbotapi.BotAPI
	chatID    int64
	messageID int
}

func NewTelegramNotifier(token string, chatID int64, messageID int) (*TelegramNotifier, error) {
	if token == "" {
		return nil, errors.New("Telegram bot token is empty")
	}
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}
	return &TelegramNotifier{bot: bot, chatID: chatID, messageID: messageID}, nil
}

func (t *TelegramNotifier) OnStart(ctx context.Context) {
	if t == nil || t.bot == nil {
		return
	}
	log.Println("Сообщение через 5 минут будет обновлено на", SERV_START)
	msg := tgbotapi.NewEditMessageText(t.chatID, t.messageID, SERV_STARTING)
	t.sendEdit(msg)

	select {
	case <-ctx.Done():
		log.Println("Контекст отменен до обновления сообщения на", SERV_START)
		return
	case <-time.After(5 * time.Minute):
		msg := tgbotapi.NewEditMessageText(t.chatID, t.messageID, SERV_START)
		t.sendEdit(msg)
	}
}

func (t *TelegramNotifier) OnStop() {
	if t == nil || t.bot == nil {
		return
	}
	msg := tgbotapi.NewEditMessageText(t.chatID, t.messageID, SERV_STOP)
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
