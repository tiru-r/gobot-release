package gobot

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"
)

// JSONRobot a JSON representation of a Robot.
type JSONRobot struct {
	Name        string            `json:"name"`
	Commands    []string          `json:"commands"`
	Connections []*JSONConnection `json:"connections"`
	Devices     []*JSONDevice     `json:"devices"`
}

// NewJSONRobot returns a JSONRobot given a Robot.
func NewJSONRobot(robot *Robot) *JSONRobot {
	jsonRobot := &JSONRobot{
		Name:        robot.Name,
		Commands:    []string{},
		Connections: []*JSONConnection{},
		Devices:     []*JSONDevice{},
	}

	for command := range robot.Commands() {
		jsonRobot.Commands = append(jsonRobot.Commands, command)
	}

	robot.Devices().Each(func(device Device) {
		jsonDevice := NewJSONDevice(device)
		jsonRobot.Connections = append(jsonRobot.Connections, NewJSONConnection(robot.Connection(jsonDevice.Connection)))
		jsonRobot.Devices = append(jsonRobot.Devices, jsonDevice)
	})
	return jsonRobot
}

// Robot is a named entity that manages a collection of connections and devices.
// It contains its own work routine and a collection of
// custom commands to control a robot remotely via the Gobot api.
type Robot struct {
	Name               string
	Work               func()
	connections        *Connections
	devices            *Devices
	trap               func(chan os.Signal)
	AutoRun            bool
	running            atomic.Bool
	done               chan bool
	workRegistry       *RobotWorkRegistry
	WorkEveryWaitGroup *sync.WaitGroup
	WorkAfterWaitGroup *sync.WaitGroup
	Commander
	Eventer
	// Context for graceful shutdown
	ctx    context.Context
	cancel context.CancelFunc
}

// Robots is a collection of Robot
type Robots []*Robot

// Len returns the amount of Robots in the collection.
func (r *Robots) Len() int {
	return len(*r)
}

// Start calls the Start method of each Robot in the collection. We return on first error.
func (r *Robots) Start(args ...any) error {
	autoRun := true
	if args[0] != nil {
		var ok bool
		if autoRun, ok = args[0].(bool); !ok {
			// we treat this as false
			autoRun = false
		}
	}
	for _, robot := range *r {
		if err := robot.Start(autoRun); err != nil {
			return err
		}
	}
	return nil
}

// Stop calls the Stop method of each Robot in the collection. We try to stop all robots and
// collect the errors.
func (r *Robots) Stop() error {
	var err error
	for _, robot := range *r {
		if e := robot.Stop(); e != nil {
			err = AppendError(err, e)
		}
	}
	return err
}

// Each enumerates through the Robots and calls specified callback function.
func (r *Robots) Each(f func(*Robot)) {
	for _, robot := range *r {
		f(robot)
	}
}

// All returns an iterator over all robots using range-over-func
func (r *Robots) All() func(func(*Robot) bool) {
	return func(yield func(*Robot) bool) {
		for _, robot := range *r {
			if !yield(robot) {
				return
			}
		}
	}
}

// NewRobot returns a new Robot. It supports the following optional params:
//
//	name:	string with the name of the Robot. A name will be automatically generated if no name is supplied.
//	[]Connection: Connections which are automatically started and stopped with the robot
//	[]Device: Devices which are automatically started and stopped with the robot
//	func(): The work routine the robot will execute once all devices and connections have been initialized and started
func NewRobot(v ...any) *Robot {
	ctx, cancel := context.WithCancel(context.Background())
	r := &Robot{
		Name:        fmt.Sprintf("%X", Rand(int(^uint(0)>>1))),
		connections: &Connections{},
		devices:     &Devices{},
		done:        make(chan bool, 1),
		trap: func(c chan os.Signal) {
			signal.Notify(c, os.Interrupt)
		},
		AutoRun:   true,
		Work:      nil,
		Eventer:   NewEventer(),
		Commander: NewCommander(),
		ctx:       ctx,
		cancel:    cancel,
	}

	for i := range v {
		switch val := v[i].(type) {
		case string:
			r.Name = val
		case []Connection:
			log.Println("Initializing connections...")
			for _, connection := range val {
				c := r.AddConnection(connection)
				log.Println("Initializing connection", c.Name(), "...")
			}
		case []Device:
			log.Println("Initializing devices...")
			for _, device := range val {
				d := r.AddDevice(device)
				log.Println("Initializing device", d.Name(), "...")
			}
		case func():
			r.Work = val
		}
	}

	r.workRegistry = &RobotWorkRegistry{
		r: make(map[string]*RobotWork),
	}
	r.WorkAfterWaitGroup = &sync.WaitGroup{}
	r.WorkEveryWaitGroup = &sync.WaitGroup{}

	r.running.Store(false)
	log.Println("Robot", r.Name, "initialized.")

	return r
}

// Start a Robot's Connections, Devices, and work. We stop initialization of
// connections and devices on first error.
func (r *Robot) Start(args ...any) error {
	if len(args) > 0 && args[0] != nil {
		var ok bool
		if r.AutoRun, ok = args[0].(bool); !ok {
			// we treat this as false
			r.AutoRun = false
		}
	}
	log.Println("Starting Robot", r.Name, "...")
	if err := r.Connections().Start(); err != nil {
		log.Println(err)
		return err
	}

	if err := r.Devices().Start(); err != nil {
		log.Println(err)
		return err
	}

	if r.Work == nil {
		r.Work = func() {}
	}

	log.Println("Starting work...")
	go func() {
		defer func() {
			select {
			case r.done <- true:
			default:
				// Channel is full, work is already done
			}
		}()
		r.Work()
	}()

	r.running.Store(true)

	if !r.AutoRun {
		return nil
	}

	c := make(chan os.Signal, 1)
	r.trap(c)

	// waiting for interrupt coming on the channel with timeout
	select {
	case <-c:
		// Stop calls the Stop method on itself, if we are "auto-running".
		return r.Stop()
	case <-r.ctx.Done():
		// Context cancelled, perform graceful shutdown
		return r.Stop()
	}
}

// Stop stops a Robot's connections and devices. We try to stop all items and
// collect all errors.
func (r *Robot) Stop() error {
	var err error
	log.Println("Stopping Robot", r.Name, "...")
	
	// Cancel context to signal shutdown
	r.cancel()
	
	// Shutdown eventer gracefully
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if e := r.Eventer.Shutdown(shutdownCtx); e != nil {
		err = AppendError(err, e)
	}
	
	if e := r.Devices().Halt(); e != nil {
		err = AppendError(err, e)
	}
	if e := r.Connections().Finalize(); e != nil {
		err = AppendError(err, e)
	}

	// Wait for work to complete with timeout
	select {
	case <-r.done:
		// Work completed normally
	case <-time.After(5 * time.Second):
		// Work didn't complete in time, continue with shutdown
		log.Println("Warning: Robot work didn't complete within timeout")
	}
	
	r.running.Store(false)
	return err
}

// Running returns if the Robot is currently started or not
func (r *Robot) Running() bool {
	return r.running.Load()
}

// Devices returns all devices associated with this Robot.
func (r *Robot) Devices() *Devices {
	return r.devices
}

// AddDevice adds a new Device to the robots collection of devices. Returns the
// added device.
func (r *Robot) AddDevice(d Device) Device {
	*r.devices = append(*r.Devices(), d)
	return d
}

// Device returns a device given a name. Returns nil if the Device does not exist.
func (r *Robot) Device(name string) Device {
	if r == nil {
		return nil
	}
	for _, device := range *r.devices {
		if device.Name() == name {
			return device
		}
	}
	return nil
}

// Connections returns all connections associated with this robot.
func (r *Robot) Connections() *Connections {
	return r.connections
}

// AddConnection adds a new connection to the robots collection of connections.
// Returns the added connection.
func (r *Robot) AddConnection(c Connection) Connection {
	*r.connections = append(*r.Connections(), c)
	return c
}

// Connection returns a connection given a name. Returns nil if the Connection
// does not exist.
func (r *Robot) Connection(name string) Connection {
	if r == nil {
		return nil
	}
	for _, connection := range *r.connections {
		if connection.Name() == name {
			return connection
		}
	}
	return nil
}
