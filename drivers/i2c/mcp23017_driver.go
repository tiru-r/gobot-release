package i2c

import (
	"fmt"
	"log"
	"strings"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/pkg/drivers/common/bit"
)

// default address for device when a2/a1/a0 pins are all tied to ground
// please consider special handling for MCP23S17
const mcp23017DefaultAddress = 0x20

const mcp23017Debug = false // toggle debugging information

// port contains all the registers for the device.
type port struct {
	IODIR   uint8 // I/O direction register: 0=output / 1=input
	IPOL    uint8 // input polarity register: 0=normal polarity / 1=inversed
	GPINTEN uint8 // interrupt on change control register: 0=disabled / 1=enabled
	DEFVAL  uint8 // default compare register for interrupt on change
	// interrupt control register: bit set to 0= use defval bit value to compare pin value/ bit set to 1= pin value
	// compared to previous pin value
	INTCON uint8
	IOCON  uint8 // configuration register
	GPPU   uint8 // pull-up resistor configuration register: 0=enabled / 1=disabled
	INTF   uint8 // interrupt flag register: 0=no interrupt / 1=pin caused interrupt
	INTCAP uint8 // interrupt capture register, captures pin values during interrupt: 0=logic low / 1=logic high
	GPIO   uint8 // port register, reading from this register reads the port
	OLAT   uint8 // output latch register, write modifies the pins: 0=logic low / 1=logic high
}

// A bank is made up of PortA and PortB pins.
// Port B pins are on the left side of the chip (starting with pin 1), while port A pins are on the right side.
type bank struct {
	portA port
	portB port
}

// mcp23017Config contains the device configuration for the IOCON register.
// These fields should only be set with values 0 or 1.
type mcp23017Config struct {
	bank   uint8
	mirror uint8
	seqop  uint8
	disslw uint8
	haen   uint8
	odr    uint8
	intpol uint8
}

type mcp23017Behavior struct {
	forceRefresh bool
	autoIODirOff bool
}

// MCP23017Driver contains the driver configuration parameters.
type MCP23017Driver struct {
	*Driver
	mcpConf  mcp23017Config
	mcpBehav mcp23017Behavior
	gobot.Eventer
}

// NewMCP23017Driver creates a new Gobot Driver to the MCP23017 i2c port expander.
// Params:
//
//	c Connector - the Adaptor to use with this Driver
//
// Optional params:
//
//	i2c.WithBus(int):	bus to use with this driver
//	i2c.WithAddress(int):	address to use with this driver
//	i2c.WithMCP23017Bank(int):	MCP23017 bank to use with this driver
//	i2c.WithMCP23017Mirror(int):	MCP23017 mirror to use with this driver
//	i2c.WithMCP23017Seqop(int):	MCP23017 seqop to use with this driver
//	i2c.WithMCP23017Disslw(int):	MCP23017 disslw to use with this driver
//	i2c.WithMCP23017Haen(int):	MCP23017 haen to use with this driver
//	i2c.WithMCP23017Odr(int):	MCP23017 odr to use with this driver
//	i2c.WithMCP23017Intpol(int):	MCP23017 intpol to use with this driver
func NewMCP23017Driver(c Connector, options ...func(Config)) *MCP23017Driver {
	d := &MCP23017Driver{
		Driver:  NewDriver(c, "MCP23017", mcp23017DefaultAddress),
		mcpConf: mcp23017Config{},
		Eventer: gobot.NewEventer(),
	}
	d.afterStart = d.initialize

	for _, option := range options {
		option(d)
	}

	//nolint:forcetypeassert // ok here
	d.AddCommand("WriteGPIO", func(params map[string]interface{}) interface{} {
		pin := params["pin"].(uint8)
		port := params["port"].(string)
		val := params["val"].(uint8)
		err := d.WriteGPIO(pin, port, val)
		return map[string]interface{}{"err": err}
	})

	//nolint:forcetypeassert // ok here
	d.AddCommand("ReadGPIO", func(params map[string]interface{}) interface{} {
		pin := params["pin"].(uint8)
		port := params["port"].(string)
		val, err := d.ReadGPIO(pin, port)
		return map[string]interface{}{"val": val, "err": err}
	})

	return d
}

// WithMCP23017Bank option sets the MCP23017Driver bank option
func WithMCP23017Bank(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.bank = val
		} else if mcp23017Debug {
			log.Printf("trying to set bank for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Mirror option sets the MCP23017Driver mirror option
func WithMCP23017Mirror(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.mirror = val
		} else if mcp23017Debug {
			log.Printf("Trying to set mirror for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Seqop option sets the MCP23017Driver seqop option
func WithMCP23017Seqop(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.seqop = val
		} else if mcp23017Debug {
			log.Printf("Trying to set seqop for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Disslw option sets the MCP23017Driver disslw option
func WithMCP23017Disslw(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.disslw = val
		} else if mcp23017Debug {
			log.Printf("Trying to set disslw for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Haen option sets the MCP23017Driver haen option
// This feature is only available for MCP23S17, because address pins are always enabled on the MCP23017.
func WithMCP23017Haen(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.haen = val
		} else if mcp23017Debug {
			log.Printf("Trying to set haen for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Odr option sets the MCP23017Driver odr option
func WithMCP23017Odr(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.odr = val
		} else if mcp23017Debug {
			log.Printf("Trying to set odr for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017Intpol option sets the MCP23017Driver intpol option
func WithMCP23017Intpol(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpConf.intpol = val
		} else if mcp23017Debug {
			log.Printf("Trying to set intpol for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017ForceRefresh option modifies the MCP23017Driver forceRefresh option
// Setting to true (1) will force refresh operation to register, although there is no change.
// Normally this is not needed, so default is off (0).
// When there is something flaky, there is a small chance to stabilize by setting this flag to true.
// However, setting this flag to true slows down each IO operation up to 100%.
func WithMCP23017ForceRefresh(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpBehav.forceRefresh = val > 0
		} else if mcp23017Debug {
			log.Printf("Trying to set forceRefresh for non-MCP23017Driver %v", c)
		}
	}
}

// WithMCP23017AutoIODirOff option modifies the MCP23017Driver autoIODirOff option
// Set IO direction at each read or write operation ensures the correct direction, which is the default setting.
// Most hardware is configured statically, so this can avoided by setting the direction using SetPinMode(),
// e.g. in the start up sequence. If this way is taken, the automatic set of direction at each call can
// be safely deactivated with this flag (set to true, 1).
// This will speedup each WriteGPIO by 50% and each ReadGPIO by 60%.
func WithMCP23017AutoIODirOff(val uint8) func(Config) {
	return func(c Config) {
		d, ok := c.(*MCP23017Driver)
		if ok {
			d.mcpBehav.autoIODirOff = val > 0
		} else if mcp23017Debug {
			log.Printf("Trying to set autoIODirOff for non-MCP23017Driver %v", c)
		}
	}
}

// SetPinMode set pin mode of a given pin immediately, based on the value:
// val = 0 output
// val = 1 input
func (m *MCP23017Driver) SetPinMode(pin uint8, portStr string, val uint8) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	selectedPort := m.getPort(portStr)
	// Set IODIR register bit for given pin to an output/input.
	return m.write(selectedPort.IODIR, pin, bitState(val))
}

// SetPullUp sets the pull up state of a given pin immediately, based on the value:
// val = 1 pull up enabled.
// val = 0 pull up disabled.
func (m *MCP23017Driver) SetPullUp(pin uint8, portStr string, val uint8) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	selectedPort := m.getPort(portStr)
	return m.write(selectedPort.GPPU, pin, bitState(val))
}

// SetGPIOPolarity will change a given pin's polarity immediately, based on the value:
// val = 1 opposite logic state of the input pin.
// val = 0 same logic state of the input pin.
func (m *MCP23017Driver) SetGPIOPolarity(pin uint8, portStr string, val uint8) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	selectedPort := m.getPort(portStr)
	return m.write(selectedPort.IPOL, pin, bitState(val))
}

// WriteGPIO writes a value to a gpio pin (0-7) and a port (A or B).
func (m *MCP23017Driver) WriteGPIO(pin uint8, portStr string, val uint8) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	selectedPort := m.getPort(portStr)
	if !m.mcpBehav.autoIODirOff {
		// Set IODIR register bit for given pin to an output by clearing bit.
		// can't call SetPinMode() because mutex will cause deadlock
		if err := m.write(selectedPort.IODIR, pin, clearBit); err != nil {
			return err
		}
	}
	// write value to OLAT register bit
	return m.write(selectedPort.OLAT, pin, bitState(val))
}

// ReadGPIO reads a value from a given gpio pin (0-7) and a port (A or B).
func (m *MCP23017Driver) ReadGPIO(pin uint8, portStr string) (uint8, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	selectedPort := m.getPort(portStr)
	if !m.mcpBehav.autoIODirOff {
		// Set IODIR register bit for given pin to an input by set bit.
		// can't call SetPinMode() because mutex will cause deadlock
		if err := m.write(selectedPort.IODIR, pin, setBit); err != nil {
			return 0, err
		}
	}
	val, err := m.read(selectedPort.GPIO)
	if err != nil {
		return val, err
	}
	val = 1 << pin & val
	if val > 1 {
		val = 1
	}
	return val, nil
}

func (m *MCP23017Driver) initialize() error {
	// Set IOCON register with MCP23017 configuration.
	ioconReg := m.getPort("A").IOCON // IOCON address is the same for Port A or B.
	ioconVal := m.mcpConf.getUint8Value()

	_, err := m.connection.Write([]uint8{ioconReg, ioconVal})
	return err
}

// write gets the value of the passed in register, and then sets the bit specified
// by the pin to the given state.
func (m *MCP23017Driver) write(reg uint8, pin uint8, state bitState) error {
	valOrg, err := m.read(reg)
	if err != nil {
		return fmt.Errorf("MCP write-read: %v", err)
	}

	var val uint8
	if state == clearBit {
		val = uint8(bit.Clear(int(valOrg), pin)) //nolint:gosec // TODO: fix later
	} else {
		val = uint8(bit.Set(int(valOrg), pin)) //nolint:gosec // TODO: fix later
	}

	if val != valOrg || m.mcpBehav.forceRefresh {
		if mcp23017Debug {
			log.Printf("write done: MCP forceRefresh: %t, address: 0x%X, register: 0x%X, name: %s, value: 0x%X\n",
				m.mcpBehav.forceRefresh, m.GetAddressOrDefault(mcp23017DefaultAddress), reg, m.getRegName(reg), val)
		}
		if err := m.connection.WriteByteData(reg, val); err != nil {
			return fmt.Errorf("MCP write-WriteByteData(reg=%d,val=%d): %v", reg, val, err)
		}
	} else if mcp23017Debug {
		log.Printf("write skipped: MCP forceRefresh: %t, address: 0x%X, register: 0x%X, name: %s, value: 0x%X\n",
			m.mcpBehav.forceRefresh, m.GetAddressOrDefault(mcp23017DefaultAddress), reg, m.getRegName(reg), val)
	}
	return nil
}

// read get the data from a given register
// it is mainly a wrapper to create additional debug messages, when activated
func (m *MCP23017Driver) read(reg uint8) (uint8, error) {
	val, err := m.connection.ReadByteData(reg)
	if err != nil {
		return val, fmt.Errorf("MCP write-ReadByteData(reg=%d): %v", reg, err)
	}
	if mcp23017Debug {
		log.Printf("reading done: MCP autoIODirOff: %t, address: 0x%X, register:0x%X, name: %s, value: 0x%X\n",
			m.mcpBehav.autoIODirOff, m.GetAddressOrDefault(mcp23017DefaultAddress), reg, m.getRegName(reg), val)
	}
	return val, nil
}

// getPort return the port (A or B) given a string and the bank.
// Port A is the default if an incorrect or no port is specified.
func (m *MCP23017Driver) getPort(portStr string) port {
	portStr = strings.ToUpper(portStr)
	switch {
	case portStr == "A":
		return mcp23017GetBank(m.mcpConf.bank).portA
	case portStr == "B":
		return mcp23017GetBank(m.mcpConf.bank).portB
	default:
		return mcp23017GetBank(m.mcpConf.bank).portA
	}
}

// getUint8Value returns the configuration data as a packed value.
func (mc *mcp23017Config) getUint8Value() uint8 {
	return mc.bank<<7 | mc.mirror<<6 | mc.seqop<<5 | mc.disslw<<4 | mc.haen<<3 | mc.odr<<2 | mc.intpol<<1
}

// getRegName returns the name of the given register related to the configured bank
// and can be used to write nice debug messages
func (m *MCP23017Driver) getRegName(reg uint8) string {
	b := mcp23017GetBank(m.mcpConf.bank)
	portStr := "A"
	regStr := "unknown"

	for i := 1; i <= 2; i++ {
		if regStr == "unknown" {
			p := b.portA
			if i == 2 {
				p = b.portB
				portStr = "B"
			}
			switch reg {
			case p.IODIR:
				regStr = "IODIR"
			case p.IPOL:
				regStr = "IPOL"
			case p.GPINTEN:
				regStr = "GPINTEN"
			case p.DEFVAL:
				regStr = "DEFVAL"
			case p.INTCON:
				regStr = "INTCON"
			case p.IOCON:
				regStr = "IOCON"
			case p.GPPU:
				regStr = "GPPU"
			case p.INTF:
				regStr = "INTF"
			case p.INTCAP:
				regStr = "INTCAP"
			case p.GPIO:
				regStr = "GPIO"
			case p.OLAT:
				regStr = "OLAT"
			}
		}
	}

	return fmt.Sprintf("%s_%s", regStr, portStr)
}

// mcp23017GetBank returns a bank's PortA and PortB registers given a bank number (0/1).
func mcp23017GetBank(bnk uint8) bank {
	if bnk == 0 {
		return bank{
			portA: port{0x00, 0x02, 0x04, 0x06, 0x08, 0x0A, 0x0C, 0x0E, 0x10, 0x12, 0x14},
			portB: port{0x01, 0x03, 0x05, 0x07, 0x09, 0x0B, 0x0D, 0x0F, 0x11, 0x13, 0x15},
		}
	}
	return bank{
		portA: port{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A},
		portB: port{0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1A},
	}
}
