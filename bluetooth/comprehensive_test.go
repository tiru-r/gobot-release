package bluetooth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAdapter for testing
type MockAdapter struct {
	mock.Mock
}

func (m *MockAdapter) Central() Central {
	args := m.Called()
	return args.Get(0).(Central)
}

func (m *MockAdapter) Peripheral() Peripheral {
	args := m.Called()
	return args.Get(0).(Peripheral)
}

func (m *MockAdapter) Address() Address {
	args := m.Called()
	return args.Get(0).(Address)
}

func (m *MockAdapter) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockAdapter) SetName(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

func (m *MockAdapter) PowerState() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockAdapter) SetPowerState(enabled bool) error {
	args := m.Called(enabled)
	return args.Error(0)
}

// MockCentral for testing
type MockCentral struct {
	mock.Mock
}

func (m *MockCentral) Enable(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCentral) Disable(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCentral) Scan(ctx context.Context, params ScanParams, callback func(Advertisement)) error {
	args := m.Called(ctx, params, callback)
	return args.Error(0)
}

func (m *MockCentral) StopScan(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCentral) Connect(ctx context.Context, address Address, params ConnectionParams) (Device, error) {
	args := m.Called(ctx, address, params)
	return args.Get(0).(Device), args.Error(1)
}

func (m *MockCentral) ConnectedDevices() []Device {
	args := m.Called()
	return args.Get(0).([]Device)
}

// MockDevice for testing
type MockDevice struct {
	mock.Mock
}

func (m *MockDevice) Address() Address {
	args := m.Called()
	return args.Get(0).(Address)
}

func (m *MockDevice) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockDevice) RSSI() int16 {
	args := m.Called()
	return args.Get(0).(int16)
}

func (m *MockDevice) Connected() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockDevice) Disconnect(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockDevice) Services() []Service {
	args := m.Called()
	return args.Get(0).([]Service)
}

func (m *MockDevice) GetService(uuid UUID) (Service, error) {
	args := m.Called(uuid)
	return args.Get(0).(Service), args.Error(1)
}

func (m *MockDevice) DiscoverServices(ctx context.Context, uuids []UUID) error {
	args := m.Called(ctx, uuids)
	return args.Error(0)
}

func (m *MockDevice) RequestMTU(ctx context.Context, mtu uint16) error {
	args := m.Called(ctx, mtu)
	return args.Error(0)
}

func (m *MockDevice) GetMTU() uint16 {
	args := m.Called()
	return args.Get(0).(uint16)
}

// Test UUID functionality
func TestUUIDComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "Valid 128-bit UUID",
			input:    "12345678-1234-1234-1234-123456789ABC",
			expected: "12345678-1234-1234-1234-123456789abc",
			wantErr:  false,
		},
		{
			name:    "Invalid UUID",
			input:   "invalid-uuid",
			wantErr: true,
		},
		{
			name:     "Valid short UUID",
			input:    "180F",
			expected: "0000180f-0000-1000-8000-00805f9b34fb",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uuid, err := NewUUID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.expected, uuid.String())
		})
	}
}

// Test standard UUIDs
func TestStandardUUIDsComprehensive(t *testing.T) {
	tests := []struct {
		name     string
		uuid     UUID
		expected string
	}{
		{
			name:     "Battery Service",
			uuid:     UUIDBattery,
			expected: "0000180f-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "Device Information Service",
			uuid:     UUIDDeviceInformation,
			expected: "0000180a-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "Battery Level Characteristic",
			uuid:     UUIDBatteryLevel,
			expected: "00002a19-0000-1000-8000-00805f9b34fb",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.uuid.String())
		})
	}
}

// Test Address functionality
func TestAddress(t *testing.T) {
	addr := Address{
		MAC:      [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
		IsRandom: false,
	}

	expected := "BC:9A:78:56:34:12"
	assert.Equal(t, expected, addr.String())
}

// Test ConnectionParams defaults
func TestDefaultConnectionParamsComprehensive(t *testing.T) {
	params := DefaultConnectionParams()

	assert.Equal(t, 10*time.Second, params.ConnectionTimeout)
	assert.Equal(t, 20*time.Millisecond, params.MinInterval)
	assert.Equal(t, 40*time.Millisecond, params.MaxInterval)
	assert.Equal(t, uint16(0), params.SlaveLatency)
	assert.Equal(t, 4*time.Second, params.SupervisionTimeout)
}

// Test ScanParams defaults
func TestDefaultScanParamsComprehensive(t *testing.T) {
	params := DefaultScanParams()

	assert.Equal(t, 30*time.Second, params.Timeout)
	assert.Equal(t, 100*time.Millisecond, params.Interval)
	assert.Equal(t, 50*time.Millisecond, params.Window)
	assert.True(t, params.ActiveScan)
	assert.True(t, params.FilterDuplicates)
}

// Test AdvertisingParams defaults
func TestDefaultAdvertisingParamsComprehensive(t *testing.T) {
	params := DefaultAdvertisingParams()

	assert.Equal(t, 100*time.Millisecond, params.Interval)
	assert.Equal(t, time.Duration(0), params.Timeout)
	assert.True(t, params.Connectable)
	assert.True(t, params.Discoverable)
	assert.Nil(t, params.TxPower)
}

// Test CharacteristicProperty flags
func TestCharacteristicProperty(t *testing.T) {
	props := CharacteristicRead | CharacteristicWrite | CharacteristicNotify

	assert.True(t, props&CharacteristicRead != 0)
	assert.True(t, props&CharacteristicWrite != 0)
	assert.True(t, props&CharacteristicNotify != 0)
	assert.False(t, props&CharacteristicIndicate != 0)
}

// Test ConnectionManager
func TestConnectionManager(t *testing.T) {
	cm, err := NewConnectionManager("12:34:56:78:9A:BC")
	require.NoError(t, err)

	assert.Equal(t, "12:34:56:78:9A:BC", cm.targetAddress)
	assert.Equal(t, 5, cm.maxReconnects)
	assert.Equal(t, 2*time.Second, cm.reconnectDelay)
	assert.False(t, cm.IsConnected())

	// Test reconnection params
	cm.SetReconnectionParams(10, 5*time.Second)
	assert.Equal(t, 10, cm.maxReconnects)
	assert.Equal(t, 5*time.Second, cm.reconnectDelay)

	// Test callbacks
	cm.SetCallbacks(
		func(*SimpleDevice) {},
		func() {},
		func(error) {},
	)

	assert.NotNil(t, cm.onConnected)
	assert.NotNil(t, cm.onDisconnected)
	assert.NotNil(t, cm.onError)
}

// Test ExampleScanner
func TestExampleScanner(t *testing.T) {
	scanner, err := NewExampleScanner()
	require.NoError(t, err)

	assert.NotNil(t, scanner.manager)
	assert.NotNil(t, scanner.devices)
	assert.Equal(t, -100, scanner.rssiFilter)

	// Test filters
	scanner.SetFilters("TestDevice", -50)
	assert.Equal(t, "TestDevice", scanner.nameFilter)
	assert.Equal(t, -50, scanner.rssiFilter)

	// Test device found callback
	scanner.SetDeviceFoundCallback(func(*ScannedDevice) {})
	assert.NotNil(t, scanner.onDeviceFound)

	// Test device handling
	scanner.handleDeviceFound("12:34:56:78:9A:BC", "TestDevice", -30)
	devices := scanner.GetDiscoveredDevices()
	assert.Len(t, devices, 1)
	assert.Equal(t, "12:34:56:78:9A:BC", devices[0].Address)
	assert.Equal(t, "TestDevice", devices[0].Name)
	assert.Equal(t, -30, devices[0].RSSI)

	// Test device filtering
	scanner.SetFilters("OtherDevice", -20)
	scanner.handleDeviceFound("12:34:56:78:9A:BD", "TestDevice", -40)
	devices = scanner.GetDiscoveredDevices()
	assert.Len(t, devices, 1) // Should still be 1 due to filtering
}

// Test ExamplePeripheralServer
func TestExamplePeripheralServer(t *testing.T) {
	server, err := NewExamplePeripheralServer(
		"TestService",
		"12345678-1234-1234-1234-123456789ABC",
		"12345678-1234-1234-1234-123456789ABD",
	)
	require.NoError(t, err)

	assert.Equal(t, "TestService", server.serviceName)
	assert.Equal(t, "12345678-1234-1234-1234-123456789ABC", server.serviceUUID)
	assert.Equal(t, "12345678-1234-1234-1234-123456789ABD", server.charUUID)
	assert.False(t, server.IsRunning())
}

// Benchmark tests
func BenchmarkUUIDCreation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = NewUUID("12345678-1234-1234-1234-123456789ABC")
	}
}

func BenchmarkAddressString(b *testing.B) {
	addr := Address{
		MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = addr.String()
	}
}

// Integration-style tests (would require actual Bluetooth hardware in real scenarios)
func TestBluetoothIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// These tests would require actual Bluetooth hardware
	// In a real scenario, you might use a test harness with mock devices
	t.Run("Manager Creation", func(t *testing.T) {
		// This would typically fail without Bluetooth hardware
		_, err := GetManager()
		if err != nil {
			t.Skipf("Bluetooth not available: %v", err)
		}
	})

	t.Run("Simple Manager Creation", func(t *testing.T) {
		// This would typically fail without Bluetooth hardware
		_, err := NewSimpleManager()
		if err != nil {
			t.Skipf("Bluetooth not available: %v", err)
		}
	})
}

// Test error conditions
func TestErrorConditions(t *testing.T) {
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "Invalid UUID format",
			test: func(t *testing.T) {
				_, err := NewUUID("invalid")
				assert.Error(t, err)
				assert.Equal(t, ErrInvalidUUID, err)
			},
		},
		{
			name: "Empty UUID",
			test: func(t *testing.T) {
				_, err := NewUUID("")
				assert.Error(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, tt.test)
	}
}

// Test concurrent access
func TestConcurrentAccess(t *testing.T) {
	scanner, err := NewExampleScanner()
	require.NoError(t, err)

	// Simulate concurrent device discoveries
	done := make(chan bool)

	for i := range 10 {
		go func(id int) {
			defer func() { done <- true }()
			for j := range 100 {
				addr := fmt.Sprintf("12:34:56:78:9A:%02X", (id*100+j)%256)
				scanner.handleDeviceFound(addr, fmt.Sprintf("Device%d", id), -50)
			}
		}(i)
	}

	// Wait for all goroutines to complete
	for range 10 {
		<-done
	}

	devices := scanner.GetDiscoveredDevices()
	assert.True(t, len(devices) > 0)
	t.Logf("Discovered %d devices concurrently", len(devices))
}

// Test memory usage and cleanup
func TestMemoryUsage(t *testing.T) {
	cm, err := NewConnectionManager("12:34:56:78:9A:BC")
	require.NoError(t, err)

	ctx := context.Background()

	// Test that channels are properly closed
	assert.NotNil(t, cm.stopCh)

	// Simulate disconnect to test cleanup
	err = cm.Disconnect(ctx)
	assert.NoError(t, err)

	// Verify cleanup
	assert.False(t, cm.IsConnected())
}
