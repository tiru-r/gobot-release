package bluetooth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobot.io/x/gobot/v2"
)

func TestClientAdaptorInterface(t *testing.T) {
	// Verify our adaptor implements the required Gobot interfaces
	var _ gobot.Adaptor = (*ClientAdaptor)(nil)
	var _ gobot.BLEConnector = (*ClientAdaptor)(nil)
}

func TestNewClientAdaptor(t *testing.T) {
	tests := []struct {
		name       string
		identifier string
		opts       []ClientAdaptorOption
	}{
		{
			name:       "address identifier",
			identifier: "12:34:56:78:9A:BC",
		},
		{
			name:       "name identifier",
			identifier: "MyDevice",
		},
		{
			name:       "with options",
			identifier: "12:34:56:78:9A:BC",
			opts: []ClientAdaptorOption{
				WithScanTimeout(5 * time.Second),
				WithSleepAfterDisconnect(100 * time.Millisecond),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			adaptor := NewClientAdaptor(tt.identifier, tt.opts...)

			assert.Equal(t, "BLEClient", adaptor.Name())
			assert.Equal(t, tt.identifier, adaptor.identifier)
			assert.False(t, adaptor.connected)
		})
	}
}

func TestClientAdaptorGobotsAdaptorInterface(t *testing.T) {
	adaptor := NewClientAdaptor("test-device")

	// Test Adaptor interface
	assert.Equal(t, "BLEClient", adaptor.Name())

	adaptor.SetName("CustomName")
	assert.Equal(t, "CustomName", adaptor.Name())

	// Test connection lifecycle
	err := adaptor.Connect()
	assert.NoError(t, err)
	assert.True(t, adaptor.connected)

	err = adaptor.Finalize()
	assert.NoError(t, err)
	assert.False(t, adaptor.connected)
}

func TestClientAdaptorBLEConnectorInterface(t *testing.T) {
	adaptor := NewClientAdaptor("12:34:56:78:9A:BC")

	// Test before connection
	assert.Equal(t, "12:34:56:78:9A:BC", adaptor.Address())

	// Connect
	err := adaptor.Connect()
	require.NoError(t, err)

	// Test address after connection
	assert.Equal(t, "12:34:56:78:9A:BC", adaptor.Address())

	// Test characteristic operations
	data, err := adaptor.ReadCharacteristic(BatteryLevelUUID)
	assert.NoError(t, err)
	assert.Len(t, data, 1)
	assert.Equal(t, byte(85), data[0]) // Mock battery level

	err = adaptor.WriteCharacteristic("1234", []byte("test"))
	assert.NoError(t, err)

	// Test subscription
	notificationReceived := make(chan bool, 1)
	err = adaptor.Subscribe("1234", func(data []byte) {
		notificationReceived <- true
	})
	assert.NoError(t, err)

	// Should receive notification within reasonable time
	select {
	case <-notificationReceived:
		// Success
	case <-time.After(2 * time.Second):
		t.Fatal("Did not receive notification")
	}

	// Test disconnection
	err = adaptor.Disconnect()
	assert.NoError(t, err)

	// Test reconnection
	err = adaptor.Reconnect()
	assert.NoError(t, err)
	assert.True(t, adaptor.connected)
}

func TestNormalizeUUID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		wantErr  bool
	}{
		{
			name:     "16-bit UUID",
			input:    "2a19",
			expected: "00002a19-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "full UUID",
			input:    "00002a19-0000-1000-8000-00805f9b34fb",
			expected: "00002a19-0000-1000-8000-00805f9b34fb",
		},
		{
			name:     "UUID without dashes",
			input:    "00002a19000010008000005f9b34fb00",
			expected: "00002a19-0000-1000-8000-005f9b34fb00",
		},
		{
			name:    "invalid UUID",
			input:   "invalid",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeUUID(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsAddress(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "valid MAC address",
			input:    "12:34:56:78:9A:BC",
			expected: true,
		},
		{
			name:     "device name",
			input:    "MyDevice",
			expected: false,
		},
		{
			name:     "invalid format",
			input:    "12-34-56-78-9A-BC",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAddress(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestClientAdaptorOptions(t *testing.T) {
	adaptor := NewClientAdaptor("test",
		WithScanTimeout(5*time.Second),
		WithSleepAfterDisconnect(200*time.Millisecond),
	)

	assert.Equal(t, 5*time.Second, adaptor.scanTimeout)
	assert.Equal(t, 200*time.Millisecond, adaptor.sleepAfterDisconnect)
}

func TestGobotStandardUUIDs(t *testing.T) {
	// Test that standard UUIDs are defined as expected
	assert.Equal(t, "180f", BatteryServiceUUID)
	assert.Equal(t, "2a19", BatteryLevelUUID)
	assert.Equal(t, "2a00", DeviceNameUUID)
	assert.Equal(t, "180a", DeviceInformationUUID)
}

func TestClientAdaptorErrorHandling(t *testing.T) {
	adaptor := NewClientAdaptor("test")

	// Test operations when not connected
	_, err := adaptor.ReadCharacteristic("1234")
	assert.Equal(t, ErrBLENotConnected, err)

	err = adaptor.WriteCharacteristic("1234", []byte("test"))
	assert.Equal(t, ErrBLENotConnected, err)

	err = adaptor.Subscribe("1234", func([]byte) {})
	assert.Equal(t, ErrBLENotConnected, err)
}

func TestClientAdaptorConcurrency(t *testing.T) {
	adaptor := NewClientAdaptor("test")
	err := adaptor.Connect()
	require.NoError(t, err)

	// Test concurrent access
	done := make(chan bool, 3)

	// Concurrent reads
	go func() {
		_, err := adaptor.ReadCharacteristic(BatteryLevelUUID)
		assert.NoError(t, err)
		done <- true
	}()

	// Concurrent writes
	go func() {
		err := adaptor.WriteCharacteristic("1234", []byte("test"))
		assert.NoError(t, err)
		done <- true
	}()

	// Concurrent address access
	go func() {
		addr := adaptor.Address()
		assert.NotEmpty(t, addr)
		done <- true
	}()

	// Wait for all operations to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
			// Success
		case <-time.After(1 * time.Second):
			t.Fatal("Concurrent operation timed out")
		}
	}
}
