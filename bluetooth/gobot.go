// Package bluetooth provides a simplified Bluetooth API that integrates with Gobot's patterns
package bluetooth

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// Common errors following Gobot's simple error pattern
var (
	ErrBLENotConnected     = errors.New("not connected")
	ErrBLENotFound         = errors.New("not found") 
	ErrBLEInvalidUUID      = errors.New("invalid UUID")
	ErrBLEScanTimeout      = errors.New("scan timeout")
	ErrBLEConnectionFailed = errors.New("connection failed")
)

// Adaptor represents the core interface for all Gobot adaptors
type Adaptor interface {
	Name() string
	SetName(string)
	Connect() error
	Finalize() error
}

// BLEConnector represents the interface for BLE functionality
type BLEConnector interface {
	Adaptor
	Reconnect() error
	Disconnect() error
	Address() string
	ReadCharacteristic(cUUID string) ([]byte, error)
	WriteCharacteristic(cUUID string, data []byte) error
	Subscribe(cUUID string, f func(data []byte)) error
	WithoutResponses(use bool)
}

// ClientAdaptor implements a Bluetooth LE client following Gobot patterns
type ClientAdaptor struct {
	name                string
	identifier          string // device address or name to connect to
	scanTimeout         time.Duration
	sleepAfterDisconnect time.Duration
	
	// Internal state
	connected        bool
	deviceAddress    string
	deviceName       string
	characteristics  map[string]string // UUID -> value cache
	
	// Configuration
	withoutResponses bool
	
	// Synchronization
	mutex sync.Mutex
	
	// For future platform implementations
	platformAdapter any
}

// ClientAdaptorOption defines the option applier pattern used throughout Gobot
type ClientAdaptorOption interface {
	apply(*ClientAdaptor)
}

type scanTimeoutOption time.Duration
type sleepAfterDisconnectOption time.Duration

func (o scanTimeoutOption) apply(a *ClientAdaptor) {
	a.scanTimeout = time.Duration(o)
}

func (o sleepAfterDisconnectOption) apply(a *ClientAdaptor) {
	a.sleepAfterDisconnect = time.Duration(o)
}

// WithScanTimeout sets the scan timeout (Gobot option pattern)
func WithScanTimeout(timeout time.Duration) ClientAdaptorOption {
	return scanTimeoutOption(timeout)
}

// WithSleepAfterDisconnect sets sleep time after disconnect
func WithSleepAfterDisconnect(sleep time.Duration) ClientAdaptorOption {
	return sleepAfterDisconnectOption(sleep)
}

// NewClientAdaptor creates a new Bluetooth LE client adaptor
// identifier can be device address (12:34:56:78:9A:BC) or device name
func NewClientAdaptor(identifier string, opts ...ClientAdaptorOption) *ClientAdaptor {
	a := &ClientAdaptor{
		name:                 "BLEClient",
		identifier:           identifier,
		scanTimeout:          10 * time.Minute,
		sleepAfterDisconnect: 500 * time.Millisecond,
		characteristics:      make(map[string]string),
	}
	
	for _, opt := range opts {
		opt.apply(a)
	}
	
	return a
}

// Name returns the adaptor name (Gobot Adaptor interface)
func (a *ClientAdaptor) Name() string {
	return a.name
}

// SetName sets the adaptor name (Gobot Adaptor interface)
func (a *ClientAdaptor) SetName(name string) {
	a.name = name
}

// Connect initiates connection to the BLE device (Gobot Adaptor interface)
func (a *ClientAdaptor) Connect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if a.connected {
		return nil
	}
	
	// For now, simulate connection - in a real implementation this would:
	// 1. Initialize platform-specific Bluetooth adapter
	// 2. Scan for the device by identifier
	// 3. Connect to the device
	// 4. Discover services and characteristics
	
	// Mock successful connection
	a.connected = true
	a.deviceAddress = a.identifier
	if isAddress(a.identifier) {
		a.deviceAddress = a.identifier
		a.deviceName = "BLE Device"
	} else {
		a.deviceName = a.identifier
		a.deviceAddress = "00:00:00:00:00:00"
	}
	
	return nil
}

// Finalize terminates the connection (Gobot Adaptor interface)
func (a *ClientAdaptor) Finalize() error {
	return a.Disconnect()
}

// Reconnect disconnects and reconnects (Gobot BLEConnector interface)
func (a *ClientAdaptor) Reconnect() error {
	if err := a.Disconnect(); err != nil {
		return err
	}
	return a.Connect()
}

// Disconnect closes the connection (Gobot BLEConnector interface)
func (a *ClientAdaptor) Disconnect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if !a.connected {
		return nil
	}
	
	a.connected = false
	time.Sleep(a.sleepAfterDisconnect)
	
	return nil
}

// Address returns the device address (Gobot BLEConnector interface)
func (a *ClientAdaptor) Address() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if a.connected {
		return a.deviceAddress
	}
	return a.identifier
}

// ReadCharacteristic reads from a characteristic (Gobot BLEConnector interface)
func (a *ClientAdaptor) ReadCharacteristic(cUUID string) ([]byte, error) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if !a.connected {
		return nil, ErrBLENotConnected
	}
	
	uuid, err := normalizeUUID(cUUID)
	if err != nil {
		return nil, err
	}
	
	// In a real implementation, this would read from the actual characteristic
	// For now, return mock data based on standard characteristics
	switch uuid {
	case "00002a19-0000-1000-8000-00805f9b34fb": // Battery Level
		return []byte{85}, nil // 85% battery
	case "00002a00-0000-1000-8000-00805f9b34fb": // Device Name
		return []byte(a.deviceName), nil
	default:
		return []byte("mock data"), nil
	}
}

// WriteCharacteristic writes to a characteristic (Gobot BLEConnector interface)
func (a *ClientAdaptor) WriteCharacteristic(cUUID string, data []byte) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if !a.connected {
		return ErrBLENotConnected
	}
	
	uuid, err := normalizeUUID(cUUID)
	if err != nil {
		return err
	}
	
	// Cache the written value
	a.characteristics[uuid] = string(data)
	
	// In a real implementation, this would write to the actual characteristic
	return nil
}

// Subscribe subscribes to characteristic notifications (Gobot BLEConnector interface)
func (a *ClientAdaptor) Subscribe(cUUID string, callback func(data []byte)) error {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	
	if !a.connected {
		return ErrBLENotConnected
	}
	
	_, err := normalizeUUID(cUUID)
	if err != nil {
		return err
	}
	
	// In a real implementation, this would enable notifications
	// For now, simulate periodic notifications
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		
		for range ticker.C {
			a.mutex.Lock()
			connected := a.connected
			a.mutex.Unlock()
			
			if !connected {
				break
			}
			
			// Send mock notification data
			callback([]byte("notification data"))
		}
	}()
	
	return nil
}

// WithoutResponses sets whether to expect responses (Gobot BLEConnector interface)
func (a *ClientAdaptor) WithoutResponses(use bool) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	a.withoutResponses = use
}

// Helper functions following Gobot's simple patterns

// normalizeUUID converts short UUIDs to full UUIDs following Bluetooth standard
func normalizeUUID(uuid string) (string, error) {
	switch len(uuid) {
	case 4:
		// 16-bit UUID: convert to full 128-bit
		return fmt.Sprintf("0000%s-0000-1000-8000-00805f9b34fb", uuid), nil
	case 36:
		// Already full UUID
		return uuid, nil
	case 32:
		// UUID without dashes - add them
		return fmt.Sprintf("%s-%s-%s-%s-%s", 
			uuid[0:8], uuid[8:12], uuid[12:16], uuid[16:20], uuid[20:32]), nil
	default:
		return "", ErrBLEInvalidUUID
	}
}

// isAddress checks if identifier looks like a MAC address
func isAddress(identifier string) bool {
	return len(identifier) == 17 && 
		identifier[2] == ':' && identifier[5] == ':' && 
		identifier[8] == ':' && identifier[11] == ':' && 
		identifier[14] == ':'
}

// Standard Bluetooth UUIDs as simple constants (Gobot style)
const (
	// Standard Services
	BatteryServiceUUID        = "180f"
	DeviceInformationUUID     = "180a"
	GenericAccessUUID         = "1800"
	GenericAttributeUUID      = "1801"
	HeartRateServiceUUID      = "180d"
	
	// Standard Characteristics  
	BatteryLevelUUID          = "2a19"
	DeviceNameUUID            = "2a00"
	ManufacturerNameUUID      = "2a29"
	ModelNumberUUID           = "2a24"
	SerialNumberUUID          = "2a25"
	HardwareRevisionUUID      = "2a27"
	FirmwareRevisionUUID      = "2a26"
	SoftwareRevisionUUID      = "2a28"
)