package bluetooth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// SimpleManager provides a simplified, high-level API for common Bluetooth operations
type SimpleManager struct {
	manager Manager
	adapter Adapter
	mu      sync.RWMutex
}

// SimpleDevice provides a simplified device interface
type SimpleDevice struct {
	device Device
	mu     sync.RWMutex
}

// SimpleService provides a simplified service interface
type SimpleService struct {
	service Service
	mu      sync.RWMutex
}

// SimpleCharacteristic provides a simplified characteristic interface
type SimpleCharacteristic struct {
	characteristic Characteristic
	mu             sync.RWMutex
}

// NewSimpleManager creates a new simplified Bluetooth manager
func NewSimpleManager() (*SimpleManager, error) {
	manager, err := GetManager()
	if err != nil {
		return nil, fmt.Errorf("failed to get Bluetooth manager: %w", err)
	}

	adapter, err := manager.DefaultAdapter()
	if err != nil {
		return nil, fmt.Errorf("failed to get default adapter: %w", err)
	}

	return &SimpleManager{
		manager: manager,
		adapter: adapter,
	}, nil
}

// EnableBluetooth enables the Bluetooth adapter
func (sm *SimpleManager) EnableBluetooth(ctx context.Context) error {
	return sm.adapter.SetPowerState(true)
}

// DisableBluetooth disables the Bluetooth adapter
func (sm *SimpleManager) DisableBluetooth(ctx context.Context) error {
	return sm.adapter.SetPowerState(false)
}

// IsBluetoothEnabled returns true if Bluetooth is enabled
func (sm *SimpleManager) IsBluetoothEnabled() bool {
	return sm.adapter.PowerState()
}

// ScanForDevices scans for nearby Bluetooth devices with a simple callback
func (sm *SimpleManager) ScanForDevices(ctx context.Context, timeout time.Duration, callback func(address string, name string, rssi int)) error {
	central := sm.adapter.Central()

	params := DefaultScanParams()
	params.Timeout = timeout

	return central.Scan(ctx, params, func(adv Advertisement) {
		callback(adv.Address.String(), adv.LocalName, int(adv.RSSI))
	})
}

// ConnectToDevice connects to a device by address
func (sm *SimpleManager) ConnectToDevice(ctx context.Context, address string) (*SimpleDevice, error) {
	addr, err := parseAddressString(address)
	if err != nil {
		return nil, fmt.Errorf("invalid address format: %w", err)
	}

	central := sm.adapter.Central()
	device, err := central.Connect(ctx, addr, DefaultConnectionParams())
	if err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", err)
	}

	return &SimpleDevice{device: device}, nil
}

// GetConnectedDevices returns a list of connected devices
func (sm *SimpleManager) GetConnectedDevices() []*SimpleDevice {
	central := sm.adapter.Central()
	devices := central.ConnectedDevices()

	simpleDevices := make([]*SimpleDevice, len(devices))
	for i, device := range devices {
		simpleDevices[i] = &SimpleDevice{device: device}
	}

	return simpleDevices
}

// StartAdvertising starts advertising with simple parameters
func (sm *SimpleManager) StartAdvertising(ctx context.Context, deviceName string, serviceUUIDs []string) error {
	peripheral := sm.adapter.Peripheral()

	// Convert string UUIDs to UUID type
	uuids := make([]UUID, len(serviceUUIDs))
	for i, uuidStr := range serviceUUIDs {
		uuid, err := NewUUID(uuidStr)
		if err != nil {
			return fmt.Errorf("invalid service UUID %s: %w", uuidStr, err)
		}
		uuids[i] = uuid
	}

	advData := AdvertisementData{
		LocalName:    deviceName,
		ServiceUUIDs: uuids,
	}

	return peripheral.StartAdvertising(ctx, DefaultAdvertisingParams(), advData)
}

// StopAdvertising stops advertising
func (sm *SimpleManager) StopAdvertising(ctx context.Context) error {
	peripheral := sm.adapter.Peripheral()
	return peripheral.StopAdvertising(ctx)
}

// SimpleDevice methods

// Address returns the device address as a string
func (sd *SimpleDevice) Address() string {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return sd.device.Address().String()
}

// Name returns the device name
func (sd *SimpleDevice) Name() string {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return sd.device.Name()
}

// IsConnected returns true if the device is connected
func (sd *SimpleDevice) IsConnected() bool {
	sd.mu.RLock()
	defer sd.mu.RUnlock()
	return sd.device.Connected()
}

// Disconnect disconnects from the device
func (sd *SimpleDevice) Disconnect(ctx context.Context) error {
	return sd.device.Disconnect(ctx)
}

// DiscoverServices discovers all services on the device
func (sd *SimpleDevice) DiscoverServices(ctx context.Context) error {
	return sd.device.DiscoverServices(ctx, nil)
}

// GetServices returns all discovered services
func (sd *SimpleDevice) GetServices() []*SimpleService {
	sd.mu.RLock()
	defer sd.mu.RUnlock()

	services := sd.device.Services()
	simpleServices := make([]*SimpleService, len(services))
	for i, service := range services {
		simpleServices[i] = &SimpleService{service: service}
	}

	return simpleServices
}

// GetServiceByUUID returns a service by UUID string
func (sd *SimpleDevice) GetServiceByUUID(uuidStr string) (*SimpleService, error) {
	uuid, err := NewUUID(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	service, err := sd.device.GetService(uuid)
	if err != nil {
		return nil, err
	}

	return &SimpleService{service: service}, nil
}

// ReadCharacteristic reads a characteristic by service and characteristic UUID strings
func (sd *SimpleDevice) ReadCharacteristic(ctx context.Context, serviceUUID, charUUID string) ([]byte, error) {
	service, err := sd.GetServiceByUUID(serviceUUID)
	if err != nil {
		return nil, err
	}

	characteristic, err := service.GetCharacteristicByUUID(charUUID)
	if err != nil {
		return nil, err
	}

	return characteristic.Read(ctx)
}

// WriteCharacteristic writes to a characteristic by service and characteristic UUID strings
func (sd *SimpleDevice) WriteCharacteristic(ctx context.Context, serviceUUID, charUUID string, data []byte) error {
	service, err := sd.GetServiceByUUID(serviceUUID)
	if err != nil {
		return err
	}

	characteristic, err := service.GetCharacteristicByUUID(charUUID)
	if err != nil {
		return err
	}

	return characteristic.Write(ctx, data)
}

// SubscribeToCharacteristic subscribes to notifications from a characteristic
func (sd *SimpleDevice) SubscribeToCharacteristic(ctx context.Context, serviceUUID, charUUID string, callback func([]byte)) error {
	service, err := sd.GetServiceByUUID(serviceUUID)
	if err != nil {
		return err
	}

	characteristic, err := service.GetCharacteristicByUUID(charUUID)
	if err != nil {
		return err
	}

	return characteristic.Subscribe(ctx, callback)
}

// SimpleService methods

// UUID returns the service UUID as a string
func (ss *SimpleService) UUID() string {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.service.UUID().String()
}

// IsPrimary returns true if this is a primary service
func (ss *SimpleService) IsPrimary() bool {
	ss.mu.RLock()
	defer ss.mu.RUnlock()
	return ss.service.Primary()
}

// GetCharacteristics returns all characteristics in this service
func (ss *SimpleService) GetCharacteristics() []*SimpleCharacteristic {
	ss.mu.RLock()
	defer ss.mu.RUnlock()

	characteristics := ss.service.Characteristics()
	simpleCharacteristics := make([]*SimpleCharacteristic, len(characteristics))
	for i, char := range characteristics {
		simpleCharacteristics[i] = &SimpleCharacteristic{characteristic: char}
	}

	return simpleCharacteristics
}

// GetCharacteristicByUUID returns a characteristic by UUID string
func (ss *SimpleService) GetCharacteristicByUUID(uuidStr string) (*SimpleCharacteristic, error) {
	uuid, err := NewUUID(uuidStr)
	if err != nil {
		return nil, fmt.Errorf("invalid UUID: %w", err)
	}

	characteristic, err := ss.service.GetCharacteristic(uuid)
	if err != nil {
		return nil, err
	}

	return &SimpleCharacteristic{characteristic: characteristic}, nil
}

// SimpleCharacteristic methods

// UUID returns the characteristic UUID as a string
func (sc *SimpleCharacteristic) UUID() string {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.characteristic.UUID().String()
}

// CanRead returns true if the characteristic supports reading
func (sc *SimpleCharacteristic) CanRead() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.characteristic.Properties()&CharacteristicRead != 0
}

// CanWrite returns true if the characteristic supports writing
func (sc *SimpleCharacteristic) CanWrite() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.characteristic.Properties()&CharacteristicWrite != 0
}

// CanNotify returns true if the characteristic supports notifications
func (sc *SimpleCharacteristic) CanNotify() bool {
	sc.mu.RLock()
	defer sc.mu.RUnlock()
	return sc.characteristic.Properties()&CharacteristicNotify != 0
}

// Read reads the characteristic value
func (sc *SimpleCharacteristic) Read(ctx context.Context) ([]byte, error) {
	return sc.characteristic.Read(ctx)
}

// Write writes to the characteristic
func (sc *SimpleCharacteristic) Write(ctx context.Context, data []byte) error {
	return sc.characteristic.Write(ctx, data)
}

// Subscribe subscribes to notifications
func (sc *SimpleCharacteristic) Subscribe(ctx context.Context, callback func([]byte)) error {
	return sc.characteristic.Subscribe(ctx, callback)
}

// Unsubscribe unsubscribes from notifications
func (sc *SimpleCharacteristic) Unsubscribe(ctx context.Context) error {
	return sc.characteristic.Unsubscribe(ctx)
}

// Helper functions

func parseAddressString(addrStr string) (Address, error) {
	if len(addrStr) != 17 { // Format: XX:XX:XX:XX:XX:XX
		return Address{}, fmt.Errorf("invalid address length")
	}

	var addr Address
	for i := 0; i < 6; i++ {
		byteStr := addrStr[i*3 : i*3+2]
		if i < 5 && addrStr[i*3+2] != ':' {
			return Address{}, fmt.Errorf("invalid address format")
		}

		val := 0
		for _, c := range byteStr {
			val *= 16
			if c >= '0' && c <= '9' {
				val += int(c - '0')
			} else if c >= 'a' && c <= 'f' {
				val += int(c - 'a' + 10)
			} else if c >= 'A' && c <= 'F' {
				val += int(c - 'A' + 10)
			} else {
				return Address{}, fmt.Errorf("invalid hex character")
			}
		}
		addr.MAC[5-i] = byte(val)
	}

	return addr, nil
}