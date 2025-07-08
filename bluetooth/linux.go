//go:build linux

package bluetooth

import (
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/godbus/dbus/v5"
	"github.com/godbus/dbus/v5/prop"
	"github.com/google/uuid"
)

const (
	bluezService      = "org.bluez"
	bluezObjectPath   = "/org/bluez"
	adapterInterface  = "org.bluez.Adapter1"
	deviceInterface   = "org.bluez.Device1"
	gattServiceInterface = "org.bluez.GattService1"
	gattCharInterface = "org.bluez.GattCharacteristic1"
	gattDescInterface = "org.bluez.GattDescriptor1"
	agentManagerInterface = "org.bluez.AgentManager1"
	profileManagerInterface = "org.bluez.ProfileManager1"
)

// linuxManager implements Manager for Linux using BlueZ
type linuxManager struct {
	conn    *dbus.Conn
	mu      sync.RWMutex
	adapters map[dbus.ObjectPath]*linuxAdapter
}

// linuxAdapter implements Adapter for Linux
type linuxAdapter struct {
	manager    *linuxManager
	path       dbus.ObjectPath
	properties map[string]dbus.Variant
	central    *linuxCentral
	peripheral *linuxPeripheral
	mu         sync.RWMutex
}

// linuxCentral implements Central for Linux
type linuxCentral struct {
	adapter   *linuxAdapter
	scanning  bool
	devices   map[dbus.ObjectPath]*linuxDevice
	mu        sync.RWMutex
}

// linuxPeripheral implements Peripheral for Linux
type linuxPeripheral struct {
	adapter      *linuxAdapter
	advertising  bool
	services     map[dbus.ObjectPath]*linuxPeripheralService
	mu           sync.RWMutex
}

// linuxDevice implements Device for Linux
type linuxDevice struct {
	central    *linuxCentral
	path       dbus.ObjectPath
	properties map[string]dbus.Variant
	services   map[dbus.ObjectPath]*linuxService
	connected  bool
	mu         sync.RWMutex
}

// linuxService implements Service for Linux
type linuxService struct {
	device         *linuxDevice
	path           dbus.ObjectPath
	uuid           UUID
	properties     map[string]dbus.Variant
	characteristics map[dbus.ObjectPath]*linuxCharacteristic
	mu             sync.RWMutex
}

// linuxCharacteristic implements Characteristic for Linux
type linuxCharacteristic struct {
	service     *linuxService
	path        dbus.ObjectPath
	uuid        UUID
	properties  map[string]dbus.Variant
	descriptors map[dbus.ObjectPath]*linuxDescriptor
	subscribed  bool
	mu          sync.RWMutex
}

// linuxDescriptor implements Descriptor for Linux
type linuxDescriptor struct {
	characteristic *linuxCharacteristic
	path           dbus.ObjectPath
	properties     map[string]dbus.Variant
	mu             sync.RWMutex
}

// linuxPeripheralService implements PeripheralService for Linux
type linuxPeripheralService struct {
	peripheral      *linuxPeripheral
	path            dbus.ObjectPath
	uuid            UUID
	primary         bool
	characteristics map[dbus.ObjectPath]*linuxPeripheralCharacteristic
	mu              sync.RWMutex
}

// linuxPeripheralCharacteristic implements PeripheralCharacteristic for Linux
type linuxPeripheralCharacteristic struct {
	service    *linuxPeripheralService
	path       dbus.ObjectPath
	uuid       UUID
	properties CharacteristicProperty
	value      []byte
	onRead     func() []byte
	onWrite    func([]byte) error
	onSubscribe func()
	onUnsubscribe func()
	mu         sync.RWMutex
}

// getPlatformManager returns the Linux implementation of Manager
func getPlatformManager() (Manager, error) {
	conn, err := dbus.SystemBus()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to system bus: %w", err)
	}

	manager := &linuxManager{
		conn:     conn,
		adapters: make(map[dbus.ObjectPath]*linuxAdapter),
	}

	if err := manager.discoverAdapters(); err != nil {
		return nil, fmt.Errorf("failed to discover adapters: %w", err)
	}

	return manager, nil
}

func (m *linuxManager) discoverAdapters() error {
	obj := m.conn.Object(bluezService, bluezObjectPath)
	
	var objects map[dbus.ObjectPath]map[string]map[string]dbus.Variant
	err := obj.Call("org.freedesktop.DBus.ObjectManager.GetManagedObjects", 0).Store(&objects)
	if err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	for path, interfaces := range objects {
		if props, ok := interfaces[adapterInterface]; ok {
			adapter := &linuxAdapter{
				manager:    m,
				path:       path,
				properties: props,
			}
			adapter.central = &linuxCentral{
				adapter: adapter,
				devices: make(map[dbus.ObjectPath]*linuxDevice),
			}
			adapter.peripheral = &linuxPeripheral{
				adapter:  adapter,
				services: make(map[dbus.ObjectPath]*linuxPeripheralService),
			}
			m.adapters[path] = adapter
		}
	}

	return nil
}

func (m *linuxManager) DefaultAdapter() (Adapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, adapter := range m.adapters {
		return adapter, nil
	}

	return nil, fmt.Errorf("no Bluetooth adapters found")
}

func (m *linuxManager) Adapters() ([]Adapter, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	adapters := make([]Adapter, 0, len(m.adapters))
	for _, adapter := range m.adapters {
		adapters = append(adapters, adapter)
	}

	return adapters, nil
}

func (m *linuxManager) OnAdapterAdded(callback func(Adapter)) {
	// TODO: Implement D-Bus signal monitoring for adapter addition
}

func (m *linuxManager) OnAdapterRemoved(callback func(Adapter)) {
	// TODO: Implement D-Bus signal monitoring for adapter removal
}

// linuxAdapter implementation
func (a *linuxAdapter) Central() Central {
	return a.central
}

func (a *linuxAdapter) Peripheral() Peripheral {
	return a.peripheral
}

func (a *linuxAdapter) Address() Address {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if addr, ok := a.properties["Address"]; ok {
		if addrStr, ok := addr.Value().(string); ok {
			return parseAddress(addrStr)
		}
	}
	return Address{}
}

func (a *linuxAdapter) Name() string {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if name, ok := a.properties["Name"]; ok {
		if nameStr, ok := name.Value().(string); ok {
			return nameStr
		}
	}
	return ""
}

func (a *linuxAdapter) SetName(name string) error {
	propsIface, _ := prop.Export(a.manager.conn, a.path, prop.Map{})
	
	err := propsIface.Set(adapterInterface, "Name", dbus.MakeVariant(name))
	if err != nil {
		return fmt.Errorf("failed to set adapter name: %w", err)
	}

	a.mu.Lock()
	a.properties["Name"] = dbus.MakeVariant(name)
	a.mu.Unlock()

	return nil
}

func (a *linuxAdapter) PowerState() bool {
	a.mu.RLock()
	defer a.mu.RUnlock()

	if powered, ok := a.properties["Powered"]; ok {
		if poweredBool, ok := powered.Value().(bool); ok {
			return poweredBool
		}
	}
	return false
}

func (a *linuxAdapter) SetPowerState(enabled bool) error {
	propsIface, _ := prop.Export(a.manager.conn, a.path, prop.Map{})
	
	err := propsIface.Set(adapterInterface, "Powered", dbus.MakeVariant(enabled))
	if err != nil {
		return fmt.Errorf("failed to set power state: %w", err)
	}

	a.mu.Lock()
	a.properties["Powered"] = dbus.MakeVariant(enabled)
	a.mu.Unlock()

	return nil
}

// linuxCentral implementation
func (c *linuxCentral) Enable(ctx context.Context) error {
	return c.adapter.SetPowerState(true)
}

func (c *linuxCentral) Disable(ctx context.Context) error {
	return c.adapter.SetPowerState(false)
}

func (c *linuxCentral) Scan(ctx context.Context, params ScanParams, callback func(Advertisement)) error {
	c.mu.Lock()
	if c.scanning {
		c.mu.Unlock()
		return fmt.Errorf("scan already in progress")
	}
	c.scanning = true
	c.mu.Unlock()

	defer func() {
		c.mu.Lock()
		c.scanning = false
		c.mu.Unlock()
	}()

	obj := c.adapter.manager.conn.Object(bluezService, c.adapter.path)

	// Set scan filter properties if needed
	propsIface, _ := prop.Export(c.adapter.manager.conn, c.adapter.path, prop.Map{})
	if params.FilterDuplicates {
		propsIface.Set(adapterInterface, "DuplicateData", dbus.MakeVariant(false))
	}

	// Start discovery
	err := obj.Call(adapterInterface+".StartDiscovery", 0).Err
	if err != nil {
		return fmt.Errorf("failed to start discovery: %w", err)
	}

	// Set up signal monitoring for device discovery
	go c.monitorDeviceDiscovery(ctx, callback)

	// Wait for timeout or context cancellation
	timer := time.NewTimer(params.Timeout)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return c.StopScan(ctx)
	}
}

func (c *linuxCentral) monitorDeviceDiscovery(ctx context.Context, callback func(Advertisement)) {
	// TODO: Implement D-Bus signal monitoring for device discovery
	// This would listen for PropertiesChanged signals on Device interfaces
	// and call the callback with Advertisement data
}

func (c *linuxCentral) StopScan(ctx context.Context) error {
	obj := c.adapter.manager.conn.Object(bluezService, c.adapter.path)
	
	err := obj.Call(adapterInterface+".StopDiscovery", 0).Err
	if err != nil {
		return fmt.Errorf("failed to stop discovery: %w", err)
	}

	c.mu.Lock()
	c.scanning = false
	c.mu.Unlock()

	return nil
}

func (c *linuxCentral) Connect(ctx context.Context, address Address, params ConnectionParams) (Device, error) {
	// Find device by address
	devicePath := dbus.ObjectPath(fmt.Sprintf("%s/dev_%s", c.adapter.path, strings.ReplaceAll(address.String(), ":", "_")))
	
	obj := c.adapter.manager.conn.Object(bluezService, devicePath)
	
	// Connect to device
	call := obj.CallWithContext(ctx, deviceInterface+".Connect", 0)
	if call.Err != nil {
		return nil, fmt.Errorf("failed to connect to device: %w", call.Err)
	}

	// Create device object
	device := &linuxDevice{
		central:   c,
		path:      devicePath,
		services:  make(map[dbus.ObjectPath]*linuxService),
		connected: true,
	}

	c.mu.Lock()
	c.devices[devicePath] = device
	c.mu.Unlock()

	return device, nil
}

func (c *linuxCentral) ConnectedDevices() []Device {
	c.mu.RLock()
	defer c.mu.RUnlock()

	devices := make([]Device, 0, len(c.devices))
	for _, device := range c.devices {
		if device.connected {
			devices = append(devices, device)
		}
	}

	return devices
}

// Helper functions
func parseAddress(addrStr string) Address {
	parts := strings.Split(addrStr, ":")
	if len(parts) != 6 {
		return Address{}
	}

	var addr Address
	for i, part := range parts {
		if b, err := hex.DecodeString(part); err == nil && len(b) == 1 {
			addr.MAC[5-i] = b[0]
		}
	}

	return addr
}

// linuxDevice implementation
func (d *linuxDevice) Address() Address {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if addr, ok := d.properties["Address"]; ok {
		if addrStr, ok := addr.Value().(string); ok {
			return parseAddress(addrStr)
		}
	}
	return Address{}
}

func (d *linuxDevice) Name() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if name, ok := d.properties["Name"]; ok {
		if nameStr, ok := name.Value().(string); ok {
			return nameStr
		}
	}
	return ""
}

func (d *linuxDevice) RSSI() int16 {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if rssi, ok := d.properties["RSSI"]; ok {
		if rssiInt, ok := rssi.Value().(int16); ok {
			return rssiInt
		}
	}
	return 0
}

func (d *linuxDevice) Connected() bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.connected
}

func (d *linuxDevice) Disconnect(ctx context.Context) error {
	obj := d.central.adapter.manager.conn.Object(bluezService, d.path)
	
	call := obj.CallWithContext(ctx, deviceInterface+".Disconnect", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to disconnect device: %w", call.Err)
	}

	d.mu.Lock()
	d.connected = false
	d.mu.Unlock()

	return nil
}

func (d *linuxDevice) Services() []Service {
	d.mu.RLock()
	defer d.mu.RUnlock()

	services := make([]Service, 0, len(d.services))
	for _, service := range d.services {
		services = append(services, service)
	}
	return services
}

func (d *linuxDevice) GetService(uuid UUID) (Service, error) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, service := range d.services {
		if service.uuid.String() == uuid.String() {
			return service, nil
		}
	}
	return nil, ErrServiceNotFound
}

func (d *linuxDevice) DiscoverServices(ctx context.Context, uuids []UUID) error {
	// Discover GATT services via D-Bus
	obj := d.central.adapter.manager.conn.Object(bluezService, d.path)
	
	call := obj.CallWithContext(ctx, deviceInterface+".Connect", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to connect for service discovery: %w", call.Err)
	}

	d.mu.Lock()
	d.connected = true
	d.mu.Unlock()

	return nil
}

func (d *linuxDevice) RequestMTU(ctx context.Context, mtu uint16) error {
	return ErrNotSupported
}

func (d *linuxDevice) GetMTU() uint16 {
	return 23 // Default ATT MTU
}

// linuxService implementation
func (s *linuxService) UUID() UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if s.uuid.UUID != (uuid.UUID{}) {
		return s.uuid
	}

	if uuid, ok := s.properties["UUID"]; ok {
		if uuidStr, ok := uuid.Value().(string); ok {
			if u, err := NewUUID(uuidStr); err == nil {
				s.uuid = u
				return u
			}
		}
	}
	return UUID{}
}

func (s *linuxService) Primary() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if primary, ok := s.properties["Primary"]; ok {
		if primaryBool, ok := primary.Value().(bool); ok {
			return primaryBool
		}
	}
	return false
}

func (s *linuxService) Characteristics() []Characteristic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	characteristics := make([]Characteristic, 0, len(s.characteristics))
	for _, char := range s.characteristics {
		characteristics = append(characteristics, char)
	}
	return characteristics
}

func (s *linuxService) GetCharacteristic(uuid UUID) (Characteristic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, char := range s.characteristics {
		if char.uuid.String() == uuid.String() {
			return char, nil
		}
	}
	return nil, ErrCharacteristicNotFound
}

// linuxCharacteristic implementation
func (c *linuxCharacteristic) UUID() UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.uuid
}

func (c *linuxCharacteristic) Properties() CharacteristicProperty {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if flags, ok := c.properties["Flags"]; ok {
		if flagsSlice, ok := flags.Value().([]string); ok {
			var props CharacteristicProperty
			for _, flag := range flagsSlice {
				switch flag {
				case "read":
					props |= CharacteristicRead
				case "write":
					props |= CharacteristicWrite
				case "write-without-response":
					props |= CharacteristicWriteWithoutResponse
				case "notify":
					props |= CharacteristicNotify
				case "indicate":
					props |= CharacteristicIndicate
				}
			}
			return props
		}
	}
	return 0
}

func (c *linuxCharacteristic) Read(ctx context.Context) ([]byte, error) {
	obj := c.service.device.central.adapter.manager.conn.Object(bluezService, c.path)
	
	var value []byte
	call := obj.CallWithContext(ctx, gattCharInterface+".ReadValue", 0, map[string]dbus.Variant{})
	if call.Err != nil {
		return nil, fmt.Errorf("failed to read characteristic: %w", call.Err)
	}
	
	err := call.Store(&value)
	if err != nil {
		return nil, fmt.Errorf("failed to store read value: %w", err)
	}
	
	return value, nil
}

func (c *linuxCharacteristic) Write(ctx context.Context, data []byte) error {
	obj := c.service.device.central.adapter.manager.conn.Object(bluezService, c.path)
	
	call := obj.CallWithContext(ctx, gattCharInterface+".WriteValue", 0, data, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("failed to write characteristic: %w", call.Err)
	}
	
	return nil
}

func (c *linuxCharacteristic) WriteWithoutResponse(ctx context.Context, data []byte) error {
	obj := c.service.device.central.adapter.manager.conn.Object(bluezService, c.path)
	
	options := map[string]dbus.Variant{
		"type": dbus.MakeVariant("command"),
	}
	
	call := obj.CallWithContext(ctx, gattCharInterface+".WriteValue", 0, data, options)
	if call.Err != nil {
		return fmt.Errorf("failed to write characteristic without response: %w", call.Err)
	}
	
	return nil
}

func (c *linuxCharacteristic) Subscribe(ctx context.Context, callback func([]byte)) error {
	obj := c.service.device.central.adapter.manager.conn.Object(bluezService, c.path)
	
	call := obj.CallWithContext(ctx, gattCharInterface+".StartNotify", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to start notifications: %w", call.Err)
	}
	
	c.mu.Lock()
	c.subscribed = true
	c.mu.Unlock()
	
	return nil
}

func (c *linuxCharacteristic) Unsubscribe(ctx context.Context) error {
	obj := c.service.device.central.adapter.manager.conn.Object(bluezService, c.path)
	
	call := obj.CallWithContext(ctx, gattCharInterface+".StopNotify", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to stop notifications: %w", call.Err)
	}
	
	c.mu.Lock()
	c.subscribed = false
	c.mu.Unlock()
	
	return nil
}

func (c *linuxCharacteristic) Descriptors() []Descriptor {
	c.mu.RLock()
	defer c.mu.RUnlock()

	descriptors := make([]Descriptor, 0, len(c.descriptors))
	for _, desc := range c.descriptors {
		descriptors = append(descriptors, desc)
	}
	return descriptors
}

// linuxDescriptor implementation
func (d *linuxDescriptor) UUID() UUID {
	d.mu.RLock()
	defer d.mu.RUnlock()

	if uuid, ok := d.properties["UUID"]; ok {
		if uuidStr, ok := uuid.Value().(string); ok {
			if u, err := NewUUID(uuidStr); err == nil {
				return u
			}
		}
	}
	return UUID{}
}

func (d *linuxDescriptor) Read(ctx context.Context) ([]byte, error) {
	obj := d.characteristic.service.device.central.adapter.manager.conn.Object(bluezService, d.path)
	
	var value []byte
	call := obj.CallWithContext(ctx, gattDescInterface+".ReadValue", 0, map[string]dbus.Variant{})
	if call.Err != nil {
		return nil, fmt.Errorf("failed to read descriptor: %w", call.Err)
	}
	
	err := call.Store(&value)
	if err != nil {
		return nil, fmt.Errorf("failed to store descriptor value: %w", err)
	}
	
	return value, nil
}

func (d *linuxDescriptor) Write(ctx context.Context, data []byte) error {
	obj := d.characteristic.service.device.central.adapter.manager.conn.Object(bluezService, d.path)
	
	call := obj.CallWithContext(ctx, gattDescInterface+".WriteValue", 0, data, map[string]dbus.Variant{})
	if call.Err != nil {
		return fmt.Errorf("failed to write descriptor: %w", call.Err)
	}
	
	return nil
}

// linuxPeripheral implementation
func (p *linuxPeripheral) Enable(ctx context.Context) error {
	return p.adapter.SetPowerState(true)
}

func (p *linuxPeripheral) Disable(ctx context.Context) error {
	return p.adapter.SetPowerState(false)
}

func (p *linuxPeripheral) AddService(uuid UUID, primary bool) (PeripheralService, error) {
	// This would require registering a GATT service with BlueZ
	service := &linuxPeripheralService{
		peripheral:      p,
		uuid:            uuid,
		primary:         primary,
		characteristics: make(map[dbus.ObjectPath]*linuxPeripheralCharacteristic),
	}
	
	servicePath := dbus.ObjectPath(fmt.Sprintf("/org/gobot/service_%s", strings.ReplaceAll(uuid.String(), "-", "_")))
	service.path = servicePath
	
	p.mu.Lock()
	p.services[servicePath] = service
	p.mu.Unlock()
	
	return service, nil
}

func (p *linuxPeripheral) GetService(uuid UUID) (PeripheralService, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	for _, service := range p.services {
		if service.uuid.String() == uuid.String() {
			return service, nil
		}
	}
	return nil, ErrServiceNotFound
}

func (p *linuxPeripheral) Services() []PeripheralService {
	p.mu.RLock()
	defer p.mu.RUnlock()

	services := make([]PeripheralService, 0, len(p.services))
	for _, service := range p.services {
		services = append(services, service)
	}
	return services
}

func (p *linuxPeripheral) StartAdvertising(ctx context.Context, params AdvertisingParams, data AdvertisementData) error {
	obj := p.adapter.manager.conn.Object(bluezService, p.adapter.path)
	
	// Set discoverable and pairable
	propsIface, _ := prop.Export(p.adapter.manager.conn, p.adapter.path, prop.Map{})
	propsIface.Set(adapterInterface, "Discoverable", dbus.MakeVariant(params.Discoverable))
	propsIface.Set(adapterInterface, "Pairable", dbus.MakeVariant(true))
	
	if data.LocalName != "" {
		propsIface.Set(adapterInterface, "Alias", dbus.MakeVariant(data.LocalName))
	}
	
	call := obj.CallWithContext(ctx, adapterInterface+".StartDiscovery", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to start advertising: %w", call.Err)
	}
	
	p.mu.Lock()
	p.advertising = true
	p.mu.Unlock()
	
	return nil
}

func (p *linuxPeripheral) StopAdvertising(ctx context.Context) error {
	obj := p.adapter.manager.conn.Object(bluezService, p.adapter.path)
	
	call := obj.CallWithContext(ctx, adapterInterface+".StopDiscovery", 0)
	if call.Err != nil {
		return fmt.Errorf("failed to stop advertising: %w", call.Err)
	}
	
	p.mu.Lock()
	p.advertising = false
	p.mu.Unlock()
	
	return nil
}

func (p *linuxPeripheral) IsAdvertising() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.advertising
}

func (p *linuxPeripheral) OnConnect(callback func(Device)) {
	// TODO: Implement D-Bus signal monitoring for device connections
}

func (p *linuxPeripheral) OnDisconnect(callback func(Device)) {
	// TODO: Implement D-Bus signal monitoring for device disconnections
}

// linuxPeripheralService implementation
func (s *linuxPeripheralService) UUID() UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.uuid
}

func (s *linuxPeripheralService) Primary() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.primary
}

func (s *linuxPeripheralService) AddCharacteristic(uuid UUID, properties CharacteristicProperty, value []byte) (PeripheralCharacteristic, error) {
	char := &linuxPeripheralCharacteristic{
		service:    s,
		uuid:       uuid,
		properties: properties,
		value:      make([]byte, len(value)),
	}
	copy(char.value, value)
	
	charPath := dbus.ObjectPath(fmt.Sprintf("%s/char_%s", s.path, strings.ReplaceAll(uuid.String(), "-", "_")))
	char.path = charPath
	
	s.mu.Lock()
	s.characteristics[charPath] = char
	s.mu.Unlock()
	
	return char, nil
}

func (s *linuxPeripheralService) GetCharacteristic(uuid UUID) (PeripheralCharacteristic, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, char := range s.characteristics {
		if char.uuid.String() == uuid.String() {
			return char, nil
		}
	}
	return nil, ErrCharacteristicNotFound
}

func (s *linuxPeripheralService) Characteristics() []PeripheralCharacteristic {
	s.mu.RLock()
	defer s.mu.RUnlock()

	characteristics := make([]PeripheralCharacteristic, 0, len(s.characteristics))
	for _, char := range s.characteristics {
		characteristics = append(characteristics, char)
	}
	return characteristics
}

// linuxPeripheralCharacteristic implementation
func (c *linuxPeripheralCharacteristic) UUID() UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.uuid
}

func (c *linuxPeripheralCharacteristic) Properties() CharacteristicProperty {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.properties
}

func (c *linuxPeripheralCharacteristic) Value() []byte {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	value := make([]byte, len(c.value))
	copy(value, c.value)
	return value
}

func (c *linuxPeripheralCharacteristic) SetValue(data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	
	c.value = make([]byte, len(data))
	copy(c.value, data)
	return nil
}

func (c *linuxPeripheralCharacteristic) NotifySubscribers(data []byte) error {
	// This would send notifications to subscribed central devices
	c.SetValue(data)
	return nil
}

func (c *linuxPeripheralCharacteristic) OnRead(callback func() []byte) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onRead = callback
}

func (c *linuxPeripheralCharacteristic) OnWrite(callback func([]byte) error) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onWrite = callback
}

func (c *linuxPeripheralCharacteristic) OnSubscribe(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onSubscribe = callback
}

func (c *linuxPeripheralCharacteristic) OnUnsubscribe(callback func()) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.onUnsubscribe = callback
}