package interfaces

// Core interfaces for Gobot - simplified and consolidated version

// Connector represents a connection to a platform or hardware
type Connector interface {
	// Name returns the name of the connection
	Name() string
	// Connect establishes the connection
	Connect() error
	// Finalize closes the connection
	Finalize() error
}

// Device represents a hardware device
type Device interface {
	// Name returns the device name
	Name() string
	// SetName sets the device name
	SetName(string)
	// Connection returns the connection for this device
	Connection() Connector
	// Start initializes the device
	Start() error
	// Halt shuts down the device
	Halt() error
}

// Driver represents a device driver with additional capabilities
type Driver interface {
	Device
	// Interval returns the polling interval
	Interval() interface{}
	// SetInterval sets the polling interval
	SetInterval(interface{})
}

// Commander provides command execution capabilities
type Commander interface {
	// Command executes a command with the given arguments
	Command(string, map[string]interface{}) interface{}
	// Commands returns available commands
	Commands() map[string]func(map[string]interface{}) interface{}
	// AddCommand adds a new command
	AddCommand(string, func(map[string]interface{}) interface{})
}

// Eventer provides event handling capabilities
type Eventer interface {
	// On registers an event handler
	On(string, func(interface{}))
	// Emit triggers an event
	Emit(string, interface{})
	// Events returns available events
	Events() []string
}

// Robot represents a complete robot with lifecycle management
type Robot interface {
	// Name returns the robot name
	Name() string
	// Start starts the robot and all its components
	Start() error
	// Stop stops the robot gracefully
	Stop() error
	// Running returns true if the robot is running
	Running() bool
	// AddDevice adds a device to the robot
	AddDevice(Device) error
	// AddConnection adds a connection to the robot
	AddConnection(Connector) error
}

// Adaptor represents a platform adaptor (combination of Connector and additional capabilities)
type Adaptor interface {
	Connector
	// Additional platform-specific methods can be added by embedding this interface
}

// DigitalPinner provides digital pin capabilities
type DigitalPinner interface {
	// DigitalRead reads from a digital pin
	DigitalRead(string) (int, error)
	// DigitalWrite writes to a digital pin
	DigitalWrite(string, byte) error
}

// AnalogReader provides analog reading capabilities
type AnalogReader interface {
	// AnalogRead reads from an analog pin
	AnalogRead(string) (int, error)
}

// PWMWriter provides PWM writing capabilities
type PWMWriter interface {
	// PwmWrite writes a PWM value to a pin
	PwmWrite(string, byte) error
}

// I2CConnector provides I2C communication capabilities
type I2CConnector interface {
	// I2CRead reads from an I2C device
	I2CRead(address int, size int) ([]byte, error)
	// I2CWrite writes to an I2C device
	I2CWrite(address int, data []byte) error
	// I2CStart starts I2C communication
	I2CStart(address int) error
}

// SPIConnector provides SPI communication capabilities
type SPIConnector interface {
	// SPIRead reads from an SPI device
	SPIRead(data []byte) error
	// SPIWrite writes to an SPI device
	SPIWrite(data []byte) error
}

// Configurable represents components that can be configured
type Configurable interface {
	// Configure applies configuration options
	Configure(options ...interface{}) error
}

// Validator represents components that can validate their configuration
type Validator interface {
	// Validate checks if the component is properly configured
	Validate() error
}

// Lifecycle represents the complete lifecycle of a component
type Lifecycle interface {
	// Start initializes and starts the component
	Start() error
	// Stop gracefully stops the component
	Stop() error
	// Restart stops and starts the component
	Restart() error
	// IsRunning returns true if the component is running
	IsRunning() bool
}

// Healthcheck represents components that can report their health status
type Healthcheck interface {
	// Health returns the health status of the component
	Health() (bool, error)
	// Ready returns true if the component is ready to serve requests
	Ready() bool
}