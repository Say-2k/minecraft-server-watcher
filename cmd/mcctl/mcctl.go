//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"minecraft-server-watcher/v2/internal/config"
	"minecraft-server-watcher/v2/internal/process"
	"minecraft-server-watcher/v2/internal/telegram"
)

func main() {
	cfg := config.LoadFromArgs(os.Args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier, err := telegram.NewTelegramNotifier(cfg)
	if err != nil {
		log.Printf("Ошибка создания Telegram notifier: %v", err)
	} else {
		log.Println("Бот запущен")
	}

	mgr := process.NewManager()
	mgr.Start(ctx, cfg.Command)
	go notifier.OnStart(ctx)

	var worker *telegram.TelegramWorker

	if notifier != nil {
		worker = telegram.NewTelegramWorker(cfg, notifier, mgr)
		go worker.Listen(ctx)
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
	log.Println("Получен сигнал завершения")

	if worker != nil {
		worker.Stop()
	}
	if notifier != nil {
		notifier.OnStop()
	}

	if err := mgr.Stop(); err != nil {
		log.Printf("Ошибка при остановке процесса: %v", err)
	}

	log.Println("Сервер остановлен 🔴")
}
