package core

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"

	"gobot.io/x/gobot/v2/internal/interfaces"
)

// Robot represents a robot that manages devices and connections.
type Robot struct {
	name        string
	work        func()
	connections []interfaces.Connector
	devices     []interfaces.Device
	trap        func(chan os.Signal)
	autoRun     bool
	running     atomic.Bool
	done        chan bool
	mu          sync.RWMutex
	commander   interfaces.Commander
	eventer     interfaces.Eventer
}

// RobotOption represents a robot configuration option.
type RobotOption func(*Robot)

// WithName sets the robot name.
func WithName(name string) RobotOption {
	return func(r *Robot) {
		r.name = name
	}
}

// WithWork sets the robot work function.
func WithWork(work func()) RobotOption {
	return func(r *Robot) {
		r.work = work
	}
}

// WithConnections sets the robot connections.
func WithConnections(connections ...interfaces.Connector) RobotOption {
	return func(r *Robot) {
		r.connections = append(r.connections, connections...)
	}
}

// WithDevices sets the robot devices.
func WithDevices(devices ...interfaces.Device) RobotOption {
	return func(r *Robot) {
		r.devices = append(r.devices, devices...)
	}
}

// WithAutoRun sets the robot auto-run flag.
func WithAutoRun(autoRun bool) RobotOption {
	return func(r *Robot) {
		r.autoRun = autoRun
	}
}

// WithCommander sets the robot commander.
func WithCommander(commander interfaces.Commander) RobotOption {
	return func(r *Robot) {
		r.commander = commander
	}
}

// WithEventer sets the robot eventer.
func WithEventer(eventer interfaces.Eventer) RobotOption {
	return func(r *Robot) {
		r.eventer = eventer
	}
}

// NewRobot creates a new robot.
func NewRobot(opts ...RobotOption) *Robot {
	r := &Robot{
		name:        generateName(),
		connections: make([]interfaces.Connector, 0),
		devices:     make([]interfaces.Device, 0),
		done:        make(chan bool, 1),
		trap: func(c chan os.Signal) {
			signal.Notify(c, os.Interrupt)
		},
		autoRun: true,
	}

	for _, opt := range opts {
		opt(r)
	}

	r.running.Store(false)
	log.Printf("Robot %s initialized", r.name)

	return r
}

// Name returns the robot name.
func (r *Robot) Name() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.name
}

// Start starts the robot.
func (r *Robot) Start() error {
	if !r.running.CompareAndSwap(false, true) {
		return fmt.Errorf("robot %s is already running", r.name)
	}

	log.Printf("Starting robot %s...", r.name)

	// Start connections
	for _, conn := range r.connections {
		if err := conn.Connect(); err != nil {
			log.Printf("Failed to connect %s: %v", conn.Name(), err)
			r.Stop() // Clean up on error
			return err
		}
		log.Printf("Connected %s", conn.Name())
	}

	// Start devices
	for _, device := range r.devices {
		if err := device.Start(); err != nil {
			log.Printf("Failed to start device %s: %v", device.Name(), err)
			r.Stop() // Clean up on error
			return err
		}
		log.Printf("Started device %s", device.Name())
	}

	// Start work routine
	if r.work != nil {
		go func() {
			log.Println("Starting work routine...")
			r.work()
			<-r.done
		}()
	}

	// Handle auto-run
	if r.autoRun {
		c := make(chan os.Signal, 1)
		r.trap(c)
		<-c
		return r.Stop()
	}

	return nil
}

// Stop stops the robot.
func (r *Robot) Stop() error {
	if !r.running.CompareAndSwap(true, false) {
		return nil // Already stopped
	}

	log.Printf("Stopping robot %s...", r.name)

	var errs []error

	// Stop devices
	for _, device := range r.devices {
		if err := device.Halt(); err != nil {
			log.Printf("Failed to halt device %s: %v", device.Name(), err)
			errs = append(errs, err)
		}
	}

	// Stop connections
	for _, conn := range r.connections {
		if err := conn.Finalize(); err != nil {
			log.Printf("Failed to finalize connection %s: %v", conn.Name(), err)
			errs = append(errs, err)
		}
	}

	// Signal work routine to stop
	if r.work != nil {
		r.done <- true
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during stop: %v", errs)
	}

	log.Printf("Robot %s stopped", r.name)
	return nil
}

// AddDevice adds a device to the robot.
func (r *Robot) AddDevice(device interfaces.Device) interfaces.Device {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.devices = append(r.devices, device)
	log.Printf("Added device %s to robot %s", device.Name(), r.name)
	return device
}

// AddConnection adds a connection to the robot.
func (r *Robot) AddConnection(conn interfaces.Connector) interfaces.Connector {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.connections = append(r.connections, conn)
	log.Printf("Added connection %s to robot %s", conn.Name(), r.name)
	return conn
}

// Device returns a device by name.
func (r *Robot) Device(name string) interfaces.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, device := range r.devices {
		if device.Name() == name {
			return device
		}
	}
	return nil
}

// Connection returns a connection by name.
func (r *Robot) Connection(name string) interfaces.Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	for _, conn := range r.connections {
		if conn.Name() == name {
			return conn
		}
	}
	return nil
}

// Devices returns all devices.
func (r *Robot) Devices() []interfaces.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()
	devices := make([]interfaces.Device, len(r.devices))
	copy(devices, r.devices)
	return devices
}

// Connections returns all connections.
func (r *Robot) Connections() []interfaces.Connector {
	r.mu.RLock()
	defer r.mu.RUnlock()
	connections := make([]interfaces.Connector, len(r.connections))
	copy(connections, r.connections)
	return connections
}

// IsRunning returns true if the robot is running.
func (r *Robot) IsRunning() bool {
	return r.running.Load()
}

// generateName generates a random name for the robot.
func generateName() string {
	return fmt.Sprintf("Robot-%d", os.Getpid())
}