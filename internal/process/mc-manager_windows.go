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
	Cfg       *config.CtlConfig
	cmd       *exec.Cmd
	isRunning bool
	Stream    grpc.BidiStreamingClient[agentpb.CtlMessage, agentpb.BotMessage]
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Start(ctx context.Context) error {
	return nil
}

func (m *Manager) Stop() error {
	return nil
}
