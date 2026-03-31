//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"minecraft-server-watcher/v2/internal/config"
	"minecraft-server-watcher/v2/internal/telegram"
)

func main() {
	cfg := config.LoadBotConfigFromEnv()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	notifier, err := telegram.NewTelegramNotifier(cfg)
	if err != nil {
		log.Printf("Ошибка создания Telegram notifier: %v", err)
	} else {
		log.Println("Бот запущен")
	}
	go notifier.OnStart(ctx)

	grpcServer := telegram.NewBotServer()
	go func() {
		if err := grpcServer.Start(); err != nil {
			log.Printf("Ошибка запуска gRPC сервера: %v", err)
		}
	}()

	var worker *telegram.TelegramWorker

	if notifier != nil {
		worker = telegram.NewTelegramWorker(cfg, notifier, grpcServer)
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

	log.Println("Бот остановлен")
}
