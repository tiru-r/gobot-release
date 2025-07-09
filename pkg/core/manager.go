package core

import (
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"slices"
	"sync"
	"sync/atomic"

	"gobot.io/x/gobot/v2/internal/interfaces"
)

// Manager manages a collection of robots.
type Manager struct {
	robots    []interfaces.Robot
	trap      func(chan os.Signal)
	autoRun   bool
	running   atomic.Bool
	mu        sync.RWMutex
	commander interfaces.Commander
	eventer   interfaces.Eventer
}

// ManagerOption represents a manager configuration option.
type ManagerOption func(*Manager)

// WithAutoRun sets the manager auto-run flag.
func WithManagerAutoRun(autoRun bool) ManagerOption {
	return func(m *Manager) {
		m.autoRun = autoRun
	}
}

// WithManagerCommander sets the manager commander.
func WithManagerCommander(commander interfaces.Commander) ManagerOption {
	return func(m *Manager) {
		m.commander = commander
	}
}

// WithManagerEventer sets the manager eventer.
func WithManagerEventer(eventer interfaces.Eventer) ManagerOption {
	return func(m *Manager) {
		m.eventer = eventer
	}
}

// NewManager creates a new manager.
func NewManager(opts ...ManagerOption) *Manager {
	m := &Manager{
		robots: make([]interfaces.Robot, 0),
		trap: func(c chan os.Signal) {
			signal.Notify(c, os.Interrupt)
		},
		autoRun: true,
	}

	for _, opt := range opts {
		opt(m)
	}

	m.running.Store(false)
	log.Println("Manager initialized")

	return m
}

// Start starts all robots.
func (m *Manager) Start() error {
	if !m.running.CompareAndSwap(false, true) {
		return fmt.Errorf("manager is already running")
	}

	log.Println("Starting manager...")

	// Start all robots
	for _, robot := range m.robots {
		if err := robot.Start(); err != nil {
			log.Printf("Failed to start robot %s: %v", robot.Name(), err)
			m.Stop() // Clean up on error
			return err
		}
		log.Printf("Started robot %s", robot.Name())
	}

	// Handle auto-run
	if m.autoRun {
		c := make(chan os.Signal, 1)
		m.trap(c)
		<-c
		return m.Stop()
	}

	return nil
}

// Stop stops all robots.
func (m *Manager) Stop() error {
	if !m.running.CompareAndSwap(true, false) {
		return nil // Already stopped
	}

	log.Println("Stopping manager...")

	var errs []error

	// Stop all robots
	for _, robot := range m.robots {
		if err := robot.Stop(); err != nil {
			log.Printf("Failed to stop robot %s: %v", robot.Name(), err)
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	log.Println("Manager stopped")
	return nil
}

// AddRobot adds a robot to the manager.
func (m *Manager) AddRobot(robot interfaces.Robot) interfaces.Robot {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.robots = append(m.robots, robot)
	log.Printf("Added robot %s to manager", robot.Name())
	return robot
}

// Robot returns a robot by name.
func (m *Manager) Robot(name string) interfaces.Robot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for _, robot := range m.robots {
		if robot.Name() == name {
			return robot
		}
	}
	return nil
}

// Robots returns all robots.
func (m *Manager) Robots() []interfaces.Robot {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return slices.Clone(m.robots)
}

// IsRunning returns true if the manager is running.
func (m *Manager) IsRunning() bool {
	return m.running.Load()
}