//go:build example
// +build example

// Simple BLE client example using the new Gobot-style Bluetooth implementation

package main

import (
	"fmt"
	"os"
	"time"

	"gobot.io/x/gobot/v2"
	"gobot.io/x/gobot/v2/bluetooth"
	"gobot.io/x/gobot/v2/drivers/ble"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run bleclient_simple_battery.go <device_address_or_name>")
		os.Exit(1)
	}

	// Create a simple BLE client adaptor
	bleAdaptor := bluetooth.NewClientAdaptor(
		os.Args[1], 
		bluetooth.WithScanTimeout(30*time.Second),
	)
	
	// Create a battery driver
	battery := ble.NewBatteryDriver(bleAdaptor)

	work := func() {
		gobot.Every(5*time.Second, func() {
			level, err := battery.GetBatteryLevel()
			if err != nil {
				fmt.Println("Error reading battery:", err)
			} else {
				fmt.Printf("Battery level: %d%%\n", level)
			}
		})
	}

	robot := gobot.NewRobot("simpleBLEBot",
		[]gobot.Connection{bleAdaptor},
		[]gobot.Device{battery},
		work,
	)

	if err := robot.Start(); err != nil {
		panic(err)
	}
}