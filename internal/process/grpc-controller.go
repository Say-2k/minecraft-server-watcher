package process

import (
	"context"
	"errors"
	"io"
	"log"
	agentpb "minecraft-server-watcher/v2/api/agent/v1"
)

func Listen(m *Manager) {
	for {
		msgStream, err := m.Stream.Recv()

		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Println("Соединение с gRPC сервером закрыто")
				return
			}

			log.Printf("Ошибка при получении сообщения: %v", err)
			return
		}
		log.Printf("Получено сообщение: %s", msgStream.Command.String())

		switch msgStream.Command {
		case agentpb.Command_START_SERVER:
			if err := m.Start(context.Background()); err != nil {
				log.Printf("Ошибка при запуске процесса: %v", err)
			}
		case agentpb.Command_STOP_SERVER:
			if err := m.Stop(); err != nil {
				log.Printf("Ошибка при остановке процесса: %v", err)
			}
		}
	}
}
