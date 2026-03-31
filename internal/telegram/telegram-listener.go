//go:build linux

package telegram

import (
	"context"
	"log"
	"slices"
	"time"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramWorker struct {
	config     *config.BotConfig
	notifier   *TelegramNotifier
	grpc       *BotServer
	cancelFunc context.CancelFunc
}

func NewTelegramWorker(cfg *config.BotConfig, notifier *TelegramNotifier, grpc *BotServer) *TelegramWorker {
	return &TelegramWorker{
		config:   cfg,
		notifier: notifier,
		grpc:     grpc,
	}
}

func (tw *TelegramWorker) Listen(ctx context.Context) {
	msgMapping := map[agentpb.ServerStatus]string{
		agentpb.ServerStatus_RUNNING:  "Сервер запущен 🟢",
		agentpb.ServerStatus_STOPPING: "Сервер остановлен 🔴",
		agentpb.ServerStatus_START:    "Сервер запускается 🟡",
	}

	ctx, cancel := context.WithCancel(ctx)
	tw.cancelFunc = cancel

	if tw.notifier != nil && tw.notifier.bot != nil {
		_, _ = tw.notifier.bot.Request(tgbotapi.DeleteWebhookConfig{})
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := tw.notifier.bot.GetUpdatesChan(updateConfig)

	var serverContext context.Context
	var serverCtxCancel context.CancelFunc

	log.Println("Telegram listener started")
	for update := range updates {
		go func() {
			if update.Message != nil {
				switch update.Message.Command() {
				case "start":
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Бот работает!")
					tw.notifier.bot.Send(msg)

				case "start_server":
					if !isHasAccess(tw, update) {
						break
					}
					if tw.grpc.lastStatus == agentpb.ServerStatus_RUNNING {
						sendReplyMessage(tw.notifier.bot, update.Message.Chat.ID, update.Message.MessageID, "Сервер уже запущен.")
					} else {
						if err := tw.grpc.sendMessage(agentpb.Command_START_SERVER); err != nil {
							sendReplyMessage(tw.notifier.bot, update.Message.Chat.ID, update.Message.MessageID, "Не удалось запустить сервер.")
						} else {
							serverContext, serverCtxCancel = context.WithCancel(ctx)
							go tw.notifier.OnStart(serverContext)
						}
					}
					deleteMsgComand := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
					go func() {
						time.Sleep(5 * time.Second)
						tw.notifier.bot.Send(deleteMsgComand)
					}()
				case "stop_server":
					if !isHasAccess(tw, update) {
						break
					}
					if tw.grpc.lastStatus == agentpb.ServerStatus_STOPPING {
						sendReplyMessage(tw.notifier.bot, update.Message.Chat.ID, update.Message.MessageID, "Сервер уже остановлен.")
					}
					if err := tw.grpc.sendMessage(agentpb.Command_STOP_SERVER); err != nil {
						log.Printf("Ошибка при остановке процесса: %v", err)
					}
					if serverCtxCancel != nil {
						serverCtxCancel()
					}
					tw.notifier.OnStop()
					deleteMsgComand := tgbotapi.NewDeleteMessage(update.Message.Chat.ID, update.Message.MessageID)
					go func() {
						time.Sleep(5 * time.Second)
						tw.notifier.bot.Send(deleteMsgComand)
					}()

				case "initial_status_message":
					if !isHasAccess(tw, update) {
						break
					}
					statusMsg := msgMapping[tw.grpc.lastStatus]
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, statusMsg)
					newMsg, err := tw.notifier.bot.Send(msg)
					if err == nil {
						tw.config.MessageID = newMsg.MessageID
					}
				}
			}
		}()
	}
}

func isHasAccess(tw *TelegramWorker, update tgbotapi.Update) bool {
	if tw.config.AdminIDs == nil || !slices.Contains(tw.config.AdminIDs, update.Message.From.UserName) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для выполнения этой команды.")
		msg.ReplyToMessageID = update.Message.MessageID
		tw.notifier.bot.Send(msg)
		return false
	}
	return true
}

func sendReplyMessage(bot *tgbotapi.BotAPI, chatID int64, replyToID int, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyToMessageID = replyToID
	newMsg, err := bot.Send(msg)
	if err == nil {
		deleteMsgComand := tgbotapi.NewDeleteMessage(chatID, newMsg.MessageID)
		go func() {
			time.Sleep(5 * time.Second)
			bot.Send(deleteMsgComand)
		}()
	}
}

func (tw *TelegramWorker) Stop() {
	tw.cancelFunc()
}
