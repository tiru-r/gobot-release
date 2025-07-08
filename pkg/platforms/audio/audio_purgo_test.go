//go:build purgo
// +build purgo

package audio

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPureGoAudioAdaptor(t *testing.T) {
	a := NewAdaptor()
	adaptor, ok := a.(*PureGoAdaptor)
	require.True(t, ok, "NewAdaptor should return *PureGoAdaptor when using purgo build tag")
	
	// Test basic adaptor functionality
	assert.Contains(t, adaptor.Name(), "PureGoAudio")
	
	adaptor.SetName("TestAudio")
	assert.Equal(t, "TestAudio", adaptor.Name())
	
	// Test connection
	err := adaptor.Connect()
	assert.NoError(t, err)
	
	// Test finalize
	err = adaptor.Finalize()
	assert.NoError(t, err)
}

func TestPureGoAudioDriver(t *testing.T) {
	a := NewAdaptor()
	adaptor, ok := a.(*PureGoAdaptor)
	require.True(t, ok)
	
	d := NewDriver(a, "test.wav")
	driver, ok := d.(*PureGoDriver)
	require.True(t, ok, "NewDriver should return *PureGoDriver when using purgo build tag")
	
	// Test basic driver functionality
	assert.Contains(t, driver.Name(), "PureGoAudio")
	assert.Equal(t, "test.wav", driver.Filename())
	
	driver.SetName("TestDriver")
	assert.Equal(t, "TestDriver", driver.Name())
	
	driver.SetFilename("test2.wav")
	assert.Equal(t, "test2.wav", driver.Filename())
	
	assert.Equal(t, adaptor, driver.Connection())
	
	// Test start and halt
	err := driver.Start()
	assert.NoError(t, err)
	
	err = driver.Halt()
	assert.NoError(t, err)
}

func TestPureGoAudioSound(t *testing.T) {
	a := NewAdaptor()
	adaptor, ok := a.(*PureGoAdaptor)
	require.True(t, ok)
	
	err := adaptor.Connect()
	require.NoError(t, err)
	defer adaptor.Finalize()
	
	// Test empty filename
	errors := adaptor.Sound("")
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "requires filename")
	
	// Test non-existent file
	errors = adaptor.Sound("nonexistent.wav")
	assert.Len(t, errors, 1)
	
	// Test unsupported format
	tmpFile, err := os.CreateTemp("", "test*.mp4")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()
	
	errors = adaptor.Sound(tmpFile.Name())
	assert.Len(t, errors, 1)
	assert.Contains(t, errors[0].Error(), "unsupported audio format")
}

func TestPureGoAudioGenerateTone(t *testing.T) {
	a := NewAdaptor()
	adaptor, ok := a.(*PureGoAdaptor)
	require.True(t, ok)
	
	err := adaptor.Connect()
	require.NoError(t, err)
	defer adaptor.Finalize()
	
	// Test tone generation
	err = adaptor.GenerateTone(440.0, 50*time.Millisecond)
	assert.NoError(t, err)
	
	// Test tone generation before initialization
	adaptor.Finalize()
	err = adaptor.GenerateTone(440.0, 50*time.Millisecond)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not initialized")
}

func TestPureGoDriverCommands(t *testing.T) {
	a := NewAdaptor()
	adaptor, ok := a.(*PureGoAdaptor)
	require.True(t, ok)
	
	err := adaptor.Connect()
	require.NoError(t, err)
	defer adaptor.Finalize()
	
	d := NewDriver(a, "test.wav")
	driver, ok := d.(*PureGoDriver)
	require.True(t, ok)
	
	// Test play command
	result := driver.Command("play")(map[string]interface{}{})
	errors, ok := result.([]error)
	assert.True(t, ok)
	assert.Len(t, errors, 1) // Should fail because file doesn't exist
	
	// Test tone command
	result = driver.Command("tone")(map[string]interface{}{
		"frequency": 440.0,
		"duration":  50 * time.Millisecond,
	})
	errors, ok = result.([]error)
	assert.True(t, ok)
	assert.Len(t, errors, 0)
	
	// Test invalid tone command
	result = driver.Command("tone")(map[string]interface{}{
		"frequency": "invalid",
	})
	errors, ok = result.([]error)
	assert.True(t, ok)
	assert.Len(t, errors, 1)
}