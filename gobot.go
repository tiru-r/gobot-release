package gobot

// Re-export core functionality to maintain backward compatibility
import (
	"gobot.io/x/gobot/v2/pkg/core"
	"gobot.io/x/gobot/v2/pkg/robot"
	"gobot.io/x/gobot/v2/pkg/adaptor"
	"gobot.io/x/gobot/v2/pkg/device"
	goboterrors "gobot.io/x/gobot/v2/internal/errors"
	gobotutils "gobot.io/x/gobot/v2/internal/utils"
)

// Core types and functions
type Robot = core.Robot
type Robots = core.Robots
type Event = core.Event
type Commander = core.Commander
type Eventer = core.Eventer
type Pinner = core.Pinner

// Connection and device types
type Connection = adaptor.Connection
type Connections = adaptor.Connections
type Device = device.Device
type Devices = device.Devices
type Adaptor = adaptor.Adaptor

// Manager
type Manager = robot.Manager

// Core functions
var NewRobot = core.NewRobot
var NewManager = robot.NewManager
var NewEvent = core.NewEvent
var NewEventer = core.NewEventer
var NewCommander = core.NewCommander

// Robot options
var WithName = core.WithName
var WithWork = core.WithWork
var WithAutoRun = core.WithAutoRun
var WithConnections = core.WithConnections
var WithDevices = core.WithDevices

// JSON types
type JSONRobot = core.JSONRobot
type JSONConnection = adaptor.JSONConnection
type JSONDevice = device.JSONDevice
var NewJSONRobot = core.NewJSONRobot
var NewJSONConnection = adaptor.NewJSONConnection
var NewJSONDevice = device.NewJSONDevice

// Utility functions
var Rand = gobotutils.Rand
var Every = gobotutils.Every
var After = gobotutils.After
var AppendError = goboterrors.AppendError
var DefaultName = gobotutils.DefaultName
var FromScale = gobotutils.FromScale
var ToScale = gobotutils.ToScale
var Rescale = gobotutils.Rescale

// Interface types
type DigitalPinner = adaptor.DigitalPinner
type DigitalPinnerProvider = adaptor.DigitalPinnerProvider
type PWMPinner = adaptor.PWMPinner
type PWMPinnerProvider = adaptor.PWMPinnerProvider
type AnalogPinner = adaptor.AnalogPinner
type I2cOperations = adaptor.I2cOperations
type SpiOperations = adaptor.SpiOperations
type OneWireOperations = adaptor.OneWireOperations
type BLEConnector = adaptor.BLEConnector
type Porter = adaptor.Porter

// Digital pin interfaces
type DigitalPinOptioner = adaptor.DigitalPinOptioner
type DigitalPinValuer = adaptor.DigitalPinValuer
type DigitalPinOptionApplier = adaptor.DigitalPinOptionApplier

// System device interfaces
type SpiSystemDevicer = adaptor.SpiSystemDevicer
type I2cSystemDevicer = adaptor.I2cSystemDevicer
type OneWireSystemDevicer = adaptor.OneWireSystemDevicer

// Manager JSON types
type JSONManager = robot.JSONManager
var NewJSONManager = robot.NewJSONManager

// Driver interface
type Driver = device.Driver