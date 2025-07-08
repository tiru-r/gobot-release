package keyboard

import (
	"os"

	"golang.org/x/term"
)

// pureGoState holds the original terminal state for pure Go implementation
var pureGoState *term.State

// configurePureGo sets up the terminal for raw input using pure Go
func configurePureGo() error {
	var err error
	pureGoState, err = term.MakeRaw(int(os.Stdin.Fd()))
	return err
}

// restorePureGo restores the terminal to its original state using pure Go
func restorePureGo() error {
	if pureGoState != nil {
		return term.Restore(int(os.Stdin.Fd()), pureGoState)
	}
	return nil
}

// Configure sets up the terminal for raw input
// This function automatically chooses between pure Go and external stty
func Configure() error {
	// Try pure Go implementation first
	err := configurePureGo()
	if err == nil {
		return nil
	}

	// Fall back to external stty if pure Go fails
	return configure()
}

// Restore returns the terminal to its original state
// This function automatically chooses the appropriate restore method
func Restore() error {
	// Try pure Go restore first
	if pureGoState != nil {
		return restorePureGo()
	}

	// Fall back to external stty restore
	return restore()
}