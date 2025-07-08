package jetson

import (
	"fmt"
	"sync"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/platforms/adaptors"
	"gobot.io/x/gobot/v2/pkg/system"
)

const (
	pwmPeriodDefault   = 3000000 // 3 ms = 333 Hz
	pwmPeriodMinimum   = 5334
	pwmDutyRateMinimum = 0.0005 // minimum duty of 1500 for default period, ~3 for minimum period

	defaultI2cBusNumber = 1

	defaultSpiBusNumber  = 0
	defaultSpiChipNumber = 0
	defaultSpiMode       = 0
	defaultSpiBitsNumber = 8
	defaultSpiMaxSpeed   = 10000000
)

// Adaptor is the Gobot adaptor for the Jetson Nano
type Adaptor struct {
	name  string
	sys   *system.Accesser // used for unit tests only
	mutex *sync.Mutex
	*adaptors.DigitalPinsAdaptor
	*adaptors.PWMPinsAdaptor
	*adaptors.I2cBusAdaptor
	*adaptors.SpiBusAdaptor
}

// NewAdaptor creates a Jetson Nano adaptor
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
		name:  gobot.DefaultName("JetsonNano"),
		sys:   sys,
		mutex: &sync.Mutex{},
	}

	var digitalPinsOpts []adaptors.DigitalPinsOptionApplier
	pwmPinsOpts := []adaptors.PwmPinsOptionApplier{
		adaptors.WithPWMDefaultPeriod(pwmPeriodDefault),
		adaptors.WithPWMMinimumPeriod(pwmPeriodMinimum),
		adaptors.WithPWMMinimumDutyRate(pwmDutyRateMinimum),
	}
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

	// Valid bus numbers are [0,1] which corresponds to /dev/i2c-0 through /dev/i2c-1.
	i2cBusNumberValidator := adaptors.NewBusNumberValidator([]int{0, 1})
	// Valid bus numbers are [0,1] which corresponds to /dev/spidev0.x through /dev/spidev1.x.
	// x is the chip number <255
	spiBusNumberValidator := adaptors.NewBusNumberValidator([]int{0, 1})

	a.DigitalPinsAdaptor = adaptors.NewDigitalPinsAdaptor(sys, a.translateDigitalPin, digitalPinsOpts...)
	a.PWMPinsAdaptor = adaptors.NewPWMPinsAdaptor(sys, a.translatePWMPin, pwmPinsOpts...)
	a.I2cBusAdaptor = adaptors.NewI2cBusAdaptor(sys, i2cBusNumberValidator.Validate, defaultI2cBusNumber)
	a.SpiBusAdaptor = adaptors.NewSpiBusAdaptor(sys, spiBusNumberValidator.Validate, defaultSpiBusNumber,
		defaultSpiChipNumber, defaultSpiMode, defaultSpiBitsNumber, defaultSpiMaxSpeed, a.DigitalPinsAdaptor, spiBusOpts...)
	return a
}

// Name returns the adaptors name
func (a *Adaptor) Name() string {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	return a.name
}

// SetName sets the adaptors name
func (a *Adaptor) SetName(n string) {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	a.name = n
}

// Connect create new connection to board and pins.
func (a *Adaptor) Connect() error {
	a.mutex.Lock()
	defer a.mutex.Unlock()

	if err := a.SpiBusAdaptor.Connect(); err != nil {
		return err
	}

	if err := a.I2cBusAdaptor.Connect(); err != nil {
		return err
	}

	if err := a.PWMPinsAdaptor.Connect(); err != nil {
		return err
	}

	return a.DigitalPinsAdaptor.Connect()
}

// Finalize closes connection to board and pins
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

	if e := a.SpiBusAdaptor.Finalize(); e != nil {
		err = gobot.AppendError(err, e)
	}
	return err
}

func (a *Adaptor) translateDigitalPin(id string) (string, int, error) {
	if line, ok := gpioPins[id]; ok {
		return "", line, nil
	}
	return "", -1, fmt.Errorf("'%s' is not a valid id for a digital pin", id)
}

func (a *Adaptor) translatePWMPin(id string) (string, int, error) {
	if channel, ok := pwmPins[id]; ok {
		return "/sys/class/pwm/pwmchip0", channel, nil
	}
	return "", 0, fmt.Errorf("'%s' is not a valid pin id for PWM on '%s'", id, a.name)
}
