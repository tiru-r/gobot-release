//go:build darwin

// Package types contains Darwin-specific type definitions for Bluetooth implementation
package types

import (
	"sync"
	"unsafe"
)

// ============================================================================
// CORE TYPE DEFINITIONS
// ============================================================================

// DarwinManager implements Manager interface for macOS
type DarwinManager struct {
	CManager unsafe.Pointer
	Adapters []*DarwinAdapter
	Mu       sync.RWMutex
}

// DarwinAdapter implements Adapter interface for macOS
type DarwinAdapter struct {
	Manager    *DarwinManager
	Central    *DarwinCentral
	Peripheral *DarwinPeripheral
	Mu         sync.RWMutex
}

// DarwinCentral implements Central interface for macOS
type DarwinCentral struct {
	Adapter      *DarwinAdapter
	CCentral     unsafe.Pointer
	Scanning     bool
	Devices      map[string]*DarwinDevice
	ScanCallback func(Advertisement)
	Mu           sync.RWMutex
}

// DarwinPeripheral implements Peripheral interface for macOS
type DarwinPeripheral struct {
	Adapter     *DarwinAdapter
	CPeripheral unsafe.Pointer
	Advertising bool
	Services    map[string]*DarwinPeripheralService
	Mu          sync.RWMutex
}

// DarwinDevice implements Device interface for macOS
type DarwinDevice struct {
	Central    *DarwinCentral
	CDevice    unsafe.Pointer
	Identifier string
	Name       string
	Connected  bool
	Services   map[string]*DarwinService
	Mu         sync.RWMutex
}

// DarwinService implements Service interface for macOS
type DarwinService struct {
	Device          *DarwinDevice
	CService        unsafe.Pointer
	UUID            UUID
	Primary         bool
	Characteristics map[string]*DarwinCharacteristic
	Mu              sync.RWMutex
}

// DarwinCharacteristic implements Characteristic interface for macOS
type DarwinCharacteristic struct {
	Service         *DarwinService
	CCharacteristic unsafe.Pointer
	UUID            UUID
	Properties      CharacteristicProperty
	Descriptors     map[string]*DarwinDescriptor
	Subscribed      bool
	Mu              sync.RWMutex
}

// DarwinDescriptor implements Descriptor interface for macOS
type DarwinDescriptor struct {
	Characteristic *DarwinCharacteristic
	CDescriptor    unsafe.Pointer
	UUID           UUID
	Mu             sync.RWMutex
}

// DarwinPeripheralService implements PeripheralService interface for macOS
type DarwinPeripheralService struct {
	Peripheral      *DarwinPeripheral
	CService        unsafe.Pointer
	UUID            UUID
	Primary         bool
	Characteristics map[string]*DarwinPeripheralCharacteristic
	Mu              sync.RWMutex
}

// DarwinPeripheralCharacteristic implements PeripheralCharacteristic interface for macOS
type DarwinPeripheralCharacteristic struct {
	Service         *DarwinPeripheralService
	CCharacteristic unsafe.Pointer
	UUID            UUID
	Properties      CharacteristicProperty
	Value           []byte
	OnRead          func() []byte
	OnWrite         func([]byte) error
	OnSubscribe     func()
	OnUnsubscribe   func()
	Mu              sync.RWMutex
}

// ============================================================================
// INTERFACE PLACEHOLDER TYPES (to be imported from actual interfaces)
// ============================================================================

// These will be replaced with actual imports from bluetooth interfaces
type (
	Advertisement          interface{}
	UUID                   interface{}
	CharacteristicProperty interface{}
	Address                interface{}
)
