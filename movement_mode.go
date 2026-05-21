package main

import "sync"

type movementMode int

const (
	movementModeIdle movementMode = iota
	movementModeEdgeWander
	movementModeFollowMouse
)

var movementModes = struct {
	sync.RWMutex
	byWindow map[string]movementMode
}{byWindow: make(map[string]movementMode)}

func setMovementMode(windowName string, mode movementMode) {
	movementModes.Lock()
	defer movementModes.Unlock()
	movementModes.byWindow[windowName] = mode
}

func currentMovementMode(windowName string) movementMode {
	movementModes.RLock()
	defer movementModes.RUnlock()
	mode, ok := movementModes.byWindow[windowName]
	if !ok {
		return movementModeEdgeWander
	}
	return mode
}
