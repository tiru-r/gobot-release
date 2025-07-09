package main

import (
	"fmt"
	"time"

	"gobot.io/x/gobot/v2/pkg/core"
)

func main() {
	// Create a simple robot
	robot := core.NewRobot(
		core.WithName("HelloBot"),
		core.WithWork(func() {
			for i := range 5 {
				fmt.Printf("Hello from Gobot! Count: %d\n", i+1)
				time.Sleep(1 * time.Second)
			}
		}),
		core.WithAutoRun(true),
	)

	// Start the robot
	if err := robot.Start(); err != nil {
		fmt.Printf("Error starting robot: %v\n", err)
	}
}