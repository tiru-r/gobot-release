package driver

import (
	"sync"
	"time"

	"gobot.io/x/gobot/v2/internal/logging"
)

// BaseDriver provides common functionality for all drivers
type BaseDriver struct {
	name          string
	connection    interface{}
	interval      time.Duration
	halt          chan bool
	commander     *Commander
	eventer       *Eventer
	mutex         sync.RWMutex
	afterConnect  func() error
	beforeHalt    func() error
	logger        *logging.Logger
	isConnected   bool
	isRunning     bool
}

// DriverConfig holds configuration options for drivers
type DriverConfig struct {
	Name         string
	Connection   interface{}
	Interval     time.Duration
	AfterConnect func() error
	BeforeHalt   func() error
}

// NewBaseDriver creates a new base driver with the given configuration
func NewBaseDriver(config DriverConfig) *BaseDriver {
	if config.Name == "" {
		config.Name = "BaseDriver"
	}
	if config.Interval == 0 {
		config.Interval = 10 * time.Millisecond
	}

	return &BaseDriver{
		name:          config.Name,
		connection:    config.Connection,
		interval:      config.Interval,
		halt:          make(chan bool),
		commander:     NewCommander(),
		eventer:       NewEventer(),
		afterConnect:  config.AfterConnect,
		beforeHalt:    config.BeforeHalt,
		logger:        logging.GetLogger("driver:" + config.Name),
	}
}

// Name returns the driver name
func (d *BaseDriver) Name() string {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.name
}

// SetName sets the driver name
func (d *BaseDriver) SetName(name string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.name = name
	d.logger = logging.GetLogger("driver:" + name)
}

// Connection returns the driver's connection
func (d *BaseDriver) Connection() interface{} {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.connection
}

// Interval returns the driver's interval
func (d *BaseDriver) Interval() time.Duration {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.interval
}

// SetInterval sets the driver's interval
func (d *BaseDriver) SetInterval(interval time.Duration) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.interval = interval
}

// IsConnected returns true if the driver is connected
func (d *BaseDriver) IsConnected() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.isConnected
}

// IsRunning returns true if the driver is running
func (d *BaseDriver) IsRunning() bool {
	d.mutex.RLock()
	defer d.mutex.RUnlock()
	return d.isRunning
}

// Logger returns the driver's logger
func (d *BaseDriver) Logger() *logging.Logger {
	return d.logger
}

// Commander returns the driver's commander
func (d *BaseDriver) Commander() *Commander {
	return d.commander
}

// Eventer returns the driver's eventer
func (d *BaseDriver) Eventer() *Eventer {
	return d.eventer
}

// Start connects to the driver and starts any required processes
func (d *BaseDriver) Start() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if d.isConnected {
		return nil
	}

	d.logger.Infof("Starting driver '%s'", d.name)

	if d.afterConnect != nil {
		if err := d.afterConnect(); err != nil {
			d.logger.Errorf("Failed to start driver '%s': %v", d.name, err)
			return err
		}
	}

	d.isConnected = true
	d.isRunning = true
	d.logger.Infof("Driver '%s' started successfully", d.name)
	return nil
}

// Halt stops the driver and terminates any running processes
func (d *BaseDriver) Halt() error {
	d.mutex.Lock()
	defer d.mutex.Unlock()

	if !d.isConnected {
		return nil
	}

	d.logger.Infof("Halting driver '%s'", d.name)

	if d.beforeHalt != nil {
		if err := d.beforeHalt(); err != nil {
			d.logger.Errorf("Failed to halt driver '%s': %v", d.name, err)
			return err
		}
	}

	d.isRunning = false
	close(d.halt)
	d.halt = make(chan bool) // Reset for potential restart

	d.isConnected = false
	d.logger.Infof("Driver '%s' halted successfully", d.name)
	return nil
}

// StartWorker starts a worker goroutine that calls the provided function at regular intervals
func (d *BaseDriver) StartWorker(workFunc func() error) {
	if workFunc == nil {
		return
	}

	go func() {
		ticker := time.NewTicker(d.Interval())
		defer ticker.Stop()

		for {
			select {
			case <-d.halt:
				d.logger.Debugf("Worker for driver '%s' stopped", d.name)
				return
			case <-ticker.C:
				if err := workFunc(); err != nil {
					d.logger.Errorf("Worker error in driver '%s': %v", d.name, err)
				}
			}
		}
	}()
}

// SafeCall executes a function safely with logging and error handling
func (d *BaseDriver) SafeCall(operation string, fn func() error) error {
	d.logger.Debugf("Executing operation '%s' on driver '%s'", operation, d.name)
	
	start := time.Now()
	err := fn()
	duration := time.Since(start)

	if err != nil {
		d.logger.Errorf("Operation '%s' failed on driver '%s': %v (took %v)", operation, d.name, err, duration)
	} else {
		d.logger.Debugf("Operation '%s' completed on driver '%s' (took %v)", operation, d.name, duration)
	}

	return err
}

// Commander provides command functionality
type Commander struct {
	commands map[string]func(map[string]interface{}) interface{}
	mutex    sync.RWMutex
}

// NewCommander creates a new commander
func NewCommander() *Commander {
	return &Commander{
		commands: make(map[string]func(map[string]interface{}) interface{}),
	}
}

// AddCommand adds a command to the commander
func (c *Commander) AddCommand(name string, command func(map[string]interface{}) interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.commands[name] = command
}

// Command executes a command
func (c *Commander) Command(name string, args map[string]interface{}) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	if cmd, exists := c.commands[name]; exists {
		return cmd(args)
	}
	return nil
}

// Commands returns all available commands
func (c *Commander) Commands() map[string]func(map[string]interface{}) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	commands := make(map[string]func(map[string]interface{}) interface{})
	for name, cmd := range c.commands {
		commands[name] = cmd
	}
	return commands
}

// Eventer provides event functionality
type Eventer struct {
	events map[string][]func(interface{})
	mutex  sync.RWMutex
}

// NewEventer creates a new eventer
func NewEventer() *Eventer {
	return &Eventer{
		events: make(map[string][]func(interface{})),
	}
}

// On adds an event handler
func (e *Eventer) On(event string, handler func(interface{})) {
	e.mutex.Lock()
	defer e.mutex.Unlock()
	e.events[event] = append(e.events[event], handler)
}

// Emit emits an event
func (e *Eventer) Emit(event string, data interface{}) {
	e.mutex.RLock()
	handlers := e.events[event]
	e.mutex.RUnlock()

	for _, handler := range handlers {
		go handler(data)
	}
}

// Events returns all available events
func (e *Eventer) Events() []string {
	e.mutex.RLock()
	defer e.mutex.RUnlock()
	events := make([]string, 0, len(e.events))
	for name := range e.events {
		events = append(events, name)
	}
	return events
}