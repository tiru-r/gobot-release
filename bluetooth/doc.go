// Package bluetooth provides a comprehensive, cross-platform Bluetooth Low Energy (BLE) API for Go.
//
// This package supports multiple platforms including:
//   - Linux (using BlueZ via D-Bus)
//   - macOS (using Core Bluetooth framework)
//   - Windows (using WinRT APIs)
//   - Nordic SoftDevice (for bare metal applications)
//
// # Architecture
//
// The package follows a layered architecture:
//
//   1. Public API Layer: High-level interfaces (Manager, Adapter, Central, Peripheral)
//   2. Platform Abstraction Layer: Platform-specific implementations
//   3. Native Integration Layer: C bridges, D-Bus, WinRT bindings
//   4. Resource Management: Lifecycle management and cleanup
//
// # Basic Usage
//
// ## Simple Device Connection
//
//	ctx := context.Background()
//	manager, err := bluetooth.NewSimpleManager()
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer manager.Close()
//
//	// Connect to a device
//	device, err := manager.ConnectToDevice(ctx, "12:34:56:78:9A:BC")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer device.Disconnect(ctx)
//
//	// Read a characteristic
//	data, err := device.ReadCharacteristic(ctx, "180F", "2A19") // Battery level
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ## Device Scanning
//
//	scanner, err := bluetooth.NewExampleScanner()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	scanner.SetDeviceFoundCallback(func(device *bluetooth.ScannedDevice) {
//	    fmt.Printf("Found: %s (%s) RSSI: %d\n", device.Name, device.Address, device.RSSI)
//	})
//
//	err = scanner.StartScan(ctx, 30*time.Second)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// ## Advertising (Peripheral Mode)
//
//	server, err := bluetooth.NewExamplePeripheralServer("My Device", "180F", "2A19")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	err = server.Start(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	defer server.Stop(ctx)
//
// # Advanced Usage
//
// ## Low-Level API
//
//	manager, err := bluetooth.GetManager()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	adapter, err := manager.DefaultAdapter()
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	central := adapter.Central()
//	err = central.Enable(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Start scanning with custom parameters
//	params := bluetooth.ScanParams{
//	    Timeout:          30 * time.Second,
//	    Interval:         100 * time.Millisecond,
//	    Window:           50 * time.Millisecond,
//	    ActiveScan:       true,
//	    FilterDuplicates: true,
//	}
//
//	err = central.Scan(ctx, params, func(adv bluetooth.Advertisement) {
//	    fmt.Printf("Device: %s, RSSI: %d\n", adv.Address.String(), adv.RSSI)
//	})
//
// ## Connection Management
//
//	connParams := bluetooth.ConnectionParams{
//	    ConnectionTimeout:  10 * time.Second,
//	    MinInterval:        20 * time.Millisecond,
//	    MaxInterval:        40 * time.Millisecond,
//	    SlaveLatency:       0,
//	    SupervisionTimeout: 4 * time.Second,
//	}
//
//	device, err := central.Connect(ctx, address, connParams)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// # Error Handling
//
// The package provides comprehensive error handling with specific error types:
//
//	err := device.Connect(ctx)
//	if err != nil {
//	    var bluetoothErr *bluetooth.BluetoothError
//	    if errors.As(err, &bluetoothErr) {
//	        switch bluetoothErr.Code {
//	        case bluetooth.ErrorCodeConnectionTimeout:
//	            // Handle timeout
//	        case bluetooth.ErrorCodeDeviceNotFound:
//	            // Handle device not found
//	        default:
//	            // Handle other errors
//	        }
//	    }
//	}
//
// # Resource Management
//
// The package automatically manages resources such as connections, scanners, and advertisers:
//
//	// Initialize resource manager
//	ctx := context.Background()
//	err := bluetooth.InitializeResourceManager(ctx)
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Cleanup all resources on exit
//	defer bluetooth.CleanupAllResources(ctx)
//
// # Platform-Specific Considerations
//
// ## Linux (BlueZ)
//
// Requires BlueZ 5.50+ and appropriate D-Bus permissions. The application may need to run
// with elevated privileges or be granted specific D-Bus permissions.
//
//	// Example systemd service file snippet
//	[Service]
//	User=bluetooth
//	Group=bluetooth
//	Capabilities=CAP_NET_RAW,CAP_NET_ADMIN
//
// ## macOS (Core Bluetooth)
//
// Requires macOS 10.13+ and appropriate entitlements. The application will prompt for
// Bluetooth permissions on first use.
//
//	// Info.plist entitlement
//	<key>NSBluetoothAlwaysUsageDescription</key>
//	<string>This app uses Bluetooth to communicate with devices</string>
//
// ## Windows (WinRT)
//
// Requires Windows 10 version 1903+ and appropriate capabilities in the app manifest.
//
//	// Package.appxmanifest
//	<Capabilities>
//	    <DeviceCapability Name="bluetooth" />
//	</Capabilities>
//
// # Thread Safety
//
// All public APIs are thread-safe unless explicitly documented otherwise. Internal
// synchronization is handled using mutexes and channels.
//
// # Performance Considerations
//
// - Connection establishment can take 1-3 seconds
// - Scanning is CPU-intensive; use appropriate intervals
// - Multiple concurrent connections may impact performance
// - Resource cleanup is performed automatically but can be triggered manually
//
// # Best Practices
//
// 1. Always use context.Context for cancellation and timeouts
// 2. Validate parameters before API calls
// 3. Handle errors appropriately with retry logic where suitable
// 4. Use resource management for cleanup
// 5. Monitor connection state for robustness
// 6. Use appropriate scan parameters to balance power and performance
//
// # Debugging
//
// Enable debug logging by setting the environment variable:
//
//	export BLUETOOTH_DEBUG=1
//
// # Migration from tinygo.org/x/bluetooth
//
// The package provides compatibility helpers for migration:
//
//	// Old tinygo code
//	adaptor := bluetooth.DefaultAdapter
//
//	// New gobot code
//	adaptor := bluetooth.NewLegacyAdaptor("device-name")
//
// For comprehensive migration guidance, see the MIGRATION.md file.
package bluetooth