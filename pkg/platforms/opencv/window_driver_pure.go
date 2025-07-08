//go:build !gocv
// +build !gocv

package opencv

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"gobot.io/x/gobot/v2"
)

// WindowDriver is the pure Go implementation of the OpenCV window driver
type WindowDriver struct {
	name   string
	window PureGoWindow
	gobot.Eventer
}

// PureGoWindow represents a pure Go window interface
type PureGoWindow interface {
	ShowImage(img image.Image) error
	WaitKey(delay int) int
	IsOpen() bool
	Close() error
}

// NewWindowDriver creates a new pure Go window driver
func NewWindowDriver() *WindowDriver {
	w := &WindowDriver{
		name:    "Window",
		Eventer: gobot.NewEventer(),
		window:  &FileWindow{},
	}
	return w
}

// Name returns the name of the device.
func (w *WindowDriver) Name() string { return w.name }

// SetName sets the name of the device.
func (w *WindowDriver) SetName(n string) { w.name = n }

// Connection returns the driver connection
func (w *WindowDriver) Connection() gobot.Connection { return nil }

// Start starts the window driver
func (w *WindowDriver) Start() error {
	return nil
}

// Halt stops the window driver
func (w *WindowDriver) Halt() error {
	if w.window != nil {
		return w.window.Close()
	}
	return nil
}

// ShowImage displays an image in the window
func (w *WindowDriver) ShowImage(img image.Image) error {
	if w.window == nil {
		return errors.New("window not initialized")
	}
	return w.window.ShowImage(img)
}

// WaitKey waits for a key press
func (w *WindowDriver) WaitKey(delay int) int {
	if w.window == nil {
		return -1
	}
	return w.window.WaitKey(delay)
}

// IsOpen returns whether the window is open
func (w *WindowDriver) IsOpen() bool {
	if w.window == nil {
		return false
	}
	return w.window.IsOpen()
}

// FileWindow implements PureGoWindow by saving images to files
type FileWindow struct {
	frameCount int
	isOpen     bool
	outputDir  string
}

func (fw *FileWindow) ShowImage(img image.Image) error {
	if fw.outputDir == "" {
		fw.outputDir = "opencv_output"
		os.MkdirAll(fw.outputDir, 0755)
	}

	fw.frameCount++
	filename := filepath.Join(fw.outputDir, fmt.Sprintf("frame_%06d.png", fw.frameCount))
	
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	err = png.Encode(file, img)
	if err != nil {
		return err
	}

	fmt.Printf("Saved frame to: %s\n", filename)
	fw.isOpen = true
	return nil
}

func (fw *FileWindow) WaitKey(delay int) int {
	// In pure Go implementation, simulate key press
	// In a real implementation, you could use terminal input
	return 27 // ESC key
}

func (fw *FileWindow) IsOpen() bool {
	return fw.isOpen
}

func (fw *FileWindow) Close() error {
	fw.isOpen = false
	return nil
}