//go:build example
// +build example

//
// Do not build by default.

package main

import (
	"fmt"
	"net/http"
	"os"
	"slices"
	"sync"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/pkg/api"
	"gocv.io/x/gocv"
)

// MJPEGStream represents an MJPEG stream using Go's standard library
type MJPEGStream struct {
	mu      sync.RWMutex
	current []byte
}

// NewMJPEGStream creates a new MJPEG stream
func NewMJPEGStream() *MJPEGStream {
	return &MJPEGStream{}
}

// UpdateJPEG updates the current frame
func (s *MJPEGStream) UpdateJPEG(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.current = slices.Clone(data)
}

// ServeHTTP implements http.Handler for serving MJPEG stream
func (s *MJPEGStream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "multipart/x-mixed-replace; boundary=frame")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for {
		s.mu.RLock()
		frame := slices.Clone(s.current)
		s.mu.RUnlock()

		if len(frame) > 0 {
			fmt.Fprintf(w, "\r\n--frame\r\n")
			fmt.Fprintf(w, "Content-Type: image/jpeg\r\n")
			fmt.Fprintf(w, "Content-Length: %d\r\n\r\n", len(frame))
			w.Write(frame)
		}

		// Check if client disconnected
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		} else {
			break
		}

		// Small delay to prevent excessive CPU usage
		// In a real implementation, you'd want to use a proper timing mechanism
		select {
		case <-r.Context().Done():
			return
		default:
		}
	}
}

var (
	deviceID int
	err      error
	webcam   *gocv.VideoCapture
	stream   *MJPEGStream
)

func main() {
	// parse args
	deviceID := os.Args[1]

	manager := gobot.NewManager()

	a := api.NewAPI(manager)

	// add the standard C3PIO API routes manually.
	a.AddC3PIORoutes()

	// starts the API without the default C3PIO API and web interface.
	// However, the C3PIO API was added manually using a.AddC3PIORoutes() which
	// means the REST API will be available, but not the web interface.
	a.StartWithoutDefaults()

	hello := manager.AddRobot(gobot.NewRobot("hello"))

	hello.AddCommand("hi_there", func(params map[string]interface{}) interface{} {
		return fmt.Sprintf("This command is attached to the robot %v", hello.Name)
	})

	// open webcam
	webcam, err = gocv.OpenVideoCapture(deviceID)
	if err != nil {
		fmt.Printf("Error opening capture device: %v\n", deviceID)
		return
	}
	defer webcam.Close()

	// create the mjpeg stream
	stream = NewMJPEGStream()
	http.Handle("/video", stream)

	// start capturing
	go mjpegCapture()

	if err := manager.Start(); err != nil {
		panic(err)
	}
}

func mjpegCapture() {
	img := gocv.NewMat()
	defer img.Close()

	for {
		if ok := webcam.Read(&img); !ok {
			fmt.Printf("Device closed: %v\n", deviceID)
			return
		}
		if img.Empty() {
			continue
		}

		buf, _ := gocv.IMEncode(".jpg", img)
		stream.UpdateJPEG(buf)
	}
}
