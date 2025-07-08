//go:build darwin

// Package managers provides Darwin-specific Bluetooth manager implementations
package managers

import (
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"gobot.io/x/gobot/v2/bluetooth/internal/cbridge"
	"gobot.io/x/gobot/v2/bluetooth/internal/types"
)

// ============================================================================
// MANAGER IMPLEMENTATION
// ============================================================================

// DarwinManager manages Bluetooth adapters on macOS
type DarwinManager struct {
	*types.DarwinManager
}

// NewDarwinManager creates a new Darwin Bluetooth manager
func NewDarwinManager() (*DarwinManager, error) {
	manager := &DarwinManager{
		DarwinManager: &types.DarwinManager{
			Adapters: make([]*types.DarwinAdapter, 0),
		},
	}

	// Create Core Bluetooth manager
	cManager := cbridge.NewCBTManager(unsafe.Pointer(manager))
	if cManager == nil {
		return nil, fmt.Errorf("failed to create Core Bluetooth manager")
	}

	manager.CManager = cManager
	runtime.SetFinalizer(manager, (*DarwinManager).finalize)

	// Initialize default adapter
	if err := manager.initializeDefaultAdapter(); err != nil {
		return nil, fmt.Errorf("failed to initialize default adapter: %w", err)
	}

	return manager, nil
}

// initializeDefaultAdapter sets up the default Bluetooth adapter
func (m *DarwinManager) initializeDefaultAdapter() error {
	// Get default adapter from Core Bluetooth
	cAdapter, err := cbridge.GetDefaultAdapter(m.CManager)
	if err != nil {
		return fmt.Errorf("failed to get default adapter: %w", err)
	}

	adapter := &types.DarwinAdapter{
		Manager: m.DarwinManager,
	}

	// Initialize central manager
	adapter.Central = &types.DarwinCentral{
		Adapter: adapter,
		Devices: make(map[string]*types.DarwinDevice),
	}
	adapter.Central.CCentral = cbridge.NewCBTCentral(cAdapter)

	// Initialize peripheral manager
	adapter.Peripheral = &types.DarwinPeripheral{
		Adapter:  adapter,
		Services: make(map[string]*types.DarwinPeripheralService),
	}
	adapter.Peripheral.CPeripheral = cbridge.NewCBTPeripheral(cAdapter)

	m.Mu.Lock()
	m.Adapters = append(m.Adapters, adapter)
	m.Mu.Unlock()

	return nil
}

// finalize cleans up the manager resources
func (m *DarwinManager) finalize() {
	if m.CManager != nil {
		cbridge.FreeCBTManager(m.CManager)
		m.CManager = nil
	}
}

// ============================================================================
// MANAGER INTERFACE IMPLEMENTATION
// ============================================================================

// DefaultAdapter returns the default Bluetooth adapter
func (m *DarwinManager) DefaultAdapter() (Adapter, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	if len(m.Adapters) == 0 {
		return nil, fmt.Errorf("no adapters available")
	}

	return NewDarwinAdapter(m.Adapters[0]), nil
}

// Adapters returns all available Bluetooth adapters
func (m *DarwinManager) Adapters() ([]Adapter, error) {
	m.Mu.RLock()
	defer m.Mu.RUnlock()

	adapters := make([]Adapter, len(m.Adapters))
	for i, adapter := range m.Adapters {
		adapters[i] = NewDarwinAdapter(adapter)
	}

	return adapters, nil
}

// OnAdapterAdded sets up callback for adapter addition events
func (m *DarwinManager) OnAdapterAdded(callback func(Adapter)) {
	// macOS typically has one adapter, so this is mostly a no-op
	// In a real implementation, this could monitor for external adapters
}

// OnAdapterRemoved sets up callback for adapter removal events
func (m *DarwinManager) OnAdapterRemoved(callback func(Adapter)) {
	// macOS typically has one adapter, so this is mostly a no-op
	// In a real implementation, this could monitor for external adapters
}

// ============================================================================
// PLATFORM MANAGER FACTORY
// ============================================================================

// GetPlatformManager returns the macOS implementation of Manager
func GetPlatformManager() (Manager, error) {
	return NewDarwinManager()
}

// ============================================================================
// INTERFACE PLACEHOLDERS
// ============================================================================

// These interfaces will be imported from the actual bluetooth package
type (
	Manager interface {
		DefaultAdapter() (Adapter, error)
		Adapters() ([]Adapter, error)
		OnAdapterAdded(callback func(Adapter))
		OnAdapterRemoved(callback func(Adapter))
	}

	Adapter interface {
		// Adapter methods will be defined here
	}
)
