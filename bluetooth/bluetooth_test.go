package bluetooth

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestUUID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid 16-bit UUID", "180D", false},
		{"valid 128-bit UUID", "12345678-1234-1234-1234-123456789ABC", false},
		{"valid 128-bit UUID without dashes", "12345678123412341234123456789ABC", false},
		{"invalid UUID", "invalid", true},
		{"empty UUID", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, err := NewUUID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewUUID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && u.UUID == (uuid.UUID{}) {
				t.Error("NewUUID() returned empty UUID for valid input")
			}
		})
	}
}

func TestAddressString(t *testing.T) {
	addr := Address{
		MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
	}
	expected := "bc:9a:78:56:34:12"
	result := addr.String()
	if result != expected {
		t.Errorf("Address.String() = %v, want %v", result, expected)
	}
}

func TestDefaultConnectionParams(t *testing.T) {
	params := DefaultConnectionParams()
	if params.ConnectionTimeout == 0 {
		t.Error("DefaultConnectionParams() returned zero ConnectionTimeout")
	}
	if params.MinInterval == 0 {
		t.Error("DefaultConnectionParams() returned zero MinInterval")
	}
	if params.MaxInterval == 0 {
		t.Error("DefaultConnectionParams() returned zero MaxInterval")
	}
	if params.SupervisionTimeout == 0 {
		t.Error("DefaultConnectionParams() returned zero SupervisionTimeout")
	}
}

func TestDefaultScanParams(t *testing.T) {
	params := DefaultScanParams()
	if params.Timeout == 0 {
		t.Error("DefaultScanParams() returned zero Timeout")
	}
	if params.Interval == 0 {
		t.Error("DefaultScanParams() returned zero Interval")
	}
	if params.Window == 0 {
		t.Error("DefaultScanParams() returned zero Window")
	}
}

func TestDefaultAdvertisingParams(t *testing.T) {
	params := DefaultAdvertisingParams()
	if params.Interval == 0 {
		t.Error("DefaultAdvertisingParams() returned zero Interval")
	}
	if !params.Connectable {
		t.Error("DefaultAdvertisingParams() returned non-connectable")
	}
	if !params.Discoverable {
		t.Error("DefaultAdvertisingParams() returned non-discoverable")
	}
}

func TestCharacteristicProperties(t *testing.T) {
	props := CharacteristicRead | CharacteristicWrite | CharacteristicNotify

	if props&CharacteristicRead == 0 {
		t.Error("CharacteristicProperties should include Read")
	}
	if props&CharacteristicWrite == 0 {
		t.Error("CharacteristicProperties should include Write")
	}
	if props&CharacteristicNotify == 0 {
		t.Error("CharacteristicProperties should include Notify")
	}
	if props&CharacteristicIndicate != 0 {
		t.Error("CharacteristicProperties should not include Indicate")
	}
}

func TestStandardUUIDs(t *testing.T) {
	tests := []struct {
		name string
		uuid UUID
	}{
		{"Generic Access", UUIDGenericAccess},
		{"Generic Attribute", UUIDGenericAttribute},
		{"Battery", UUIDBattery},
		{"Device Information", UUIDDeviceInformation},
		{"Heart Rate", UUIDHeartRate},
		{"Device Name", UUIDDeviceName},
		{"Battery Level", UUIDBatteryLevel},
		{"Manufacturer Name", UUIDManufacturerNameString},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.uuid.UUID == (uuid.UUID{}) {
				t.Errorf("Standard UUID %s is empty", tt.name)
			}
		})
	}
}

func TestHexDigit(t *testing.T) {
	tests := []struct {
		input    byte
		expected byte
	}{
		{0, '0'},
		{9, '9'},
		{10, 'a'},
		{15, 'f'},
	}

	for _, tt := range tests {
		result := hexDigit(tt.input)
		if result != tt.expected {
			t.Errorf("hexDigit(%d) = %c, want %c", tt.input, result, tt.expected)
		}
	}
}

// Mock implementations for testing
type mockManager struct {
	adapters []Adapter
}

func (m *mockManager) DefaultAdapter() (Adapter, error) {
	if len(m.adapters) == 0 {
		return nil, ErrNotSupported
	}
	return m.adapters[0], nil
}

func (m *mockManager) Adapters() ([]Adapter, error) {
	return m.adapters, nil
}

func (m *mockManager) OnAdapterAdded(callback func(Adapter))   {}
func (m *mockManager) OnAdapterRemoved(callback func(Adapter)) {}

type mockAdapter struct {
	name       string
	address    Address
	powered    bool
	central    Central
	peripheral Peripheral
}

func (a *mockAdapter) Central() Central                 { return a.central }
func (a *mockAdapter) Peripheral() Peripheral           { return a.peripheral }
func (a *mockAdapter) Address() Address                 { return a.address }
func (a *mockAdapter) Name() string                     { return a.name }
func (a *mockAdapter) SetName(name string) error        { a.name = name; return nil }
func (a *mockAdapter) PowerState() bool                 { return a.powered }
func (a *mockAdapter) SetPowerState(enabled bool) error { a.powered = enabled; return nil }

type mockCentral struct {
	adapter  *mockAdapter
	enabled  bool
	scanning bool
	devices  []Device
}

func (c *mockCentral) Enable(ctx context.Context) error  { c.enabled = true; return nil }
func (c *mockCentral) Disable(ctx context.Context) error { c.enabled = false; return nil }
func (c *mockCentral) Scan(ctx context.Context, params ScanParams, callback func(Advertisement)) error {
	if !c.enabled {
		return ErrNotSupported
	}
	c.scanning = true
	// Simulate finding a device
	go func() {
		time.Sleep(100 * time.Millisecond)
		callback(Advertisement{
			Address:   Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}},
			RSSI:      -60,
			LocalName: "Test Device",
		})
	}()
	return nil
}
func (c *mockCentral) StopScan(ctx context.Context) error { c.scanning = false; return nil }
func (c *mockCentral) Connect(ctx context.Context, address Address, params ConnectionParams) (Device, error) {
	if !c.enabled {
		return nil, ErrNotSupported
	}
	device := &mockDevice{
		central:   c,
		address:   address,
		name:      "Test Device",
		connected: true,
	}
	c.devices = append(c.devices, device)
	return device, nil
}
func (c *mockCentral) ConnectedDevices() []Device { return c.devices }

type mockDevice struct {
	central   *mockCentral
	address   Address
	name      string
	connected bool
	services  []Service
	mtu       uint16
}

func (d *mockDevice) Address() Address                     { return d.address }
func (d *mockDevice) Name() string                         { return d.name }
func (d *mockDevice) RSSI() int16                          { return -60 }
func (d *mockDevice) Connected() bool                      { return d.connected }
func (d *mockDevice) Disconnect(ctx context.Context) error { d.connected = false; return nil }
func (d *mockDevice) Services() []Service                  { return d.services }
func (d *mockDevice) GetService(uuid UUID) (Service, error) {
	for _, service := range d.services {
		if service.UUID() == uuid {
			return service, nil
		}
	}
	return nil, ErrServiceNotFound
}
func (d *mockDevice) DiscoverServices(ctx context.Context, uuids []UUID) error { return nil }
func (d *mockDevice) RequestMTU(ctx context.Context, mtu uint16) error         { d.mtu = mtu; return nil }
func (d *mockDevice) GetMTU() uint16                                           { return d.mtu }

func TestMockAdapter(t *testing.T) {
	adapter := &mockAdapter{
		name:    "Test Adapter",
		address: Address{MAC: [6]byte{0x11, 0x22, 0x33, 0x44, 0x55, 0x66}},
	}

	central := &mockCentral{adapter: adapter}
	adapter.central = central

	// Test adapter methods
	if adapter.Name() != "Test Adapter" {
		t.Error("Adapter name mismatch")
	}

	if adapter.PowerState() {
		t.Error("Adapter should start powered off")
	}

	err := adapter.SetPowerState(true)
	if err != nil {
		t.Errorf("Failed to set power state: %v", err)
	}

	if !adapter.PowerState() {
		t.Error("Adapter should be powered on")
	}

	// Test central methods
	ctx := context.Background()

	err = central.Enable(ctx)
	if err != nil {
		t.Errorf("Failed to enable central: %v", err)
	}

	// Test scanning
	scanDone := make(chan bool)
	err = central.Scan(ctx, DefaultScanParams(), func(adv Advertisement) {
		if adv.LocalName == "Test Device" {
			scanDone <- true
		}
	})
	if err != nil {
		t.Errorf("Failed to start scan: %v", err)
	}

	select {
	case <-scanDone:
		// Success
	case <-time.After(1 * time.Second):
		t.Error("Scan timeout")
	}

	// Test connection
	address := Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}}
	device, err := central.Connect(ctx, address, DefaultConnectionParams())
	if err != nil {
		t.Errorf("Failed to connect: %v", err)
	}

	if !device.Connected() {
		t.Error("Device should be connected")
	}

	if device.Address() != address {
		t.Error("Device address mismatch")
	}

	// Test disconnection
	err = device.Disconnect(ctx)
	if err != nil {
		t.Errorf("Failed to disconnect: %v", err)
	}

	if device.Connected() {
		t.Error("Device should be disconnected")
	}
}
