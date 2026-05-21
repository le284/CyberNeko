//go:build !darwin && !windows

package main

func activeWindowTarget() (targetRect, bool) {
	return targetRect{}, false
}

func windowEdgeTargetAt(feetX int, feetY int) (targetRect, bool) {
	return targetRect{}, false
}
