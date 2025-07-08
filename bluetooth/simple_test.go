package bluetooth

import (
	"context"
	"testing"
	"time"
)

func TestParseAddressString(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		expected Address
	}{
		{
			name: "valid address uppercase",
			input: "12:34:56:78:9A:BC",
			wantErr: false,
			expected: Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}},
		},
		{
			name: "valid address lowercase",
			input: "12:34:56:78:9a:bc",
			wantErr: false,
			expected: Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}},
		},
		{
			name: "invalid length",
			input: "12:34:56",
			wantErr: true,
		},
		{
			name: "invalid format",
			input: "12-34-56-78-9A-BC",
			wantErr: true,
		},
		{
			name: "invalid hex",
			input: "GG:34:56:78:9A:BC",
			wantErr: true,
		},
		{
			name: "empty string",
			input: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseAddressString(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseAddressString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && result != tt.expected {
				t.Errorf("parseAddressString() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// Mock simple manager for testing
type mockSimpleManager struct {
	manager *mockManager
	adapter *mockAdapter
	enabled bool
}

func newMockSimpleManager() *mockSimpleManager {
	adapter := &mockAdapter{
		name: "Mock Adapter",
		address: Address{MAC: [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}},
		powered: true,
	}
	
	central := &mockCentral{
		adapter: adapter,
		enabled: true,
	}
	adapter.central = central

	manager := &mockManager{
		adapters: []Adapter{adapter},
	}

	return &mockSimpleManager{
		manager: manager,
		adapter: adapter,
		enabled: true,
	}
}

func TestSimpleManagerBasics(t *testing.T) {
	// Test with mock implementations
	mockMgr := newMockSimpleManager()
	
	// Create a SimpleManager manually for testing
	sm := &SimpleManager{
		manager: mockMgr.manager,
		adapter: mockMgr.adapter,
	}

	// Test Bluetooth state
	if !sm.IsBluetoothEnabled() {
		t.Error("Mock adapter should be enabled")
	}

	// Test enabling (should be no-op since already enabled)
	ctx := context.Background()
	err := sm.EnableBluetooth(ctx)
	if err != nil {
		t.Errorf("EnableBluetooth() failed: %v", err)
	}

	// Test disabling
	err = sm.DisableBluetooth(ctx)
	if err != nil {
		t.Errorf("DisableBluetooth() failed: %v", err)
	}

	if sm.IsBluetoothEnabled() {
		t.Error("Adapter should be disabled")
	}

	// Re-enable for other tests
	err = sm.EnableBluetooth(ctx)
	if err != nil {
		t.Errorf("EnableBluetooth() failed: %v", err)
	}
}

func TestSimpleManagerScan(t *testing.T) {
	mockMgr := newMockSimpleManager()
	sm := &SimpleManager{
		manager: mockMgr.manager,
		adapter: mockMgr.adapter,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	found := false
	err := sm.ScanForDevices(ctx, 1*time.Second, func(address, name string, rssi int) {
		if name == "Test Device" {
			found = true
		}
		if rssi > 0 {
			t.Error("RSSI should be negative")
		}
	})

	if err != nil {
		t.Errorf("ScanForDevices() failed: %v", err)
	}

	if !found {
		t.Error("Should have found test device")
	}
}

func TestSimpleManagerConnect(t *testing.T) {
	mockMgr := newMockSimpleManager()
	sm := &SimpleManager{
		manager: mockMgr.manager,
		adapter: mockMgr.adapter,
	}

	ctx := context.Background()

	// Test connecting to a device
	address := "12:34:56:78:9A:BC"
	device, err := sm.ConnectToDevice(ctx, address)
	if err != nil {
		t.Errorf("ConnectToDevice() failed: %v", err)
	}

	if device == nil {
		t.Error("ConnectToDevice() returned nil device")
	}

	if !device.IsConnected() {
		t.Error("Device should be connected")
	}

	if device.Address() != address {
		t.Errorf("Device address = %v, want %v", device.Address(), address)
	}

	// Test getting connected devices
	devices := sm.GetConnectedDevices()
	if len(devices) != 1 {
		t.Errorf("GetConnectedDevices() returned %d devices, want 1", len(devices))
	}

	// Test disconnecting
	err = device.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() failed: %v", err)
	}

	if device.IsConnected() {
		t.Error("Device should be disconnected")
	}
}

func TestSimpleManagerAdvertising(t *testing.T) {
	mockMgr := newMockSimpleManager()
	sm := &SimpleManager{
		manager: mockMgr.manager,
		adapter: mockMgr.adapter,
	}

	// Add a mock peripheral to the adapter
	peripheral := &mockPeripheral{
		adapter: mockMgr.adapter,
	}
	mockMgr.adapter.peripheral = peripheral

	ctx := context.Background()

	// Test starting advertising
	err := sm.StartAdvertising(ctx, "Test Device", []string{"180D"})
	if err != nil {
		t.Errorf("StartAdvertising() failed: %v", err)
	}

	// Test stopping advertising
	err = sm.StopAdvertising(ctx)
	if err != nil {
		t.Errorf("StopAdvertising() failed: %v", err)
	}
}

func TestSimpleDeviceOperations(t *testing.T) {
	// Create a mock device
	mockDev := &mockDevice{
		address: Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}},
		name: "Test Device",
		connected: true,
	}

	device := &SimpleDevice{device: mockDev}

	// Test basic properties
	if device.Address() != "bc:9a:78:56:34:12" {
		t.Errorf("Device.Address() = %v, want bc:9a:78:56:34:12", device.Address())
	}

	if device.Name() != "Test Device" {
		t.Errorf("Device.Name() = %v, want Test Device", device.Name())
	}

	if !device.IsConnected() {
		t.Error("Device should be connected")
	}

	// Test service discovery
	ctx := context.Background()
	err := device.DiscoverServices(ctx)
	if err != nil {
		t.Errorf("DiscoverServices() failed: %v", err)
	}

	services := device.GetServices()
	if len(services) != 0 {
		t.Errorf("GetServices() returned %d services, want 0", len(services))
	}

	// Test disconnect
	err = device.Disconnect(ctx)
	if err != nil {
		t.Errorf("Disconnect() failed: %v", err)
	}

	if device.IsConnected() {
		t.Error("Device should be disconnected")
	}
}

// Additional mock implementations for testing
type mockPeripheral struct {
	adapter     *mockAdapter
	advertising bool
	services    []PeripheralService
}

func (p *mockPeripheral) Enable(ctx context.Context) error { return nil }
func (p *mockPeripheral) Disable(ctx context.Context) error { return nil }
func (p *mockPeripheral) AddService(uuid UUID, primary bool) (PeripheralService, error) {
	service := &mockPeripheralService{
		peripheral: p,
		uuid: uuid,
		primary: primary,
	}
	p.services = append(p.services, service)
	return service, nil
}
func (p *mockPeripheral) GetService(uuid UUID) (PeripheralService, error) {
	for _, service := range p.services {
		if service.UUID() == uuid {
			return service, nil
		}
	}
	return nil, ErrServiceNotFound
}
func (p *mockPeripheral) Services() []PeripheralService { return p.services }
func (p *mockPeripheral) StartAdvertising(ctx context.Context, params AdvertisingParams, data AdvertisementData) error {
	p.advertising = true
	return nil
}
func (p *mockPeripheral) StopAdvertising(ctx context.Context) error {
	p.advertising = false
	return nil
}
func (p *mockPeripheral) IsAdvertising() bool { return p.advertising }
func (p *mockPeripheral) OnConnect(callback func(Device)) {}
func (p *mockPeripheral) OnDisconnect(callback func(Device)) {}

type mockPeripheralService struct {
	peripheral      *mockPeripheral
	uuid            UUID
	primary         bool
	characteristics []PeripheralCharacteristic
}

func (s *mockPeripheralService) UUID() UUID { return s.uuid }
func (s *mockPeripheralService) Primary() bool { return s.primary }
func (s *mockPeripheralService) AddCharacteristic(uuid UUID, properties CharacteristicProperty, value []byte) (PeripheralCharacteristic, error) {
	char := &mockPeripheralCharacteristic{
		service: s,
		uuid: uuid,
		properties: properties,
		value: value,
	}
	s.characteristics = append(s.characteristics, char)
	return char, nil
}
func (s *mockPeripheralService) GetCharacteristic(uuid UUID) (PeripheralCharacteristic, error) {
	for _, char := range s.characteristics {
		if char.UUID() == uuid {
			return char, nil
		}
	}
	return nil, ErrCharacteristicNotFound
}
func (s *mockPeripheralService) Characteristics() []PeripheralCharacteristic { return s.characteristics }

type mockPeripheralCharacteristic struct {
	service       *mockPeripheralService
	uuid          UUID
	properties    CharacteristicProperty
	value         []byte
	onRead        func() []byte
	onWrite       func([]byte) error
	onSubscribe   func()
	onUnsubscribe func()
}

func (c *mockPeripheralCharacteristic) UUID() UUID { return c.uuid }
func (c *mockPeripheralCharacteristic) Properties() CharacteristicProperty { return c.properties }
func (c *mockPeripheralCharacteristic) Value() []byte { return c.value }
func (c *mockPeripheralCharacteristic) SetValue(data []byte) error { c.value = data; return nil }
func (c *mockPeripheralCharacteristic) NotifySubscribers(data []byte) error { return nil }
func (c *mockPeripheralCharacteristic) OnRead(callback func() []byte) { c.onRead = callback }
func (c *mockPeripheralCharacteristic) OnWrite(callback func([]byte) error) { c.onWrite = callback }
func (c *mockPeripheralCharacteristic) OnSubscribe(callback func()) { c.onSubscribe = callback }
func (c *mockPeripheralCharacteristic) OnUnsubscribe(callback func()) { c.onUnsubscribe = callback }