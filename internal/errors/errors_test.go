package errors

import (
	"errors"
	"testing"
)

func TestGobotError(t *testing.T) {
	err := NewConnectionError("test connection failed", nil)
	
	if err.Code != ErrConnectionFailed {
		t.Errorf("Expected error code %v, got %v", ErrConnectionFailed, err.Code)
	}
	
	expectedMsg := "test connection failed"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestGobotErrorWithCause(t *testing.T) {
	cause := errors.New("underlying error")
	err := NewConnectionError("connection failed", cause)
	
	if err.Unwrap() != cause {
		t.Error("Expected unwrapped error to be the cause")
	}
	
	expectedMsg := "connection failed: underlying error"
	if err.Error() != expectedMsg {
		t.Errorf("Expected error message '%s', got '%s'", expectedMsg, err.Error())
	}
}

func TestErrorIs(t *testing.T) {
	err1 := NewConnectionError("test", nil)
	err2 := NewConnectionError("another test", nil)
	err3 := NewDeviceNotFoundError("device")
	
	if !errors.Is(err1, err2) {
		t.Error("Expected errors with same code to match with errors.Is")
	}
	
	if errors.Is(err1, err3) {
		t.Error("Expected errors with different codes not to match")
	}
}

func TestSpecificErrorConstructors(t *testing.T) {
	tests := []struct {
		name     string
		create   func() *GobotError
		expected ErrorCode
	}{
		{"DeviceNotFound", func() *GobotError { return NewDeviceNotFoundError("test") }, ErrDeviceNotFound},
		{"InvalidPin", func() *GobotError { return NewInvalidPinError(13) }, ErrInvalidPin},
		{"InvalidPort", func() *GobotError { return NewInvalidPortError("/dev/ttyUSB0") }, ErrInvalidPort},
		{"InvalidBus", func() *GobotError { return NewInvalidBusError(1) }, ErrInvalidBus},
		{"InvalidAddress", func() *GobotError { return NewInvalidAddressError(0x48) }, ErrInvalidAddress},
		{"PermissionDenied", func() *GobotError { return NewPermissionDeniedError("/dev/mem") }, ErrPermissionDenied},
		{"Timeout", func() *GobotError { return NewTimeoutError("connection") }, ErrTimeout},
		{"NotSupported", func() *GobotError { return NewNotSupportedError("PWM") }, ErrNotSupported},
		{"InvalidArgument", func() *GobotError { return NewInvalidArgumentError("frequency") }, ErrInvalidArgument},
		{"AlreadyStarted", func() *GobotError { return NewAlreadyStartedError("device") }, ErrAlreadyStarted},
		{"NotStarted", func() *GobotError { return NewNotStartedError("device") }, ErrNotStarted},
	}
	
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.create()
			if err.Code != test.expected {
				t.Errorf("Expected error code %v, got %v", test.expected, err.Code)
			}
			if err.Error() == "" {
				t.Error("Expected non-empty error message")
			}
		})
	}
}