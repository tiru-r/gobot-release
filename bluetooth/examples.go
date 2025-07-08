package bluetooth

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// ConnectionManager provides automatic reconnection and robust error handling
type ConnectionManager struct {
	manager           *SimpleManager
	targetAddress     string
	device            *SimpleDevice
	reconnectAttempts int
	maxReconnects     int
	reconnectDelay    time.Duration
	connected         bool
	mu                sync.RWMutex
	onConnected       func(*SimpleDevice)
	onDisconnected    func()
	onError           func(error)
	stopCh            chan struct{}
}

// NewConnectionManager creates a new connection manager with automatic reconnection
func NewConnectionManager(targetAddress string) (*ConnectionManager, error) {
	manager, err := NewSimpleManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create Bluetooth manager: %w", err)
	}

	return &ConnectionManager{
		manager:        manager,
		targetAddress:  targetAddress,
		maxReconnects:  5,
		reconnectDelay: 2 * time.Second,
		stopCh:         make(chan struct{}),
	}, nil
}

// SetReconnectionParams configures reconnection behavior
func (cm *ConnectionManager) SetReconnectionParams(maxReconnects int, delay time.Duration) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.maxReconnects = maxReconnects
	cm.reconnectDelay = delay
}

// SetCallbacks sets event callbacks
func (cm *ConnectionManager) SetCallbacks(onConnected func(*SimpleDevice), onDisconnected func(), onError func(error)) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	cm.onConnected = onConnected
	cm.onDisconnected = onDisconnected
	cm.onError = onError
}

// Connect establishes connection with automatic retry
func (cm *ConnectionManager) Connect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.manager.IsBluetoothEnabled() {
		if err := cm.manager.EnableBluetooth(ctx); err != nil {
			return fmt.Errorf("failed to enable Bluetooth: %w", err)
		}
	}

	return cm.connectWithRetry(ctx)
}

func (cm *ConnectionManager) connectWithRetry(ctx context.Context) error {
	for attempt := 0; attempt <= cm.maxReconnects; attempt++ {
		device, err := cm.manager.ConnectToDevice(ctx, cm.targetAddress)
		if err == nil {
			cm.device = device
			cm.connected = true
			cm.reconnectAttempts = 0
			
			if cm.onConnected != nil {
				go cm.onConnected(device)
			}
			
			// Start monitoring connection
			go cm.monitorConnection(ctx)
			return nil
		}

		if attempt < cm.maxReconnects {
			if cm.onError != nil {
				go cm.onError(fmt.Errorf("connection attempt %d failed: %w", attempt+1, err))
			}
			
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(cm.reconnectDelay):
				continue
			}
		}
	}

	return fmt.Errorf("failed to connect after %d attempts", cm.maxReconnects+1)
}

func (cm *ConnectionManager) monitorConnection(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-cm.stopCh:
			return
		case <-ticker.C:
			cm.mu.RLock()
			device := cm.device
			cm.mu.RUnlock()

			if device == nil || !device.IsConnected() {
				cm.mu.Lock()
				if cm.connected {
					cm.connected = false
					if cm.onDisconnected != nil {
						go cm.onDisconnected()
					}
					
					// Attempt reconnection
					go func() {
						cm.mu.Lock()
						defer cm.mu.Unlock()
						cm.connectWithRetry(ctx)
					}()
				}
				cm.mu.Unlock()
			}
		}
	}
}

// Device returns the current device if connected
func (cm *ConnectionManager) Device() *SimpleDevice {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	if cm.connected {
		return cm.device
	}
	return nil
}

// IsConnected returns true if device is connected
func (cm *ConnectionManager) IsConnected() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.connected
}

// Disconnect closes the connection
func (cm *ConnectionManager) Disconnect(ctx context.Context) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	close(cm.stopCh)
	
	if cm.device != nil && cm.connected {
		err := cm.device.Disconnect(ctx)
		cm.connected = false
		cm.device = nil
		return err
	}
	return nil
}

// ExampleScanner demonstrates device scanning with filtering
type ExampleScanner struct {
	manager     *SimpleManager
	devices     map[string]*ScannedDevice
	nameFilter  string
	rssiFilter  int
	mu          sync.RWMutex
	onDeviceFound func(*ScannedDevice)
}

type ScannedDevice struct {
	Address   string
	Name      string
	RSSI      int
	FirstSeen time.Time
	LastSeen  time.Time
}

// NewExampleScanner creates a new scanner with filtering capabilities
func NewExampleScanner() (*ExampleScanner, error) {
	manager, err := NewSimpleManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create Bluetooth manager: %w", err)
	}

	return &ExampleScanner{
		manager: manager,
		devices: make(map[string]*ScannedDevice),
		rssiFilter: -100, // Accept all by default
	}, nil
}

// SetFilters configures device filtering
func (es *ExampleScanner) SetFilters(nameFilter string, rssiThreshold int) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.nameFilter = nameFilter
	es.rssiFilter = rssiThreshold
}

// SetDeviceFoundCallback sets callback for when devices are found
func (es *ExampleScanner) SetDeviceFoundCallback(callback func(*ScannedDevice)) {
	es.mu.Lock()
	defer es.mu.Unlock()
	es.onDeviceFound = callback
}

// StartScan begins scanning for devices
func (es *ExampleScanner) StartScan(ctx context.Context, duration time.Duration) error {
	if !es.manager.IsBluetoothEnabled() {
		if err := es.manager.EnableBluetooth(ctx); err != nil {
			return fmt.Errorf("failed to enable Bluetooth: %w", err)
		}
	}

	return es.manager.ScanForDevices(ctx, duration, es.handleDeviceFound)
}

func (es *ExampleScanner) handleDeviceFound(address, name string, rssi int) {
	es.mu.Lock()
	defer es.mu.Unlock()

	// Apply filters
	if es.nameFilter != "" && name != es.nameFilter {
		return
	}
	if rssi < es.rssiFilter {
		return
	}

	now := time.Now()
	
	if device, exists := es.devices[address]; exists {
		device.LastSeen = now
		device.RSSI = rssi
		if device.Name == "" && name != "" {
			device.Name = name
		}
	} else {
		device := &ScannedDevice{
			Address:   address,
			Name:      name,
			RSSI:      rssi,
			FirstSeen: now,
			LastSeen:  now,
		}
		es.devices[address] = device
		
		if es.onDeviceFound != nil {
			go es.onDeviceFound(device)
		}
	}
}

// GetDiscoveredDevices returns all discovered devices
func (es *ExampleScanner) GetDiscoveredDevices() []*ScannedDevice {
	es.mu.RLock()
	defer es.mu.RUnlock()

	devices := make([]*ScannedDevice, 0, len(es.devices))
	for _, device := range es.devices {
		devices = append(devices, device)
	}
	return devices
}

// ExamplePeripheralServer demonstrates creating a BLE peripheral server
type ExamplePeripheralServer struct {
	manager     *SimpleManager
	serviceName string
	serviceUUID string
	charUUID    string
	isRunning   bool
	mu          sync.RWMutex
}

// NewExamplePeripheralServer creates a new peripheral server
func NewExamplePeripheralServer(serviceName, serviceUUID, charUUID string) (*ExamplePeripheralServer, error) {
	manager, err := NewSimpleManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create Bluetooth manager: %w", err)
	}

	return &ExamplePeripheralServer{
		manager:     manager,
		serviceName: serviceName,
		serviceUUID: serviceUUID,
		charUUID:    charUUID,
	}, nil
}

// Start begins advertising and serving
func (eps *ExamplePeripheralServer) Start(ctx context.Context) error {
	eps.mu.Lock()
	defer eps.mu.Unlock()

	if eps.isRunning {
		return fmt.Errorf("server is already running")
	}

	if !eps.manager.IsBluetoothEnabled() {
		if err := eps.manager.EnableBluetooth(ctx); err != nil {
			return fmt.Errorf("failed to enable Bluetooth: %w", err)
		}
	}

	// Start advertising
	err := eps.manager.StartAdvertising(ctx, eps.serviceName, []string{eps.serviceUUID})
	if err != nil {
		return fmt.Errorf("failed to start advertising: %w", err)
	}

	eps.isRunning = true
	return nil
}

// Stop stops the peripheral server
func (eps *ExamplePeripheralServer) Stop(ctx context.Context) error {
	eps.mu.Lock()
	defer eps.mu.Unlock()

	if !eps.isRunning {
		return nil
	}

	err := eps.manager.StopAdvertising(ctx)
	eps.isRunning = false
	return err
}

// IsRunning returns true if the server is running
func (eps *ExamplePeripheralServer) IsRunning() bool {
	eps.mu.RLock()
	defer eps.mu.RUnlock()
	return eps.isRunning
}

// ExampleHeartRateMonitor demonstrates connecting to and reading from a heart rate monitor
func ExampleHeartRateMonitor() {
	ctx := context.Background()
	
	// Create connection manager
	cm, err := NewConnectionManager("12:34:56:78:9A:BC") // Replace with actual device address
	if err != nil {
		log.Fatal("Failed to create connection manager:", err)
	}
	defer cm.Disconnect(ctx)

	// Set up callbacks
	cm.SetCallbacks(
		func(device *SimpleDevice) {
			log.Printf("Connected to heart rate monitor: %s", device.Name())
			
			// Discover services
			if err := device.DiscoverServices(ctx); err != nil {
				log.Printf("Failed to discover services: %v", err)
				return
			}

			// Subscribe to heart rate measurements
			err := device.SubscribeToCharacteristic(ctx, "180D", "2A37", func(data []byte) {
				if len(data) >= 2 {
					heartRate := int(data[1])
					log.Printf("Heart rate: %d BPM", heartRate)
				}
			})
			if err != nil {
				log.Printf("Failed to subscribe to heart rate: %v", err)
			}
		},
		func() {
			log.Println("Disconnected from heart rate monitor")
		},
		func(err error) {
			log.Printf("Connection error: %v", err)
		},
	)

	// Connect with retry
	if err := cm.Connect(ctx); err != nil {
		log.Fatal("Failed to connect:", err)
	}

	// Keep running
	time.Sleep(60 * time.Second)
}

// ExampleDeviceScanner demonstrates device scanning with filtering
func ExampleDeviceScanner() {
	ctx := context.Background()
	
	scanner, err := NewExampleScanner()
	if err != nil {
		log.Fatal("Failed to create scanner:", err)
	}

	// Filter for devices with strong signal
	scanner.SetFilters("", -60) // Only devices with RSSI > -60 dBm

	// Set up device found callback
	scanner.SetDeviceFoundCallback(func(device *ScannedDevice) {
		log.Printf("Found device: %s (%s) RSSI: %d dBm", 
			device.Name, device.Address, device.RSSI)
	})

	// Scan for 30 seconds
	if err := scanner.StartScan(ctx, 30*time.Second); err != nil {
		log.Fatal("Failed to start scan:", err)
	}

	// Print discovered devices
	devices := scanner.GetDiscoveredDevices()
	log.Printf("Discovered %d devices", len(devices))
	for _, device := range devices {
		log.Printf("  - %s (%s) RSSI: %d dBm", 
			device.Name, device.Address, device.RSSI)
	}
}

// ExamplePeripheralService demonstrates creating a BLE peripheral service
func ExamplePeripheralService() {
	ctx := context.Background()
	
	server, err := NewExamplePeripheralServer(
		"Gobot Device",
		"12345678-1234-1234-1234-123456789ABC",
		"12345678-1234-1234-1234-123456789ABD",
	)
	if err != nil {
		log.Fatal("Failed to create peripheral server:", err)
	}

	// Start the server
	if err := server.Start(ctx); err != nil {
		log.Fatal("Failed to start server:", err)
	}
	defer server.Stop(ctx)

	log.Println("BLE peripheral server is running...")
	log.Println("Advertising as 'Gobot Device'")
	
	// Keep running
	time.Sleep(60 * time.Second)
}