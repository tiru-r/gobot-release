package errors

import (
	"errors"
	"fmt"
)

// AppendError appends an error to an existing error using Go's native errors.Join.
// This replaces the functionality of github.com/hashicorp/go-multierror.
// If baseErr is nil, it returns newErr. If newErr is nil, it returns baseErr.
// If both are non-nil, it joins them into a single error.
//
// Usage pattern:
//   var err error
//   if e := operation1(); e != nil {
//       err = AppendError(err, e)
//   }
//   if e := operation2(); e != nil {
//       err = AppendError(err, e)
//   }
//   return err
func AppendError(baseErr, newErr error) error {
	if newErr == nil {
		return baseErr
	}
	if baseErr == nil {
		return newErr
	}
	return errors.Join(baseErr, newErr)
}

// AppendErrorf appends a formatted error to an existing error.
// This is a convenience function for common error formatting patterns.
func AppendErrorf(baseErr error, format string, args ...interface{}) error {
	newErr := fmt.Errorf(format, args...)
	return AppendError(baseErr, newErr)
}

// Additional error constructors for common scenarios

// NewConnectionError creates a connection error
func NewConnectionError(message string, cause error) error {
	if cause != nil {
		return fmt.Errorf("connection error: %s: %w", message, cause)
	}
	return fmt.Errorf("connection error: %s", message)
}

// NewDeviceNotFoundError creates a device not found error
func NewDeviceNotFoundError(deviceName string) error {
	return fmt.Errorf("device not found: %s", deviceName)
}

// NewInvalidPinError creates an invalid pin error
func NewInvalidPinError(pin int) error {
	return fmt.Errorf("invalid pin: %d", pin)
}

// NewInvalidPortError creates an invalid port error
func NewInvalidPortError(port string) error {
	return fmt.Errorf("invalid port: %s", port)
}

// NewInvalidBusError creates an invalid bus error
func NewInvalidBusError(bus int) error {
	return fmt.Errorf("invalid bus: %d", bus)
}

// NewInvalidAddressError creates an invalid address error
func NewInvalidAddressError(address int) error {
	return fmt.Errorf("invalid address: %d", address)
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(resource string) error {
	return fmt.Errorf("permission denied: %s", resource)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) error {
	return fmt.Errorf("timeout: %s", operation)
}

// NewNotSupportedError creates a not supported error
func NewNotSupportedError(operation string) error {
	return fmt.Errorf("not supported: %s", operation)
}

// NewInvalidArgumentError creates an invalid argument error
func NewInvalidArgumentError(argument string) error {
	return fmt.Errorf("invalid argument: %s", argument)
}

// NewAlreadyStartedError creates an already started error
func NewAlreadyStartedError(component string) error {
	return fmt.Errorf("already started: %s", component)
}

// NewNotStartedError creates a not started error
func NewNotStartedError(component string) error {
	return fmt.Errorf("not started: %s", component)
}