//go:build darwin

package main

/*
#cgo LDFLAGS: -framework CoreGraphics
#include <CoreGraphics/CoreGraphics.h>
#include <stdbool.h>

typedef struct CursorPoint {
	int x;
	int y;
	bool ok;
} CursorPoint;

CursorPoint getCursorLocation(void) {
	CursorPoint result = {0, 0, false};
	CGEventRef event = CGEventCreate(NULL);
	if (event == NULL) {
		return result;
	}
	CGPoint point = CGEventGetLocation(event);
	CFRelease(event);
	result.x = (int)point.x;
	result.y = (int)point.y;
	result.ok = true;
	return result;
}
*/
import "C"

func cursorPosition() (screenPoint, bool) {
	point := C.getCursorLocation()
	if !bool(point.ok) {
		return screenPoint{}, false
	}
	return screenPoint{X: int(point.x), Y: int(point.y)}, true
}
