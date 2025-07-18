//go:build utils
// +build utils

// Do not build by default.
//
// Joystick scanner
// Based on original code from
// https://github.com/0xcafed00d/joystick/blob/master/joysticktest/joysticktest.go
// Simple program that displays the state of the specified joystick
//
//	go run joysticktest.go 2
//
// displays state of joystick id 2
package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"gobot.io/x/gobot/v2/platforms/joystick"
	"golang.org/x/term"
)

var termState *term.State

func printAt(x, y int, s string) {
	// Move cursor to position and print text
	fmt.Printf("\033[%d;%dH%s", y+1, x+1, s)
}

func clearScreen() {
	fmt.Print("\033[2J\033[H")
}

func hideCursor() {
	fmt.Print("\033[?25l")
}

func showCursor() {
	fmt.Print("\033[?25h")
}

func initTerminal() error {
	var err error
	termState, err = term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return err
	}
	clearScreen()
	hideCursor()
	return nil
}

func closeTerminal() {
	if termState != nil {
		term.Restore(int(os.Stdin.Fd()), termState)
	}
	showCursor()
	clearScreen()
}

func checkKeyPress() bool {
	// Non-blocking read from stdin
	buffer := make([]byte, 1)
	os.Stdin.SetReadDeadline(time.Now())
	n, err := os.Stdin.Read(buffer)
	if err != nil || n == 0 {
		return false
	}
	return buffer[0] == 'q'
}

func readJoystick(js joystick.Joystick) {
	jinfo, err := js.Read()
	if err != nil {
		printAt(1, 5, "Error: "+err.Error())
		return
	}

	printAt(1, 5, "Buttons:")
	for button := 0; button < js.ButtonCount(); button++ {
		//nolint:gosec // TODO: fix later
		if jinfo.Buttons&(1<<uint32(button)) != 0 {
			printAt(10+button, 5, "X")
			printAt(1, 6, fmt.Sprintf("Button %2d Pressed", button))
		} else {
			printAt(10+button, 5, ".")
		}
	}

	for axis := 0; axis < js.AxisCount(); axis++ {
		printAt(1, axis+8, fmt.Sprintf("Axis %2d Value: %7d", axis, jinfo.AxisData[axis]))
	}
}

func main() {
	jsid := 0
	if len(os.Args) > 1 {
		i, err := strconv.Atoi(os.Args[1])
		if err != nil {
			fmt.Println(err)
			return
		}
		jsid = i
	}

	js, jserr := joystick.Open(jsid)

	if jserr != nil {
		fmt.Println(jserr)
		return
	}

	err := initTerminal()
	if err != nil {
		panic(err)
	}
	defer closeTerminal()

	ticker := time.NewTicker(time.Millisecond * 40)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			printAt(1, 0, "-- Press 'q' to Exit --")
			printAt(1, 1, fmt.Sprintf("Joystick Name: %s", js.Name())) //nolint:perfsprint // ok here
			printAt(1, 2, fmt.Sprintf("   Axis Count: %d", js.AxisCount()))
			printAt(1, 3, fmt.Sprintf(" Button Count: %d", js.ButtonCount()))
			readJoystick(js)

			// Check for quit key
			if checkKeyPress() {
				return
			}
		}
	}
}
