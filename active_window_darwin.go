//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework CoreGraphics -framework Foundation -framework AppKit

#include <CoreGraphics/CoreGraphics.h>
#include <Foundation/Foundation.h>
#include <AppKit/AppKit.h>
#include <stdbool.h>
#include <math.h>

typedef struct ActiveWindowBounds {
	int x;
	int y;
	int width;
	int height;
	int screenX;
	int screenY;
	int screenWidth;
	int screenHeight;
	bool ok;
} ActiveWindowBounds;

typedef struct ScreenBounds {
	int x;
	int y;
	int width;
	int height;
	bool ok;
} ScreenBounds;

static bool isUsableWindow(CFDictionaryRef window, CGRect* bounds) {
	CFNumberRef layerRef = (CFNumberRef)CFDictionaryGetValue(window, kCGWindowLayer);
	int layer = 0;
	if (layerRef == NULL || !CFNumberGetValue(layerRef, kCFNumberIntType, &layer) || layer != 0) {
		return false;
	}

	CFStringRef owner = (CFStringRef)CFDictionaryGetValue(window, kCGWindowOwnerName);
	if (owner != NULL && CFStringCompare(owner, CFSTR("CyberNeko"), 0) == kCFCompareEqualTo) {
		return false;
	}

	CFDictionaryRef boundsRef = (CFDictionaryRef)CFDictionaryGetValue(window, kCGWindowBounds);
	if (boundsRef == NULL || !CGRectMakeWithDictionaryRepresentation(boundsRef, bounds)) {
		return false;
	}

	return bounds->size.width >= 160 && bounds->size.height >= 32;
}

static ScreenBounds screenContainingPoint(CGPoint point) {
	ScreenBounds result = {0, 0, 0, 0, false};
	NSArray<NSScreen *> *screens = [NSScreen screens];
	for (NSScreen *screen in screens) {
		NSRect frame = [screen frame];
		if (NSPointInRect(NSMakePoint(point.x, point.y), frame)) {
			result.x = (int)frame.origin.x;
			result.y = (int)frame.origin.y;
			result.width = (int)frame.size.width;
			result.height = (int)frame.size.height;
			result.ok = true;
			return result;
		}
	}

	NSScreen *screen = [NSScreen mainScreen];
	if (screen != NULL) {
		NSRect frame = [screen frame];
		result.x = (int)frame.origin.x;
		result.y = (int)frame.origin.y;
		result.width = (int)frame.size.width;
		result.height = (int)frame.size.height;
		result.ok = true;
	}
	return result;
}

static ActiveWindowBounds boundsToResult(CGRect bounds) {
	ActiveWindowBounds result = {0, 0, 0, 0, 0, 0, 0, 0, true};
	result.x = (int)bounds.origin.x;
	result.y = (int)bounds.origin.y;
	result.width = (int)bounds.size.width;
	result.height = (int)bounds.size.height;

	ScreenBounds screen = screenContainingPoint(bounds.origin);
	if (screen.ok) {
		result.screenX = screen.x;
		result.screenY = screen.y;
		result.screenWidth = screen.width;
		result.screenHeight = screen.height;
	}
	return result;
}

ActiveWindowBounds getActiveWindowBounds(void) {
	ActiveWindowBounds result = {0, 0, 0, 0, 0, 0, 0, 0, false};
	CFArrayRef windows = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
	if (windows == NULL) {
		return result;
	}

	CFIndex count = CFArrayGetCount(windows);
	for (CFIndex i = 0; i < count; i++) {
		CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windows, i);
		CGRect bounds;
		if (!isUsableWindow(window, &bounds)) {
			continue;
		}

		result = boundsToResult(bounds);
		break;
	}

	CFRelease(windows);
	return result;
}

ActiveWindowBounds getWindowEdgeTarget(double feetX, double feetY, double tolerance) {
	ActiveWindowBounds result = {0, 0, 0, 0, 0, 0, 0, 0, false};
	CFArrayRef windows = CGWindowListCopyWindowInfo(kCGWindowListOptionOnScreenOnly | kCGWindowListExcludeDesktopElements, kCGNullWindowID);
	if (windows == NULL) {
		return result;
	}

	double bestDistance = tolerance + 1.0;
	CFIndex count = CFArrayGetCount(windows);
	for (CFIndex i = 0; i < count; i++) {
		CFDictionaryRef window = (CFDictionaryRef)CFArrayGetValueAtIndex(windows, i);
		CGRect bounds;
		if (!isUsableWindow(window, &bounds)) {
			continue;
		}

		double left = bounds.origin.x;
		double right = bounds.origin.x + bounds.size.width;
		if (feetX < left - tolerance || feetX > right + tolerance) {
			continue;
		}

		double topDistance = fabs(feetY - bounds.origin.y);
		double bottomDistance = fabs(feetY - (bounds.origin.y + bounds.size.height));
		double distance = topDistance < bottomDistance ? topDistance : bottomDistance;
		if (distance <= tolerance && distance < bestDistance) {
			bestDistance = distance;
			result = boundsToResult(bounds);
			if (distance <= 3.0) {
				break;
			}
		}
	}

	CFRelease(windows);
	return result;
}
*/
import "C"

func activeWindowTarget() (targetRect, bool) {
	return cBoundsToTarget(C.getActiveWindowBounds())
}

func windowEdgeTargetAt(feetX int, feetY int) (targetRect, bool) {
	return cBoundsToTarget(C.getWindowEdgeTarget(C.double(feetX), C.double(feetY), C.double(edgeLockTolerance)))
}

func cBoundsToTarget(bounds C.ActiveWindowBounds) (targetRect, bool) {
	if !bool(bounds.ok) {
		return targetRect{}, false
	}

	return targetRect{
		X:      int(bounds.x),
		Y:      int(bounds.y),
		Width:  int(bounds.width),
		Height: int(bounds.height),

		ScreenX:      int(bounds.screenX),
		ScreenY:      int(bounds.screenY),
		ScreenWidth:  int(bounds.screenWidth),
		ScreenHeight: int(bounds.screenHeight),
	}, true
}
