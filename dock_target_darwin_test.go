//go:build darwin

package main

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestDockEdgeTargetUsesVisibleBottomDock(t *testing.T) {
	target, ok := dockEdgeTargetFromScreen(
		application.Rect{X: 0, Y: 0, Width: 1512, Height: 982},
		application.Rect{X: 0, Y: 80, Width: 1512, Height: 869},
		2,
	)
	if !ok {
		t.Fatal("expected bottom Dock target")
	}

	line := lineForTarget(target)
	if line.Y != 1496 {
		t.Fatalf("expected paws to align with Dock top, got y=%d", line.Y)
	}
	if line.Edge != "dock" {
		t.Fatalf("expected Dock edge perch, got %q", line.Edge)
	}
}

func TestDockEdgeTargetUsesBottomScreenEdgeWhenDockAutohides(t *testing.T) {
	target, ok := dockEdgeTargetFromScreen(
		application.Rect{X: 0, Y: 0, Width: 1512, Height: 982},
		application.Rect{X: 0, Y: 0, Width: 1512, Height: 949},
		2,
	)
	if !ok {
		t.Fatal("expected hidden Dock to use the bottom screen rail")
	}

	line := lineForTarget(target)
	if line.Y != 1656 {
		t.Fatalf("expected paws on bottom rail, got y=%d", line.Y)
	}
}

func TestDockEdgeTargetRejectsSideDock(t *testing.T) {
	_, ok := dockEdgeTargetFromScreen(
		application.Rect{X: 0, Y: 0, Width: 1512, Height: 982},
		application.Rect{X: 80, Y: 0, Width: 1432, Height: 949},
		2,
	)
	if ok {
		t.Fatal("expected side Dock to fall back to active-window edge behavior")
	}
}
