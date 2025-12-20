//go:build linux

package telegram

import (
	"context"
	"log"
	"minecraft-server-watcher/internal/config"
	"minecraft-server-watcher/internal/process"
	"slices"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramWorker struct {
	config     *config.Config
	notifier   *TelegramNotifier
	manager    *process.Manager
	cancelFunc context.CancelFunc
}

func NewTelegramWorker(cfg *config.Config, notifier *TelegramNotifier, mgr *process.Manager) *TelegramWorker {
	return &TelegramWorker{
		config:   cfg,
		notifier: notifier,
		manager:  mgr,
	}
}

func (tw *TelegramWorker) Listen(ctx context.Context) {
	ctx, cancel := context.WithCancel(ctx)
	tw.cancelFunc = cancel

	if tw.notifier != nil && tw.notifier.bot != nil {
		_, _ = tw.notifier.bot.Request(tgbotapi.DeleteWebhookConfig{})
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := tw.notifier.bot.GetUpdatesChan(updateConfig)

	log.Println("Telegram listener started")
	for update := range updates {
		go func() {
			if update.Message != nil {
				switch update.Message.Text {
				case "/start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот работает!")
					tw.notifier.bot.Send(msg)

				case "/start_server":
					if tw.config.AdminIDs == nil || !slices.Contains(tw.config.AdminIDs, update.Message.From.ID) {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для выполнения этой команды.")
						msg.ReplyToMessageID = update.Message.MessageID
						tw.notifier.bot.Send(msg)
						break
					}
					if err := tw.manager.Start(ctx, tw.config.Command); err != nil {
						msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Не удалось запустить сервер.")
						msg.ReplyToMessageID = update.Message.MessageID
						tw.notifier.bot.Send(msg)
					} else {
						tw.notifier.OnStart(ctx)
					}

				case "/stop_server":
					if err := tw.manager.Stop(); err != nil {
						log.Printf("Ошибка при остановке процесса: %v", err)
					}
					tw.notifier.OnStop()
				}
			}
		}()
	}
}

func (tw *TelegramWorker) Stop() {
	tw.cancelFunc()
}
