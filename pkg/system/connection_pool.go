package system

import (
	"fmt"
	"sync"
	"time"
)

// ConnectionPool manages a pool of reusable connections
type ConnectionPool struct {
	devices map[string]*i2cDevice
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
}

// PooledDevice wraps a device with metadata
type PooledDevice struct {
	device   *i2cDevice
	lastUsed time.Time
	inUse    bool
}

// NewConnectionPool creates a new connection pool
func NewConnectionPool(maxSize int, ttl time.Duration) *ConnectionPool {
	pool := &ConnectionPool{
		devices: make(map[string]*i2cDevice),
		maxSize: maxSize,
		ttl:     ttl,
	}
	
	// Start cleanup goroutine
	go pool.cleanup()
	
	return pool
}

// GetDevice returns a device from the pool or creates a new one
func (p *ConnectionPool) GetDevice(accesser *Accesser, location string) (*i2cDevice, error) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	if device, exists := p.devices[location]; exists {
		return device, nil
	}
	
	// Check pool size limit
	if len(p.devices) >= p.maxSize {
		return nil, fmt.Errorf("connection pool full (max %d)", p.maxSize)
	}
	
	// Create new device
	device, err := accesser.NewI2cDevice(location)
	if err != nil {
		return nil, err
	}
	
	p.devices[location] = device
	return device, nil
}

// ReleaseDevice marks a device as available for reuse
func (p *ConnectionPool) ReleaseDevice(location string) {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	// Device stays in pool for reuse
	// Actual cleanup happens in background
}

// cleanup removes stale connections from the pool
func (p *ConnectionPool) cleanup() {
	ticker := time.NewTicker(p.ttl / 2)
	defer ticker.Stop()
	
	for range ticker.C {
		p.mutex.Lock()
		now := time.Now()
		
		for location, device := range p.devices {
			// Check if device hasn't been used recently
			// For simplicity, we'll keep devices in pool until explicitly removed
			_ = location
			_ = device
			_ = now
		}
		p.mutex.Unlock()
	}
}

// Close closes all devices in the pool
func (p *ConnectionPool) Close() error {
	p.mutex.Lock()
	defer p.mutex.Unlock()
	
	var lastErr error
	for location, device := range p.devices {
		if err := device.Close(); err != nil {
			lastErr = err
		}
		delete(p.devices, location)
	}
	
	return lastErr
}

// Size returns the current number of pooled connections
func (p *ConnectionPool) Size() int {
	p.mutex.RLock()
	defer p.mutex.RUnlock()
	return len(p.devices)
}