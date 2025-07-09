# Gobot

[![Go Reference](https://pkg.go.dev/badge/gobot.io/x/gobot/v2.svg)](https://pkg.go.dev/gobot.io/x/gobot/v2)
[![License](https://img.shields.io/badge/License-Apache_2.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

Gobot is a framework for robotics, IoT, and the connected machine, written in the Go programming language. It provides a simple, unified API for controlling hardware devices across multiple platforms and supports concurrent programming patterns with an event-driven architecture.

## Features

- **Simple, unified API** for controlling hardware devices
- **Multi-platform support** - Arduino, Raspberry Pi, BeagleBone, Intel Edison, and more
- **Concurrent programming** with built-in support for Go's concurrency patterns
- **Event-driven architecture** for responsive hardware interactions
- **Extensible plugin system** with rich ecosystem of drivers and adaptors
- **RESTful API** (C3PIO-compatible) for external integrations
- **Real-time communication** with WebSocket support

## Quick Start

### Installation

```bash
go get -u gobot.io/x/gobot/v2
```

### Hello World Example

```go
package main

import (
    "fmt"
    "time"
    "gobot.io/x/gobot/v2/pkg/core"
)

func main() {
    robot := core.NewRobot(
        core.WithName("HelloBot"),
        core.WithWork(func() {
            for i := range 5 {
                fmt.Printf("Hello from Gobot! Count: %d\n", i+1)
                time.Sleep(1 * time.Second)
            }
        }),
        core.WithAutoRun(true),
    )

    if err := robot.Start(); err != nil {
        fmt.Printf("Error starting robot: %v\n", err)
    }
}
```

### LED Blink Example

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

    robot := gobot.NewRobot("blinkBot",
        []gobot.Connection{firmataAdaptor},
        []gobot.Device{led},
        work,
    )

    robot.Start()
}
```

## Architecture

Gobot follows a simple, modular architecture with three main components:

### 1. Robot
The main controller that orchestrates connections and devices. Features include:
- Automatic connection and device lifecycle management
- Graceful shutdown handling
- Event system for device communication
- Work function for main program logic

### 2. Connections (Adaptors)
Hardware adaptors that provide interfaces to platforms:
- **GPIO** - General Purpose Input/Output
- **I2C** - Inter-Integrated Circuit protocol
- **SPI** - Serial Peripheral Interface
- **Serial** - Serial communication protocols
- **Bluetooth** - Wireless communication
- **WiFi** - Network connectivity

### 3. Devices (Drivers)
Software drivers that control specific hardware components:
- **Sensors** - Temperature, humidity, accelerometer, etc.
- **Actuators** - LEDs, motors, servos, etc.
- **Communication** - MQTT, WebSockets, REST APIs
- **Storage** - Database interfaces, file systems

## Supported Platforms

### Single Board Computers
- **Raspberry Pi** - All models with GPIO support
- **BeagleBone** - BeagleBone Black, PocketBeagle
- **Intel Edison** - Arduino and Mini breakout boards
- **Intel Joule** - Development platform
- **CHIP** - Next Thing Co. computer
- **Tinker Board** - ASUS single-board computer

### Microcontrollers
- **Arduino** - Uno, Mega, Leonardo, etc. (via Firmata)
- **ESP32/ESP8266** - WiFi-enabled microcontrollers
- **Particle** - Photon, Electron, Argon
- **Digispark** - ATtiny85-based microcontroller

### Specialized Hardware
- **Parrot Drones** - AR.Drone, Bebop, Minidrone
- **Sphero** - Robotic balls and toys
- **MQTT** - Message queue telemetry transport
- **Keyboard/Joystick** - Human interface devices
- **OpenCV** - Computer vision and image processing

## Core Concepts

### Robot Lifecycle

1. **Initialization** - Robot is created with connections and devices
2. **Connection Start** - All connections are established
3. **Device Start** - All devices are initialized
4. **Work Execution** - Main work function runs
5. **Graceful Shutdown** - Devices and connections are properly closed

### Events

Gobot provides an event system for handling asynchronous communication:

```go
// Subscribe to device events
gobot.On(robot.Event("button"), func(data interface{}) {
    fmt.Println("Button pressed!")
})

// Publish custom events
gobot.Publish(robot.Event("sensor"), sensorData)
```

### Utility Functions

```go
// Execute function repeatedly at intervals
gobot.Every(500*time.Millisecond, func() {
    fmt.Println("This runs every 500ms")
})

// Execute function once after delay
gobot.After(2*time.Second, func() {
    fmt.Println("This runs after 2 seconds")
})
```

## Web Interface & API

Gobot provides a RESTful API (C3PIO-compatible) for external integrations:

```go
package main

import (
    "gobot.io/x/gobot/v2"
    "gobot.io/x/gobot/v2/pkg/api"
)

func main() {
    manager := gobot.NewManager()
    
    // Start API server
    api := api.NewAPI(manager)
    api.Start() // Serves on :3000 by default
    
    // Add your robots
    manager.AddRobot(gobot.NewRobot("mybot"))
    manager.Start()
}
```

### Recommended External Tools

For modern web interfaces, we recommend:
- **[ThingsBoard](https://thingsboard.io/)** - Comprehensive IoT platform with professional dashboards
- **[Node-RED](https://nodered.org/)** - Browser-based flow editor for IoT applications
- **Custom applications** - Build with React, Vue.js, or other frameworks connecting to Gobot's API

## Examples

The `examples/` directory contains over 100 examples demonstrating various use cases:

### Basic Examples
- **Hello World** - Simple robot introduction
- **LED Blink** - Basic GPIO control
- **Button Input** - Reading digital inputs

### Advanced Examples
- **Computer Vision** - OpenCV integration for image processing
- **Robot Swarm** - Coordinating multiple robots
- **Web Control** - Browser-based robot control interfaces
- **Sensor Networks** - IoT sensor data collection and processing

### Platform-Specific Examples
- **Arduino/Firmata** - Microcontroller programming
- **Raspberry Pi** - GPIO, I2C, SPI device control
- **Drone Control** - Flying robot automation
- **Bluetooth/BLE** - Wireless device communication

## Documentation

- **[API Reference](https://pkg.go.dev/gobot.io/x/gobot/v2)** - Complete API documentation
- **[Migration Guide](docs/MIGRATION.md)** - Upgrading between versions
- **[Contributing](docs/CONTRIBUTING.md)** - Development guidelines
- **[Code of Conduct](docs/CODE_OF_CONDUCT.md)** - Community standards
- **[Changelog](docs/CHANGELOG.md)** - Version history
- **[Maintainability](docs/MAINTAINABILITY_IMPROVEMENTS.md)** - Code quality improvements

## Testing

Gobot includes comprehensive test coverage:

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific package tests
go test ./drivers/gpio
```

## Contributing

We welcome contributions! Here's how to get started:

1. **Fork the repository** on GitHub
2. **Create a feature branch** from main
3. **Make your changes** with appropriate tests
4. **Run the test suite** to ensure everything works
5. **Submit a pull request** with a clear description

Please read our [Contributing Guide](docs/CONTRIBUTING.md) for detailed information about development setup, coding standards, and the contribution process.

## Community

- **GitHub Issues** - Bug reports and feature requests
- **GitHub Discussions** - Community support and general questions
- **Examples** - Learn from the extensive example library

## License

Copyright (c) 2013-2020 The Hybrid Group

Licensed under the Apache License, Version 2.0 (the "License"); you may not use this file except in compliance with the License. You may obtain a copy of the License at:

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software distributed under the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the License for the specific language governing permissions and limitations under the License.

## Support

For support and questions:
- Check the [examples](examples/) directory for code samples
- Review the [API documentation](https://pkg.go.dev/gobot.io/x/gobot/v2)
- Open an issue on GitHub for bugs or feature requests
- Start a discussion on GitHub for general questions

---

**Ready to build robots with Go?** Start with the [Quick Start](#quick-start) guide and explore the examples to see what's possible with Gobot!