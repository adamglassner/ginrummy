//go:build !windows && !js && !headless

package main

import (
	"fmt"
	"os"
)

func fatal(text string) {
	fmt.Fprintln(os.Stderr, text)
	os.Exit(1)
}

func singleInstance() bool { return true }

func testFlags() (bool, bool) {
	return os.Getenv("GINRUMMY_AUTOPILOT") != "", os.Getenv("GINRUMMY_FAST") != ""
}
