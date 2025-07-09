package bluetooth

import (
	"errors"
	"fmt"
	"time"
)

// Core Bluetooth errors
var (
	// Connection errors
	ErrNotConnected           = NewBluetoothErrorWithCode(ErrorCodeNotConnected, "device not connected")
	ErrConnectionFailed       = NewBluetoothErrorWithCode(ErrorCodeConnectionFailed, "connection failed")
	ErrConnectionTimeout      = NewBluetoothErrorWithCode(ErrorCodeConnectionTimeout, "connection timeout")
	ErrDisconnected           = NewBluetoothErrorWithCode(ErrorCodeDisconnected, "device disconnected")
	ErrConnectionLost         = NewBluetoothErrorWithCode(ErrorCodeConnectionLost, "connection lost unexpectedly")
	ErrMaxConnectionsReached  = NewBluetoothErrorWithCode(ErrorCodeMaxConnections, "maximum connections reached")

	// Discovery errors
	ErrScanFailed             = NewBluetoothErrorWithCode(ErrorCodeScanFailed, "scan failed to start")
	ErrScanTimeout            = NewBluetoothErrorWithCode(ErrorCodeScanTimeout, "scan timeout")
	ErrDeviceNotFound         = NewBluetoothErrorWithCode(ErrorCodeDeviceNotFound, "device not found")
	ErrServiceNotFound        = NewBluetoothErrorWithCode(ErrorCodeServiceNotFound, "service not found")
	ErrCharacteristicNotFound = NewBluetoothErrorWithCode(ErrorCodeCharacteristicNotFound, "characteristic not found")
	ErrDescriptorNotFound     = NewBluetoothErrorWithCode(ErrorCodeDescriptorNotFound, "descriptor not found")

	// Operation errors
	ErrOperationFailed        = NewBluetoothErrorWithCode(ErrorCodeOperationFailed, "operation failed")
	ErrOperationTimeout       = NewBluetoothErrorWithCode(ErrorCodeOperationTimeout, "operation timeout")
	ErrOperationCancelled     = NewBluetoothErrorWithCode(ErrorCodeOperationCancelled, "operation cancelled")
	ErrOperationNotSupported  = NewBluetoothErrorWithCode(ErrorCodeNotSupported, "operation not supported")
	ErrInvalidOperation       = NewBluetoothErrorWithCode(ErrorCodeInvalidOperation, "invalid operation")

	// Read/Write errors
	ErrReadFailed             = NewBluetoothErrorWithCode(ErrorCodeReadFailed, "read operation failed")
	ErrWriteFailed            = NewBluetoothErrorWithCode(ErrorCodeWriteFailed, "write operation failed")
	ErrInsufficientPermissions = NewBluetoothErrorWithCode(ErrorCodeInsufficientPermissions, "insufficient permissions")
	ErrInvalidLength          = NewBluetoothErrorWithCode(ErrorCodeInvalidLength, "invalid data length")
	ErrInvalidData            = NewBluetoothErrorWithCode(ErrorCodeInvalidData, "invalid data")

	// Adapter/Manager errors
	ErrAdapterNotFound        = NewBluetoothErrorWithCode(ErrorCodeAdapterNotFound, "adapter not found")
	ErrAdapterNotEnabled      = NewBluetoothErrorWithCode(ErrorCodeAdapterNotEnabled, "adapter not enabled")
	ErrBluetoothUnavailable   = NewBluetoothErrorWithCode(ErrorCodeBluetoothUnavailable, "bluetooth unavailable")
	ErrPoweringOn             = NewBluetoothErrorWithCode(ErrorCodePoweringOn, "adapter is powering on")
	ErrPoweringOff            = NewBluetoothErrorWithCode(ErrorCodePoweringOff, "adapter is powering off")
	ErrUnauthorized           = NewBluetoothErrorWithCode(ErrorCodeUnauthorized, "bluetooth access unauthorized")

	// Advertising errors
	ErrAdvertisingFailed      = NewBluetoothErrorWithCode(ErrorCodeAdvertisingFailed, "advertising failed to start")
	ErrAdvertisingNotSupported = NewBluetoothErrorWithCode(ErrorCodeAdvertisingNotSupported, "advertising not supported")
	ErrAdvertisingTimeout     = NewBluetoothErrorWithCode(ErrorCodeAdvertisingTimeout, "advertising timeout")

	// Validation errors (already defined in validation.go)
	ErrInvalidUUID            = NewBluetoothErrorWithCode(ErrorCodeInvalidUUID, "invalid UUID format")
	ErrInvalidAddress         = NewBluetoothErrorWithCode(ErrorCodeInvalidAddress, "invalid address format")
	ErrInvalidParameter       = NewBluetoothErrorWithCode(ErrorCodeInvalidParameter, "invalid parameter")

	// Resource errors
	ErrResourceBusy           = NewBluetoothErrorWithCode(ErrorCodeResourceBusy, "resource is busy")
	ErrOutOfMemory            = NewBluetoothErrorWithCode(ErrorCodeOutOfMemory, "out of memory")
	ErrTooManyRequests        = NewBluetoothErrorWithCode(ErrorCodeTooManyRequests, "too many requests")

	// Platform-specific errors
	ErrPlatformNotSupported   = NewBluetoothErrorWithCode(ErrorCodePlatformNotSupported, "platform not supported")
)

// Error codes for categorizing errors
type ErrorCode int

const (
	ErrorCodeUnknown ErrorCode = iota

	// Connection error codes (1000-1999)
	ErrorCodeNotConnected = 1000 + iota
	ErrorCodeConnectionFailed
	ErrorCodeConnectionTimeout
	ErrorCodeDisconnected
	ErrorCodeConnectionLost
	ErrorCodeMaxConnections

	// Discovery error codes (2000-2999)
	ErrorCodeScanFailed = 2000 + iota
	ErrorCodeScanTimeout
	ErrorCodeDeviceNotFound
	ErrorCodeServiceNotFound
	ErrorCodeCharacteristicNotFound
	ErrorCodeDescriptorNotFound

	// Operation error codes (3000-3999)
	ErrorCodeOperationFailed = 3000 + iota
	ErrorCodeOperationTimeout
	ErrorCodeOperationCancelled
	ErrorCodeNotSupported
	ErrorCodeInvalidOperation

	// Read/Write error codes (4000-4999)
	ErrorCodeReadFailed = 4000 + iota
	ErrorCodeWriteFailed
	ErrorCodeInsufficientPermissions
	ErrorCodeInvalidLength
	ErrorCodeInvalidData

	// Adapter/Manager error codes (5000-5999)
	ErrorCodeAdapterNotFound = 5000 + iota
	ErrorCodeAdapterNotEnabled
	ErrorCodeBluetoothUnavailable
	ErrorCodePoweringOn
	ErrorCodePoweringOff
	ErrorCodeUnauthorized

	// Advertising error codes (6000-6999)
	ErrorCodeAdvertisingFailed = 6000 + iota
	ErrorCodeAdvertisingNotSupported
	ErrorCodeAdvertisingTimeout

	// Validation error codes (7000-7999)
	ErrorCodeInvalidUUID = 7000 + iota
	ErrorCodeInvalidAddress
	ErrorCodeInvalidParameter

	// Resource error codes (8000-8999)
	ErrorCodeResourceBusy = 8000 + iota
	ErrorCodeOutOfMemory
	ErrorCodeTooManyRequests

	// Platform error codes (9000-9999)
	ErrorCodePlatformNotSupported = 9000 + iota
)

// BluetoothError represents a comprehensive Bluetooth error with context
type BluetoothError struct {
	Code      ErrorCode
	Message   string
	Cause     error
	Timestamp time.Time
	Context   map[string]any
}

// NewBluetoothError creates a new Bluetooth error
func NewBluetoothError(message string) *BluetoothError {
	return &BluetoothError{
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]any),
	}
}

// NewBluetoothErrorWithCode creates a new Bluetooth error with error code
func NewBluetoothErrorWithCode(code ErrorCode, message string) *BluetoothError {
	return &BluetoothError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Context:   make(map[string]any),
	}
}

// Error implements the error interface
func (e *BluetoothError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// WithCause adds a causing error
func (e *BluetoothError) WithCause(cause error) *BluetoothError {
	e.Cause = cause
	return e
}

// WithContext adds contextual information
func (e *BluetoothError) WithContext(key string, value any) *BluetoothError {
	if e.Context == nil {
		e.Context = make(map[string]any)
	}
	e.Context[key] = value
	return e
}

// WithAddress adds device address context
func (e *BluetoothError) WithAddress(addr Address) *BluetoothError {
	return e.WithContext("address", addr.String())
}

// WithUUID adds UUID context
func (e *BluetoothError) WithUUID(uuid UUID) *BluetoothError {
	return e.WithContext("uuid", uuid.String())
}

// WithDevice adds device context
func (e *BluetoothError) WithDevice(name string) *BluetoothError {
	return e.WithContext("device", name)
}

// WithOperation adds operation context
func (e *BluetoothError) WithOperation(operation string) *BluetoothError {
	return e.WithContext("operation", operation)
}

// WithTimeout adds timeout context
func (e *BluetoothError) WithTimeout(timeout time.Duration) *BluetoothError {
	return e.WithContext("timeout", timeout.String())
}

// Is implements error matching for Go 1.13+ error handling
func (e *BluetoothError) Is(target error) bool {
	if target == nil {
		return false
	}

	if bluetoothErr, ok := target.(*BluetoothError); ok {
		return e.Code == bluetoothErr.Code
	}

	// Check against standard errors
	switch target {
	case ErrNotConnected:
		return e.Code == ErrorCodeNotConnected
	case ErrConnectionFailed:
		return e.Code == ErrorCodeConnectionFailed
	case ErrScanTimeout:
		return e.Code == ErrorCodeScanTimeout
	// Add more cases as needed
	default:
		return false
	}
}

// Unwrap returns the underlying cause for Go 1.13+ error unwrapping
func (e *BluetoothError) Unwrap() error {
	return e.Cause
}

// IsTemporary indicates if the error is temporary and operation can be retried
func (e *BluetoothError) IsTemporary() bool {
	switch e.Code {
	case ErrorCodeConnectionTimeout,
		ErrorCodeOperationTimeout,
		ErrorCodeResourceBusy,
		ErrorCodeTooManyRequests,
		ErrorCodeConnectionLost:
		return true
	default:
		return false
	}
}

// IsRetryable indicates if the operation should be retried
func (e *BluetoothError) IsRetryable() bool {
	return e.IsTemporary() && e.Code != ErrorCodeOperationCancelled
}

// Severity indicates the error severity level
type Severity int

const (
	SeverityInfo Severity = iota
	SeverityWarning
	SeverityError
	SeverityCritical
)

// Severity returns the severity level of the error
func (e *BluetoothError) Severity() Severity {
	switch e.Code {
	case ErrorCodeNotConnected,
		ErrorCodeDeviceNotFound,
		ErrorCodeServiceNotFound,
		ErrorCodeCharacteristicNotFound:
		return SeverityInfo
	case ErrorCodeConnectionTimeout,
		ErrorCodeOperationTimeout,
		ErrorCodeScanTimeout:
		return SeverityWarning
	case ErrorCodeConnectionFailed,
		ErrorCodeOperationFailed,
		ErrorCodeReadFailed,
		ErrorCodeWriteFailed:
		return SeverityError
	case ErrorCodeBluetoothUnavailable,
		ErrorCodeAdapterNotFound,
		ErrorCodePlatformNotSupported,
		ErrorCodeOutOfMemory:
		return SeverityCritical
	default:
		return SeverityError
	}
}

// Error creation helpers for common scenarios

// NewConnectionError creates a connection-related error
func NewConnectionError(message string, cause error) *BluetoothError {
	return NewBluetoothErrorWithCode(ErrorCodeConnectionFailed, message).WithCause(cause)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string, timeout time.Duration) *BluetoothError {
	return NewBluetoothErrorWithCode(ErrorCodeOperationTimeout, fmt.Sprintf("%s timeout", operation)).
		WithOperation(operation).
		WithTimeout(timeout)
}

// NewNotFoundError creates a not found error
func NewNotFoundError(resource, identifier string) *BluetoothError {
	var code ErrorCode
	switch resource {
	case "device":
		code = ErrorCodeDeviceNotFound
	case "service":
		code = ErrorCodeServiceNotFound
	case "characteristic":
		code = ErrorCodeCharacteristicNotFound
	case "descriptor":
		code = ErrorCodeDescriptorNotFound
	default:
		code = ErrorCodeOperationFailed
	}

	return NewBluetoothErrorWithCode(code, fmt.Sprintf("%s not found: %s", resource, identifier)).
		WithContext("resource", resource).
		WithContext("identifier", identifier)
}

// NewValidationErrorBT creates a validation error as BluetoothError
func NewValidationErrorBT(parameter string, value any, reason string) *BluetoothError {
	return NewBluetoothErrorWithCode(ErrorCodeInvalidParameter, fmt.Sprintf("invalid %s: %s", parameter, reason)).
		WithContext("parameter", parameter).
		WithContext("value", value).
		WithContext("reason", reason)
}

// NewPlatformError creates a platform-specific error
func NewPlatformError(platform, operation string, cause error) *BluetoothError {
	return NewBluetoothErrorWithCode(ErrorCodePlatformNotSupported, 
		fmt.Sprintf("%s not supported on %s", operation, platform)).
		WithCause(cause).
		WithContext("platform", platform).
		WithOperation(operation)
}

// Error utilities

// IsBluetoothError checks if an error is a BluetoothError
func IsBluetoothError(err error) bool {
	_, ok := err.(*BluetoothError)
	return ok
}

// GetBluetoothError extracts BluetoothError from an error chain
func GetBluetoothError(err error) *BluetoothError {
	if bluetoothErr, ok := err.(*BluetoothError); ok {
		return bluetoothErr
	}

	// Try to unwrap and find BluetoothError
	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return GetBluetoothError(unwrapped)
	}

	return nil
}

// WrapError wraps an existing error as a BluetoothError
func WrapError(err error, code ErrorCode, message string) *BluetoothError {
	if err == nil {
		return nil
	}

	// If it's already a BluetoothError, return it
	if bluetoothErr, ok := err.(*BluetoothError); ok {
		return bluetoothErr
	}

	return NewBluetoothErrorWithCode(code, message).WithCause(err)
}

// CombineErrors combines multiple errors into a single error using errors.Join
func CombineErrors(errs ...error) error {
	var validErrors []error
	for _, err := range errs {
		if err != nil {
			validErrors = append(validErrors, err)
		}
	}

	switch len(validErrors) {
	case 0:
		return nil
	case 1:
		return validErrors[0]
	default:
		return errors.Join(validErrors...)
	}
}

// RetryableOperation represents an operation that can be retried
type RetryableOperation func() error

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts int
	InitialDelay time.Duration
	MaxDelay     time.Duration
	Multiplier   float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     5 * time.Second,
		Multiplier:   2.0,
	}
}

// RetryWithBackoff executes an operation with exponential backoff retry
func RetryWithBackoff(operation RetryableOperation, config RetryConfig) error {
	var lastError error
	delay := config.InitialDelay

	for attempt := 0; attempt < config.MaxAttempts; attempt++ {
		err := operation()
		if err == nil {
			return nil
		}

		lastError = err

		// Check if error is retryable
		if bluetoothErr := GetBluetoothError(err); bluetoothErr != nil && !bluetoothErr.IsRetryable() {
			return err
		}

		// Don't delay after the last attempt
		if attempt < config.MaxAttempts-1 {
			time.Sleep(delay)
			delay = time.Duration(float64(delay) * config.Multiplier)
			if delay > config.MaxDelay {
				delay = config.MaxDelay
			}
		}
	}

	return NewBluetoothErrorWithCode(ErrorCodeOperationFailed, "operation failed after all retry attempts").
		WithCause(lastError).
		WithContext("attempts", config.MaxAttempts)
}