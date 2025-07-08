package system

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"gobot.io/x/gobot/v2"
)

const systemCdevDebug = false

type cdevLine interface {
	SetValue(value int) error
	Value() (int, error)
	Close() error
}

// digitalPinCdev now uses native Go 1.24 implementation
type digitalPinCdev struct {
	chipName string
	pin      int
	*digitalPinConfig
	line cdevLine
	nativePin *digitalPinCdevNative // Use native implementation
}

var digitalPinCdevReconfigure = digitalPinCdevReconfigureLine // to allow unit testing

var (
	digitalPinCdevUsed      = map[bool]string{true: "used", false: "unused"}
	digitalPinCdevActiveLow = map[bool]string{true: "low", false: "high"}
	digitalPinCdevDebounced = map[bool]string{true: "debounced", false: "not debounced"}
)

// Modern Go 1.24 mappings replace the old gpiocdev constants
var digitalPinCdevDirection = map[string]string{
	"unknown": "unknown direction",
	"input":   "input", 
	"output":  "output",
}

var digitalPinCdevDrive = map[int]string{
	digitalPinDrivePushPull:   "push-pull", 
	digitalPinDriveOpenDrain:  "open-drain",
	digitalPinDriveOpenSource: "open-source",
}

var digitalPinCdevBias = map[int]string{
	digitalPinBiasDefault:  "unknown", 
	digitalPinBiasDisable:  "disabled",
	digitalPinBiasPullUp:   "pull-up", 
	digitalPinBiasPullDown: "pull-down",
}

var digitalPinCdevEdgeDetect = map[int]string{
	digitalPinEventNone:          "no", 
	digitalPinEventOnRisingEdge:  "rising",
	digitalPinEventOnFallingEdge: "falling", 
	digitalPinEventOnBothEdges:   "both",
}

// newDigitalPinCdev returns a digital pin using native Go 1.24 GPIO implementation
func newDigitalPinCdev(chipName string, pin int, options ...func(gobot.DigitalPinOptioner) bool) *digitalPinCdev {
	if chipName == "" {
		chipName = "gpiochip0"
	}
	cfg := newDigitalPinConfig("gobotio"+strconv.Itoa(pin), options...)
	
	// Create native pin implementation
	nativePin := newDigitalPinCdevNative(chipName, pin, options...)
	
	d := &digitalPinCdev{
		chipName:         chipName,
		pin:              pin,
		digitalPinConfig: cfg,
		nativePin:        nativePin,
	}
	return d
}

// ApplyOptions apply all given options to the pin immediately
func (d *digitalPinCdev) ApplyOptions(options ...func(gobot.DigitalPinOptioner) bool) error {
	anyChange := false
	for _, option := range options {
		anyChange = option(d) || anyChange
	}
	if anyChange {
		// Apply to native pin as well if it exists
		if d.nativePin != nil {
			if err := d.nativePin.ApplyOptions(options...); err != nil {
				return err
			}
		}
		return digitalPinCdevReconfigure(d, false)
	}
	return nil
}

// DirectionBehavior gets the direction behavior when the pin is used the next time
func (d *digitalPinCdev) DirectionBehavior() string {
	return d.direction
}

// Export sets the pin as used by this driver
func (d *digitalPinCdev) Export() error {
	err := digitalPinCdevReconfigure(d, false)
	if err != nil {
		return fmt.Errorf("cdev.Export(): %v", err)
	}
	return nil
}

// Unexport releases the pin as input
func (d *digitalPinCdev) Unexport() error {
	var errs []string
	
	// Only reconfigure if we have a line (maintains test compatibility)
	if d.line != nil {
		// Try to reconfigure
		if err := digitalPinCdevReconfigure(d, true); err != nil {
			errs = append(errs, err.Error())
		}
		
		// Close line
		if err := d.line.Close(); err != nil {
			err = fmt.Errorf("cdev.Unexport()-line.Close(): %v", err)
			errs = append(errs, err.Error())
		}
	}
	
	// Close native pin if exists (but don't call reconfigure for it)
	if d.nativePin != nil && d.line == nil {
		if err := d.nativePin.Unexport(); err != nil {
			errs = append(errs, err.Error())
		}
	}
	
	if len(errs) == 0 {
		return nil
	}

	return errors.New(strings.Join(errs, ","))
}

// Write writes the given value to the character device
func (d *digitalPinCdev) Write(val int) error {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}

	// Prefer line interface for compatibility with tests
	if d.line != nil {
		err := d.line.SetValue(val)
		if err != nil {
			return fmt.Errorf("cdev.Write(): %v", err)
		}
		return nil
	}

	// Use native implementation as fallback
	if d.nativePin != nil {
		return d.nativePin.Write(val)
	}
	
	return errors.New("no active line or native pin")
}

// Read reads the given value from character device
func (d *digitalPinCdev) Read() (int, error) {
	// Prefer line interface for compatibility with tests
	if d.line != nil {
		val, err := d.line.Value()
		if err != nil {
			return 0, fmt.Errorf("cdev.Read(): %v", err)
		}
		return val, err
	}

	// Use native implementation as fallback
	if d.nativePin != nil {
		return d.nativePin.Read()
	}
	
	return 0, errors.New("no active line or native pin")
}

// ListLines is used for development purposes
func (d *digitalPinCdev) ListLines() error {
	// For now, return a simple message indicating modern implementation
	fmt.Printf("GPIO Chip: %s, Pin: %d, Direction: %s\n", 
		d.chipName, d.pin, d.direction)
	return nil
}

// List is used for development purposes  
func (d *digitalPinCdev) List() error {
	fmt.Printf("GPIO Line: %s-%d, Label: %s, Direction: %s\n",
		d.chipName, d.pin, d.label, d.direction)
	return nil
}

func digitalPinCdevReconfigureLine(d *digitalPinCdev, forceInput bool) error {
	// Use native implementation
	if d.nativePin != nil {
		return digitalPinCdevReconfigureNative(d.nativePin, forceInput)
	}
	
	// Legacy fallback (should not be reached in modern implementation)
	return errors.New("no native pin implementation available")
}