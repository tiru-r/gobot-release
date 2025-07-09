package bluetooth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestValidateAddress(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		address   Address
		expectErr bool
	}{
		{
			name:      "valid address",
			address:   Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}},
			expectErr: false,
		},
		{
			name:      "zero address",
			address:   Address{MAC: [6]byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00}},
			expectErr: true,
		},
		{
			name:      "all same bytes",
			address:   Address{MAC: [6]byte{0xFF, 0xFF, 0xFF, 0xFF, 0xFF, 0xFF}},
			expectErr: true,
		},
		{
			name:      "valid random address",
			address:   Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}, IsRandom: true},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddress(tt.address)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAddressString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		address   string
		expectErr bool
	}{
		{
			name:      "valid address",
			address:   "12:34:56:78:9A:BC",
			expectErr: false,
		},
		{
			name:      "valid lowercase address",
			address:   "12:34:56:78:9a:bc",
			expectErr: false,
		},
		{
			name:      "wrong length",
			address:   "12:34:56:78:9A",
			expectErr: true,
		},
		{
			name:      "wrong separator",
			address:   "12-34-56-78-9A-BC",
			expectErr: true,
		},
		{
			name:      "invalid hex",
			address:   "GG:34:56:78:9A:BC",
			expectErr: true,
		},
		{
			name:      "empty string",
			address:   "",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAddressString(tt.address)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateUUIDString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		uuid      string
		expectErr bool
	}{
		{
			name:      "valid 16-bit UUID",
			uuid:      "180F",
			expectErr: false,
		},
		{
			name:      "valid 32-char UUID",
			uuid:      "12345678123412341234123456789ABC",
			expectErr: false,
		},
		{
			name:      "valid full UUID",
			uuid:      "12345678-1234-1234-1234-123456789ABC",
			expectErr: false,
		},
		{
			name:      "empty string",
			uuid:      "",
			expectErr: true,
		},
		{
			name:      "invalid 16-bit UUID",
			uuid:      "180G",
			expectErr: true,
		},
		{
			name:      "invalid length",
			uuid:      "123",
			expectErr: true,
		},
		{
			name:      "invalid full UUID",
			uuid:      "12345678-1234-1234-1234-123456789XYZ",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateUUIDString(tt.uuid)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateConnectionParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		params    ConnectionParams
		expectErr bool
	}{
		{
			name:      "valid params",
			params:    DefaultConnectionParams(),
			expectErr: false,
		},
		{
			name: "invalid connection timeout",
			params: ConnectionParams{
				ConnectionTimeout:  500 * time.Millisecond, // too short
				MinInterval:        20 * time.Millisecond,
				MaxInterval:        40 * time.Millisecond,
				SupervisionTimeout: 4 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid interval range",
			params: ConnectionParams{
				ConnectionTimeout:  10 * time.Second,
				MinInterval:        50 * time.Millisecond,
				MaxInterval:        30 * time.Millisecond, // max < min
				SupervisionTimeout: 4 * time.Second,
			},
			expectErr: true,
		},
		{
			name: "invalid supervision timeout",
			params: ConnectionParams{
				ConnectionTimeout:  10 * time.Second,
				MinInterval:        20 * time.Millisecond,
				MaxInterval:        40 * time.Millisecond,
				SupervisionTimeout: 1 * time.Millisecond, // too short
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateConnectionParams(tt.params)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateScanParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		params    ScanParams
		expectErr bool
	}{
		{
			name: "valid params",
			params: ScanParams{
				Timeout:  30 * time.Second,
				Interval: 30 * time.Millisecond,
				Window:   20 * time.Millisecond,
			},
			expectErr: false,
		},
		{
			name: "invalid timeout",
			params: ScanParams{
				Timeout:  500 * time.Millisecond, // too short
				Interval: 30 * time.Millisecond,
				Window:   20 * time.Millisecond,
			},
			expectErr: true,
		},
		{
			name: "invalid window",
			params: ScanParams{
				Timeout:  30 * time.Second,
				Interval: 30 * time.Millisecond,
				Window:   40 * time.Millisecond, // window > interval
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScanParams(tt.params)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdvertisingParams(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		params    AdvertisingParams
		expectErr bool
	}{
		{
			name:      "valid params",
			params:    DefaultAdvertisingParams(),
			expectErr: false,
		},
		{
			name: "invalid interval",
			params: AdvertisingParams{
				Interval: 10 * time.Millisecond, // too short
			},
			expectErr: true,
		},
		{
			name: "invalid tx power",
			params: AdvertisingParams{
				Interval: 100 * time.Millisecond,
				TxPower:  func() *int8 { v := int8(50); return &v }(), // too high
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdvertisingParams(tt.params)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateAdvertisementData(t *testing.T) {
	tests := []struct {
		name      string
		data      AdvertisementData
		expectErr bool
	}{
		{
			name: "valid data",
			data: AdvertisementData{
				LocalName: "Test Device",
			},
			expectErr: false,
		},
		{
			name: "too long name",
			data: AdvertisementData{
				LocalName: string(make([]byte, 300)), // too long
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAdvertisementData(tt.data)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCharacteristicData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		data      []byte
		expectErr bool
	}{
		{
			name:      "valid data",
			data:      []byte("Hello World"),
			expectErr: false,
		},
		{
			name:      "too long data",
			data:      make([]byte, 600), // too long
			expectErr: true,
		},
		{
			name:      "empty data",
			data:      []byte{},
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCharacteristicData(tt.data)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateCharacteristicProperties(t *testing.T) {
	tests := []struct {
		name      string
		props     CharacteristicProperty
		expectErr bool
	}{
		{
			name:      "valid properties",
			props:     CharacteristicRead | CharacteristicWrite,
			expectErr: false,
		},
		{
			name:      "invalid combination broadcast + write",
			props:     CharacteristicBroadcast | CharacteristicWrite,
			expectErr: true,
		},
		{
			name:      "valid notify",
			props:     CharacteristicNotify,
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCharacteristicProperties(tt.props)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateDeviceName(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		devName   string
		expectErr bool
	}{
		{
			name:      "valid name",
			devName:   "My Device",
			expectErr: false,
		},
		{
			name:      "too long name",
			devName:   string(make([]byte, 300)),
			expectErr: true,
		},
		{
			name:      "name with control characters",
			devName:   "Device\x00Name",
			expectErr: true,
		},
		{
			name:      "name with allowed whitespace",
			devName:   "Device\tName\nLine2",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDeviceName(tt.devName)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateMTU(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		mtu       uint16
		expectErr bool
	}{
		{
			name:      "valid MTU",
			mtu:       247,
			expectErr: false,
		},
		{
			name:      "minimum MTU",
			mtu:       23,
			expectErr: false,
		},
		{
			name:      "maximum MTU",
			mtu:       517,
			expectErr: false,
		},
		{
			name:      "too small MTU",
			mtu:       22,
			expectErr: true,
		},
		{
			name:      "too large MTU",
			mtu:       600,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMTU(tt.mtu)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidateTimeout(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		timeout   time.Duration
		operation string
		expectErr bool
	}{
		{
			name:      "valid timeout",
			timeout:   30 * time.Second,
			operation: "scan",
			expectErr: false,
		},
		{
			name:      "negative timeout",
			timeout:   -1 * time.Second,
			operation: "scan",
			expectErr: true,
		},
		{
			name:      "zero timeout",
			timeout:   0,
			operation: "scan",
			expectErr: true,
		},
		{
			name:      "too long timeout",
			timeout:   10 * time.Minute,
			operation: "scan",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateTimeout(tt.timeout, tt.operation)
			if tt.expectErr {
				assert.Error(t, err)
				assert.IsType(t, &ValidationError{}, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestValidationError(t *testing.T) {
	t.Parallel()
	err := NewValidationError("test_param", "test_value", "test reason")
	
	assert.Equal(t, "test_param", err.Parameter)
	assert.Equal(t, "test_value", err.Value)
	assert.Equal(t, "test reason", err.Reason)
	assert.Contains(t, err.Error(), "test_param")
	assert.Contains(t, err.Error(), "test_value")
	assert.Contains(t, err.Error(), "test reason")
}

func TestIsHexChar(t *testing.T) {
	tests := []struct {
		char     rune
		expected bool
	}{
		{'0', true},
		{'9', true},
		{'A', true},
		{'F', true},
		{'a', true},
		{'f', true},
		{'G', false},
		{'g', false},
		{'z', false},
		{'@', false},
	}

	for _, tt := range tests {
		t.Run(string(tt.char), func(t *testing.T) {
			result := isHexChar(tt.char)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Benchmark tests for validation functions
func BenchmarkValidateAddress(b *testing.B) {
	addr := Address{MAC: [6]byte{0x12, 0x34, 0x56, 0x78, 0x9A, 0xBC}}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAddress(addr)
	}
}

func BenchmarkValidateAddressString(b *testing.B) {
	addrStr := "12:34:56:78:9A:BC"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateAddressString(addrStr)
	}
}

func BenchmarkValidateUUIDString(b *testing.B) {
	uuid := "12345678-1234-1234-1234-123456789ABC"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateUUIDString(uuid)
	}
}

func BenchmarkValidateConnectionParams(b *testing.B) {
	params := DefaultConnectionParams()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateConnectionParams(params)
	}
}