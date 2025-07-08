//go:build !libusb
// +build !libusb

package digispark

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"gobot.io/x/gobot/v2/drivers/i2c"
)

const (
	availableI2cAddress = 0x40
	maxUint8            = ^uint8(0)
)

var (
	_       i2c.Connector = (*Adaptor)(nil)
	i2cData               = []byte{5, 4, 3, 2, 1, 0}
)

func initTestAdaptorI2cPure() *Adaptor {
	a := NewAdaptor()
	err := a.Connect()
	if err != nil {
		panic(err)
	}
	return a
}

func TestPureGoDigisparkAdaptorI2cGetI2cConnection(t *testing.T) {
	// arrange
	var c i2c.Connection
	var err error
	a := initTestAdaptorI2cPure()

	// act
	c, err = a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// assert
	require.NoError(t, err)
	assert.NotNil(t, c)
}

func TestPureGoDigisparkAdaptorI2cGetI2cConnectionFailWithInvalidBus(t *testing.T) {
	// arrange
	a := initTestAdaptorI2cPure()

	// act
	c, err := a.GetI2cConnection(0x40, 1)

	// assert
	require.ErrorContains(t, err, "Invalid bus number 1, only 0 is supported")
	assert.Nil(t, c)
}

func TestPureGoDigisparkAdaptorI2cStartFailWithWrongAddress(t *testing.T) {
	// arrange
	data := []byte{0, 1, 2, 3, 4}
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(0x39, a.DefaultI2cBus())

	// act
	count, err := c.Write(data)

	// assert
	assert.Equal(t, 0, count)
	require.ErrorContains(t, err, fmt.Sprintf("Mock I2C: address 0x%02x not available", 0x39))
}

func TestPureGoDigisparkAdaptorI2cWrite(t *testing.T) {
	// arrange
	data := []byte{0, 1, 2, 3, 4}
	dataLen := len(data)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	count, err := c.Write(data)

	// assert
	require.NoError(t, err)
	assert.Equal(t, dataLen, count)
}

func TestPureGoDigisparkAdaptorI2cWriteByte(t *testing.T) {
	// arrange
	data := byte(0x02)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	err := c.WriteByte(data)

	// assert
	require.NoError(t, err)
}

func TestPureGoDigisparkAdaptorI2cWriteByteData(t *testing.T) {
	// arrange
	reg := uint8(0x03)
	data := byte(0x09)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	err := c.WriteByteData(reg, data)

	// assert
	require.NoError(t, err)
}

func TestPureGoDigisparkAdaptorI2cWriteWordData(t *testing.T) {
	// arrange
	reg := uint8(0x04)
	data := uint16(0x0508)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	err := c.WriteWordData(reg, data)

	// assert
	require.NoError(t, err)
}

func TestPureGoDigisparkAdaptorI2cWriteBlockData(t *testing.T) {
	// arrange
	reg := uint8(0x05)
	data := []byte{0x80, 0x81, 0x82}
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	err := c.WriteBlockData(reg, data)

	// assert
	require.NoError(t, err)
}

func TestPureGoDigisparkAdaptorI2cRead(t *testing.T) {
	// arrange
	data := []byte{0, 1, 2, 3, 4}
	dataLen := len(data)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	count, err := c.Read(data)

	// assert
	assert.Equal(t, dataLen, count)
	require.NoError(t, err)
	// Data should have been populated by mock
	assert.NotEqual(t, []byte{0, 1, 2, 3, 4}, data)
}

func TestPureGoDigisparkAdaptorI2cReadByte(t *testing.T) {
	// arrange
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	data, err := c.ReadByte()

	// assert
	require.NoError(t, err)
	// Should return mock data
	assert.NotEqual(t, byte(0), data)
}

func TestPureGoDigisparkAdaptorI2cReadByteData(t *testing.T) {
	// arrange
	reg := uint8(0x04)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	data, err := c.ReadByteData(reg)

	// assert
	require.NoError(t, err)
	// Should return mock data
	assert.NotEqual(t, uint8(0), data)
}

func TestPureGoDigisparkAdaptorI2cReadWordData(t *testing.T) {
	// arrange
	reg := uint8(0x05)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	data, err := c.ReadWordData(reg)

	// assert
	require.NoError(t, err)
	// Should return mock data
	assert.NotEqual(t, uint16(0), data)
}

func TestPureGoDigisparkAdaptorI2cReadBlockData(t *testing.T) {
	// arrange
	reg := uint8(0x05)
	data := []byte{0, 0, 0, 0, 0}
	dataLen := len(data)
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	err := c.ReadBlockData(reg, data)

	// assert
	require.NoError(t, err)
	// Data should have been populated by mock
	assert.NotEqual(t, []byte{0, 0, 0, 0, 0}, data[:dataLen])
}

func TestPureGoDigisparkAdaptorI2cUpdateDelay(t *testing.T) {
	// arrange
	a := initTestAdaptorI2cPure()
	c, _ := a.GetI2cConnection(availableI2cAddress, a.DefaultI2cBus())

	// act
	conn := c.(*digisparkI2cConnection)
	err := conn.UpdateDelay(uint(100))

	// assert
	require.NoError(t, err)
}