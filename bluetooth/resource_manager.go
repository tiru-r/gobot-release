package bluetooth

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ResourceManager manages the lifecycle of Bluetooth resources
type ResourceManager struct {
	mu          sync.RWMutex
	resources   map[string]Resource
	cleanupChan chan string
	stopChan    chan struct{}
	started     bool
}

// Resource represents a managed Bluetooth resource
type Resource interface {
	// ID returns the unique identifier for this resource
	ID() string
	
	// Type returns the resource type (e.g., "connection", "scanner", "advertiser")
	Type() string
	
	// Cleanup performs resource cleanup
	Cleanup(ctx context.Context) error
	
	// IsActive returns true if the resource is currently active
	IsActive() bool
	
	// LastActivity returns the last time this resource was active
	LastActivity() time.Time
}

// NewResourceManager creates a new resource manager
func NewResourceManager() *ResourceManager {
	return &ResourceManager{
		resources:   make(map[string]Resource),
		cleanupChan: make(chan string, 100),
		stopChan:    make(chan struct{}),
	}
}

// Start starts the resource manager
func (rm *ResourceManager) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.started {
		return NewBluetoothErrorWithCode(ErrorCodeInvalidOperation, "resource manager already started")
	}

	rm.started = true
	go rm.cleanupLoop(ctx)
	return nil
}

// Stop stops the resource manager and cleans up all resources
func (rm *ResourceManager) Stop(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.started {
		return nil
	}

	rm.started = false
	close(rm.stopChan)

	// Cleanup all remaining resources
	var errors []error
	for id, resource := range rm.resources {
		if err := resource.Cleanup(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to cleanup resource %s: %w", id, err))
		}
	}

	// Clear resources map
	rm.resources = make(map[string]Resource)

	if len(errors) > 0 {
		return CombineErrors(errors...)
	}

	return nil
}

// Register registers a resource for management
func (rm *ResourceManager) Register(resource Resource) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.started {
		return NewBluetoothErrorWithCode(ErrorCodeInvalidOperation, "resource manager not started")
	}

	id := resource.ID()
	if _, exists := rm.resources[id]; exists {
		return NewBluetoothErrorWithCode(ErrorCodeResourceBusy, fmt.Sprintf("resource %s already registered", id))
	}

	rm.resources[id] = resource
	return nil
}

// Unregister unregisters and cleans up a resource
func (rm *ResourceManager) Unregister(ctx context.Context, id string) error {
	rm.mu.Lock()
	resource, exists := rm.resources[id]
	if exists {
		delete(rm.resources, id)
	}
	rm.mu.Unlock()

	if !exists {
		return NewNotFoundError("resource", id)
	}

	return resource.Cleanup(ctx)
}

// Get retrieves a resource by ID
func (rm *ResourceManager) Get(id string) (Resource, bool) {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resource, exists := rm.resources[id]
	return resource, exists
}

// List returns all registered resources
func (rm *ResourceManager) List() []Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	resources := make([]Resource, 0, len(rm.resources))
	for _, resource := range rm.resources {
		resources = append(resources, resource)
	}
	return resources
}

// ListByType returns all resources of a specific type
func (rm *ResourceManager) ListByType(resourceType string) []Resource {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	var resources []Resource
	for _, resource := range rm.resources {
		if resource.Type() == resourceType {
			resources = append(resources, resource)
		}
	}
	return resources
}

// CleanupInactive schedules cleanup of an inactive resource
func (rm *ResourceManager) CleanupInactive(id string) {
	select {
	case rm.cleanupChan <- id:
	default:
		// Channel is full, skip this cleanup request
	}
}

// CleanupExpired cleans up resources that have been inactive for too long
func (rm *ResourceManager) CleanupExpired(ctx context.Context, maxAge time.Duration) error {
	rm.mu.RLock()
	expiredResources := make([]Resource, 0)
	cutoff := time.Now().Add(-maxAge)

	for _, resource := range rm.resources {
		if !resource.IsActive() && resource.LastActivity().Before(cutoff) {
			expiredResources = append(expiredResources, resource)
		}
	}
	rm.mu.RUnlock()

	var errors []error
	for _, resource := range expiredResources {
		if err := rm.Unregister(ctx, resource.ID()); err != nil {
			errors = append(errors, err)
		}
	}

	if len(errors) > 0 {
		return CombineErrors(errors...)
	}

	return nil
}

// cleanupLoop runs the background cleanup process
func (rm *ResourceManager) cleanupLoop(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second) // Periodic cleanup every 30 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-rm.stopChan:
			return
		case id := <-rm.cleanupChan:
			rm.handleCleanupRequest(ctx, id)
		case <-ticker.C:
			rm.periodicCleanup(ctx)
		}
	}
}

// handleCleanupRequest handles a specific cleanup request
func (rm *ResourceManager) handleCleanupRequest(ctx context.Context, id string) {
	rm.mu.RLock()
	resource, exists := rm.resources[id]
	rm.mu.RUnlock()

	if !exists {
		return
	}

	// Only cleanup if resource is inactive
	if !resource.IsActive() {
		if err := rm.Unregister(ctx, id); err != nil {
			// Log error (in a real implementation, use proper logging)
			fmt.Printf("Failed to cleanup resource %s: %v\n", id, err)
		}
	}
}

// periodicCleanup performs periodic cleanup of expired resources
func (rm *ResourceManager) periodicCleanup(ctx context.Context) {
	// Cleanup resources inactive for more than 5 minutes
	if err := rm.CleanupExpired(ctx, 5*time.Minute); err != nil {
		// Log error (in a real implementation, use proper logging)
		fmt.Printf("Periodic cleanup failed: %v\n", err)
	}
}

// Statistics returns resource management statistics
func (rm *ResourceManager) Statistics() ResourceStatistics {
	rm.mu.RLock()
	defer rm.mu.RUnlock()

	stats := ResourceStatistics{
		TotalResources: len(rm.resources),
		ResourcesByType: make(map[string]int),
	}

	for _, resource := range rm.resources {
		stats.ResourcesByType[resource.Type()]++
		if resource.IsActive() {
			stats.ActiveResources++
		}
	}

	return stats
}

// ResourceStatistics contains resource management statistics
type ResourceStatistics struct {
	TotalResources  int
	ActiveResources int
	ResourcesByType map[string]int
}

// ConnectionResource represents a managed connection
type ConnectionResource struct {
	id           string
	device       Device
	lastActivity time.Time
	active       bool
	mu           sync.RWMutex
}

// NewConnectionResource creates a new connection resource
func NewConnectionResource(device Device) *ConnectionResource {
	return &ConnectionResource{
		id:           fmt.Sprintf("connection-%s", device.Address().String()),
		device:       device,
		lastActivity: time.Now(),
		active:       true,
	}
}

// ID returns the connection ID
func (cr *ConnectionResource) ID() string {
	return cr.id
}

// Type returns the resource type
func (cr *ConnectionResource) Type() string {
	return "connection"
}

// Cleanup disconnects the device
func (cr *ConnectionResource) Cleanup(ctx context.Context) error {
	cr.mu.Lock()
	defer cr.mu.Unlock()

	if !cr.active {
		return nil
	}

	cr.active = false
	if cr.device.Connected() {
		return cr.device.Disconnect(ctx)
	}
	return nil
}

// IsActive returns true if the connection is active
func (cr *ConnectionResource) IsActive() bool {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.active && cr.device.Connected()
}

// LastActivity returns the last activity time
func (cr *ConnectionResource) LastActivity() time.Time {
	cr.mu.RLock()
	defer cr.mu.RUnlock()
	return cr.lastActivity
}

// UpdateActivity updates the last activity time
func (cr *ConnectionResource) UpdateActivity() {
	cr.mu.Lock()
	defer cr.mu.Unlock()
	cr.lastActivity = time.Now()
}

// ScannerResource represents a managed scanner
type ScannerResource struct {
	id           string
	central      Central
	lastActivity time.Time
	active       bool
	mu           sync.RWMutex
}

// NewScannerResource creates a new scanner resource
func NewScannerResource(central Central) *ScannerResource {
	return &ScannerResource{
		id:           fmt.Sprintf("scanner-%p", central),
		central:      central,
		lastActivity: time.Now(),
		active:       true,
	}
}

// ID returns the scanner ID
func (sr *ScannerResource) ID() string {
	return sr.id
}

// Type returns the resource type
func (sr *ScannerResource) Type() string {
	return "scanner"
}

// Cleanup stops the scanner
func (sr *ScannerResource) Cleanup(ctx context.Context) error {
	sr.mu.Lock()
	defer sr.mu.Unlock()

	if !sr.active {
		return nil
	}

	sr.active = false
	return sr.central.StopScan(ctx)
}

// IsActive returns true if the scanner is active
func (sr *ScannerResource) IsActive() bool {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.active
}

// LastActivity returns the last activity time
func (sr *ScannerResource) LastActivity() time.Time {
	sr.mu.RLock()
	defer sr.mu.RUnlock()
	return sr.lastActivity
}

// UpdateActivity updates the last activity time
func (sr *ScannerResource) UpdateActivity() {
	sr.mu.Lock()
	defer sr.mu.Unlock()
	sr.lastActivity = time.Now()
}

// AdvertiserResource represents a managed advertiser
type AdvertiserResource struct {
	id           string
	peripheral   Peripheral
	lastActivity time.Time
	active       bool
	mu           sync.RWMutex
}

// NewAdvertiserResource creates a new advertiser resource
func NewAdvertiserResource(peripheral Peripheral) *AdvertiserResource {
	return &AdvertiserResource{
		id:           fmt.Sprintf("advertiser-%p", peripheral),
		peripheral:   peripheral,
		lastActivity: time.Now(),
		active:       true,
	}
}

// ID returns the advertiser ID
func (ar *AdvertiserResource) ID() string {
	return ar.id
}

// Type returns the resource type
func (ar *AdvertiserResource) Type() string {
	return "advertiser"
}

// Cleanup stops advertising
func (ar *AdvertiserResource) Cleanup(ctx context.Context) error {
	ar.mu.Lock()
	defer ar.mu.Unlock()

	if !ar.active {
		return nil
	}

	ar.active = false
	return ar.peripheral.StopAdvertising(ctx)
}

// IsActive returns true if the advertiser is active
func (ar *AdvertiserResource) IsActive() bool {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return ar.active && ar.peripheral.IsAdvertising()
}

// LastActivity returns the last activity time
func (ar *AdvertiserResource) LastActivity() time.Time {
	ar.mu.RLock()
	defer ar.mu.RUnlock()
	return ar.lastActivity
}

// UpdateActivity updates the last activity time
func (ar *AdvertiserResource) UpdateActivity() {
	ar.mu.Lock()
	defer ar.mu.Unlock()
	ar.lastActivity = time.Now()
}

// Global resource manager instance
var globalResourceManager *ResourceManager
var resourceManagerOnce sync.Once

// GetResourceManager returns the global resource manager instance
func GetResourceManager() *ResourceManager {
	resourceManagerOnce.Do(func() {
		globalResourceManager = NewResourceManager()
	})
	return globalResourceManager
}

// InitializeResourceManager initializes the global resource manager
func InitializeResourceManager(ctx context.Context) error {
	rm := GetResourceManager()
	return rm.Start(ctx)
}

// CleanupAllResources cleans up all resources using the global resource manager
func CleanupAllResources(ctx context.Context) error {
	rm := GetResourceManager()
	return rm.Stop(ctx)
}