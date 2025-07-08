# Bluetooth Package - Restructured Architecture

This package provides cross-platform Bluetooth Low Energy (BLE) support for Gobot with a focus on clean architecture and maintainability.

## 🏗️ Architecture Overview

The bluetooth package has been completely restructured from a monolithic 1,221-line file into a clean, modular architecture:

```
bluetooth/
├── RESTRUCTURE.md                     # This documentation
├── darwin_main.go                     # Main Darwin platform interface
├── darwin_old.go.bak                  # Backup of original monolithic file
│
├── central/                           # Central (client) role implementation
│   └── darwin_central.go              # macOS Central manager
│
├── device/                            # Device management
│   └── darwin_device.go               # macOS Device implementation
│
├── peripheral/                        # Peripheral (server) role implementation
│   └── darwin_peripheral.go           # macOS Peripheral manager
│
├── service/                           # Service and characteristic management
│   └── darwin_service.go              # macOS Service implementation
│
└── internal/                          # Internal implementation details
    ├── types/                         # Type definitions
    │   └── darwin_types.go            # Darwin-specific types
    ├── cbridge/                       # C/Objective-C bridge
    │   └── darwin_cbridge.go          # Core Bluetooth bridge
    └── managers/                      # Platform managers
        └── darwin_manager.go          # Darwin Bluetooth manager
```

## ✨ Key Improvements

### **1. Separation of Concerns**
- **C Bridge Layer**: All Objective-C and C code isolated in `internal/cbridge/`
- **Type Definitions**: Clean type definitions in `internal/types/`
- **Role-Based Modules**: Central, Peripheral, Device, and Service in separate packages
- **Platform Management**: Adapter and manager logic separated

### **2. Memory Management**
- **Proper Resource Cleanup**: Finalizers and explicit cleanup methods
- **Safe Pointer Handling**: Encapsulated unsafe pointer management
- **Error Handling**: Comprehensive error types and handling

### **3. Testability**
- **Modular Design**: Each component can be unit tested independently
- **Interface Segregation**: Clean interfaces for mocking
- **Dependency Injection**: Components accept dependencies rather than creating them

### **4. Maintainability**
- **Single Responsibility**: Each file has a single, clear purpose
- **Documentation**: Comprehensive documentation for each component
- **Consistent Patterns**: Standard patterns used throughout

## 🔧 Component Responsibilities

### **Central (`central/darwin_central.go`)**
- BLE scanning and device discovery
- Connection management
- Callback handling for scan results
- Device lifecycle management

### **Device (`device/darwin_device.go`)**
- Device connection state management
- Service discovery
- Device information (name, RSSI, address)
- Connected device registry

### **Peripheral (`peripheral/darwin_peripheral.go`)**
- Advertising management
- Service publication
- Characteristic management
- Central connection handling

### **Service (`service/darwin_service.go`)**
- Service and characteristic discovery
- Read/write operations
- Notification handling
- Descriptor management

### **Internal Components**

#### **Types (`internal/types/darwin_types.go`)**
- Darwin-specific type definitions
- Struct definitions for all BLE entities
- Thread-safe data structures

#### **C Bridge (`internal/cbridge/darwin_cbridge.go`)**
- Objective-C interface implementations
- C function wrappers
- Callback bridge functions
- Memory management helpers

#### **Managers (`internal/managers/darwin_manager.go`)**
- Platform-specific manager implementations
- Adapter management
- Platform initialization

## 🚀 Usage Example

```go
// Get platform manager
manager, err := bluetooth.GetPlatformManager()
if err != nil {
    panic(err)
}

// Get default adapter
adapter, err := manager.DefaultAdapter()
if err != nil {
    panic(err)
}

// Use central for scanning
central := adapter.Central()
err = central.Scan(ctx, scanParams, func(ad bluetooth.Advertisement) {
    fmt.Printf("Found device: %s\n", ad.LocalName)
})
```

## 🔒 Thread Safety

All components are designed to be thread-safe:
- Mutex protection for shared state
- Atomic operations where appropriate
- Proper synchronization for C callbacks

## 🛠️ Platform-Specific Notes

### **macOS/Darwin**
- Uses Core Bluetooth framework
- Privacy restrictions limit MAC address access
- Adapter power control not available programmatically
- Requires proper entitlements for full functionality

## 📈 Performance Improvements

1. **Reduced Memory Footprint**: Modular loading of components
2. **Better Concurrency**: Fine-grained locking strategies
3. **Efficient Callbacks**: Direct C callback handling without Go overhead
4. **Resource Management**: Proper cleanup prevents memory leaks

## 🧪 Testing Strategy

Each component can be tested independently:
- Unit tests for individual components
- Integration tests for component interaction
- Mock implementations for platform-independent testing
- Benchmark tests for performance validation

## 🔮 Future Enhancements

The new architecture enables:
- Easy addition of new platforms (Linux, Windows)
- Plugin system for custom device types
- Advanced features like mesh networking
- Better debugging and monitoring tools

## 📝 Migration Guide

For users of the old monolithic implementation:
- All public APIs remain the same
- Internal structure is completely different
- Better error messages and debugging info
- Improved performance and stability

This restructured architecture provides a solid foundation for future Bluetooth development while maintaining backward compatibility and improving maintainability.