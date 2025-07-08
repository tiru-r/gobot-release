package joule

import (
	"fmt"
	"sync"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/pkg/platforms/adaptors"
	"gobot.io/x/gobot/v2/pkg/system"
)

const (
	defaultI2cBusNumber = 0

	defaultSpiMaxSpeed = 5000 // 5kHz (more than 15kHz not possible with SPI over GPIO)
)

type sysfsPin struct {
	pin    int
	pwmPin int
}

// Adaptor represents an Intel Joule
type Adaptor struct {
	name  string
	sys   *system.Accesser // used for unit tests only
	mutex sync.Mutex
	*adaptors.DigitalPinsAdaptor
	*adaptors.PWMPinsAdaptor
	*adaptors.I2cBusAdaptor
	*adaptors.SpiBusAdaptor // for usage of "adaptors.WithSpiGpioAccess()"
}

// NewAdaptor returns a new Joule Adaptor
//
// Optional parameters:
//
//	adaptors.WithGpioCdevAccess():	use character device driver instead of sysfs
//	adaptors.WithSpiGpioAccess(sclk, ncs, sdo, sdi):	use GPIO's instead of /dev/spidev#.#
//
//	Optional parameters for PWM, see [adaptors.NewPWMPinsAdaptor]
func NewAdaptor(opts ...interface{}) *Adaptor {
	sys := system.NewAccesser(system.WithDigitalPinSysfsAccess())
	a := &Adaptor{
		name: gobot.DefaultName("Joule"),
		sys:  sys,
	}

	var digitalPinsOpts []adaptors.DigitalPinsOptionApplier
	pwmPinsOpts := []adaptors.PwmPinsOptionApplier{adaptors.WithPWMPinInitializer(pwmPinInitializer)}
	var spiBusOpts []adaptors.SpiBusOptionApplier
	for _, opt := range opts {
		switch o := opt.(type) {
		case adaptors.DigitalPinsOptionApplier:
			digitalPinsOpts = append(digitalPinsOpts, o)
		case adaptors.PwmPinsOptionApplier:
			pwmPinsOpts = append(pwmPinsOpts, o)
		case adaptors.SpiBusOptionApplier:
			spiBusOpts = append(spiBusOpts, o)
		default:
			panic(fmt.Sprintf("'%s' can not be applied on adaptor '%s'", opt, a.name))
		}
	}

	// Valid bus numbers are [0..2] which corresponds to /dev/i2c-0 through /dev/i2c-2.
	i2cBusNumberValidator := adaptors.NewBusNumberValidator([]int{0, 1, 2})

	a.DigitalPinsAdaptor = adaptors.NewDigitalPinsAdaptor(sys, a.translateDigitalPin, digitalPinsOpts...)
	a.PWMPinsAdaptor = adaptors.NewPWMPinsAdaptor(sys, a.translatePWMPin, pwmPinsOpts...)
	a.I2cBusAdaptor = adaptors.NewI2cBusAdaptor(sys, i2cBusNumberValidator.Validate, defaultI2cBusNumber)

	// SPI is only supported when "adaptors.WithSpiGpioAccess()" is given
	if len(spiBusOpts) > 0 {
		a.SpiBusAdaptor = adaptors.NewSpiBusAdaptor(sys, func(int) error { return nil }, 0, 0, 0, 0, defaultSpiMaxSpeed,
			a.DigitalPinsAdaptor, spiBusOpts...)
	}

	return a
}

// Name returns the adaptors name
func (a *Adaptor) Name() string { return a.name }

// SetName sets the adaptors name
func (a *Adaptor) SetName(n string) { a.name = n }

// Connect create new connection to board and pins.
func (a *Adaptor) Connect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.I2cBusAdaptor.Connect(); err != nil {
		return err
	}

	if err := a.PWMPinsAdaptor.Connect(); err != nil {
		return err
	}
	return a.DigitalPinsAdaptor.Connect()
}

// Finalize releases all i2c devices and exported digital and pwm pins.
func (a *Adaptor) Finalize() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	err := a.DigitalPinsAdaptor.Finalize()

	if e := a.PWMPinsAdaptor.Finalize(); e != nil {
		err = gobot.AppendError(err, e)
	}

	if e := a.I2cBusAdaptor.Finalize(); e != nil {
		err = gobot.AppendError(err, e)
	}

	return err
}

func (a *Adaptor) translateDigitalPin(id string) (string, int, error) {
	if val, ok := sysfsPinMap[id]; ok {
		return "", val.pin, nil
	}
	return "", -1, fmt.Errorf("'%s' is not a valid id for a digital pin", id)
}

func (a *Adaptor) translatePWMPin(id string) (string, int, error) {
	sysPin, ok := sysfsPinMap[id]
	if !ok {
		return "", -1, fmt.Errorf("'%s' is not a valid id for a pin", id)
	}
	if sysPin.pwmPin == -1 {
		return "", -1, fmt.Errorf("'%s' is not a valid id for a PWM pin", id)
	}
	return "/sys/class/pwm/pwmchip0", sysPin.pwmPin, nil
}

func pwmPinInitializer(_ string, pin gobot.PWMPinner) error {
	if err := pin.Export(); err != nil {
		return err
	}
	if err := pin.SetPeriod(10000000); err != nil {
		return err
	}
	return pin.SetEnabled(true)
}
