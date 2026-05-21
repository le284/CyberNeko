//go:build !darwin && !windows

package main

func cursorPosition() (screenPoint, bool) {
	return screenPoint{}, false
}
