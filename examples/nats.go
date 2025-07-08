//go:build example
// +build example

//
// Do not build by default.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nats-io/nats.go/jetstream"
	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/platforms/nats"
)

func main() {
	natsAdaptor := nats.NewAdaptorWithAuth("localhost:4222", 1234, "user", "pass")
	// Set a reasonable timeout for modern context operations
	natsAdaptor.SetTimeout(10 * time.Second)

	work := func() {
		// Modern NATS with context support
		ctx := context.Background()
		
		natsAdaptor.On("hello", func(msg nats.Message) {
			fmt.Printf("Received hello: %s\n", string(msg.Data))
		})
		natsAdaptor.On("hola", func(msg nats.Message) {
			fmt.Printf("Received hola: %s\n", string(msg.Data))
		})
		
		data := []byte("modern NATS message")
		gobot.Every(1*time.Second, func() {
			if !natsAdaptor.Publish("hello", data) {
				fmt.Println("Failed to publish hello")
			}
		})
		gobot.Every(5*time.Second, func() {
			if !natsAdaptor.Publish("hola", data) {
				fmt.Println("Failed to publish hola")
			}
		})
		
		// Example of modern JetStream usage if available
		if js := natsAdaptor.JetStream(); js != nil {
			fmt.Println("JetStream is available for advanced messaging")
			// Create a stream (optional example)
			ctxTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()
			if _, err := js.CreateOrUpdateStream(ctxTimeout, jetstream.StreamConfig{
				Name:     "EVENTS",
				Subjects: []string{"hello", "hola"},
			}); err != nil {
				fmt.Printf("Stream creation failed: %v\n", err)
			}
		}
	}

	robot := gobot.NewRobot("natsBot",
		[]gobot.Connection{natsAdaptor},
		work,
	)

	if err := robot.Start(); err != nil {
		panic(err)
	}
}
