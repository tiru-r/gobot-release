# Gobot

Gobot is a framework for robotics, IoT, and the connected machine, written in the Go programming language.

## Features

- Simple, unified API for controlling hardware devices
- Support for multiple platforms and devices
- Built-in support for concurrent programming patterns
- Event-driven architecture
- Extensible plugin system
- Rich ecosystem of drivers and adaptors

## Architecture

Gobot follows a simple architecture with three main components:

- **Robot**: The main controller that orchestrates connections and devices
- **Connections**: Hardware adaptors that provide interfaces to platforms (GPIO, I2C, SPI, etc.)
- **Devices**: Software drivers that control specific hardware components

## Quick Start

### Installation

```bash
go get -u gobot.io/x/gobot/v2
```

### Basic Example

```go
package main

import (
    "time"
    "gobot.io/x/gobot/v2"
    "gobot.io/x/gobot/v2/drivers/gpio"
    "gobot.io/x/gobot/v2/platforms/firmata"
)

func main() {
    firmataAdaptor := firmata.NewAdaptor("/dev/ttyACM0")
    led := gpio.NewLedDriver(firmataAdaptor, "13")

    work := func() {
        gobot.Every(1*time.Second, func() {
            led.Toggle()
        })
    }

    robot := gobot.NewRobot("bot",
        []gobot.Connection{firmataAdaptor},
        []gobot.Device{led},
        work,
    )

    robot.Start()
}
```

## Core Concepts

### Robot

The `Robot` type is the main entry point for controlling your hardware. It manages connections to hardware platforms and the devices connected to them.

Key features:
- Automatic connection and device lifecycle management
- Graceful shutdown handling
- Event system for device communication
- Work function for main program logic

### Connections

Connections represent the interface to a hardware platform (Arduino, Raspberry Pi, etc.). They handle the low-level communication protocols.

### Devices

Devices are software drivers that control specific hardware components (LEDs, sensors, motors, etc.). They use connections to communicate with the actual hardware.

### Events

Gobot provides an event system for handling asynchronous communication between devices and your application code.

## Robot Lifecycle

1. **Initialization**: Robot is created with connections and devices
2. **Connection Start**: All connections are established
3. **Device Start**: All devices are initialized
4. **Work Execution**: Main work function runs
5. **Graceful Shutdown**: Devices and connections are properly closed

## Web Interface

Gobot provides a RESTful API (C3PIO-compatible) for external integrations. For modern web interfaces, we recommend using external tools:

- **[ThingsBoard](https://thingsboard.io/)** - Comprehensive IoT platform with professional dashboards
- **[Node-RED](https://nodered.org/)** - Browser-based flow editor for IoT applications  
- **Custom applications** - Build with React, Vue.js, or other modern frameworks connecting to Gobot's API

### API Example

```go
package main

import (
    "gobot.io/x/gobot/v2"
    "gobot.io/x/gobot/v2/pkg/api"
)

func main() {
    manager := gobot.NewManager()
    
    // Start API server with web interface
    api := api.NewAPI(manager)
    api.Start() // Serves on :3000 by default
    
    // Add your robots...
    manager.AddRobot(gobot.NewRobot("mybot"))
    manager.Start()
}
```

## API Reference

### Creating a Robot

```go
robot := gobot.NewRobot(name, connections, devices, work)
```

### Robot Methods

- `Start()`: Start the robot and begin executing work
- `Stop()`: Gracefully stop the robot
- `Running()`: Check if the robot is currently running
- `AddDevice(device)`: Add a device to the robot
- `AddConnection(connection)`: Add a connection to the robot

### Utility Functions

- `gobot.Every(duration, func())`: Execute function repeatedly at intervals
- `gobot.After(duration, func())`: Execute function once after delay

## Error Handling

Gobot follows Go's idiomatic error handling patterns. Most operations return an error that should be checked:

```go
if err := robot.Start(); err != nil {
    log.Fatal(err)
}
```

## Concurrency

Gobot is designed with Go's concurrency model in mind. The framework handles goroutines internally, but you can also use standard Go concurrency patterns in your work functions.

## Migration

For information about migrating between versions, see [MIGRATION.md](MIGRATION.md).

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## Changelog

See [CHANGELOG.md](CHANGELOG.md) for version history and changes.

## Code of Conduct

This project follows a [Code of Conduct](CODE_OF_CONDUCT.md). By participating, you agree to abide by its terms.

## License

Licensed under the Apache 2.0 License. See LICENSE file for details.

## Maintainability

For information about ongoing maintainability improvements, see [MAINTAINABILITY_IMPROVEMENTS.md](MAINTAINABILITY_IMPROVEMENTS.md).