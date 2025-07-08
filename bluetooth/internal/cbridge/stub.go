//go:build !darwin

// Package cbridge provides a stub implementation for non-Darwin platforms
package cbridge

import (
	"unsafe"
)

// This file provides stub implementations for platforms other than Darwin
// to ensure the package can be imported without build errors.

// ============================================================================
// ERROR DEFINITIONS
// ============================================================================

// BluetoothError represents a Bluetooth-specific error
type BluetoothError struct {
	Message string
}

func NewBluetoothError(message string) *BluetoothError {
	return &BluetoothError{Message: message}
}

func (e *BluetoothError) Error() string {
	return e.Message
}

var (
	ErrAdapterNotFound  = NewBluetoothError("adapter not found")
	ErrEnableFailed     = NewBluetoothError("failed to enable")
	ErrDisableFailed    = NewBluetoothError("failed to disable")
	ErrScanFailed       = NewBluetoothError("scan failed")
	ErrScanStopFailed   = NewBluetoothError("scan stop failed")
	ErrConnectionFailed = NewBluetoothError("connection failed")
)

// ============================================================================
// MANAGER FUNCTIONS
// ============================================================================

// NewCBTManager creates a new CBT manager (stub implementation)
func NewCBTManager(goManager unsafe.Pointer) unsafe.Pointer {
	return nil
}

// FreeCBTManager frees a CBT manager (stub implementation)
func FreeCBTManager(manager unsafe.Pointer) {
	// No-op for stub
}

// GetDefaultAdapter gets the default adapter (stub implementation)
func GetDefaultAdapter(manager unsafe.Pointer) (unsafe.Pointer, error) {
	return nil, ErrAdapterNotFound
}

// ============================================================================
// CENTRAL FUNCTIONS
// ============================================================================

// NewCBTCentral creates a new CBT central (stub implementation)
func NewCBTCentral(adapter unsafe.Pointer) unsafe.Pointer {
	return nil
}

// FreeCBTCentral frees a CBT central (stub implementation)
func FreeCBTCentral(central unsafe.Pointer) {
	// No-op for stub
}

// EnableCentral enables the central manager (stub implementation)
func EnableCentral(central unsafe.Pointer) error {
	return ErrEnableFailed
}

// DisableCentral disables the central manager (stub implementation)
func DisableCentral(central unsafe.Pointer) error {
	return ErrDisableFailed
}

// StartScan starts scanning for devices (stub implementation)
func StartScan(central unsafe.Pointer, timeout int) error {
	return ErrScanFailed
}

// StopScan stops scanning for devices (stub implementation)
func StopScan(central unsafe.Pointer) error {
	return ErrScanStopFailed
}

// ConnectDevice connects to a device (stub implementation)
func ConnectDevice(central unsafe.Pointer, identifier string) (unsafe.Pointer, error) {
	return nil, ErrConnectionFailed
}

// ============================================================================
// PERIPHERAL FUNCTIONS
// ============================================================================

// NewCBTPeripheral creates a new CBT peripheral (stub implementation)
func NewCBTPeripheral(adapter unsafe.Pointer) unsafe.Pointer {
	return nil
}

// FreeCBTPeripheral frees a CBT peripheral (stub implementation)
func FreeCBTPeripheral(peripheral unsafe.Pointer) {
	// No-op for stub
}

// EnablePeripheral enables the peripheral manager (stub implementation)
func EnablePeripheral(peripheral unsafe.Pointer) error {
	return ErrEnableFailed
}

// DisablePeripheral disables the peripheral manager (stub implementation)
func DisablePeripheral(peripheral unsafe.Pointer) error {
	return ErrDisableFailed
}

// StartAdvertising starts advertising (stub implementation)
func StartAdvertising(peripheral unsafe.Pointer, name, serviceUUID string) error {
	return NewBluetoothError("advertising start failed")
}

// StopAdvertising stops advertising (stub implementation)
func StopAdvertising(peripheral unsafe.Pointer) error {
	return NewBluetoothError("advertising stop failed")
}

// ============================================================================
// DEVICE FUNCTIONS
// ============================================================================

// DisconnectDevice disconnects from a device (stub implementation)
func DisconnectDevice(device unsafe.Pointer) error {
	return ErrConnectionFailed
}

// DiscoverServices discovers services on a device (stub implementation)
func DiscoverServices(device unsafe.Pointer) error {
	return NewBluetoothError("service discovery failed")
}

// GetDeviceName gets the name of a device (stub implementation)
func GetDeviceName(device unsafe.Pointer) string {
	return ""
}

// GetDeviceRSSI gets the RSSI of a device (stub implementation)
func GetDeviceRSSI(device unsafe.Pointer) int16 {
	return 0
}

// ============================================================================
// SERVICE AND CHARACTERISTIC FUNCTIONS
// ============================================================================

// AddService adds a service to the peripheral (stub implementation)
func AddService(peripheral, service unsafe.Pointer) error {
	return NewBluetoothError("add service failed")
}

// RemoveService removes a service from the peripheral (stub implementation)
func RemoveService(peripheral, service unsafe.Pointer) error {
	return NewBluetoothError("remove service failed")
}

// AddCharacteristic adds a characteristic to a service (stub implementation)
func AddCharacteristic(service, characteristic unsafe.Pointer) error {
	return NewBluetoothError("add characteristic failed")
}

// SendNotification sends a notification for a characteristic (stub implementation)
func SendNotification(characteristic unsafe.Pointer, data []byte) error {
	return NewBluetoothError("send notification failed")
}

// ============================================================================
// CALLBACK BRIDGE FUNCTIONS
// ============================================================================

// scanResultCallbackBridge is a callback bridge for scan results (stub implementation)
func scanResultCallbackBridge(userData unsafe.Pointer, identifier *byte, name *byte, rssi int) {
	// No-op for stub
}

// connectionCallbackBridge is a callback bridge for connections (stub implementation)
func connectionCallbackBridge(userData unsafe.Pointer, cDevice unsafe.Pointer) {
	// No-op for stub
}

// disconnectionCallbackBridge is a callback bridge for disconnections (stub implementation)
func disconnectionCallbackBridge(userData unsafe.Pointer, cDevice unsafe.Pointer) {
	// No-op for stub
}
