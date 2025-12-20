package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"minecraft-server-watcher/internal/config"
	"minecraft-server-watcher/internal/notify"
	"minecraft-server-watcher/internal/process"
)

func main() {
	cfg := config.LoadFromArgs(os.Args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier := notify.NewTelegram(cfg.BotToken, cfg.ChatID, cfg.MessageID)

	mgr := process.NewManager()
	if err := mgr.Start(ctx, cfg.Command); err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
	}

	log.Println("Сервер запущен 🟢")

	if notifier != nil {
		go notifier.OnStart(ctx)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
	log.Println("Получен сигнал завершения")

	if notifier != nil {
		notifier.OnStop()
	}

	if err := mgr.Stop(); err != nil {
		log.Printf("Ошибка при остановке процесса: %v", err)
	}

	log.Println("Сервер остановлен 🔴")
}
