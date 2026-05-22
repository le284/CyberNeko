//go:build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AppKit

#import <AppKit/AppKit.h>
#include <stdbool.h>

typedef struct DockScreenInfo {
	int x;
	int y;
	int width;
	int height;
	int workX;
	int workY;
	int workWidth;
	int workHeight;
	double scale;
	bool ok;
} DockScreenInfo;

static DockScreenInfo getDockScreenInfo(void) {
	DockScreenInfo result = {0, 0, 0, 0, 0, 0, 0, 0, 1.0, false};
	NSArray<NSScreen *> *screens = [NSScreen screens];
	NSScreen *screen = screens.count > 0 ? [screens objectAtIndex:0] : [NSScreen mainScreen];
	if (screen == NULL) {
		return result;
	}

	NSRect frame = [screen frame];
	NSRect visible = [screen visibleFrame];
	result.x = (int)frame.origin.x;
	result.y = (int)frame.origin.y;
	result.width = (int)frame.size.width;
	result.height = (int)frame.size.height;
	result.workX = (int)visible.origin.x;
	result.workY = (int)visible.origin.y;
	result.workWidth = (int)visible.size.width;
	result.workHeight = (int)visible.size.height;
	result.scale = (double)screen.backingScaleFactor;
	result.ok = true;
	return result;
}
*/
import "C"

import (
	"math"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func dockEdgeTarget(app *application.App) (targetRect, bool) {
	info := C.getDockScreenInfo()
	if !bool(info.ok) {
		return targetRect{}, false
	}

	return dockEdgeTargetFromScreen(
		application.Rect{X: int(info.x), Y: int(info.y), Width: int(info.width), Height: int(info.height)},
		application.Rect{X: int(info.workX), Y: int(info.workY), Width: int(info.workWidth), Height: int(info.workHeight)},
		float64(info.scale),
	)
}

func dockEdgeTargetFromScreen(bounds application.Rect, workArea application.Rect, scale float64) (targetRect, bool) {
	if bounds.Width <= 0 || bounds.Height <= 0 {
		return targetRect{}, false
	}
	if scale <= 0 {
		scale = 1
	}

	bottomInset := workArea.Y - bounds.Y
	if bottomInset < 0 {
		bottomInset = 0
	}

	leftInset := workArea.X - bounds.X
	rightInset := bounds.X + bounds.Width - (workArea.X + workArea.Width)
	if bottomInset == 0 && (leftInset > 12 || rightInset > 12) {
		// 当前动画只支持水平行走。Dock 在左/右侧时没有可趴的水平 Dock 上沿，
		// 因此回退到普通窗口边缘逻辑，避免把猫错误地放到屏幕底部。
		return targetRect{}, false
	}

	dockLineY := bounds.Height
	dockHeight := 1
	if bottomInset > 0 {
		dockLineY = bounds.Height - bottomInset
		dockHeight = bottomInset
	}

	return targetRect{
		X:                   scaleCoordinate(bounds.X, scale),
		Y:                   scaleCoordinate(dockLineY, scale),
		Width:               scaleCoordinate(bounds.Width, scale),
		Height:              max(1, scaleCoordinate(dockHeight, scale)),
		ScreenX:             scaleCoordinate(bounds.X, scale),
		ScreenY:             0,
		ScreenWidth:         scaleCoordinate(bounds.Width, scale),
		ScreenHeight:        scaleCoordinate(bounds.Height, scale),
		Scale:               scale,
		VisualEdge:          "dock",
		AllowBottomOverflow: true,
	}, true
}

func scaleCoordinate(value int, scale float64) int {
	return int(math.Round(float64(value) * scale))
}
