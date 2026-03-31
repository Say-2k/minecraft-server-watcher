//go:build linux

package process

import (
	"context"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

type Manager struct {
	cmd       *exec.Cmd
	isRunning bool
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Start(ctx context.Context, command string) error {
	if command == "" {
		return nil
	}
	log.Printf("Запуск команды: %s", command)
	m.cmd = exec.CommandContext(ctx, "sh", "-c", command)
	m.cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}
	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	err := m.cmd.Start()
	m.isRunning = true
	return err
}

func (m *Manager) Stop() error {
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
	}
	return nil
}
