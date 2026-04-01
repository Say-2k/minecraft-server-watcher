//go:build windows

package process

import (
	"context"
	agentpb "minecraft-server-watcher/v2/api/agent/v1"
	"minecraft-server-watcher/v2/internal/config"
	"os/exec"

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
	return nil
}

func (m *Manager) SetStream(stream grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]) {
}

func (m *Manager) GetCfg() *config.CtlConfig {
	return nil
}

func (m *Manager) SetCfg(cfg *config.CtlConfig) {
}

func (m *Manager) Start(ctx context.Context) error {
	return nil
}

func (m *Manager) Stop() error {
	return nil
}
