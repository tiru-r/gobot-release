//go:build (!cgo || purgo)
// +build !cgo purgo

// Package audio provides pure Go audio playback without external dependencies
package audio

import (
	"errors"
	"time"

	"gobot.io/x/gobot/v2"
)

// PureGoDriver is gobot software device for pure Go audio playback
type PureGoDriver struct {
	name       string
	connection gobot.Connection
	interval   time.Duration
	halt       chan bool
	gobot.Eventer
	gobot.Commander
	filename string
}

// NewPureGoDriver returns a new pure Go audio Driver. It accepts:
//
// *PureGoAdaptor: The pure Go audio adaptor to use for the driver
//
//	string: The filename of the audio to start playing
func NewPureGoDriver(a *PureGoAdaptor, filename string) *PureGoDriver {
	d := &PureGoDriver{
		name:       gobot.DefaultName("PureGoAudio"),
		connection: a,
		interval:   500 * time.Millisecond,
		filename:   filename,
		halt:       make(chan bool),
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
	}

	// Add commander interface methods
	d.AddCommand("play", func(params map[string]interface{}) interface{} {
		return d.Play()
	})

	d.AddCommand("sound", func(params map[string]interface{}) interface{} {
		if filename, ok := params["filename"].(string); ok {
			return d.Sound(filename)
		}
		return []error{errors.New("invalid filename parameter")}
	})

	d.AddCommand("tone", func(params map[string]interface{}) interface{} {
		freq, freqOk := params["frequency"].(float64)
		dur, durOk := params["duration"].(time.Duration)
		if freqOk && durOk {
			return d.GenerateTone(freq, dur)
		}
		return []error{errors.New("invalid tone parameters")}
	})

	return d
}

// Name returns the Driver Name
func (d *PureGoDriver) Name() string { return d.name }

// SetName sets the Driver Name
func (d *PureGoDriver) SetName(n string) { d.name = n }

// Filename returns the file name for the driver to playback
func (d *PureGoDriver) Filename() string { return d.filename }

// SetFilename sets the file name for the driver to playback
func (d *PureGoDriver) SetFilename(filename string) { d.filename = filename }

// Connection returns the Driver Connection
func (d *PureGoDriver) Connection() gobot.Connection {
	return d.connection
}

// Sound plays back a sound file. It accepts:
//
//	string: The filename of the audio to start playing
func (d *PureGoDriver) Sound(fileName string) []error {
	if adaptor, ok := d.Connection().(*PureGoAdaptor); ok {
		return adaptor.Sound(fileName)
	}
	return []error{errors.New("invalid connection type")}
}

// Play plays back the current sound file.
func (d *PureGoDriver) Play() []error {
	return d.Sound(d.Filename())
}

// GenerateTone generates a pure tone for testing
func (d *PureGoDriver) GenerateTone(frequency float64, duration time.Duration) []error {
	if adaptor, ok := d.Connection().(*PureGoAdaptor); ok {
		err := adaptor.GenerateTone(frequency, duration)
		if err != nil {
			return []error{err}
		}
		return nil
	}
	return []error{errors.New("invalid connection type")}
}

// Start starts the Driver
func (d *PureGoDriver) Start() error {
	return nil
}

// Halt halts the Driver
func (d *PureGoDriver) Halt() error {
	close(d.halt)
	return nil
}

// NewDriver is an alias for NewPureGoDriver when using pure Go build
func NewDriver(a gobot.Adaptor, filename string) gobot.Driver {
	if adaptor, ok := a.(*PureGoAdaptor); ok {
		return NewPureGoDriver(adaptor, filename)
	}
	// Fallback to interface-based approach
	return &PureGoDriver{
		name:       gobot.DefaultName("PureGoAudio"),
		connection: a,
		interval:   500 * time.Millisecond,
		filename:   filename,
		halt:       make(chan bool),
		Eventer:    gobot.NewEventer(),
		Commander:  gobot.NewCommander(),
	}
}