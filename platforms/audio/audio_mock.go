//go:build (!cgo || purgo)
// +build !cgo purgo

package audio

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"time"

	"gobot.io/x/gobot/v2"
)

// PureGoAdaptor is a pure Go audio playback adaptor using mock implementation
type PureGoAdaptor struct {
	name        string
	initialized bool
}

// NewPureGoAdaptor returns a new pure Go audio Adaptor
func NewPureGoAdaptor() *PureGoAdaptor {
	return &PureGoAdaptor{
		name: gobot.DefaultName("PureGoAudio"),
	}
}

// Name returns the Adaptor Name
func (a *PureGoAdaptor) Name() string { return a.name }

// SetName sets the Adaptor Name
func (a *PureGoAdaptor) SetName(n string) { a.name = n }

// Connect establishes a connection to the Pure Go Audio adaptor
func (a *PureGoAdaptor) Connect() error {
	a.initialized = true
	log.Println("Pure Go Audio mock adaptor initialized")
	return nil
}

// Finalize terminates the connection to the Pure Go Audio adaptor
func (a *PureGoAdaptor) Finalize() error {
	a.initialized = false
	log.Println("Pure Go Audio mock adaptor finalized")
	return nil
}

// Sound plays a sound file using pure Go mock implementation and accepts:
//
//	string: The filename of the audio to start playing
func (a *PureGoAdaptor) Sound(fileName string) []error {
	var errorsList []error

	if fileName == "" {
		log.Println("Requires filename for audio file.")
		errorsList = append(errorsList, errors.New("requires filename for audio file"))
		return errorsList
	}

	_, err := os.Stat(fileName)
	if err != nil {
		log.Println(err)
		errorsList = append(errorsList, err)
		return errorsList
	}

	if !a.initialized {
		err := errors.New("audio context not initialized")
		log.Println(err)
		errorsList = append(errorsList, err)
		return errorsList
	}

	// Mock play audio file based on file type
	err = a.playFile(fileName)
	if err != nil {
		log.Println(err)
		errorsList = append(errorsList, err)
		return errorsList
	}

	return nil
}

// playFile mocks playing an audio file
func (a *PureGoAdaptor) playFile(fileName string) error {
	fileType := path.Ext(fileName)
	
	switch fileType {
	case ".wav":
		return a.mockPlayWAV(fileName)
	case ".mp3":
		return a.mockPlayMP3(fileName)
	case ".raw", ".pcm":
		return a.mockPlayRaw(fileName)
	default:
		return fmt.Errorf("unsupported audio format: %s (supported: .wav, .mp3, .raw, .pcm)", fileType)
	}
}

// mockPlayWAV mocks playing a WAV file
func (a *PureGoAdaptor) mockPlayWAV(fileName string) error {
	log.Printf("Mock: Playing WAV file: %s", fileName)
	
	// Simulate playback duration
	time.Sleep(100 * time.Millisecond)
	
	log.Printf("Mock: Finished playing WAV file: %s", fileName)
	return nil
}

// mockPlayMP3 mocks playing an MP3 file
func (a *PureGoAdaptor) mockPlayMP3(fileName string) error {
	log.Printf("Mock: Playing MP3 file: %s", fileName)
	
	// Simulate playback duration
	time.Sleep(100 * time.Millisecond)
	
	log.Printf("Mock: Finished playing MP3 file: %s", fileName)
	return nil
}

// mockPlayRaw mocks playing a raw PCM file
func (a *PureGoAdaptor) mockPlayRaw(fileName string) error {
	log.Printf("Mock: Playing raw PCM file: %s", fileName)
	
	// Simulate playback duration
	time.Sleep(100 * time.Millisecond)
	
	log.Printf("Mock: Finished playing raw PCM file: %s", fileName)
	return nil
}

// GenerateTone generates a mock tone for testing
func (a *PureGoAdaptor) GenerateTone(frequency float64, duration time.Duration) error {
	if !a.initialized {
		return errors.New("audio context not initialized")
	}

	log.Printf("Mock: Generating tone at %.2f Hz for %v", frequency, duration)
	
	// Simulate tone generation
	time.Sleep(duration)
	
	log.Printf("Mock: Finished generating tone")
	return nil
}

// NewAdaptor is an alias for NewPureGoAdaptor when using pure Go build
func NewAdaptor() gobot.Adaptor {
	return NewPureGoAdaptor()
}