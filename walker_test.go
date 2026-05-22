package main

import (
	"math/rand"
	"testing"
)

func TestLineForTargetUsesTopEdgeWhenThereIsRoom(t *testing.T) {
	line := lineForTarget(targetRect{X: 100, Y: 300, Width: 800, Height: 500, ScreenX: 0, ScreenY: 0, ScreenWidth: 1200, ScreenHeight: 900})

	if line.Y != 146 {
		t.Fatalf("expected pet paws on top edge, got y=%d", line.Y)
	}
	if line.Left != 100 || line.Right != 680 {
		t.Fatalf("unexpected walking range: %+v", line)
	}
	if line.Edge != "top" {
		t.Fatalf("expected top edge, got %q", line.Edge)
	}
}

func TestLineForTargetFallsBackToBottomEdgeWhenTopHasNoRoom(t *testing.T) {
	line := lineForTarget(targetRect{X: 100, Y: 33, Width: 800, Height: 500, ScreenX: 0, ScreenY: 0, ScreenWidth: 1200, ScreenHeight: 900})

	if line.Y != 379 {
		t.Fatalf("expected pet paws on bottom edge, got y=%d", line.Y)
	}
	if line.Edge != "bottom" {
		t.Fatalf("expected bottom edge, got %q", line.Edge)
	}
}

func TestChooseNextWaypointStaysInsideLine(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	line := walkLine{Left: 100, Right: 300, Y: 50, Edge: "top"}

	for i := 0; i < 64; i++ {
		waypoint := chooseNextWaypoint(rng, line, 180)
		if waypoint < line.Left || waypoint > line.Right {
			t.Fatalf("waypoint out of line: %d", waypoint)
		}
	}
}

func TestNextFollowPositionMovesTowardCursor(t *testing.T) {
	bounds := targetRect{ScreenX: 0, ScreenY: 0, ScreenWidth: 1200, ScreenHeight: 900}
	nextX, nextY, moving := nextFollowPosition(100, 100, screenPoint{X: 600, Y: 500}, bounds)

	if !moving {
		t.Fatal("expected pet to move toward distant cursor")
	}
	if nextX <= 100 || nextY <= 100 {
		t.Fatalf("expected movement down and right, got (%d,%d)", nextX, nextY)
	}
}

func TestNextFollowPositionStopsNearCursor(t *testing.T) {
	bounds := targetRect{ScreenX: 0, ScreenY: 0, ScreenWidth: 1200, ScreenHeight: 900}
	nextX, nextY, moving := nextFollowPosition(100, 100, screenPoint{X: 210, Y: 220}, bounds)

	if moving {
		t.Fatal("expected pet to stop near cursor")
	}
	if nextX != 100 || nextY != 100 {
		t.Fatalf("expected position to stay still, got (%d,%d)", nextX, nextY)
	}
}

func TestMovementModeIsPerWindow(t *testing.T) {
	first := "test-pet-first"
	second := "test-pet-second"

	if currentMovementMode(first) != movementModeEdgeWander {
		t.Fatal("expected unknown pets to default to edge wander mode")
	}

	setMovementMode(first, movementModeFollowMouse)
	setMovementMode(second, movementModeIdle)

	if currentMovementMode(first) != movementModeFollowMouse {
		t.Fatal("expected first pet to keep follow mouse mode")
	}
	if currentMovementMode(second) != movementModeIdle {
		t.Fatal("expected second pet to keep idle mode")
	}
}
