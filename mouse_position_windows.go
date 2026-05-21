//go:build windows

package main

import (
	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
)

func cursorPosition() (screenPoint, bool) {
	x, y, ok := w32.GetCursorPos()
	if !ok {
		return screenPoint{}, false
	}
	point := application.PhysicalToDipPoint(application.Point{X: x, Y: y})
	return screenPoint{X: point.X, Y: point.Y}, true
}
