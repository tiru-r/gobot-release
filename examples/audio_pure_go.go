//go:build purgo && example
// +build purgo,example

// Pure Go Audio Example
// This example demonstrates pure Go audio playback without external dependencies
// Build with: CGO_ENABLED=0 go build -tags 'purgo,example' ./examples/audio_pure_go.go

package main

import (
	"fmt"
	"log"
	"time"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/pkg/platforms/audio"
)

func main() {
	fmt.Println("🎵 Pure Go Audio Example")
	fmt.Println("========================")
	fmt.Println("This example demonstrates pure Go audio playback without external dependencies.")
	fmt.Println()

	// Create pure Go audio adaptor
	audioAdaptor := audio.NewAdaptor()
	
	// Create audio driver with a test file
	audioDriver := audio.NewDriver(audioAdaptor, "./examples/laser.mp3")
	
	work := func() {
		fmt.Println("🎵 Starting audio demo...")
		
		// Generate test tones
		if pureAdaptor, ok := audioAdaptor.(*audio.PureGoAdaptor); ok {
			fmt.Println("🎶 Generating test tones...")
			
			// Generate different frequency tones
			frequencies := []float64{440.0, 523.25, 659.25, 783.99} // A4, C5, E5, G5
			for i, freq := range frequencies {
				fmt.Printf("   Playing tone %d: %.2f Hz\n", i+1, freq)
				err := pureAdaptor.GenerateTone(freq, 500*time.Millisecond)
				if err != nil {
					log.Printf("Error generating tone: %v", err)
				}
				time.Sleep(100 * time.Millisecond)
			}
			
			fmt.Println("🎵 Tones complete!")
		}
		
		// Try to play the laser sound file (will simulate playback)
		fmt.Println("🔫 Attempting to play laser sound...")
		if driver, ok := audioDriver.(*audio.PureGoDriver); ok {
			errors := driver.Play()
			if len(errors) > 0 {
				fmt.Printf("   Note: %v (file may not exist - this is expected)\n", errors[0])
			}
		}
		
		// Create a simple WAV file for testing
		fmt.Println("🎵 Creating test WAV file...")
		wavFile := createTestWAVFile()
		if wavFile != "" {
			fmt.Printf("   Created: %s\n", wavFile)
			fmt.Println("   Playing test WAV file...")
			if driver, ok := audioDriver.(*audio.PureGoDriver); ok {
				errors := driver.Sound(wavFile)
				if len(errors) == 0 {
					fmt.Println("   ✅ WAV file playback successful!")
				} else {
					fmt.Printf("   ❌ Error playing WAV: %v\n", errors[0])
				}
			}
		}
		
		// Demonstrate commander interface
		fmt.Println("🎮 Testing commander interface...")
		if driver, ok := audioDriver.(*audio.PureGoDriver); ok {
			result := driver.Command("tone")(map[string]interface{}{
				"frequency": 880.0,
				"duration":  200 * time.Millisecond,
			})
			if errors, ok := result.([]error); ok && len(errors) == 0 {
				fmt.Println("   ✅ Commander tone generation successful!")
			} else {
				fmt.Printf("   ❌ Commander error: %v\n", result)
			}
		}
		
		fmt.Println("🎵 Audio demo complete!")
	}

	robot := gobot.NewRobot("pureGoAudioBot",
		[]gobot.Connection{audioAdaptor},
		[]gobot.Device{audioDriver},
		work,
	)

	fmt.Println("🤖 Starting robot...")
	if err := robot.Start(); err != nil {
		log.Fatal(err)
	}
}

// createTestWAVFile creates a simple WAV file for testing
func createTestWAVFile() string {
	// This is a simplified version - in a real implementation,
	// you might want to use a more sophisticated WAV generation
	return "" // For this example, we'll just return empty string
}