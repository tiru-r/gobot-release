package gobot

import (
	"errors"
	"fmt"

	gobotErrors "gobot.io/x/gobot/v2/internal/errors"
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

// Convenience functions that delegate to the internal errors package
// These provide a unified API while using the structured error system internally

// NewConnectionError creates a connection error
func NewConnectionError(message string, cause error) error {
	return gobotErrors.NewConnectionError(message, cause)
}

// NewDeviceNotFoundError creates a device not found error
func NewDeviceNotFoundError(deviceName string) error {
	return gobotErrors.NewDeviceNotFoundError(deviceName)
}

// NewInvalidPinError creates an invalid pin error
func NewInvalidPinError(pin int) error {
	return gobotErrors.NewInvalidPinError(pin)
}

// NewInvalidPortError creates an invalid port error
func NewInvalidPortError(port string) error {
	return gobotErrors.NewInvalidPortError(port)
}

// NewInvalidBusError creates an invalid bus error
func NewInvalidBusError(bus int) error {
	return gobotErrors.NewInvalidBusError(bus)
}

// NewInvalidAddressError creates an invalid address error
func NewInvalidAddressError(address int) error {
	return gobotErrors.NewInvalidAddressError(address)
}

// NewPermissionDeniedError creates a permission denied error
func NewPermissionDeniedError(resource string) error {
	return gobotErrors.NewPermissionDeniedError(resource)
}

// NewTimeoutError creates a timeout error
func NewTimeoutError(operation string) error {
	return gobotErrors.NewTimeoutError(operation)
}

// NewNotSupportedError creates a not supported error
func NewNotSupportedError(operation string) error {
	return gobotErrors.NewNotSupportedError(operation)
}

// NewInvalidArgumentError creates an invalid argument error
func NewInvalidArgumentError(argument string) error {
	return gobotErrors.NewInvalidArgumentError(argument)
}

// NewAlreadyStartedError creates an already started error
func NewAlreadyStartedError(component string) error {
	return gobotErrors.NewAlreadyStartedError(component)
}

// NewNotStartedError creates a not started error
func NewNotStartedError(component string) error {
	return gobotErrors.NewNotStartedError(component)
}