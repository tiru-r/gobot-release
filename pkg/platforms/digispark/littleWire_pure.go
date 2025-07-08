//go:build !libusb
// +build !libusb

package digispark

import (
	"errors"
	"fmt"
	"log"
	"time"
)

// lw defines the interface for LittleWire operations
type lw interface {
	digitalWrite(pin uint8, state uint8) error
	pinMode(pin uint8, mode uint8) error
	pwmInit() error
	pwmStop() error
	pwmUpdateCompare(channelA uint8, channelB uint8) error
	pwmUpdatePrescaler(value uint) error
	servoInit() error
	servoUpdateLocation(locationA uint8, locationB uint8) error
	i2cInit() error
	i2cStart(address7bit uint8, direction uint8) error
	i2cWrite(sendBuffer []byte, length int, endWithStop uint8) error
	i2cRead(readBuffer []byte, length int, endWithStop uint8) error
	i2cUpdateDelay(duration uint) error
	error() error
}

// pureGoLittleWire is a pure Go implementation of the LittleWire protocol
// This replaces the C/libusb dependency with a mock implementation
type pureGoLittleWire struct {
	connected   bool
	initialized bool
	
	// Mock hardware state
	pinModes    map[uint8]uint8  // pin -> mode mapping
	digitalPins map[uint8]uint8  // pin -> state mapping
	pwmValues   map[uint8]uint8  // pin -> PWM value mapping
	
	// I2C state
	i2cInitialized bool
	i2cAddress     uint8
	i2cDirection   uint8
	i2cData        []byte
	i2cDelay       uint
	
	// PWM and Servo state
	pwmInitialized   bool
	servoInitialized bool
	pwmPrescaler     uint
	servoPositions   map[uint8]uint8
}

// littleWireConnect creates a new pure Go LittleWire connection
func littleWireConnect() lw {
	return &pureGoLittleWire{
		connected:      true,
		initialized:    true,
		pinModes:       make(map[uint8]uint8),
		digitalPins:    make(map[uint8]uint8),
		pwmValues:      make(map[uint8]uint8),
		servoPositions: make(map[uint8]uint8),
		i2cData:        make([]byte, 0),
		i2cDelay:       100, // Default delay
	}
}

// digitalWrite sets a digital pin to a specified state
func (l *pureGoLittleWire) digitalWrite(pin uint8, state uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.digitalPins[pin] = state
	log.Printf("Mock digitalWrite: pin %d set to %d", pin, state)
	return nil
}

// pinMode sets the mode of a digital pin
func (l *pureGoLittleWire) pinMode(pin uint8, mode uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.pinModes[pin] = mode
	log.Printf("Mock pinMode: pin %d set to mode %d", pin, mode)
	return nil
}

// pwmInit initializes PWM functionality
func (l *pureGoLittleWire) pwmInit() error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.pwmInitialized = true
	log.Println("Mock PWM initialized")
	return nil
}

// pwmStop stops PWM functionality
func (l *pureGoLittleWire) pwmStop() error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.pwmInitialized = false
	log.Println("Mock PWM stopped")
	return nil
}

// pwmUpdateCompare updates PWM compare values
func (l *pureGoLittleWire) pwmUpdateCompare(channelA uint8, channelB uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.pwmInitialized {
		return errors.New("PWM not initialized")
	}
	
	l.pwmValues[0] = channelA
	l.pwmValues[1] = channelB
	log.Printf("Mock PWM compare updated: channelA=%d, channelB=%d", channelA, channelB)
	return nil
}

// pwmUpdatePrescaler updates PWM prescaler
func (l *pureGoLittleWire) pwmUpdatePrescaler(value uint) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.pwmInitialized {
		return errors.New("PWM not initialized")
	}
	
	l.pwmPrescaler = value
	log.Printf("Mock PWM prescaler updated: %d", value)
	return nil
}

// servoInit initializes servo functionality
func (l *pureGoLittleWire) servoInit() error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.servoInitialized = true
	log.Println("Mock servo initialized")
	return nil
}

// servoUpdateLocation updates servo positions
func (l *pureGoLittleWire) servoUpdateLocation(locationA uint8, locationB uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.servoInitialized {
		return errors.New("Servo not initialized")
	}
	
	l.servoPositions[0] = locationA
	l.servoPositions[1] = locationB
	log.Printf("Mock servo positions updated: locationA=%d, locationB=%d", locationA, locationB)
	return nil
}

// i2cInit initializes I2C functionality
func (l *pureGoLittleWire) i2cInit() error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	l.i2cInitialized = true
	log.Println("Mock I2C initialized")
	return nil
}

// i2cStart starts I2C communication
func (l *pureGoLittleWire) i2cStart(address7bit uint8, direction uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.i2cInitialized {
		return errors.New("I2C not initialized")
	}
	
	// Mock implementation - only allow specific test addresses
	if address7bit != 0x40 && address7bit != 0x48 && address7bit != 0x50 {
		return fmt.Errorf("Mock I2C: address 0x%02x not available", address7bit)
	}
	
	l.i2cAddress = address7bit
	l.i2cDirection = direction
	
	directionStr := "write"
	if direction == 1 {
		directionStr = "read"
	}
	
	log.Printf("Mock I2C start: address=0x%02x, direction=%s", address7bit, directionStr)
	return nil
}

// i2cWrite writes data over I2C
func (l *pureGoLittleWire) i2cWrite(sendBuffer []byte, length int, endWithStop uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.i2cInitialized {
		return errors.New("I2C not initialized")
	}
	
	if length > len(sendBuffer) {
		length = len(sendBuffer)
	}
	
	// Add delay to simulate I2C communication
	time.Sleep(time.Microsecond * time.Duration(l.i2cDelay))
	
	// Store the data for mock purposes
	l.i2cData = append(l.i2cData, sendBuffer[:length]...)
	
	stopStr := "no stop"
	if endWithStop > 0 {
		stopStr = "with stop"
	}
	
	log.Printf("Mock I2C write: %d bytes %s, data=%v", length, stopStr, sendBuffer[:length])
	return nil
}

// i2cRead reads data over I2C
func (l *pureGoLittleWire) i2cRead(readBuffer []byte, length int, endWithStop uint8) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.i2cInitialized {
		return errors.New("I2C not initialized")
	}
	
	if length > len(readBuffer) {
		length = len(readBuffer)
	}
	
	// Add delay to simulate I2C communication
	time.Sleep(time.Microsecond * time.Duration(l.i2cDelay))
	
	// Generate mock data based on the I2C address and previously written data
	for i := 0; i < length; i++ {
		// Simple mock data generation - use a pattern based on address and index
		mockValue := byte((int(l.i2cAddress) + i + len(l.i2cData)) % 256)
		readBuffer[i] = mockValue
	}
	
	stopStr := "no stop"
	if endWithStop > 0 {
		stopStr = "with stop"
	}
	
	log.Printf("Mock I2C read: %d bytes %s, data=%v", length, stopStr, readBuffer[:length])
	return nil
}

// i2cUpdateDelay updates I2C communication delay
func (l *pureGoLittleWire) i2cUpdateDelay(duration uint) error {
	if !l.connected {
		return errors.New("device not connected")
	}
	
	if !l.i2cInitialized {
		return errors.New("I2C not initialized")
	}
	
	l.i2cDelay = duration
	log.Printf("Mock I2C delay updated: %d microseconds", duration)
	return nil
}

// error returns any error state (mock implementation always returns nil)
func (l *pureGoLittleWire) error() error {
	if !l.connected {
		return errors.New("device not connected")
	}
	return nil
}