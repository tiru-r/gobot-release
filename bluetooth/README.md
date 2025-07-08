# Cross-Platform Bluetooth API for Go 1.24

A pure Go Bluetooth Low Energy (BLE) API that supports Linux, macOS, Windows, and bare metal (Nordic SoftDevice) platforms.

## Features

- **Cross-platform**: Works on Linux (BlueZ), macOS (Core Bluetooth), Windows (WinRT), and Nordic SoftDevice
- **Pure Go**: No external dependencies except platform-specific system libraries
- **Modern Go**: Built for Go 1.24 with modern language features
- **Both Central and Peripheral**: Support for both client and server roles
- **Simple API**: High-level wrapper for common operations
- **Thread-safe**: All operations are safe for concurrent use

## Platform Support

| Platform | Central (Client) | Peripheral (Server) | Notes |
|----------|------------------|-------------------|-------|
| Linux | ✅ BlueZ D-Bus | ✅ BlueZ D-Bus | Requires BlueZ 5.x |
| macOS | ✅ Core Bluetooth | ✅ Core Bluetooth | macOS 10.13+ |
| Windows | ✅ WinRT | ✅ WinRT | Windows 10 1803+ |
| Nordic SoftDevice | ✅ Native C API | ✅ Native C API | nRF52 series |

## Installation

```bash
go get gobot.io/x/gobot/v2/bluetooth
```

## Quick Start

### Simple Device Scanner

```go
package main

import (
    "context"
    "fmt"
    "log"
    "time"
    
    "gobot.io/x/gobot/v2/bluetooth"
)

func main() {
    manager, err := bluetooth.NewSimpleManager()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Enable Bluetooth if needed
    if !manager.IsBluetoothEnabled() {
        manager.EnableBluetooth(ctx)
    }
    
    // Scan for devices
    manager.ScanForDevices(ctx, 10*time.Second, func(address, name string, rssi int) {
        fmt.Printf("Found: %s (%s) RSSI: %d dBm\\n", name, address, rssi)
    })
}
```

### Connect and Read Characteristic

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "gobot.io/x/gobot/v2/bluetooth"
)

func main() {
    manager, err := bluetooth.NewSimpleManager()
    if err != nil {
        log.Fatal(err)
    }
    
    ctx := context.Background()
    
    // Connect to device
    device, err := manager.ConnectToDevice(ctx, "12:34:56:78:9A:BC")
    if err != nil {
        log.Fatal(err)
    }
    defer device.Disconnect(ctx)
    
    // Discover services
    if err := device.DiscoverServices(ctx); err != nil {
        log.Fatal(err)
    }
    
    // Read battery level
    data, err := device.ReadCharacteristic(ctx, "180F", "2A19")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Printf("Battery level: %d%%\\n", data[0])
}
```

### Create BLE Peripheral Server

```go
package main

import (
    "context"
    "fmt"
    "log"
    
    "gobot.io/x/gobot/v2/bluetooth"
)

func main() {
    manager, err := bluetooth.GetManager()
    if err != nil {
        log.Fatal(err)
    }
    
    adapter, err := manager.DefaultAdapter()
    if err != nil {
        log.Fatal(err)
    }
    
    peripheral := adapter.Peripheral()
    
    // Create a custom service
    serviceUUID, _ := bluetooth.NewUUID("12345678-1234-1234-1234-123456789ABC")
    service, err := peripheral.AddService(serviceUUID, true)
    if err != nil {
        log.Fatal(err)
    }
    
    // Add a readable characteristic
    charUUID, _ := bluetooth.NewUUID("12345678-1234-1234-1234-123456789ABD")
    char, err := service.AddCharacteristic(charUUID, bluetooth.CharacteristicRead, []byte("Hello World"))
    if err != nil {
        log.Fatal(err)
    }
    
    // Set up read callback
    char.OnRead(func() []byte {
        return []byte(fmt.Sprintf("Time: %s", time.Now().Format("15:04:05")))
    })
    
    // Start advertising
    ctx := context.Background()
    advData := bluetooth.AdvertisementData{
        LocalName: "Go BLE Server",
        ServiceUUIDs: []bluetooth.UUID{serviceUUID},
    }
    
    err = peripheral.StartAdvertising(ctx, bluetooth.DefaultAdvertisingParams(), advData)
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("BLE server running...")
    select {} // Run forever
}
```

## API Reference

### Core Types

- `Manager`: Top-level interface for accessing Bluetooth adapters
- `Adapter`: Represents a Bluetooth adapter (supports both Central and Peripheral)
- `Central`: Client-side operations (scanning, connecting)
- `Peripheral`: Server-side operations (advertising, serving)
- `Device`: Represents a connected BLE device
- `Service`: GATT service interface
- `Characteristic`: GATT characteristic interface

### Simple API

For common use cases, use the simplified API:

- `SimpleManager`: High-level manager with convenience methods
- `SimpleDevice`: Simplified device interface
- `SimpleService`: Simplified service interface
- `SimpleCharacteristic`: Simplified characteristic interface

### UUIDs

The package includes standard Bluetooth UUIDs:

```go
// Standard services
bluetooth.UUIDGenericAccess       // 1800
bluetooth.UUIDBattery             // 180F
bluetooth.UUIDDeviceInformation   // 180A
bluetooth.UUIDHeartRate           // 180D

// Standard characteristics
bluetooth.UUIDDeviceName          // 2A00
bluetooth.UUIDBatteryLevel        // 2A19
bluetooth.UUIDManufacturerNameString // 2A29
```

## Platform-Specific Notes

### Linux (BlueZ)

- Requires BlueZ 5.x or later
- Uses D-Bus for communication with BlueZ daemon
- Requires appropriate permissions (usually `bluetooth` group)
- May need to disable ModemManager for some devices

### macOS (Core Bluetooth)

- Requires macOS 10.13 or later
- Uses Core Bluetooth framework via CGO
- Privacy restrictions apply (no adapter MAC address access)
- Requires app entitlements for distribution

### Windows (WinRT)

- Requires Windows 10 version 1803 or later
- Uses Windows Runtime APIs via winrt-go
- May require pairing for some operations
- Administrator privileges may be required

### Nordic SoftDevice

- Supports nRF52 series microcontrollers
- Requires Nordic SoftDevice S132/S140
- Uses direct C API calls via CGO
- Designed for bare metal/embedded use

## Error Handling

The API uses Go's standard error handling patterns:

```go
device, err := manager.ConnectToDevice(ctx, address)
if err != nil {
    if err == bluetooth.ErrNotSupported {
        // Handle unsupported operation
    } else if err == bluetooth.ErrConnectionFailed {
        // Handle connection failure
    } else {
        // Handle other errors
    }
}
```

Common errors:
- `ErrNotSupported`: Operation not supported on this platform
- `ErrNotConnected`: Device not connected
- `ErrConnectionFailed`: Failed to connect to device
- `ErrScanTimeout`: Scan operation timed out
- `ErrServiceNotFound`: Requested service not found
- `ErrCharacteristicNotFound`: Requested characteristic not found

## Thread Safety

All API operations are thread-safe. You can safely call methods from multiple goroutines:

```go
// Safe to use from multiple goroutines
go func() {
    manager.ScanForDevices(ctx, timeout, callback)
}()

go func() {
    device.ReadCharacteristic(ctx, serviceUUID, charUUID)
}()
```

## Testing

Run the full test suite:

```bash
go test ./bluetooth/...
```

Platform-specific tests:

```bash
# Linux only
go test -tags linux ./bluetooth/...

# macOS only  
go test -tags darwin ./bluetooth/...

# Windows only
go test -tags windows ./bluetooth/...

# Nordic only
go test -tags "tinygo,nrf52" ./bluetooth/...
```

## Examples

See the `examples/` directory for complete working examples:

- `simple_scanner.go`: Basic device scanning
- `heart_rate_monitor.go`: Connect to heart rate monitor
- `peripheral_server.go`: BLE peripheral server

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass on your target platform
5. Submit a pull request

## License

This project is licensed under the same license as Gobot.

## Acknowledgments

- TinyGo Bluetooth library for inspiration and Nordic SoftDevice integration
- BlueZ project for Linux Bluetooth stack
- Microsoft for WinRT Go bindings
- Apple for Core Bluetooth documentation