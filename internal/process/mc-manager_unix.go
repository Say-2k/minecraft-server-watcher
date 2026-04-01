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
	cfg       *config.CtlConfig
	cmd       *exec.Cmd
	isRunning bool
	stream    grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]
}

type IManager interface {
	Start(ctx context.Context) error
	Stop() error
	GetStream() grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]
	SetStream(stream grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage])
	GetCfg() *config.CtlConfig
	SetCfg(cfg *config.CtlConfig)
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) GetStream() grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage] {
	return m.stream
}

func (m *Manager) SetStream(stream grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]) {
	m.stream = stream
}

func (m *Manager) GetCfg() *config.CtlConfig {
	return m.cfg
}

func (m *Manager) SetCfg(cfg *config.CtlConfig) {
	m.cfg = cfg
}

func (m *Manager) Start(ctx context.Context) error {
	if m.cfg.Command == "" {
		return nil
	}
	log.Printf("Запуск команды: %s", m.cfg.Command)
	m.cmd = exec.CommandContext(ctx, "sh", "-c", m.cfg.Command)
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	err := m.cmd.Start()
	if err != nil {
		log.Printf("Ошибка при запуске команды: %v", err)
		return err
	}
	log.Printf("Сервер запущен: %s", m.cfg.Command)
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
	if m.cmd != nil && m.cmd.Process != nil {
		log.Printf("Остановка команды: %s %d", m.cfg.Command, m.cmd.Process.Pid)

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
	if m.stream != nil {
		err := m.stream.Send(&agentpb.CtlMessage{Status: status})
		if err != nil {
			log.Printf("Ошибка при отправке статуса: %v", err)
		}
	}
}
