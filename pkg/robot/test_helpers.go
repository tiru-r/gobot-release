package robot

import (
	"os"

	"gobot.io/x/gobot/v2/pkg/core"
)

type NullReadWriteCloser struct{}

func (NullReadWriteCloser) Write(p []byte) (int, error) {
	return len(p), nil
}

func (NullReadWriteCloser) Read(b []byte) (int, error) {
	return len(b), nil
}

func (NullReadWriteCloser) Close() error {
	return nil
}

// Type aliases for test helpers
type Connection = core.Connection
type Device = core.Device

type testDriver struct {
	name       string
	pin        string
	connection Connection
	core.Commander
}

var (
	testDriverStart = func() error { return nil }
	testDriverHalt  = func() error { return nil }
)

func (t *testDriver) Start() error           { return testDriverStart() }
func (t *testDriver) Halt() error            { return testDriverHalt() }
func (t *testDriver) Name() string           { return t.name }
func (t *testDriver) SetName(n string)       { t.name = n }
func (t *testDriver) Pin() string            { return t.pin }
func (t *testDriver) Connection() Connection { return t.connection }

func newTestDriver(adaptor *testAdaptor, name string, pin string) *testDriver {
	t := &testDriver{
		name:       name,
		connection: adaptor,
		pin:        pin,
		Commander:  core.NewCommander(),
	}

	t.AddCommand("DriverCommand", func(params map[string]any) any { return nil })

	return t
}

type testAdaptor struct {
	name string
	port string
}

var (
	testAdaptorConnect  = func() error { return nil }
	testAdaptorFinalize = func() error { return nil }
)

func (t *testAdaptor) Finalize() error  { return testAdaptorFinalize() }
func (t *testAdaptor) Connect() error   { return testAdaptorConnect() }
func (t *testAdaptor) Name() string     { return t.name }
func (t *testAdaptor) SetName(n string) { t.name = n }
func (t *testAdaptor) Port() string     { return t.port }

func newTestAdaptor(name string, port string) *testAdaptor { //nolint:unparam // keep for tests
	return &testAdaptor{
		name: name,
		port: port,
	}
}

func newTestRobot(name string) *core.Robot {
	adaptor1 := newTestAdaptor("Connection1", "/dev/null")
	adaptor2 := newTestAdaptor("Connection2", "/dev/null")
	adaptor3 := newTestAdaptor("", "/dev/null")
	driver1 := newTestDriver(adaptor1, "Device1", "0")
	driver2 := newTestDriver(adaptor2, "Device2", "2")
	driver3 := newTestDriver(adaptor3, "", "1")
	work := func() {}
	r := core.NewRobot(name,
		[]Connection{adaptor1, adaptor2, adaptor3},
		[]Device{driver1, driver2, driver3},
		work,
	)
	r.AddCommand("RobotCommand", func(params map[string]any) any { return nil })
	r.SetTrap(func(c chan os.Signal) {
		c <- os.Interrupt
	})

	return r
}

// AppendError adds the test helper
var AppendError = core.AppendError