//go:build example
// +build example

//
// Do not build by default.

// TO RUN:
//
//	go run ./examples/nats_driver_ping.go <SERVER>
//
// EXAMPLE:
//
//	go run ./examples/nats_driver_ping.go tls://nats.demo.io:4443
package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/pkg/platforms/nats"
)

func main() {
	natsAdaptor := nats.NewAdaptor(os.Args[1], 1234)
	// Set timeout for modern context operations
	natsAdaptor.SetTimeout(10 * time.Second)

	holaDriver := nats.NewDriver(natsAdaptor, "hola")
	helloDriver := nats.NewDriver(natsAdaptor, "hello")

	work := func() {
		ctx := context.Background()
		
		_ = helloDriver.On(nats.Data, func(msg nats.Message) {
			fmt.Printf("Hello received: %s\n", string(msg.Data))
		})

		_ = holaDriver.On(nats.Data, func(msg nats.Message) {
			fmt.Printf("Hola received: %s\n", string(msg.Data))
		})

		data := []byte("modern NATS driver message")
		gobot.Every(1*time.Second, func() {
			if !helloDriver.Publish(data) {
				fmt.Println("Failed to publish hello")
			}
		})

		gobot.Every(5*time.Second, func() {
			if !holaDriver.Publish(data) {
				fmt.Println("Failed to publish hola")
			}
		})
		
		// Example of modern JetStream consumer
		go func() {
			time.Sleep(2 * time.Second) // Wait for connection
			if err := helloDriver.ConsumeMessages(ctx, 10, func(msg nats.Message) {
				fmt.Printf("JetStream hello: %s\n", string(msg.Data))
			}); err != nil {
				fmt.Printf("JetStream consumer failed: %v\n", err)
			}
		}()
	}

	robot := gobot.NewRobot("natsBot",
		[]gobot.Connection{natsAdaptor},
		[]gobot.Device{helloDriver, holaDriver},
		work,
	)

	if err := robot.Start(); err != nil {
		panic(err)
	}
}
