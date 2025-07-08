package joystick

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"gobot.io/x/gobot/v2"
)

// JoystickEvent represents a joystick input event
type JoystickEvent struct {
	Time   uint32 // event timestamp in milliseconds
	Value  int16  // event value
	Type   uint8  // event type
	Number uint8  // axis/button number
}

// State represents joystick state
type State struct {
	Buttons  uint32 // button state bitmask
	AxisData []int  // axis values
}

// Joystick represents a joystick interface
type Joystick interface {
	Read() (State, error)
	ButtonCount() int
	AxisCount() int
	Name() string
	Close() error
}

// joystick implements the Joystick interface using Linux joystick API
type joystick struct {
	file        *os.File
	buttonCount int
	axisCount   int
	name        string
	buttonState uint32
	axisState   []int
}

// Open opens a joystick device
func Open(id int) (Joystick, error) {
	devicePath := fmt.Sprintf("/dev/input/js%d", id)
	file, err := os.OpenFile(devicePath, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to open joystick device %s: %w", devicePath, err)
	}

	j := &joystick{
		file: file,
	}

	// Get joystick information using ioctl
	if err := j.getJoystickInfo(); err != nil {
		file.Close()
		return nil, fmt.Errorf("failed to get joystick info: %w", err)
	}

	j.axisState = make([]int, j.axisCount)
	return j, nil
}

// getJoystickInfo retrieves joystick capabilities using ioctl
func (j *joystick) getJoystickInfo() error {
	// JSIOCGAXES - get number of axes
	const JSIOCGAXES = 0x80016a11
	// JSIOCGBUTTONS - get number of buttons
	const JSIOCGBUTTONS = 0x80016a12
	// JSIOCGNAME - get joystick name
	const JSIOCGNAME = 0x80016a13

	// Get number of axes
	var axes uint8
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(j.file.Fd()), JSIOCGAXES, uintptr(unsafe.Pointer(&axes)))
	if errno != 0 {
		return fmt.Errorf("ioctl JSIOCGAXES failed: %v", errno)
	}
	j.axisCount = int(axes)

	// Get number of buttons
	var buttons uint8
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(j.file.Fd()), JSIOCGBUTTONS, uintptr(unsafe.Pointer(&buttons)))
	if errno != 0 {
		return fmt.Errorf("ioctl JSIOCGBUTTONS failed: %v", errno)
	}
	j.buttonCount = int(buttons)

	// Get joystick name
	nameBuf := make([]byte, 128)
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(j.file.Fd()), JSIOCGNAME, uintptr(unsafe.Pointer(&nameBuf[0])))
	if errno != 0 {
		j.name = "Unknown"
	} else {
		// Find null terminator
		nullIndex := 0
		for i, b := range nameBuf {
			if b == 0 {
				nullIndex = i
				break
			}
		}
		j.name = string(nameBuf[:nullIndex])
	}

	return nil
}

// Read reads the current joystick state
func (j *joystick) Read() (State, error) {
	// Read joystick events
	for {
		var event JoystickEvent
		err := binary.Read(j.file, binary.LittleEndian, &event)
		if err != nil {
			if err == io.EOF {
				break
			}
			return State{}, fmt.Errorf("failed to read joystick event: %w", err)
		}

		// Handle different event types
		switch event.Type & 0x7F { // Mask out the init bit
		case 0x01: // Button event
			if event.Number < 32 { // Support up to 32 buttons
				if event.Value != 0 {
					j.buttonState |= 1 << event.Number
				} else {
					j.buttonState &^= 1 << event.Number
				}
			}
		case 0x02: // Axis event
			if int(event.Number) < len(j.axisState) {
				j.axisState[event.Number] = int(event.Value)
			}
		}

		// If this was an init event, continue reading
		if event.Type&0x80 != 0 {
			continue
		}

		// Return current state after processing one non-init event
		break
	}

	return State{
		Buttons:  j.buttonState,
		AxisData: append([]int(nil), j.axisState...), // Copy slice
	}, nil
}

// ButtonCount returns the number of buttons
func (j *joystick) ButtonCount() int {
	return j.buttonCount
}

// AxisCount returns the number of axes
func (j *joystick) AxisCount() int {
	return j.axisCount
}

// Name returns the joystick name
func (j *joystick) Name() string {
	return j.name
}

// Close closes the joystick device
func (j *joystick) Close() error {
	return j.file.Close()
}

// Adaptor represents a connection to a joystick
type Adaptor struct {
	name     string
	id       string
	joystick Joystick
	connect  func(*Adaptor) error
}

// NewAdaptor returns a new Joystick Adaptor.
// Pass in the ID of the joystick you wish to connect to.
func NewAdaptor(id string) *Adaptor {
	return &Adaptor{
		name: gobot.DefaultName("Joystick"),
		connect: func(j *Adaptor) error {
			i, err := strconv.Atoi(id)
			if err != nil {
				return fmt.Errorf("invalid joystick ID: %v", err)
			}

			joy, err := Open(i)
			if err != nil {
				return fmt.Errorf("no joystick available: %v", err)
			}

			j.id = id
			j.joystick = joy
			return nil
		},
	}
}

// Name returns the adaptors name
func (j *Adaptor) Name() string { return j.name }

// SetName sets the adaptors name
func (j *Adaptor) SetName(n string) { j.name = n }

// Connect connects to the joystick
func (j *Adaptor) Connect() error {
	return j.connect(j)
}

// Finalize closes connection to joystick
func (j *Adaptor) Finalize() error {
	j.joystick.Close()
	return nil
}
