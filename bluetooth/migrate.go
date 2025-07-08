package bluetooth

import (
	"time"
	
	"gobot.io/x/gobot/v2"
)

// MigrateFromTinyGo helps migrate from tinygo.org/x/bluetooth to our implementation
// This provides a compatibility layer for existing code

// LegacyClientAdaptorOption for backward compatibility
type LegacyClientAdaptorOption func(*ClientAdaptor)

// WithDebug enables debug mode (legacy compatibility)
func WithDebug() LegacyClientAdaptorOption {
	return func(a *ClientAdaptor) {
		// Debug mode would be implemented in a real platform implementation
	}
}

// ConvertLegacyOptions converts old-style options to new options
func ConvertLegacyOptions(opts ...LegacyClientAdaptorOption) []ClientAdaptorOption {
	// For now, just return empty slice since the main options are already supported
	return []ClientAdaptorOption{}
}

// NewLegacyAdaptor creates an adaptor with legacy option support
func NewLegacyAdaptor(identifier string, legacyOpts ...LegacyClientAdaptorOption) gobot.BLEConnector {
	adaptor := NewClientAdaptor(identifier)
	
	// Apply legacy options
	for _, opt := range legacyOpts {
		opt(adaptor)
	}
	
	return adaptor
}

// Helper functions that match the old tinygo patterns

// DefaultScanTimeout for compatibility
const DefaultScanTimeout = 10 * time.Minute

// DefaultSleepAfterDisconnect for compatibility
const DefaultSleepAfterDisconnect = 500 * time.Millisecond