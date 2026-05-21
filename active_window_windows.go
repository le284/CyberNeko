//go:build windows

package main

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/w32"
)

func activeWindowTarget() (targetRect, bool) {
	hwnd := w32.GetForegroundWindow()
	if hwnd == 0 || !w32.IsWindowVisible(hwnd) {
		return targetRect{}, false
	}

	_, processID := w32.GetWindowThreadProcessId(hwnd)
	if processID == os.Getpid() {
		// 桌宠自己被点击或拖拽时也会短暂成为前台窗口。
		// 这里返回 false，让行走器保持上一次外部目标，避免猫咪跳到自己的窗口边缘。
		return targetRect{}, false
	}

	return targetFromPhysicalWindow(hwnd, extendedFrameBounds(hwnd))
}

func windowEdgeTargetAt(feetX int, feetY int) (targetRect, bool) {
	physicalFeet := application.DipToPhysicalPoint(application.Point{X: feetX, Y: feetY})
	tolerance := edgeLockTolerance * 2
	type candidate struct {
		hwnd     w32.HWND
		bounds   w32.RECT
		distance int
	}

	best := candidate{distance: tolerance + 1}
	callback := syscall.NewCallback(func(hwnd uintptr, lparam uintptr) uintptr {
		window := w32.HWND(hwnd)
		if !w32.IsWindowVisible(window) {
			return 1
		}

		_, processID := w32.GetWindowThreadProcessId(window)
		if processID == os.Getpid() {
			return 1
		}

		bounds := extendedFrameBounds(window)
		width := int(bounds.Right - bounds.Left)
		height := int(bounds.Bottom - bounds.Top)
		if width < 160 || height < 32 {
			return 1
		}

		if physicalFeet.X < int(bounds.Left)-tolerance || physicalFeet.X > int(bounds.Right)+tolerance {
			return 1
		}

		topDistance := abs(physicalFeet.Y - int(bounds.Top))
		bottomDistance := abs(physicalFeet.Y - int(bounds.Bottom))
		distance := min(topDistance, bottomDistance)
		if distance <= tolerance && distance < best.distance {
			best = candidate{hwnd: window, bounds: bounds, distance: distance}
		}

		return 1
	})

	w32.EnumWindows(callback, 0)
	if best.hwnd == 0 {
		return targetRect{}, false
	}

	return targetFromPhysicalWindow(best.hwnd, best.bounds)
}

func targetFromPhysicalWindow(hwnd w32.HWND, physicalWindow w32.RECT) (targetRect, bool) {
	if physicalWindow.Right <= physicalWindow.Left || physicalWindow.Bottom <= physicalWindow.Top {
		return targetRect{}, false
	}

	windowRect := rectToAppRect(physicalWindow)
	if windowRect.Width < 160 || windowRect.Height < 32 {
		return targetRect{}, false
	}

	monitor := w32.MonitorFromWindow(hwnd, w32.MONITOR_DEFAULTTONEAREST)
	if monitor == 0 {
		return targetRect{}, false
	}

	monitorInfo := w32.MONITORINFO{CbSize: uint32(unsafe.Sizeof(w32.MONITORINFO{}))}
	if !w32.GetMonitorInfo(monitor, &monitorInfo) {
		return targetRect{}, false
	}

	// Wails 的 Window.SetPosition 使用 DIP 坐标，不是 Win32 的物理像素。
	// GetWindowRect / DwmGetWindowAttribute / GetMonitorInfo 拿到的是物理像素，
	// 所以必须通过 Wails ScreenManager 转换，否则在 125%/150% 缩放的 Windows 上会明显偏位。
	dipWindow := application.PhysicalToDipRect(windowRect)
	dipWorkArea := application.PhysicalToDipRect(rectToAppRect(monitorInfo.RcWork))

	return targetRect{
		X:            dipWindow.X,
		Y:            dipWindow.Y,
		Width:        dipWindow.Width,
		Height:       dipWindow.Height,
		ScreenX:      dipWorkArea.X,
		ScreenY:      dipWorkArea.Y,
		ScreenWidth:  dipWorkArea.Width,
		ScreenHeight: dipWorkArea.Height,
	}, true
}

func extendedFrameBounds(hwnd w32.HWND) w32.RECT {
	var rect w32.RECT
	hr := w32.DwmGetWindowAttribute(
		hwnd,
		w32.DWMWA_EXTENDED_FRAME_BOUNDS,
		unsafe.Pointer(&rect),
		unsafe.Sizeof(rect),
	)
	if w32.SUCCEEDED(hr) && rect.Right > rect.Left && rect.Bottom > rect.Top {
		return rect
	}

	fallback := w32.GetWindowRect(hwnd)
	if fallback == nil {
		return w32.RECT{}
	}
	return *fallback
}

func rectToAppRect(rect w32.RECT) application.Rect {
	return application.Rect{
		X:      int(rect.Left),
		Y:      int(rect.Top),
		Width:  int(rect.Right - rect.Left),
		Height: int(rect.Bottom - rect.Top),
	}
}
