//go:build !libusb
// +build !libusb

package digispark

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPureGoDigisparkAdaptorName(t *testing.T) {
	a := NewAdaptor()
	assert.True(t, strings.HasPrefix(a.Name(), "Digispark"))
	a.SetName("NewName")
	assert.Equal(t, "NewName", a.Name())
}

func TestPureGoDigisparkAdaptorConnect(t *testing.T) {
	a := NewAdaptor()
	assert.NoError(t, a.Connect())
}

func TestPureGoDigisparkAdaptorFinalize(t *testing.T) {
	a := NewAdaptor()
	assert.NoError(t, a.Finalize())
}

func TestPureGoDigisparkAdaptorDigitalWrite(t *testing.T) {
	a := NewAdaptor()
	err := a.Connect()
	require.NoError(t, err)

	err = a.DigitalWrite("1", 1)
	assert.NoError(t, err)
}

func TestPureGoDigisparkAdaptorDigitalWriteError(t *testing.T) {
	a := NewAdaptor()
	err := a.Connect()
	require.NoError(t, err)

	err = a.DigitalWrite("foo", 1)
	assert.Error(t, err)
}

func TestPureGoDigisparkAdaptorPwmWrite(t *testing.T) {
	a := NewAdaptor()
	err := a.Connect()
	require.NoError(t, err)

	err = a.PwmWrite("1", 100)
	assert.NoError(t, err)
}

func TestPureGoDigisparkAdaptorServoWrite(t *testing.T) {
	a := NewAdaptor()
	err := a.Connect()
	require.NoError(t, err)

	err = a.ServoWrite("1", 50)
	assert.NoError(t, err)
}

func TestPureGoDigisparkAdaptorDefaultI2cBus(t *testing.T) {
	a := NewAdaptor()
	assert.Equal(t, 0, a.DefaultI2cBus())
}

func TestPureGoDigisparkAdaptorGetI2cConnection(t *testing.T) {
	a := NewAdaptor()
	err := a.Connect()
	require.NoError(t, err)

	con, err := a.GetI2cConnection(0x40, 0)
	assert.NoError(t, err)
	assert.NotNil(t, con)

	con, err = a.GetI2cConnection(0x40, 1)
	assert.Error(t, err)
	assert.Nil(t, con)
}