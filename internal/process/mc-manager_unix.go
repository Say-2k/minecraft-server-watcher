//go:build !windows

package process

import (
	"context"
	"log"
	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"
	"os"
	"os/exec"
	"syscall"
	"time"

	"google.golang.org/grpc"
)

type Manager struct {
	Cfg       *config.CtlConfig
	cmd       *exec.Cmd
	isRunning bool
	Stream    grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Start(ctx context.Context) error {
	if m.Cfg.Command == "" {
		return nil
	}
	log.Printf("Запуск команды: %s", m.Cfg.Command)
	m.cmd = exec.CommandContext(ctx, "sh", "-c", m.Cfg.Command)
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	err := m.cmd.Start()
	log.Printf("Сервер запущен: %s %d", m.Cfg.Command, m.cmd.Process.Pid)
	m.isRunning = true

	go func() {
		m.sendStatus(agentpb.ServerStatus_START)

		time.AfterFunc(5*time.Minute, func() {
			if m.isRunning {
				m.sendStatus(agentpb.ServerStatus_RUNNING)
			}
		})
	}()

	return err
}

func (m *Manager) Stop() error {
	log.Printf("Остановка команды: %s %d", m.Cfg.Command, m.cmd.Process.Pid)

	if m.cmd != nil && m.cmd.Process != nil {
		pgid, err := syscall.Getpgid(m.cmd.Process.Pid)
		if err != nil {
			return err
		}

		syscall.Kill(-pgid, syscall.SIGTERM)

		done := make(chan struct{})
		go func() {
			_ = m.cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
			m.isRunning = false
			return nil
		case <-time.After(8 * time.Second):
			syscall.Kill(-pgid, syscall.SIGKILL)
			m.isRunning = false
		}

		go m.sendStatus(agentpb.ServerStatus_STOPPING)
	}
	return nil
}

func (m *Manager) sendStatus(status agentpb.ServerStatus) {
	if m.Stream != nil {
		err := m.Stream.Send(&agentpb.CtlMessage{Status: status})
		if err != nil {
			log.Printf("Ошибка при отправке статуса: %v", err)
		}
	}
}
