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
		return
	} else {
		log.Println("Бот запущен")
	}

	grpcServer, err := telegram.NewBotServer(cfg, notifier)
	if err != nil {
		log.Printf("Ошибка создания gRPC сервера: %v", err)
		return
	}

	if err := grpcServer.Start(); err != nil {
		log.Printf("Ошибка запуска gRPC сервера: %v", err)
		return
	}

	var worker telegram.ITelegramWorker

	if notifier != nil {
		worker, err = telegram.NewTelegramWorker(cfg, notifier, grpcServer)
		if err != nil {
			log.Printf("Ошибка создания Telegram worker: %v", err)
		} else {
			go worker.Listen(ctx)
		}
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
	grpcServer.Disconnect()

	log.Println("Бот остановлен")
}
