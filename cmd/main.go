package main

import (
	"context"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

const (
	SERV_START    = "Сервер запущен 🟢"
	SERV_STOP     = "Сервер остановлен 🔴"
	SERV_STARTING = "Сервер запускается 🟡"
)

func main() {
	var err error
	var botId string
	var chatId int64
	var messageId int
	var command string
	if len(os.Args) > 4 {
		botId = os.Args[1]
		if chatId, err = strconv.ParseInt(os.Args[2], 10, 64); err != nil {
		}
		if messageId, err = strconv.Atoi(os.Args[3]); err != nil {
		}
		command = os.Args[4]
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	bot, err := tgbotapi.NewBotAPI(botId)
	if err != nil {
		log.Printf("Ошибка создания бота: %v", err)
	}

	// === СИГНАЛЫ ===
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)

	// === ЗАПУСК СЕРВЕРА ===
	cmd := exec.CommandContext(ctx, command)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		log.Fatal("Minecraft failed to start:", err)
	}

	log.Println(SERV_START)

	if bot != nil {
		go onStart(ctx, bot, chatId, messageId)
	}

	// === ЖДЁМ SIGTERM ОТ DOCKER ===
	<-sigc
	log.Println("Получен SIGTERM")

	if bot != nil {
		onStop(bot, chatId, messageId)
	}

	// корректно останавливаем дочерний процесс
	cancel()
	_ = cmd.Process.Signal(os.Interrupt)
	if err = cmd.Wait(); err != nil {
		log.Printf("Ошибка при завершении майнкрафт: %v", err)
	}

	log.Println(SERV_STOP)
}

func onStart(ctx context.Context, bot *tgbotapi.BotAPI, chatId int64, messageId int) {
	timer := time.NewTimer(5 * time.Minute)
	log.Println("Сообщение через 5 минут будет обновлено на", SERV_START)
	msg := tgbotapi.NewEditMessageText(chatId, messageId, SERV_STARTING)
	sendMessage(bot, msg)

	select {
	case <-ctx.Done():
		return
	case <-timer.C:
		msg := tgbotapi.NewEditMessageText(chatId, messageId, SERV_START)
		sendMessage(bot, msg)
	}
}

func onStop(bot *tgbotapi.BotAPI, chatId int64, messageId int) {
	msg := tgbotapi.NewEditMessageText(chatId, messageId, SERV_STOP)
	sendMessage(bot, msg)
}

func sendMessage(bot *tgbotapi.BotAPI, msg tgbotapi.EditMessageTextConfig) {
	_, err := bot.Send(msg)
	if err != nil {
		log.Printf("Сообщение не обновлено. %v", err)
		return
	}

	log.Println("Сообщение обновлено на", msg.Text)
}
