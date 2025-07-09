// Package bluetooth provides a cross-platform Bluetooth Low Energy API for Go 1.24
// with support for Linux, macOS, Windows, and bare metal using Nordic SoftDevice.
package bluetooth

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)


// Address represents a Bluetooth device address
type Address struct {
	MAC      [6]byte
	IsRandom bool
}

func (a Address) String() string {
	return sprintf("%02x:%02x:%02x:%02x:%02x:%02x",
		a.MAC[5], a.MAC[4], a.MAC[3], a.MAC[2], a.MAC[1], a.MAC[0])
}

// NewAddress creates a new Address from a MAC address string with validation
func NewAddress(mac string) (Address, error) {
	if err := ValidateAddressString(mac); err != nil {
		return Address{}, err
	}

	addr, err := parseAddressString(mac)
	if err != nil {
		return Address{}, NewBluetoothErrorWithCode(ErrorCodeInvalidAddress, "failed to parse address").WithCause(err)
	}

	if err := ValidateAddress(addr); err != nil {
		return Address{}, err
	}

	return addr, nil
}

// UUID represents a Bluetooth UUID (16-bit, 32-bit, or 128-bit)
type UUID struct {
	uuid.UUID
}

// NewUUID creates a new UUID from string with validation
func NewUUID(s string) (UUID, error) {
	// Validate the UUID string format first
	if err := ValidateUUIDString(s); err != nil {
		return UUID{}, err
	}

	// Handle different UUID formats
	var fullUUIDString string
	switch len(s) {
	case 4:
		// 16-bit UUID: Convert to full 128-bit UUID
		fullUUIDString = fmt.Sprintf("0000%s-0000-1000-8000-00805f9b34fb", s)
	case 32:
		// 32-character UUID without dashes: add dashes
		fullUUIDString = fmt.Sprintf("%s-%s-%s-%s-%s",
			s[0:8], s[8:12], s[12:16], s[16:20], s[20:32])
	case 36:
		// Full UUID with dashes
		fullUUIDString = s
	default:
		return UUID{}, NewValidationErrorBT("uuid", s, "invalid length")
	}

	u, err := uuid.Parse(fullUUIDString)
	if err != nil {
		return UUID{}, NewBluetoothErrorWithCode(ErrorCodeInvalidUUID, "failed to parse UUID").WithCause(err)
	}
	
	return UUID{UUID: u}, nil
}

// Advertisement represents advertising data
type Advertisement struct {
	Address          Address
	RSSI             int16
	LocalName        string
	ServiceUUIDs     []UUID
	ServiceData      map[UUID][]byte
	ManufacturerData map[uint16][]byte
	TxPowerLevel     *int8
	Connectable      bool
}

// ConnectionParams defines connection parameters
type ConnectionParams struct {
	ConnectionTimeout  time.Duration
	MinInterval        time.Duration
	MaxInterval        time.Duration
	SlaveLatency       uint16
	SupervisionTimeout time.Duration
}

// DefaultConnectionParams returns default connection parameters
func DefaultConnectionParams() ConnectionParams {
	return ConnectionParams{
		ConnectionTimeout:  10 * time.Second,
		MinInterval:        20 * time.Millisecond,
		MaxInterval:        40 * time.Millisecond,
		SlaveLatency:       0,
		SupervisionTimeout: 4 * time.Second,
	}
}

// Validate validates the connection parameters
func (cp ConnectionParams) Validate() error {
	return ValidateConnectionParams(cp)
}

// ScanParams defines scanning parameters
type ScanParams struct {
	Timeout          time.Duration
	Interval         time.Duration
	Window           time.Duration
	ActiveScan       bool
	FilterDuplicates bool
}

// DefaultScanParams returns default scan parameters
func DefaultScanParams() ScanParams {
	return ScanParams{
		Timeout:          30 * time.Second,
		Interval:         100 * time.Millisecond,
		Window:           50 * time.Millisecond,
		ActiveScan:       true,
		FilterDuplicates: true,
	}
}

// Validate validates the scan parameters
func (sp ScanParams) Validate() error {
	return ValidateScanParams(sp)
}

// AdvertisingParams defines advertising parameters
type AdvertisingParams struct {
	Interval     time.Duration
	Timeout      time.Duration
	Connectable  bool
	Discoverable bool
	TxPower      *int8
}

// DefaultAdvertisingParams returns default advertising parameters
func DefaultAdvertisingParams() AdvertisingParams {
	return AdvertisingParams{
		Interval:     100 * time.Millisecond,
		Timeout:      0, // infinite
		Connectable:  true,
		Discoverable: true,
		TxPower:      nil, // use default
	}
}

// Validate validates the advertising parameters
func (ap AdvertisingParams) Validate() error {
	return ValidateAdvertisingParams(ap)
}

// CharacteristicProperty defines characteristic properties
type CharacteristicProperty uint8

const (
	CharacteristicBroadcast CharacteristicProperty = 1 << iota
	CharacteristicRead
	CharacteristicWriteWithoutResponse
	CharacteristicWrite
	CharacteristicNotify
	CharacteristicIndicate
	CharacteristicAuthenticatedSignedWrites
	CharacteristicExtendedProperties
)

// Descriptor represents a characteristic descriptor
type Descriptor interface {
	UUID() UUID
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
}

// Characteristic represents a GATT characteristic
type Characteristic interface {
	UUID() UUID
	Properties() CharacteristicProperty
	Read(ctx context.Context) ([]byte, error)
	Write(ctx context.Context, data []byte) error
	WriteWithoutResponse(ctx context.Context, data []byte) error
	Subscribe(ctx context.Context, callback func([]byte)) error
	Unsubscribe(ctx context.Context) error
	Descriptors() []Descriptor
}

// Service represents a GATT service
type Service interface {
	UUID() UUID
	Primary() bool
	Characteristics() []Characteristic
	GetCharacteristic(uuid UUID) (Characteristic, error)
}

// Device represents a connected Bluetooth device
type Device interface {
	Address() Address
	Name() string
	RSSI() int16
	Connected() bool
	Disconnect(ctx context.Context) error
	Services() []Service
	GetService(uuid UUID) (Service, error)
	DiscoverServices(ctx context.Context, uuids []UUID) error
	RequestMTU(ctx context.Context, mtu uint16) error
	GetMTU() uint16
}

// Central represents a Bluetooth Central (client) role
type Central interface {
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	Scan(ctx context.Context, params ScanParams, callback func(Advertisement)) error
	StopScan(ctx context.Context) error
	Connect(ctx context.Context, address Address, params ConnectionParams) (Device, error)
	ConnectedDevices() []Device
}

// PeripheralCharacteristic represents a characteristic in peripheral mode
type PeripheralCharacteristic interface {
	UUID() UUID
	Properties() CharacteristicProperty
	Value() []byte
	SetValue(data []byte) error
	NotifySubscribers(data []byte) error
	OnRead(callback func() []byte)
	OnWrite(callback func([]byte) error)
	OnSubscribe(callback func())
	OnUnsubscribe(callback func())
}

// PeripheralService represents a service in peripheral mode
type PeripheralService interface {
	UUID() UUID
	Primary() bool
	AddCharacteristic(uuid UUID, properties CharacteristicProperty, value []byte) (PeripheralCharacteristic, error)
	GetCharacteristic(uuid UUID) (PeripheralCharacteristic, error)
	Characteristics() []PeripheralCharacteristic
}

// Peripheral represents a Bluetooth Peripheral (server) role
type Peripheral interface {
	Enable(ctx context.Context) error
	Disable(ctx context.Context) error
	AddService(uuid UUID, primary bool) (PeripheralService, error)
	GetService(uuid UUID) (PeripheralService, error)
	Services() []PeripheralService
	StartAdvertising(ctx context.Context, params AdvertisingParams, data AdvertisementData) error
	StopAdvertising(ctx context.Context) error
	IsAdvertising() bool
	OnConnect(callback func(Device))
	OnDisconnect(callback func(Device))
}

// AdvertisementData represents data to be advertised
type AdvertisementData struct {
	LocalName        string
	ServiceUUIDs     []UUID
	ServiceData      map[UUID][]byte
	ManufacturerData map[uint16][]byte
	TxPowerLevel     *int8
	Appearance       *uint16
}

// Adapter represents a Bluetooth adapter that can operate in both central and peripheral modes
type Adapter interface {
	Central() Central
	Peripheral() Peripheral
	Address() Address
	Name() string
	SetName(name string) error
	PowerState() bool
	SetPowerState(enabled bool) error
}

// Manager provides access to Bluetooth adapters
type Manager interface {
	DefaultAdapter() (Adapter, error)
	Adapters() ([]Adapter, error)
	OnAdapterAdded(callback func(Adapter))
	OnAdapterRemoved(callback func(Adapter))
}

// GetManager returns the platform-specific Bluetooth manager
func GetManager() (Manager, error) {
	return getPlatformManager()
}

// Standard UUIDs
var (
	// Standard GATT Services
	UUIDGenericAccess     = mustParseUUID("1800")
	UUIDGenericAttribute  = mustParseUUID("1801")
	UUIDBattery           = mustParseUUID("180F")
	UUIDDeviceInformation = mustParseUUID("180A")
	UUIDHeartRate         = mustParseUUID("180D")

	// Standard GATT Characteristics
	UUIDDeviceName                              = mustParseUUID("2A00")
	UUIDAppearance                              = mustParseUUID("2A01")
	UUIDPeripheralPrivacyFlag                   = mustParseUUID("2A02")
	UUIDReconnectionAddress                     = mustParseUUID("2A03")
	UUIDPeripheralPreferredConnectionParameters = mustParseUUID("2A04")
	UUIDBatteryLevel                            = mustParseUUID("2A19")
	UUIDManufacturerNameString                  = mustParseUUID("2A29")
	UUIDModelNumberString                       = mustParseUUID("2A24")
	UUIDSerialNumberString                      = mustParseUUID("2A25")
	UUIDHardwareRevisionString                  = mustParseUUID("2A27")
	UUIDFirmwareRevisionString                  = mustParseUUID("2A26")
	UUIDSoftwareRevisionString                  = mustParseUUID("2A28")
	UUIDSystemID                                = mustParseUUID("2A23")

	// Standard GATT Descriptors
	UUIDCharacteristicExtendedProperties  = mustParseUUID("2900")
	UUIDCharacteristicUserDescription     = mustParseUUID("2901")
	UUIDClientCharacteristicConfiguration = mustParseUUID("2902")
	UUIDServerCharacteristicConfiguration = mustParseUUID("2903")
	UUIDCharacteristicPresentationFormat  = mustParseUUID("2904")
	UUIDCharacteristicAggregateFormat     = mustParseUUID("2905")
)

func mustParseUUID(s string) UUID {
	// Convert short UUIDs to full UUIDs first
	var fullUUID string
	switch len(s) {
	case 4:
		// 16-bit UUID: Convert to full 128-bit UUID
		fullUUID = fmt.Sprintf("0000%s-0000-1000-8000-00805f9b34fb", s)
	case 36:
		// Already a full UUID
		fullUUID = s
	default:
		panic("invalid UUID format: " + s)
	}

	u, err := NewUUID(fullUUID)
	if err != nil {
		panic(err)
	}
	return u
}

// Helper function for sprintf - implemented to avoid imports
func sprintf(format string, args ...any) string {
	// Basic sprintf implementation for MAC address formatting
	if format == "%02x:%02x:%02x:%02x:%02x:%02x" && len(args) == 6 {
		bytes := make([]byte, 6)
		for i, arg := range args {
			if b, ok := arg.(byte); ok {
				bytes[i] = b
			}
		}
		return string([]byte{
			hexDigitUpper(bytes[0] >> 4), hexDigitUpper(bytes[0] & 0xF), ':',
			hexDigitUpper(bytes[1] >> 4), hexDigitUpper(bytes[1] & 0xF), ':',
			hexDigitUpper(bytes[2] >> 4), hexDigitUpper(bytes[2] & 0xF), ':',
			hexDigitUpper(bytes[3] >> 4), hexDigitUpper(bytes[3] & 0xF), ':',
			hexDigitUpper(bytes[4] >> 4), hexDigitUpper(bytes[4] & 0xF), ':',
			hexDigitUpper(bytes[5] >> 4), hexDigitUpper(bytes[5] & 0xF),
		})
	}
	return ""
}

func hexDigit(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'a' + b - 10
}

func hexDigitUpper(b byte) byte {
	if b < 10 {
		return '0' + b
	}
	return 'A' + b - 10
}
