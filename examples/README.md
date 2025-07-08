# Gobot Examples

This directory contains examples demonstrating various features and use cases of Gobot.

## Directory Structure

### Basic Examples (`basic/`)
Simple examples to get started with Gobot:

- **hello/**: Basic robot that prints messages
- **blink/**: LED blinking example
- **button/**: Button press detection

### Platform Examples (`platforms/`)
Platform-specific examples:

- **raspi/**: Raspberry Pi examples
- **arduino/**: Arduino/Firmata examples
- **edison/**: Intel Edison examples
- **beaglebone/**: BeagleBone examples
- **firmata/**: Firmata protocol examples

### Advanced Examples (`advanced/`)
Complex examples demonstrating advanced features:

- **robot-swarm/**: Multiple robot coordination
- **web-control/**: Web-based robot control
- **computer-vision/**: Computer vision integration

## Running Examples

### Prerequisites
- Go 1.24 or later
- Hardware platform (Raspberry Pi, Arduino, etc.) for platform-specific examples
- Required drivers and libraries for your platform

### Basic Usage
```bash
# Run a basic example
cd examples/basic/hello
go run main.go

# Run a platform-specific example
cd examples/platforms/raspi/blink
go run main.go
```

### Cross-Compilation
Some examples can be cross-compiled for different platforms:

```bash
# For Raspberry Pi
GOOS=linux GOARCH=arm64 go build -o example-arm64 main.go

# For Arduino (requires specific build tags)
go build -tags=firmata -o example-firmata main.go
```

## Example Categories

### 1. Basic Examples
Start here if you're new to Gobot. These examples demonstrate:
- Creating and starting robots
- Basic work functions
- Simple device interactions

### 2. Platform Examples
Hardware-specific examples showing:
- Platform adaptor usage
- GPIO, I2C, SPI communication
- Platform-specific drivers

### 3. Advanced Examples
Complex scenarios including:
- Multiple robot coordination
- Web APIs and remote control
- Computer vision integration
- Real-time data processing

## Contributing Examples

When adding new examples:

1. Choose the appropriate category (basic/platforms/advanced)
2. Create a descriptive directory name
3. Include a `main.go` file with clear comments
4. Add a `README.md` explaining the example
5. Test on relevant hardware platforms

## Getting Help

- Check the [main documentation](../../docs/)
- Visit the [Gobot community](https://gobot.io/community/)
- Open an issue on [GitHub](https://github.com/hybridgroup/gobot/issues)