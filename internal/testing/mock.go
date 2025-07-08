package testing

import (
	"sync"

	"time"
)

// MockConnection is a mock connection for testing.
type MockConnection struct {
	name      string
	connected bool
	mu        sync.RWMutex
	ConnectFunc   func() error
	FinalizeFunc  func() error
}

// NewMockConnection creates a new mock connection.
func NewMockConnection(name string) *MockConnection {
	return &MockConnection{
		name:      name,
		connected: false,
	}
}

// Name returns the connection name.
func (m *MockConnection) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// Connect connects the mock connection.
func (m *MockConnection) Connect() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.ConnectFunc != nil {
		if err := m.ConnectFunc(); err != nil {
			return err
		}
	}
	
	m.connected = true
	return nil
}

// Finalize finalizes the mock connection.
func (m *MockConnection) Finalize() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.FinalizeFunc != nil {
		if err := m.FinalizeFunc(); err != nil {
			return err
		}
	}
	
	m.connected = false
	return nil
}

// Connected returns true if the connection is connected.
func (m *MockConnection) Connected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connected
}

// MockDevice is a mock device for testing.
type MockDevice struct {
	name       string
	connection interface{}
	started    bool
	mu         sync.RWMutex
	StartFunc  func() error
	HaltFunc   func() error
}

// NewMockDevice creates a new mock device.
func NewMockDevice(name string, conn interface{}) *MockDevice {
	return &MockDevice{
		name:       name,
		connection: conn,
		started:    false,
	}
}

// Name returns the device name.
func (m *MockDevice) Name() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.name
}

// SetName sets the device name.
func (m *MockDevice) SetName(name string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.name = name
}

// Connection returns the device connection.
func (m *MockDevice) Connection() interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.connection
}

// Start starts the mock device.
func (m *MockDevice) Start() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.StartFunc != nil {
		if err := m.StartFunc(); err != nil {
			return err
		}
	}
	
	m.started = true
	return nil
}

// Halt halts the mock device.
func (m *MockDevice) Halt() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.HaltFunc != nil {
		if err := m.HaltFunc(); err != nil {
			return err
		}
	}
	
	m.started = false
	return nil
}

// IsStarted returns true if the device is started.
func (m *MockDevice) IsStarted() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.started
}

// MockAdaptor is a mock adaptor for testing.
type MockAdaptor struct {
	*MockConnection
	port string
	mu   sync.RWMutex
}

// NewMockAdaptor creates a new mock adaptor.
func NewMockAdaptor(name, port string) *MockAdaptor {
	return &MockAdaptor{
		MockConnection: NewMockConnection(name),
		port:          port,
	}
}

// Port returns the adaptor port.
func (m *MockAdaptor) Port() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.port
}

// SetPort sets the adaptor port.
func (m *MockAdaptor) SetPort(port string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.port = port
}

// MockPin is a mock pin for testing.
type MockPin struct {
	number    int
	exported  bool
	direction string
	value     int
	mu        sync.RWMutex
	ExportFunc    func() error
	UnexportFunc  func() error
	DirectionFunc func(string) error
	ReadFunc      func() (int, error)
	WriteFunc     func(int) error
}

// NewMockPin creates a new mock pin.
func NewMockPin(number int) *MockPin {
	return &MockPin{
		number:    number,
		exported:  false,
		direction: "in",
		value:     0,
	}
}

// Export exports the mock pin.
func (m *MockPin) Export() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.ExportFunc != nil {
		if err := m.ExportFunc(); err != nil {
			return err
		}
	}
	
	m.exported = true
	return nil
}

// Unexport unexports the mock pin.
func (m *MockPin) Unexport() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.UnexportFunc != nil {
		if err := m.UnexportFunc(); err != nil {
			return err
		}
	}
	
	m.exported = false
	return nil
}

// Direction sets the pin direction.
func (m *MockPin) Direction(dir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.DirectionFunc != nil {
		if err := m.DirectionFunc(dir); err != nil {
			return err
		}
	}
	
	m.direction = dir
	return nil
}

// Read reads the pin value.
func (m *MockPin) Read() (int, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.ReadFunc != nil {
		return m.ReadFunc()
	}
	
	return m.value, nil
}

// Write writes a value to the pin.
func (m *MockPin) Write(value int) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	if m.WriteFunc != nil {
		if err := m.WriteFunc(value); err != nil {
			return err
		}
	}
	
	m.value = value
	return nil
}

// IsExported returns true if the pin is exported.
func (m *MockPin) IsExported() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.exported
}

// GetDirection returns the pin direction.
func (m *MockPin) GetDirection() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.direction
}

// GetValue returns the pin value.
func (m *MockPin) GetValue() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.value
}

// TestHelper provides utilities for testing
type TestHelper struct {
	startTime time.Time
}

// NewTestHelper creates a new test helper
func NewTestHelper() *TestHelper {
	return &TestHelper{
		startTime: time.Now(),
	}
}

// ElapsedTime returns the time elapsed since the helper was created
func (t *TestHelper) ElapsedTime() time.Duration {
	return time.Since(t.startTime)
}

// WaitFor waits for a condition to be true or timeout
func (t *TestHelper) WaitFor(condition func() bool, timeout time.Duration) bool {
	start := time.Now()
	for time.Since(start) < timeout {
		if condition() {
			return true
		}
		time.Sleep(10 * time.Millisecond)
	}
	return false
}