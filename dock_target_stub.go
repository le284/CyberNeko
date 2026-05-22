//go:build !darwin

package main

import "github.com/wailsapp/wails/v3/pkg/application"

func dockEdgeTarget(app *application.App) (targetRect, bool) {
	return targetRect{}, false
}
