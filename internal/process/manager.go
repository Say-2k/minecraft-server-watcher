package process

import (
	"context"
	"log"
	"os"
	"os/exec"
	"runtime"
	"time"
)

type Manager struct {
	cmd    *exec.Cmd
	cancel context.CancelFunc
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) Start(parentCtx context.Context, command string) error {
	if command == "" {
		return nil
	}
	ctx, cancel := context.WithCancel(parentCtx)
	m.cancel = cancel

	if runtime.GOOS == "windows" {
		m.cmd = exec.CommandContext(ctx, "cmd", "/C", command)
	} else {
		m.cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	m.cmd.Stdout = os.Stdout
	m.cmd.Stderr = os.Stderr

	if err := m.cmd.Start(); err != nil {
		return err
	}

	go func() {
		if err := m.cmd.Wait(); err != nil {
			log.Printf("Процесс завершился с ошибкой: %v", err)
		} else {
			log.Println("Процесс завершён")
		}
		time.Sleep(100 * time.Millisecond)
	}()

	return nil
}

func (m *Manager) Stop() error {
	if m.cancel != nil {
		m.cancel()
	}
	if m.cmd != nil && m.cmd.Process != nil {
		_ = m.cmd.Process.Signal(os.Interrupt)

		done := make(chan struct{})
		go func() {
			_ = m.cmd.Wait()
			close(done)
		}()

		select {
		case <-done:
			return nil
		case <-time.After(10 * time.Second):
			_ = m.cmd.Process.Kill()
		}
	}
	return nil
}
