package bleclient

import (
	"gobot.io/x/gobot/v2/bluetooth"
)

// NewSimpleAdaptor creates a new simple BLE client adaptor following Gobot patterns
// This adaptor is designed to be more straightforward and match Gobot's architectural style
func NewSimpleAdaptor(identifier string, opts ...bluetooth.ClientAdaptorOption) *bluetooth.ClientAdaptor {
	return bluetooth.NewClientAdaptor(identifier, opts...)
}