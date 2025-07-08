package errors

import (
	"errors"
	"fmt"
)

// ErrorCode represents different types of gobot errors.
type ErrorCode int

const (
	// ErrConnectionFailed indicates a connection failure.
	ErrConnectionFailed ErrorCode = iota
	// ErrDeviceNotFound indicates a device was not found.
	ErrDeviceNotFound
	// ErrInvalidPin indicates an invalid pin number.
	ErrInvalidPin
	// ErrInvalidPort indicates an invalid port.
	ErrInvalidPort
	// ErrInvalidBus indicates an invalid bus number.
	ErrInvalidBus
	// ErrInvalidAddress indicates an invalid address.
	ErrInvalidAddress
	// ErrPermissionDenied indicates permission denied.
	ErrPermissionDenied
	// ErrTimeout indicates a timeout occurred.
	ErrTimeout
	// ErrNotSupported indicates an operation is not supported.
	ErrNotSupported
	// ErrInvalidArgument indicates an invalid argument.
	ErrInvalidArgument
	// ErrAlreadyStarted indicates the component is already started.
	ErrAlreadyStarted
	// ErrNotStarted indicates the component is not started.
	ErrNotStarted
)

// GobotError represents a gobot-specific error.
type GobotError struct {
	Code    ErrorCode
	Message string
	Cause   error
}

// Error returns the error message.
func (e *GobotError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Cause)
	}
	return e.Message
}

// Unwrap returns the underlying error.
func (e *GobotError) Unwrap() error {
	return e.Cause
}

// Is checks if the error matches the target.
func (e *GobotError) Is(target error) bool {
	var gobotErr *GobotError
	if errors.As(target, &gobotErr) {
		return e.Code == gobotErr.Code
	}
	return false
}

// NewError creates a new GobotError.
func NewError(code ErrorCode, message string, cause error) *GobotError {
	return &GobotError{
		Code:    code,
		Message: message,
		Cause:   cause,
	}
}

// NewConnectionError creates a connection error.
func NewConnectionError(message string, cause error) *GobotError {
	return NewError(ErrConnectionFailed, message, cause)
}

// NewDeviceNotFoundError creates a device not found error.
func NewDeviceNotFoundError(deviceName string) *GobotError {
	return NewError(ErrDeviceNotFound, fmt.Sprintf("device '%s' not found", deviceName), nil)
}

// NewInvalidPinError creates an invalid pin error.
func NewInvalidPinError(pin int) *GobotError {
	return NewError(ErrInvalidPin, fmt.Sprintf("invalid pin: %d", pin), nil)
}

// NewInvalidPortError creates an invalid port error.
func NewInvalidPortError(port string) *GobotError {
	return NewError(ErrInvalidPort, fmt.Sprintf("invalid port: %s", port), nil)
}

// NewInvalidBusError creates an invalid bus error.
func NewInvalidBusError(bus int) *GobotError {
	return NewError(ErrInvalidBus, fmt.Sprintf("invalid bus: %d", bus), nil)
}

// NewInvalidAddressError creates an invalid address error.
func NewInvalidAddressError(address int) *GobotError {
	return NewError(ErrInvalidAddress, fmt.Sprintf("invalid address: 0x%02X", address), nil)
}

// NewPermissionDeniedError creates a permission denied error.
func NewPermissionDeniedError(resource string) *GobotError {
	return NewError(ErrPermissionDenied, fmt.Sprintf("permission denied: %s", resource), nil)
}

// NewTimeoutError creates a timeout error.
func NewTimeoutError(operation string) *GobotError {
	return NewError(ErrTimeout, fmt.Sprintf("timeout: %s", operation), nil)
}

// NewNotSupportedError creates a not supported error.
func NewNotSupportedError(operation string) *GobotError {
	return NewError(ErrNotSupported, fmt.Sprintf("not supported: %s", operation), nil)
}

// NewInvalidArgumentError creates an invalid argument error.
func NewInvalidArgumentError(argument string) *GobotError {
	return NewError(ErrInvalidArgument, fmt.Sprintf("invalid argument: %s", argument), nil)
}

// NewAlreadyStartedError creates an already started error.
func NewAlreadyStartedError(component string) *GobotError {
	return NewError(ErrAlreadyStarted, fmt.Sprintf("already started: %s", component), nil)
}

// NewNotStartedError creates a not started error.
func NewNotStartedError(component string) *GobotError {
	return NewError(ErrNotStarted, fmt.Sprintf("not started: %s", component), nil)
}

// AppendError appends an error to an existing error.
func AppendError(err, newErr error) error {
	if err == nil {
		return newErr
	}
	if newErr == nil {
		return err
	}
	return fmt.Errorf("%v; %v", err, newErr)
}