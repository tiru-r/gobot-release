//go:build !gocv
// +build !gocv

package opencv

import (
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"

	"gobot.io/x/gobot/v2"
)

const (
	// Frame event
	Frame = "frame"
)

// PureGoCamera represents a pure Go camera interface
type PureGoCamera interface {
	Read() (image.Image, error)
	Close() error
}

// CameraDriver is the pure Go implementation of the OpenCV camera driver
type CameraDriver struct {
	name   string
	camera PureGoCamera
	Source interface{}
	start  func(*CameraDriver) error
	gobot.Eventer
}

// NewCameraDriver creates a new pure Go camera driver with specified source.
func NewCameraDriver(source interface{}) *CameraDriver {
	c := &CameraDriver{
		name:    "Camera",
		Eventer: gobot.NewEventer(),
		Source:  source,
		start: func(c *CameraDriver) error {
			switch v := c.Source.(type) {
			case string:
				// For file sources, create a file-based camera
				c.camera = &FileCamera{filename: v}
			case int:
				// For device sources, we'll use a mock camera for pure Go
				c.camera = &MockCamera{deviceID: v}
			default:
				return errors.New("Unknown camera source")
			}
			return nil
		},
	}

	c.AddEvent(Frame)
	return c
}

// Name returns the name of the device.
func (c *CameraDriver) Name() string { return c.name }

// SetName sets the name of the device.
func (c *CameraDriver) SetName(n string) { c.name = n }

// Connection returns the driver connection
func (c *CameraDriver) Connection() gobot.Connection { return nil }

// Start starts the camera driver
func (c *CameraDriver) Start() error {
	return c.start(c)
}

// Halt stops the camera driver
func (c *CameraDriver) Halt() error {
	if c.camera != nil {
		return c.camera.Close()
	}
	return nil
}

// Read returns the next frame from the camera
func (c *CameraDriver) Read() (image.Image, error) {
	if c.camera == nil {
		return nil, errors.New("camera not initialized")
	}
	return c.camera.Read()
}

// FileCamera implements PureGoCamera for file-based sources
type FileCamera struct {
	filename string
	file     *os.File
}

func (fc *FileCamera) Read() (image.Image, error) {
	if fc.file == nil {
		var err error
		fc.file, err = os.Open(fc.filename)
		if err != nil {
			return nil, err
		}
	}

	// Try to decode as JPEG first, then PNG
	fc.file.Seek(0, 0)
	img, err := jpeg.Decode(fc.file)
	if err != nil {
		fc.file.Seek(0, 0)
		img, err = png.Decode(fc.file)
		if err != nil {
			return nil, err
		}
	}
	return img, nil
}

func (fc *FileCamera) Close() error {
	if fc.file != nil {
		return fc.file.Close()
	}
	return nil
}

// MockCamera implements PureGoCamera for device sources (mock implementation)
type MockCamera struct {
	deviceID int
}

func (mc *MockCamera) Read() (image.Image, error) {
	// Create a simple colored rectangle as a mock frame
	img := image.NewRGBA(image.Rect(0, 0, 640, 480))
	// Fill with a simple pattern based on device ID
	baseColor := uint8((mc.deviceID * 50) % 255)
	for y := 0; y < 480; y++ {
		for x := 0; x < 640; x++ {
			img.Set(x, y, color.RGBA{baseColor, uint8(x%255), uint8(y%255), 255})
		}
	}
	return img, nil
}

func (mc *MockCamera) Close() error {
	return nil
}