//go:build !linux && !darwin && !windows

package bluetooth

import "context"

// For backward compatibility, implement a simple stub version
type stubManager struct{}
type stubAdapter struct{}
type stubCentral struct{}

func (m *stubManager) DefaultAdapter() (Adapter, error) {
	return &stubAdapter{}, nil
}

func (m *stubManager) Adapters() ([]Adapter, error) {
	return []Adapter{&stubAdapter{}}, nil
}

func (m *stubManager) OnAdapterAdded(callback func(Adapter)) {}
func (m *stubManager) OnAdapterRemoved(callback func(Adapter)) {}

func (a *stubAdapter) Central() Central { return &stubCentral{} }
func (a *stubAdapter) Peripheral() Peripheral { return nil }
func (a *stubAdapter) Address() Address { return Address{} }
func (a *stubAdapter) Name() string { return "Stub Adapter" }
func (a *stubAdapter) SetName(name string) error { return nil }
func (a *stubAdapter) PowerState() bool { return true }
func (a *stubAdapter) SetPowerState(enabled bool) error { return nil }

func (c *stubCentral) Enable(ctx context.Context) error { return nil }
func (c *stubCentral) Disable(ctx context.Context) error { return nil }
func (c *stubCentral) Scan(ctx context.Context, params ScanParams, callback func(Advertisement)) error { return nil }
func (c *stubCentral) StopScan(ctx context.Context) error { return nil }
func (c *stubCentral) Connect(ctx context.Context, address Address, params ConnectionParams) (Device, error) { return nil, ErrNotSupported }
func (c *stubCentral) ConnectedDevices() []Device { return nil }

// getPlatformManager returns a stub implementation for unsupported platforms
func getPlatformManager() (Manager, error) {
	return &stubManager{}, nil
}