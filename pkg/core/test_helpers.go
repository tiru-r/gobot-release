package core

import "os"

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

type testDriver struct {
	name       string
	pin        string
	connection Connection
	Commander
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
		Commander:  NewCommander(),
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

func newTestRobot(name string) *Robot {
	adaptor1 := newTestAdaptor("Connection1", "/dev/null")
	adaptor2 := newTestAdaptor("Connection2", "/dev/null")
	adaptor3 := newTestAdaptor("", "/dev/null")
	driver1 := newTestDriver(adaptor1, "Device1", "0")
	driver2 := newTestDriver(adaptor2, "Device2", "2")
	driver3 := newTestDriver(adaptor3, "", "1")
	work := func() {}
	r := NewRobot(name,
		[]Connection{adaptor1, adaptor2, adaptor3},
		[]Device{driver1, driver2, driver3},
		work,
	)
	r.AddCommand("RobotCommand", func(params map[string]any) any { return nil })
	// Set up trap function for testing
	setTestTrap(r)

	return r
}

// setTestTrap is a helper function to set the trap for testing
func setTestTrap(r *Robot) {
	r.SetTrap(func(c chan os.Signal) {
		c <- os.Interrupt
	})
}