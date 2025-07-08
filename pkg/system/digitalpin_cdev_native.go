package system

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
	"unsafe"

	"gobot.io/x/gobot/v2"
)

// Native Go 1.24 implementation of GPIO character device interface
// This replaces the github.com/warthog618/go-gpiocdev dependency

const systemCdevNativeDebug = false

// GPIO character device ioctl commands and structures
const (
	// GPIO chip info ioctl
	_GPIO_GET_CHIPINFO_IOCTL = 0x8044b401

	// GPIO line info ioctl  
	_GPIO_GET_LINEINFO_IOCTL = 0xc048b402

	// GPIO line request ioctl (v2 API)
	_GPIO_V2_GET_LINE_IOCTL = 0xc100b407

	// GPIO line value get/set ioctls
	_GPIO_V2_LINE_GET_VALUES_IOCTL = 0xc010b40e
	_GPIO_V2_LINE_SET_VALUES_IOCTL = 0xc010b40f
)

// GPIO line flags
const (
	_GPIO_V2_LINE_FLAG_USED         = 1 << 0
	_GPIO_V2_LINE_FLAG_ACTIVE_LOW   = 1 << 1
	_GPIO_V2_LINE_FLAG_INPUT        = 1 << 2
	_GPIO_V2_LINE_FLAG_OUTPUT       = 1 << 3
	_GPIO_V2_LINE_FLAG_EDGE_RISING  = 1 << 4
	_GPIO_V2_LINE_FLAG_EDGE_FALLING = 1 << 5
	_GPIO_V2_LINE_FLAG_OPEN_DRAIN   = 1 << 6
	_GPIO_V2_LINE_FLAG_OPEN_SOURCE  = 1 << 7
	_GPIO_V2_LINE_FLAG_BIAS_PULL_UP = 1 << 8
	_GPIO_V2_LINE_FLAG_BIAS_PULL_DOWN = 1 << 9
	_GPIO_V2_LINE_FLAG_BIAS_DISABLED  = 1 << 10
)

// Structs for GPIO ioctl operations
type gpioChipInfo struct {
	Name  [32]byte
	Label [32]byte
	Lines uint32
}

type gpioLineInfo struct {
	Offset uint32
	Name   [32]byte
	Consumer [32]byte
	Flags  uint64
	_      [4]uint64 // padding for future use
}

type gpioV2LineRequest struct {
	Offsets    [64]uint32
	Consumer   [32]byte
	Config     gpioV2LineConfig
	NumLines   uint32
	Fd         int32
	_          [5]uint32 // padding
}

type gpioV2LineConfig struct {
	Flags        uint64
	OutputValues uint64
	NumAttrs     uint32
	_            uint32 // padding
	Attrs        [10]gpioV2LineConfigAttribute
}

type gpioV2LineConfigAttribute struct {
	Attr gpioV2LineAttribute
	Mask uint64
}

type gpioV2LineAttribute struct {
	ID   uint32
	_    uint32 // padding
	Data [8]byte
}

type gpioV2LineValues struct {
	Bits uint64
	Mask uint64
}

// Native GPIO chip
type nativeGpioChip struct {
	name string
	fd   *os.File
	info gpioChipInfo
}

// Native GPIO line
type nativeGpioLine struct {
	chip   *nativeGpioChip
	offset uint32
	fd     *os.File
	config *digitalPinConfig
}

// Native digital pin implementation
type digitalPinCdevNative struct {
	chipName string
	pin      int
	*digitalPinConfig
	line *nativeGpioLine
}

// Interface compliance check
var _ cdevLine = (*nativeGpioLine)(nil)

// NewNativeChip opens a GPIO chip character device
func newNativeChip(chipName string, consumer string) (*nativeGpioChip, error) {
	if chipName == "" {
		chipName = "gpiochip0"
	}
	
	devicePath := "/dev/" + chipName
	fd, err := os.OpenFile(devicePath, os.O_RDWR, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open %s: %v", devicePath, err)
	}

	chip := &nativeGpioChip{
		name: chipName,
		fd:   fd,
	}

	// Get chip info
	if err := chip.getChipInfo(); err != nil {
		fd.Close()
		return nil, fmt.Errorf("failed to get chip info: %v", err)
	}

	return chip, nil
}

func (c *nativeGpioChip) Close() error {
	if c.fd != nil {
		return c.fd.Close()
	}
	return nil
}

func (c *nativeGpioChip) getChipInfo() error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, 
		uintptr(c.fd.Fd()), 
		_GPIO_GET_CHIPINFO_IOCTL, 
		uintptr(unsafe.Pointer(&c.info)))
	if errno != 0 {
		return errno
	}
	return nil
}

func (c *nativeGpioChip) requestLine(offset uint32, config *digitalPinConfig) (*nativeGpioLine, error) {
	var req gpioV2LineRequest
	
	// Set line offset
	req.Offsets[0] = offset
	req.NumLines = 1
	
	// Set consumer label
	copy(req.Consumer[:], config.label)
	
	// Configure line flags
	var flags uint64
	
	if config.direction == IN {
		flags |= _GPIO_V2_LINE_FLAG_INPUT
	} else {
		flags |= _GPIO_V2_LINE_FLAG_OUTPUT
		if config.outInitialState != 0 {
			req.Config.OutputValues |= 1
		}
	}
	
	if config.activeLow {
		flags |= _GPIO_V2_LINE_FLAG_ACTIVE_LOW
	}
	
	// Set bias
	switch config.bias {
	case digitalPinBiasPullUp:
		flags |= _GPIO_V2_LINE_FLAG_BIAS_PULL_UP
	case digitalPinBiasPullDown:
		flags |= _GPIO_V2_LINE_FLAG_BIAS_PULL_DOWN
	case digitalPinBiasDisable:
		flags |= _GPIO_V2_LINE_FLAG_BIAS_DISABLED
	}
	
	// Set drive mode for outputs
	if config.direction == OUT {
		switch config.drive {
		case digitalPinDriveOpenDrain:
			flags |= _GPIO_V2_LINE_FLAG_OPEN_DRAIN
		case digitalPinDriveOpenSource:
			flags |= _GPIO_V2_LINE_FLAG_OPEN_SOURCE
		}
	}
	
	// Set edge detection for inputs
	if config.direction == IN {
		switch config.edge {
		case digitalPinEventOnRisingEdge:
			flags |= _GPIO_V2_LINE_FLAG_EDGE_RISING
		case digitalPinEventOnFallingEdge:
			flags |= _GPIO_V2_LINE_FLAG_EDGE_FALLING
		case digitalPinEventOnBothEdges:
			flags |= _GPIO_V2_LINE_FLAG_EDGE_RISING | _GPIO_V2_LINE_FLAG_EDGE_FALLING
		}
	}
	
	req.Config.Flags = flags
	
	// Request the line
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(c.fd.Fd()),
		_GPIO_V2_GET_LINE_IOCTL,
		uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		return nil, errno
	}
	
	// Create file descriptor for the line
	lineFd := os.NewFile(uintptr(req.Fd), fmt.Sprintf("gpio-line-%d", offset))
	if lineFd == nil {
		return nil, errors.New("failed to create line file descriptor")
	}
	
	line := &nativeGpioLine{
		chip:   c,
		offset: offset,
		fd:     lineFd,
		config: config,
	}
	
	return line, nil
}

// Native GPIO line methods
func (l *nativeGpioLine) SetValue(value int) error {
	var values gpioV2LineValues
	values.Mask = 1 // Set mask for line 0
	if value != 0 {
		values.Bits = 1
	}
	
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(l.fd.Fd()),
		_GPIO_V2_LINE_SET_VALUES_IOCTL,
		uintptr(unsafe.Pointer(&values)))
	if errno != 0 {
		return errno
	}
	return nil
}

func (l *nativeGpioLine) Value() (int, error) {
	var values gpioV2LineValues
	values.Mask = 1 // Set mask for line 0
	
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(l.fd.Fd()),
		_GPIO_V2_LINE_GET_VALUES_IOCTL,
		uintptr(unsafe.Pointer(&values)))
	if errno != 0 {
		return 0, errno
	}
	
	if values.Bits&1 != 0 {
		return 1, nil
	}
	return 0, nil
}

func (l *nativeGpioLine) Close() error {
	if l.fd != nil {
		return l.fd.Close()
	}
	return nil
}

// Native digital pin implementation
func newDigitalPinCdevNative(chipName string, pin int, options ...func(gobot.DigitalPinOptioner) bool) *digitalPinCdevNative {
	if chipName == "" {
		chipName = "gpiochip0"
	}
	cfg := newDigitalPinConfig("gobotio"+strconv.Itoa(pin), options...)
	d := &digitalPinCdevNative{
		chipName:         chipName,
		pin:              pin,
		digitalPinConfig: cfg,
	}
	return d
}

// ApplyOptions apply all given options to the pin immediately
func (d *digitalPinCdevNative) ApplyOptions(options ...func(gobot.DigitalPinOptioner) bool) error {
	if d == nil {
		return errors.New("native pin is nil")
	}
	anyChange := false
	for _, option := range options {
		anyChange = option(d) || anyChange
	}
	if anyChange {
		return d.reconfigure(false)
	}
	return nil
}

// DirectionBehavior gets the direction behavior when the pin is used the next time
func (d *digitalPinCdevNative) DirectionBehavior() string {
	return d.direction
}

// Export sets the pin as used by this driver
func (d *digitalPinCdevNative) Export() error {
	err := d.reconfigure(false)
	if err != nil {
		return fmt.Errorf("native cdev.Export(): %v", err)
	}
	return nil
}

// Unexport releases the pin as input
func (d *digitalPinCdevNative) Unexport() error {
	var errs []string
	if d.line != nil {
		if err := d.reconfigure(true); err != nil {
			errs = append(errs, err.Error())
		}
		if err := d.line.Close(); err != nil {
			err = fmt.Errorf("native cdev.Unexport()-line.Close(): %v", err)
			errs = append(errs, err.Error())
		}
	}
	if len(errs) == 0 {
		return nil
	}
	return errors.New(strings.Join(errs, ","))
}

// Write writes the given value to the character device
func (d *digitalPinCdevNative) Write(val int) error {
	if val < 0 {
		val = 0
	}
	if val > 1 {
		val = 1
	}

	err := d.line.SetValue(val)
	if err != nil {
		return fmt.Errorf("native cdev.Write(): %v", err)
	}
	return nil
}

// Read reads the given value from character device
func (d *digitalPinCdevNative) Read() (int, error) {
	val, err := d.line.Value()
	if err != nil {
		return 0, fmt.Errorf("native cdev.Read(): %v", err)
	}
	return val, err
}

func (d *digitalPinCdevNative) reconfigure(forceInput bool) error {
	if d == nil {
		return errors.New("native pin is nil")
	}
	// cleanup old line
	if d.line != nil {
		d.line.Close()
	}
	d.line = nil

	// open chip
	chip, err := newNativeChip(d.chipName, d.label)
	if err != nil {
		return fmt.Errorf("native cdev.reconfigure()-newNativeChip(%s): %v", d.chipName, err)
	}
	defer chip.Close()

	// configure direction if forcing input
	if forceInput {
		d.direction = IN
	}

	// request line with configuration
	line, err := chip.requestLine(uint32(d.pin), d.digitalPinConfig)
	if err != nil {
		return fmt.Errorf("native cdev.reconfigure()-requestLine(%d): %v", d.pin, err)
	}
	
	d.line = line

	// start discrete polling function if configured
	if (d.direction == IN || forceInput) && d.pollInterval > 0 {
		if err := startEdgePolling(d.label, d.Read, d.pollInterval, d.edge, d.edgeEventHandler,
			d.pollQuitChan); err != nil {
			return err
		}
	}

	return nil
}

// Create alias for the new native implementation
var digitalPinCdevReconfigureNative = func(d *digitalPinCdevNative, forceInput bool) error {
	return d.reconfigure(forceInput)
}