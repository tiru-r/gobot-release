package i2c

import (
	"fmt"

	"gobot.io/x/gobot/v2"
)

const (
	// Error event
	Error = "error"
)

const (
	// BusNotInitialized is the initial value for a bus
	BusNotInitialized = -1

	// AddressNotInitialized is the initial value for an address
	AddressNotInitialized = -1
)

var (
	// ErrNotEnoughBytes is used when the count of read bytes was too small
	ErrNotEnoughBytes = fmt.Errorf("Not enough bytes read")
	// ErrNotReady is used when the device is not ready
	ErrNotReady = fmt.Errorf("Device is not ready")
)

type bitState uint8

const (
	clearBit bitState = 0x00
	setBit   bitState = 0x01
)

// Connection is a connection to an I2C device with a specified address
// on a specific bus. Used as an alternative to the I2c interface.
// Implements I2cOperations to talk to the device, wrapping the
// calls in SetAddress to always target the specified device.
// Provided by an Adaptor by implementing the I2cConnector interface.
type Connection gobot.I2cOperations

type i2cConnection struct {
	bus     gobot.I2cSystemDevicer
	address int
}

// NewConnection creates and returns a new connection to a specific i2c device on a bus and address.
func NewConnection(bus gobot.I2cSystemDevicer, address int) *i2cConnection {
	return &i2cConnection{bus: bus, address: address}
}

// Read data from an i2c device.
func (c *i2cConnection) Read(data []byte) (int, error) {
	return c.bus.Read(c.address, data)
}

// Write data to an i2c device.
func (c *i2cConnection) Write(data []byte) (int, error) {
	return c.bus.Write(c.address, data)
}

// Close connection to i2c device. The bus was created by adaptor and will be closed there.
func (c *i2cConnection) Close() error {
	return nil
}

// ReadByte reads a single byte from the i2c device.
func (c *i2cConnection) ReadByte() (byte, error) {
	// Set address through Write operation first to ensure address is set
	if _, err := c.bus.Write(c.address, []byte{}); err != nil {
		return 0, err
	}
	return c.bus.ReadByte()
}

// ReadByteData reads a byte value for a register on the i2c device.
func (c *i2cConnection) ReadByteData(reg uint8) (uint8, error) {
	return c.bus.ReadByteData(c.address, reg)
}

// ReadWordData reads a word value for a register on the i2c device.
func (c *i2cConnection) ReadWordData(reg uint8) (uint16, error) {
	return c.bus.ReadWordData(c.address, reg)
}

// ReadBlockData reads a block of bytes from a register on the i2c device.
func (c *i2cConnection) ReadBlockData(reg uint8, b []byte) error {
	return c.bus.ReadBlockData(c.address, reg, b)
}

// WriteByte writes a single byte to the i2c device.
func (c *i2cConnection) WriteByte(val byte) error {
	// Set address through Write operation first
	if _, err := c.bus.Write(c.address, []byte{}); err != nil {
		return err
	}
	return c.bus.WriteByte(val)
}

// WriteByteData writes a byte value to a register on the i2c device.
func (c *i2cConnection) WriteByteData(reg uint8, val uint8) error {
	return c.bus.WriteByteData(c.address, reg, val)
}

// WriteWordData writes a word value to a register on the i2c device.
func (c *i2cConnection) WriteWordData(reg uint8, val uint16) error {
	return c.bus.WriteWordData(c.address, reg, val)
}

// WriteBlockData writes a block of bytes to a register on the i2c device.
func (c *i2cConnection) WriteBlockData(reg uint8, b []byte) error {
	return c.bus.WriteBlockData(c.address, reg, b)
}

// WriteBytes writes a block of bytes to the current register on the i2c device.
func (c *i2cConnection) WriteBytes(b []byte) error {
	return c.bus.WriteBytes(c.address, b)
}

func twosComplement16Bit(uValue uint16) int16 {
	result := int32(uValue)
	if result&0x8000 != 0 {
		result -= 1 << 16
	}
	return int16(result) //nolint:gosec // ok here
}

func swapBytes(value uint16) uint16 {
	return (value << 8) | (value >> 8)
}
