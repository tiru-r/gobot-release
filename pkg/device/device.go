package device

import (
	"gobot.io/x/gobot/v2/pkg/core"
)

// Type aliases for convenience - re-export core types for backward compatibility
type Commander = core.Commander
type Driver = core.Driver
type Pinner = core.Pinner
type Connection = core.Connection
type Device = core.Device
type Devices = core.Devices
type JSONDevice = core.JSONDevice

// Function aliases
var NewJSONDevice = core.NewJSONDevice