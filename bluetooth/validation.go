package bluetooth

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Validation patterns and constants
const (
	// Address validation
	MacAddressLength = 17
	MacAddressRegex  = `^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$`

	// UUID validation
	UUIDShortLength = 4
	UUIDFullLength  = 36
	UUIDBareLength  = 32

	// Connection parameter limits
	MinConnectionTimeout  = 1 * time.Second
	MaxConnectionTimeout  = 60 * time.Second
	MinInterval          = 7500 * time.Microsecond  // 7.5ms
	MaxInterval          = 4 * time.Second
	MaxSlaveLatency      = 499
	MinSupervisionTimeout = 100 * time.Millisecond
	MaxSupervisionTimeout = 32 * time.Second

	// Scan parameter limits
	MinScanTimeout   = 1 * time.Second
	MaxScanTimeout   = 10 * time.Minute
	MinScanInterval  = 2500 * time.Microsecond // 2.5ms
	MaxScanInterval  = 40959 * time.Microsecond
	MinScanWindow    = 2500 * time.Microsecond
	MaxScanWindow    = 40959 * time.Microsecond

	// Advertising parameter limits
	MinAdvertisingInterval = 20 * time.Millisecond
	MaxAdvertisingInterval = 10240 * time.Millisecond
	MinTxPower            = -127
	MaxTxPower            = 20

	// Data length limits
	MaxAdvertisementDataLength = 31
	MaxCharacteristicLength    = 512
	MaxDeviceNameLength        = 248
)

var (
	macAddressRegex = regexp.MustCompile(MacAddressRegex)
)

// ValidationError represents a parameter validation error
type ValidationError struct {
	Parameter string
	Value     any
	Reason    string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("invalid parameter %s: %v (%s)", e.Parameter, e.Value, e.Reason)
}

// NewValidationError creates a new validation error
func NewValidationError(parameter string, value any, reason string) *ValidationError {
	return &ValidationError{
		Parameter: parameter,
		Value:     value,
		Reason:    reason,
	}
}

// ValidateAddress validates a Bluetooth MAC address
func ValidateAddress(addr Address) error {
	// Check for zero address
	if addr.MAC == [6]byte{} {
		return NewValidationError("address", addr, "cannot be zero address")
	}

	// Additional validation for specific invalid patterns
	allSame := true
	for i := 1; i < 6; i++ {
		if addr.MAC[i] != addr.MAC[0] {
			allSame = false
			break
		}
	}
	if allSame {
		return NewValidationError("address", addr, "cannot have all identical bytes")
	}

	return nil
}

// ValidateAddressString validates a MAC address string format
func ValidateAddressString(addrStr string) error {
	if len(addrStr) != MacAddressLength {
		return NewValidationError("address", addrStr, fmt.Sprintf("must be %d characters long", MacAddressLength))
	}

	if !macAddressRegex.MatchString(addrStr) {
		return NewValidationError("address", addrStr, "must be in format XX:XX:XX:XX:XX:XX")
	}

	return nil
}

// ValidateUUID validates a UUID string or object
func ValidateUUID(u UUID) error {
	if u.UUID == (uuid.UUID{}) {
		return NewValidationError("uuid", u, "cannot be nil UUID")
	}

	// Additional validation for reserved UUIDs if needed
	return nil
}

// ValidateUUIDString validates a UUID string format
func ValidateUUIDString(uuidStr string) error {
	if uuidStr == "" {
		return NewValidationError("uuid", uuidStr, "cannot be empty")
	}

	// Remove any whitespace
	uuidStr = strings.TrimSpace(uuidStr)

	switch len(uuidStr) {
	case UUIDShortLength:
		// 16-bit UUID: validate hex
		for _, r := range uuidStr {
			if !isHexChar(r) {
				return NewValidationError("uuid", uuidStr, "short UUID must contain only hex characters")
			}
		}
	case UUIDBareLength:
		// 32-character UUID without dashes
		for _, r := range uuidStr {
			if !isHexChar(r) {
				return NewValidationError("uuid", uuidStr, "bare UUID must contain only hex characters")
			}
		}
	case UUIDFullLength:
		// Full UUID with dashes: validate with google/uuid
		if _, err := uuid.Parse(uuidStr); err != nil {
			return NewValidationError("uuid", uuidStr, fmt.Sprintf("invalid UUID format: %v", err))
		}
	default:
		return NewValidationError("uuid", uuidStr, "invalid UUID length")
	}

	return nil
}

// ValidateConnectionParams validates connection parameters
func ValidateConnectionParams(params ConnectionParams) error {
	if params.ConnectionTimeout < MinConnectionTimeout || params.ConnectionTimeout > MaxConnectionTimeout {
		return NewValidationError("connectionTimeout", params.ConnectionTimeout,
			fmt.Sprintf("must be between %v and %v", MinConnectionTimeout, MaxConnectionTimeout))
	}

	if params.MinInterval < MinInterval || params.MinInterval > MaxInterval {
		return NewValidationError("minInterval", params.MinInterval,
			fmt.Sprintf("must be between %v and %v", MinInterval, MaxInterval))
	}

	if params.MaxInterval < MinInterval || params.MaxInterval > MaxInterval {
		return NewValidationError("maxInterval", params.MaxInterval,
			fmt.Sprintf("must be between %v and %v", MinInterval, MaxInterval))
	}

	if params.MinInterval > params.MaxInterval {
		return NewValidationError("intervals", fmt.Sprintf("min=%v, max=%v", params.MinInterval, params.MaxInterval),
			"minInterval cannot be greater than maxInterval")
	}

	if params.SlaveLatency > MaxSlaveLatency {
		return NewValidationError("slaveLatency", params.SlaveLatency,
			fmt.Sprintf("must be <= %d", MaxSlaveLatency))
	}

	if params.SupervisionTimeout < MinSupervisionTimeout || params.SupervisionTimeout > MaxSupervisionTimeout {
		return NewValidationError("supervisionTimeout", params.SupervisionTimeout,
			fmt.Sprintf("must be between %v and %v", MinSupervisionTimeout, MaxSupervisionTimeout))
	}

	// Supervision timeout must be larger than effective connection interval
	effectiveInterval := params.MaxInterval * time.Duration(1+params.SlaveLatency)
	minSupervisionTimeout := effectiveInterval * 6 // BLE spec recommendation
	if params.SupervisionTimeout < minSupervisionTimeout {
		return NewValidationError("supervisionTimeout", params.SupervisionTimeout,
			fmt.Sprintf("must be at least %v (6x effective connection interval)", minSupervisionTimeout))
	}

	return nil
}

// ValidateScanParams validates scan parameters
func ValidateScanParams(params ScanParams) error {
	if params.Timeout < MinScanTimeout || params.Timeout > MaxScanTimeout {
		return NewValidationError("timeout", params.Timeout,
			fmt.Sprintf("must be between %v and %v", MinScanTimeout, MaxScanTimeout))
	}

	if params.Interval < MinScanInterval || params.Interval > MaxScanInterval {
		return NewValidationError("interval", params.Interval,
			fmt.Sprintf("must be between %v and %v", MinScanInterval, MaxScanInterval))
	}

	if params.Window < MinScanWindow || params.Window > MaxScanWindow {
		return NewValidationError("window", params.Window,
			fmt.Sprintf("must be between %v and %v", MinScanWindow, MaxScanWindow))
	}

	if params.Window > params.Interval {
		return NewValidationError("window", fmt.Sprintf("window=%v, interval=%v", params.Window, params.Interval),
			"scan window cannot be greater than scan interval")
	}

	return nil
}

// ValidateAdvertisingParams validates advertising parameters
func ValidateAdvertisingParams(params AdvertisingParams) error {
	if params.Interval < MinAdvertisingInterval || params.Interval > MaxAdvertisingInterval {
		return NewValidationError("interval", params.Interval,
			fmt.Sprintf("must be between %v and %v", MinAdvertisingInterval, MaxAdvertisingInterval))
	}

	if params.TxPower != nil {
		if *params.TxPower < MinTxPower || *params.TxPower > MaxTxPower {
			return NewValidationError("txPower", *params.TxPower,
				fmt.Sprintf("must be between %d and %d dBm", MinTxPower, MaxTxPower))
		}
	}

	return nil
}

// ValidateAdvertisementData validates advertisement data
func ValidateAdvertisementData(data AdvertisementData) error {
	if len(data.LocalName) > MaxDeviceNameLength {
		return NewValidationError("localName", data.LocalName,
			fmt.Sprintf("must be <= %d bytes", MaxDeviceNameLength))
	}

	// Validate service UUIDs
	for i, serviceUUID := range data.ServiceUUIDs {
		if err := ValidateUUID(serviceUUID); err != nil {
			return NewValidationError(fmt.Sprintf("serviceUUIDs[%d]", i), serviceUUID, err.Error())
		}
	}

	// Validate service data UUIDs
	for serviceUUID, serviceData := range data.ServiceData {
		if err := ValidateUUID(serviceUUID); err != nil {
			return NewValidationError("serviceData UUID", serviceUUID, err.Error())
		}
		if len(serviceData) > MaxCharacteristicLength {
			return NewValidationError("serviceData", serviceData,
				fmt.Sprintf("service data must be <= %d bytes", MaxCharacteristicLength))
		}
	}

	// Calculate total advertisement data size (simplified)
	totalSize := len(data.LocalName)
	for _, serviceData := range data.ServiceData {
		totalSize += len(serviceData) + 3 // UUID + length byte + type byte
	}
	if totalSize > MaxAdvertisementDataLength {
		return NewValidationError("advertisementData", totalSize,
			fmt.Sprintf("total advertisement data must be <= %d bytes", MaxAdvertisementDataLength))
	}

	return nil
}

// ValidateCharacteristicData validates characteristic data
func ValidateCharacteristicData(data []byte) error {
	if len(data) > MaxCharacteristicLength {
		return NewValidationError("characteristicData", len(data),
			fmt.Sprintf("must be <= %d bytes", MaxCharacteristicLength))
	}
	return nil
}

// ValidateCharacteristicProperties validates characteristic properties
func ValidateCharacteristicProperties(props CharacteristicProperty) error {
	// Check for invalid property combinations
	if props&CharacteristicBroadcast != 0 && props&CharacteristicWrite != 0 {
		return NewValidationError("properties", props, "broadcast and write properties are mutually exclusive")
	}

	if props&CharacteristicNotify != 0 && props&CharacteristicIndicate != 0 {
		return NewValidationError("properties", props, "notify and indicate should not both be set")
	}

	return nil
}

// ValidateDeviceName validates a device name
func ValidateDeviceName(name string) error {
	if len(name) > MaxDeviceNameLength {
		return NewValidationError("deviceName", name,
			fmt.Sprintf("must be <= %d bytes", MaxDeviceNameLength))
	}

	// Check for invalid characters (control characters)
	for i, r := range name {
		if r < 32 && r != 9 && r != 10 && r != 13 { // Allow tab, LF, CR
			return NewValidationError("deviceName", name,
				fmt.Sprintf("contains invalid control character at position %d", i))
		}
	}

	return nil
}

// Helper functions

func isHexChar(r rune) bool {
	return (r >= '0' && r <= '9') || (r >= 'A' && r <= 'F') || (r >= 'a' && r <= 'f')
}

// ValidateContext validates context parameters
func ValidateContext(operation string) error {
	// Context validation not implemented - would check for cancellation, deadlines, etc.
	return nil
}

// ValidateMTU validates MTU size
func ValidateMTU(mtu uint16) error {
	const (
		MinMTU = 23   // Minimum ATT MTU
		MaxMTU = 517  // Maximum ATT MTU per BLE spec
	)

	if mtu < MinMTU || mtu > MaxMTU {
		return NewValidationError("mtu", mtu,
			fmt.Sprintf("must be between %d and %d", MinMTU, MaxMTU))
	}

	return nil
}

// ValidateTimeout validates a timeout duration
func ValidateTimeout(timeout time.Duration, operation string) error {
	if timeout < 0 {
		return NewValidationError("timeout", timeout, "cannot be negative")
	}

	if timeout == 0 {
		return NewValidationError("timeout", timeout, "cannot be zero for "+operation)
	}

	maxTimeout := 5 * time.Minute // Reasonable maximum for most operations
	if timeout > maxTimeout {
		return NewValidationError("timeout", timeout,
			fmt.Sprintf("cannot exceed %v for %s", maxTimeout, operation))
	}

	return nil
}