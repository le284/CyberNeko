package main

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

const (
	petWindowWidth  = 220
	petWindowHeight = 240

	// petGroundOffset 是从透明窗口左上角到“脚底/阴影底部”的垂直距离。
	// 这个值必须和 frontend/public/style.css 中的角色布局保持一致：窗口高 240，主体视觉底部约在 224。
	// 用 target.Y - petGroundOffset 可以让猫咪视觉脚底踩在目标窗口的上边缘。
	petGroundOffset = 224

	walkStepPixels      = 5
	followMinStepPixels = 22
	followMaxStepPixels = 96
	followStopDistance  = 44
	followAcceleration  = 0.3
	walkTick            = 45 * time.Millisecond
	targetRefresh       = 500 * time.Millisecond
	edgeLockTolerance   = 56

	minWaypointDistance = 72
	pauseMin            = 350 * time.Millisecond
	pauseJitter         = 950 * time.Millisecond
)

type screenPoint struct {
	X int
	Y int
}

type targetRect struct {
	X            int
	Y            int
	Width        int
	Height       int
	ScreenX      int
	ScreenY      int
	ScreenWidth  int
	ScreenHeight int
}

type walkLine struct {
	Left  int
	Right int
	Y     int
	Edge  string
}

type petVisual struct {
	window    *application.WebviewWindow
	state     string
	direction string
}

func startTopEdgeWalker(app *application.App, window *application.WebviewWindow, windowName string, startRatio float64) {
	startWindowEdgeWanderer(app, window, windowName, startRatio)
}

func startWindowEdgeWanderer(app *application.App, window *application.WebviewWindow, windowName string, startRatio float64) {
	go func() {
		rng := rand.New(rand.NewSource(time.Now().UnixNano() + int64(startRatio*1000)))
		ticker := time.NewTicker(walkTick)
		defer ticker.Stop()

		target := screenTarget(app)
		if activeTarget, ok := activeWindowTarget(); ok {
			target = activeTarget
		}

		line := lineForTarget(target)
		x := initialXForLine(line, startRatio)
		destination := chooseNextWaypoint(rng, line, x)
		lastTargetRefresh := time.Now()
		waitUntil := time.Now().Add(randomPause(rng))
		visual := petVisual{window: window}

		visual.emitState("idle")
		visual.emitDirection("right")
		movePetWindow(window, x, line.Y)

		for now := range ticker.C {
			switch currentMovementMode(windowName) {
			case movementModeIdle:
				visual.emitState("idle")
				continue
			case movementModeFollowMouse:
				followMouse(app, window, &visual)
				continue
			}

			if now.Sub(lastTargetRefresh) >= targetRefresh {
				actualX, actualY := window.Position()
				x = actualX

				feetX := actualX + petWindowWidth/2
				feetY := actualY + petGroundOffset
				if edgeTarget, ok := windowEdgeTargetAt(feetX, feetY); ok {
					target = edgeTarget
				} else if nextTarget, ok := activeWindowTarget(); ok {
					target = nextTarget
				}

				nextLine := lineForTarget(target)
				if nextLine != line {
					line = nextLine
					x = clamp(x, line.Left, line.Right)
					destination = chooseNextWaypoint(rng, line, x)
					waitUntil = now.Add(randomPause(rng))
				}
				lastTargetRefresh = now
			}

			if line.Right <= line.Left {
				x = line.Left
				visual.emitState("idle")
				movePetWindow(window, x, line.Y)
				continue
			}

			if now.Before(waitUntil) {
				visual.emitState("idle")
				movePetWindow(window, x, line.Y)
				continue
			}

			if abs(destination-x) <= walkStepPixels {
				x = destination
				visual.emitState("idle")
				movePetWindow(window, x, line.Y)
				destination = chooseNextWaypoint(rng, line, x)
				waitUntil = now.Add(randomPause(rng))
				continue
			}

			direction := 1
			if destination < x {
				direction = -1
				visual.emitDirection("left")
			} else {
				visual.emitDirection("right")
			}

			visual.emitState("walking")
			x += direction * walkStepPixels
			if direction > 0 && x > destination {
				x = destination
			}
			if direction < 0 && x < destination {
				x = destination
			}
			x = clamp(x, line.Left, line.Right)
			movePetWindow(window, x, line.Y)
		}
	}()
}

func followMouse(app *application.App, window *application.WebviewWindow, visual *petVisual) {
	cursor, ok := cursorPosition()
	if !ok {
		visual.emitState("idle")
		return
	}

	x, y := window.Position()
	bounds := screenTargetForPoint(app, cursor.X, cursor.Y)
	nextX, nextY, moving := nextFollowPosition(x, y, cursor, bounds)

	centerX := x + petWindowWidth/2
	if cursor.X < centerX {
		visual.emitDirection("left")
	} else {
		visual.emitDirection("right")
	}

	if moving {
		visual.emitState("walking")
	} else {
		visual.emitState("idle")
	}
	movePetWindow(window, nextX, nextY)
}

func nextFollowPosition(currentX int, currentY int, cursor screenPoint, bounds targetRect) (int, int, bool) {
	centerX := currentX + petWindowWidth/2
	centerY := currentY + petWindowHeight/2
	dx := cursor.X - centerX
	dy := cursor.Y - centerY
	distance := math.Hypot(float64(dx), float64(dy))

	if distance <= followStopDistance {
		return currentX, currentY, false
	}

	step := followStep(distance)
	nextX := currentX + int(math.Round(float64(dx)/distance*step))
	nextY := currentY + int(math.Round(float64(dy)/distance*step))

	maxX := bounds.ScreenX + bounds.ScreenWidth - petWindowWidth
	maxY := bounds.ScreenY + bounds.ScreenHeight - petWindowHeight
	nextX = clamp(nextX, bounds.ScreenX, max(bounds.ScreenX, maxX))
	nextY = clamp(nextY, bounds.ScreenY, max(bounds.ScreenY, maxY))

	return nextX, nextY, nextX != currentX || nextY != currentY
}

func followStep(distance float64) float64 {
	remaining := distance - float64(followStopDistance)
	if remaining <= 0 {
		return 0
	}

	step := math.Max(float64(followMinStepPixels), distance*followAcceleration)
	step = math.Min(step, float64(followMaxStepPixels))
	return math.Min(step, remaining)
}

func (visual *petVisual) emitState(state string) {
	if visual.state == state {
		return
	}
	visual.state = state
	visual.window.ExecJS(fmt.Sprintf("window.__cyberNekoSetState?.(%s)", strconv.Quote(state)))
}

func (visual *petVisual) emitDirection(direction string) {
	if visual.direction == direction {
		return
	}
	visual.direction = direction
	visual.window.ExecJS(fmt.Sprintf("window.__cyberNekoSetDirection?.(%s)", strconv.Quote(direction)))
}

func initialXForLine(line walkLine, ratio float64) int {
	if line.Right <= line.Left {
		return line.Left
	}

	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	span := line.Right - line.Left
	return line.Left + int(math.Round(float64(span)*ratio))
}

func chooseNextWaypoint(rng *rand.Rand, line walkLine, currentX int) int {
	if line.Right <= line.Left {
		return line.Left
	}

	span := line.Right - line.Left
	if span <= minWaypointDistance {
		return line.Left + rng.Intn(span+1)
	}

	requiredDistance := min(minWaypointDistance, span/2)
	for range 8 {
		candidate := line.Left + rng.Intn(span+1)
		if abs(candidate-currentX) >= requiredDistance {
			return candidate
		}
	}

	if currentX < line.Left+span/2 {
		return line.Right
	}
	return line.Left
}

func randomPause(rng *rand.Rand) time.Duration {
	return pauseMin + time.Duration(rng.Int63n(int64(pauseJitter)))
}

func setPetAnimationState(window *application.WebviewWindow, state string) {
	window.ExecJS(fmt.Sprintf("window.__cyberNekoSetState?.(%s)", strconv.Quote(state)))
}

func movePetWindow(window *application.WebviewWindow, x int, y int) {
	window.ExecJS(fmt.Sprintf("window.__cyberNekoSetPosition?.(%d,%d)", x, y))
}

func screenTarget(app *application.App) targetRect {
	screen := app.Screen.GetPrimary()
	if screen == nil {
		return targetRect{X: 0, Y: 0, Width: 1280, Height: 720, ScreenX: 0, ScreenY: 0, ScreenWidth: 1280, ScreenHeight: 720}
	}

	return targetRect{
		X:            screen.WorkArea.X,
		Y:            screen.WorkArea.Y,
		Width:        screen.WorkArea.Width,
		Height:       screen.WorkArea.Height,
		ScreenX:      screen.WorkArea.X,
		ScreenY:      screen.WorkArea.Y,
		ScreenWidth:  screen.WorkArea.Width,
		ScreenHeight: screen.WorkArea.Height,
	}
}

func screenTargetForPoint(app *application.App, x int, y int) targetRect {
	for _, screen := range app.Screen.GetAll() {
		if x >= screen.WorkArea.X && x <= screen.WorkArea.X+screen.WorkArea.Width &&
			y >= screen.WorkArea.Y && y <= screen.WorkArea.Y+screen.WorkArea.Height {
			return targetRect{
				X:            screen.WorkArea.X,
				Y:            screen.WorkArea.Y,
				Width:        screen.WorkArea.Width,
				Height:       screen.WorkArea.Height,
				ScreenX:      screen.WorkArea.X,
				ScreenY:      screen.WorkArea.Y,
				ScreenWidth:  screen.WorkArea.Width,
				ScreenHeight: screen.WorkArea.Height,
			}
		}
	}
	return screenTarget(app)
}

func lineForTarget(target targetRect) walkLine {
	left := target.X
	right := target.X + max(0, target.Width-petWindowWidth)

	// 上边缘空间足够时，让猫咪脚底踩在当前窗口上边缘；否则改为踩在当前窗口下边缘。
	y := target.Y - petGroundOffset
	edge := "top"
	if y < target.ScreenY {
		y = target.Y + target.Height - petGroundOffset
		edge = "bottom"
	}

	bottomLimit := target.ScreenY + target.ScreenHeight - petWindowHeight
	y = clamp(y, target.ScreenY, max(target.ScreenY, bottomLimit))

	return walkLine{Left: left, Right: right, Y: y, Edge: edge}
}

func clamp(value int, low int, high int) int {
	if value < low {
		return low
	}
	if value > high {
		return high
	}
	return value
}

func abs(value int) int {
	if value < 0 {
		return -value
	}
	return value
}
