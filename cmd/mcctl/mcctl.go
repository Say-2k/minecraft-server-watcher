//go:build linux

package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"
	"minecraft-server-watcher/v2/internal/process"

	"google.golang.org/grpc"
)

func main() {
	cfg := config.LoadCtlConfigFromArgs(os.Args)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	botHost, exists := os.LookupEnv("BOT_HOST")
	if !exists {
		botHost = "localhost"
	}
	botPort, exists := os.LookupEnv("BOT_PORT")
	if !exists {
		botPort = "50051"
	}

	connect, err := grpc.NewClient(botHost + ":" + botPort)
	if err != nil {
		log.Printf("Ошибка создания gRPC клиента: %v", err)
	}

	defer connect.Close()

	client := agentpb.NewAgentServiceClient(connect)
	stream, err := client.Connect(ctx)
	if err != nil {
		log.Printf("Ошибка подключения к gRPC серверу: %v", err)
	}

	err = stream.Send(&agentpb.CtlMessage{Status: agentpb.ServerStatus_START})
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}

	mgr := process.NewManager()
	mgr.Start(ctx, cfg.Command)

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGTERM, syscall.SIGINT)
	<-sigc
	log.Println("Получен сигнал завершения")

	err = stream.Send(&agentpb.CtlMessage{Status: agentpb.ServerStatus_STOPPING})
	if err != nil {
		log.Printf("Ошибка отправки сообщения: %v", err)
	}
	if err := mgr.Stop(); err != nil {
		log.Printf("Ошибка при остановке процесса: %v", err)
	}

	log.Println("Сервер остановлен 🔴")
}
