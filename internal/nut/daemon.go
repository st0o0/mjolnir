package nut

import (
	"log"
	"os"
	"os/exec"
)

type Manager struct {
	confPath string
	started  bool
	upsmon   *exec.Cmd
}

func NewManager(confPath string) *Manager {
	return &Manager{confPath: confPath}
}

func (m *Manager) Start() error {
	env := append(os.Environ(), "NUT_CONFPATH="+m.confPath)

	log.Printf("[mjolnir] starting drivers")
	if err := m.run(env, "upsdrvctl", "-u", "root", "start"); err != nil {
		return err
	}

	log.Printf("[mjolnir] starting upsd")
	if err := m.run(env, "upsd", "-u", "nut"); err != nil {
		return err
	}

	log.Printf("[mjolnir] starting upsmon")
	cmd := exec.Command("upsmon", "-F")
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}

	m.upsmon = cmd
	m.started = true
	return nil
}

func (m *Manager) Stop() {
	log.Printf("[mjolnir] shutting down...")

	if err := m.run(nil, "upsmon", "-c", "stop"); err != nil {
		log.Printf("[mjolnir] warning: upsmon stop: %v", err)
	}
	if err := m.run(nil, "upsd", "-c", "stop"); err != nil {
		log.Printf("[mjolnir] warning: upsd stop: %v", err)
	}
	if err := m.run(nil, "upsdrvctl", "stop"); err != nil {
		log.Printf("[mjolnir] warning: upsdrvctl stop: %v", err)
	}

	log.Printf("[mjolnir] shutdown complete")
}

func (m *Manager) Wait() error {
	if m.upsmon == nil {
		return nil
	}
	return m.upsmon.Wait()
}

func (m *Manager) Started() bool {
	return m.started
}

func (m *Manager) run(env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if env != nil {
		cmd.Env = env
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
