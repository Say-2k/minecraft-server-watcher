package telegram

import (
	"context"
	"fmt"
	"log"
	"slices"
	"strconv"
	"time"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type TelegramWorker struct {
	config     *config.BotConfig
	notifier   ITelegramNotifier
	grpc       IBotServer
	cancelFunc context.CancelFunc
}

type ITelegramWorker interface {
	Listen(ctx context.Context)
	Stop()
}

func NewTelegramWorker(cfg *config.BotConfig, notifier ITelegramNotifier, grpc IBotServer) (ITelegramWorker, error) {
	if cfg == nil {
		return nil, fmt.Errorf("BotConfig is nil")
	}
	if notifier == nil {
		return nil, fmt.Errorf("TelegramNotifier is nil")
	}
	if grpc == nil {
		return nil, fmt.Errorf("gRPC server is nil")
	}
	return &TelegramWorker{
		config:   cfg,
		notifier: notifier,
		grpc:     grpc,
	}, nil
}

func (tw *TelegramWorker) Listen(ctx context.Context) {
	msgMapping := map[agentpb.ServerStatus]string{
		agentpb.ServerStatus_RUNNING:  "Сервер запущен 🟢",
		agentpb.ServerStatus_STOPPING: "Сервер остановлен 🔴",
		agentpb.ServerStatus_START:    "Сервер запускается 🟡",
	}

	ctx, cancel := context.WithCancel(ctx)
	tw.cancelFunc = cancel

	if tw.notifier != nil && tw.notifier.GetBot() != nil {
		_, _ = tw.notifier.GetBot().Request(tgbotapi.DeleteWebhookConfig{})
	}

	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := tw.notifier.GetBot().GetUpdatesChan(updateConfig)

	log.Println("Telegram listener started")
	for {
		select {
		case <-ctx.Done():
			log.Println("Telegram listener stopped")
			return
		case update := <-updates:
			go func(upd tgbotapi.Update) {
				if upd.Message != nil {
					switch upd.Message.Command() {
					case "start":
						msg := tgbotapi.NewMessage(upd.Message.Chat.ID, "Бот работает!")
						tw.notifier.GetBot().Send(msg)

					case "start_server":
						if !isHasAccess(tw, upd) {
							break
						}

						if tw.grpc.GetLastStatus() == agentpb.ServerStatus_RUNNING || tw.grpc.GetLastStatus() == agentpb.ServerStatus_START {
							sendReplyMessage(tw.notifier.GetBot(), upd.Message.Chat.ID, upd.Message.MessageID, "Сервер уже запущен.")
						} else if err := tw.grpc.SendMessage(agentpb.Command_START_SERVER); err != nil {
							sendReplyMessage(tw.notifier.GetBot(), upd.Message.Chat.ID, upd.Message.MessageID, "Не удалось запустить сервер.")
						} else {
							go tw.notifier.OnStart()
						}

						deleteMsgComand := tgbotapi.NewDeleteMessage(upd.Message.Chat.ID, upd.Message.MessageID)

						go func() {
							time.Sleep(5 * time.Second)
							tw.notifier.GetBot().Send(deleteMsgComand)
						}()

					case "stop_server":
						if !isHasAccess(tw, upd) {
							break
						}
						if tw.grpc.GetLastStatus() == agentpb.ServerStatus_STOPPING {
							sendReplyMessage(tw.notifier.GetBot(), upd.Message.Chat.ID, upd.Message.MessageID, "Сервер уже остановлен.")
						} else if err := tw.grpc.SendMessage(agentpb.Command_STOP_SERVER); err != nil {
							log.Printf("Ошибка при остановке процесса: %v", err)
						} else {
							go tw.notifier.OnStop()
						}

						deleteMsgComand := tgbotapi.NewDeleteMessage(upd.Message.Chat.ID, upd.Message.MessageID)

						go func() {
							time.Sleep(5 * time.Second)
							tw.notifier.GetBot().Send(deleteMsgComand)
						}()

					case "initial_status_message":
						if !isHasAccess(tw, upd) {
							break
						}
						statusMsg := msgMapping[tw.grpc.GetLastStatus()]
						msg := tgbotapi.NewMessage(upd.Message.Chat.ID, statusMsg)
						newMsg, err := tw.notifier.GetBot().Send(msg)
						if err == nil {
							tw.config.MessageID = newMsg.MessageID
						}
					}
				}
			}(update)
		}
	}
}

func isHasAccess(tw *TelegramWorker, update tgbotapi.Update) bool {
	if tw.config.AdminIDs == nil || !slices.Contains(tw.config.AdminIDs, strconv.Itoa(int(update.Message.From.ID))) {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "У вас нет прав для выполнения этой команды.")
		msg.ReplyToMessageID = update.Message.MessageID
		tw.notifier.GetBot().Send(msg)
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
