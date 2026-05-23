package main

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

type petProfile struct {
	ID         string
	Slot       int
	WindowName string
	Title      string
	MenuName   string
	StartRatio float64
}

type basePetProfile struct {
	ID         string
	Title      string
	WindowName string
	MenuName   string
	StartRatio float64
}

type petRuntime struct {
	profile petProfile
	window  *application.WebviewWindow
	visible bool
}

type PetManager struct {
	app          *application.App
	settingsPath string

	mu             sync.Mutex
	settings       AppSettings
	pets           map[int]*petRuntime
	settingsWindow *application.WebviewWindow
	shortcutKeys   map[string]string
}

var basePetProfiles = []basePetProfile{
	{ID: "neko", Title: "CyberNeko", WindowName: "pet-neko", MenuName: "neko-menu", StartRatio: 0.32},
}

func newPetManager(app *application.App, settings AppSettings, settingsPath string) *PetManager {
	return &PetManager{
		app:          app,
		settingsPath: settingsPath,
		settings:     normalizeAppSettings(settings),
		pets:         make(map[int]*petRuntime),
	}
}

func (m *PetManager) Settings() AppSettings {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.settings
}

func (m *PetManager) SetPetCount(count int) (AppSettings, error) {
	m.mu.Lock()
	m.settings.PetCount = clamp(count, 1, maxPetCount)
	settings := m.settings
	settingsPath := m.settingsPath
	m.mu.Unlock()

	if err := saveAppSettings(settingsPath, settings); err != nil {
		return settings, err
	}

	m.applyPetCount(settings.PetCount)
	m.broadcastSettings(settings)
	return settings, nil
}

func (m *PetManager) SetShortcuts(shortcuts ShortcutSettings) (AppSettings, error) {
	shortcuts = normalizeShortcutSettings(shortcuts)

	m.mu.Lock()
	m.settings.Shortcuts = shortcuts
	settings := m.settings
	settingsPath := m.settingsPath
	m.mu.Unlock()

	if err := saveAppSettings(settingsPath, settings); err != nil {
		return settings, err
	}

	m.applyShortcuts(settings.Shortcuts)
	m.broadcastSettings(settings)
	return settings, nil
}

func (m *PetManager) BroadcastVisualsChanged(revision int64) {
	m.app.Event.Emit("pet:visuals", PetVisualsChangedEvent{Revision: revision})
}

func (m *PetManager) applyInitialPetCount() {
	m.mu.Lock()
	count := m.settings.PetCount
	m.mu.Unlock()
	m.applyPetCount(count)
}

func (m *PetManager) applyPetCount(count int) {
	count = clamp(count, 1, maxPetCount)

	m.mu.Lock()
	defer m.mu.Unlock()

	for slot := 0; slot < maxPetCount; slot++ {
		runtime := m.pets[slot]
		shouldShow := slot < count

		if shouldShow {
			if runtime == nil {
				runtime = m.createPetRuntime(slot, true)
				m.pets[slot] = runtime
				continue
			}

			if !runtime.visible {
				runtime.window.Show()
				runtime.visible = true
				setMovementMode(runtime.profile.WindowName, movementModeEdgeWander)
			}
			continue
		}

		if runtime != nil && runtime.visible {
			setMovementMode(runtime.profile.WindowName, movementModeIdle)
			runtime.window.Hide()
			runtime.visible = false
		}
	}
}

func (m *PetManager) createPetRuntime(slot int, visible bool) *petRuntime {
	profile := petProfileForSlot(slot)
	petWindow := newPetWindow(m.app, profile, !visible)
	setMovementMode(profile.WindowName, movementModeEdgeWander)
	registerPetContextMenu(m.app, m, profile, petWindow)
	startTopEdgeWalker(m.app, petWindow, profile.WindowName, profile.StartRatio)
	return &petRuntime{profile: profile, window: petWindow, visible: visible}
}

func (m *PetManager) applyShortcuts(shortcuts ShortcutSettings) {
	shortcuts = normalizeShortcutSettings(shortcuts)
	nextKeys := make(map[string]string)
	for action, shortcut := range shortcutActionMap(shortcuts) {
		binding, err := shortcutBindingKey(shortcut)
		if err == nil && binding != "" {
			nextKeys[action] = binding
		}
	}

	m.mu.Lock()
	previousKeys := m.shortcutKeys
	m.shortcutKeys = nextKeys
	m.mu.Unlock()

	for _, binding := range previousKeys {
		m.app.KeyBinding.Remove(binding)
	}

	for action, binding := range nextKeys {
		action := action
		m.app.KeyBinding.Add(binding, func(window application.Window) {
			m.applyShortcutAction(action, window.Name())
		})
	}
}

func (m *PetManager) applyShortcutAction(action string, sourceWindowName string) {
	targets := m.shortcutTargets(sourceWindowName)
	for _, target := range targets {
		mode := movementModeForShortcutAction(action, currentMovementMode(target.profile.WindowName))
		m.setPetRuntimeMode(target, mode)
	}
}

func (m *PetManager) shortcutTargets(sourceWindowName string) []*petRuntime {
	m.mu.Lock()
	defer m.mu.Unlock()

	if runtime := m.petRuntimeByWindowName(sourceWindowName); runtime != nil && runtime.visible {
		return []*petRuntime{runtime}
	}

	targets := make([]*petRuntime, 0, len(m.pets))
	for slot := 0; slot < maxPetCount; slot++ {
		if runtime := m.pets[slot]; runtime != nil && runtime.visible {
			targets = append(targets, runtime)
		}
	}
	return targets
}

func (m *PetManager) petRuntimeByWindowName(windowName string) *petRuntime {
	for _, runtime := range m.pets {
		if runtime != nil && runtime.profile.WindowName == windowName {
			return runtime
		}
	}
	return nil
}

func (m *PetManager) setPetRuntimeMode(runtime *petRuntime, mode movementMode) {
	setMovementMode(runtime.profile.WindowName, mode)
	if mode == movementModeIdle {
		setPetAnimationState(runtime.window, "idle")
	}
}

func movementModeForShortcutAction(action string, currentMode movementMode) movementMode {
	switch action {
	case shortcutActionIdle:
		return movementModeIdle
	case shortcutActionEdgeWander:
		return movementModeEdgeWander
	case shortcutActionFollowMouse:
		return movementModeFollowMouse
	case shortcutActionCycle:
		return nextMovementMode(currentMode)
	default:
		return currentMode
	}
}

func nextMovementMode(currentMode movementMode) movementMode {
	switch currentMode {
	case movementModeIdle:
		return movementModeEdgeWander
	case movementModeEdgeWander:
		return movementModeFollowMouse
	default:
		return movementModeIdle
	}
}

func (m *PetManager) OpenSettingsWindow() {
	m.mu.Lock()
	settingsWindow := m.settingsWindow
	m.mu.Unlock()

	if settingsWindow != nil {
		settingsWindow.Show()
		settingsWindow.Focus()
		return
	}

	settingsWindow = m.app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:             "settings",
		Title:            "CyberNeko 设置",
		Width:            460,
		Height:           700,
		MinWidth:         420,
		MinHeight:        620,
		AlwaysOnTop:      true,
		BackgroundColour: application.NewRGB(248, 250, 252),
		URL:              "/?view=settings",
		Mac: application.MacWindow{
			TitleBar: application.MacTitleBarDefault,
		},
	})
	settingsWindow.RegisterHook(events.Common.WindowClosing, func(event *application.WindowEvent) {
		settingsWindow.Hide()
		event.Cancel()
	})
	settingsWindow.Center()
	settingsWindow.Show()
	settingsWindow.Focus()

	m.mu.Lock()
	m.settingsWindow = settingsWindow
	m.mu.Unlock()
}

func (m *PetManager) broadcastSettings(settings AppSettings) {
	m.app.Event.Emit("pet:settings", AppSettingsChangedEvent{
		PetCount:  settings.PetCount,
		MaxPets:   settings.MaxPets,
		Shortcuts: settings.Shortcuts,
	})
}

func petProfileForSlot(slot int) petProfile {
	base := basePetProfiles[slot%len(basePetProfiles)]
	if slot < len(basePetProfiles) {
		return petProfile{
			ID:         base.ID,
			Slot:       slot,
			WindowName: base.WindowName,
			Title:      base.Title,
			MenuName:   base.MenuName,
			StartRatio: base.StartRatio,
		}
	}

	index := slot + 1
	return petProfile{
		ID:         base.ID,
		Slot:       slot,
		WindowName: fmt.Sprintf("pet-%d", index),
		Title:      fmt.Sprintf("CyberNeko %d", index),
		MenuName:   fmt.Sprintf("pet-%d-menu", index),
		StartRatio: float64(index) / float64(maxPetCount+1),
	}
}

func newPetWindow(app *application.App, profile petProfile, hidden bool) *application.WebviewWindow {
	query := url.Values{}
	query.Set("pet", profile.ID)
	query.Set("slot", fmt.Sprintf("%d", profile.Slot))

	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:  profile.WindowName,
		Title: profile.Title,

		Width:  petWindowWidth,
		Height: petWindowHeight,
		Hidden: hidden,

		// Frameless 去掉系统标题栏和边框，让桌宠只剩前端绘制的角色。
		Frameless: true,

		// AlwaysOnTop 让窗口进入更高 z-order，浮在浏览器、IDE 等普通窗口之上。
		AlwaysOnTop: true,

		DisableResize: true,

		// 透明背景是桌宠窗口的核心：OS 不绘制底色，WebView 只露出猫咪 DOM。
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),

		DefaultContextMenuDisabled: true,
		IgnoreMouseEvents:          false,
		URL:                        "/?" + query.Encode(),

		Mac: application.MacWindow{
			Backdrop: application.MacBackdropTransparent,
			TitleBar: application.MacTitleBar{
				Hide:            true,
				HideTitle:       true,
				FullSizeContent: true,
			},
			DisableShadow: true,
			WindowLevel:   application.MacWindowLevelFloating,
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
				application.MacWindowCollectionBehaviorFullScreenAuxiliary |
				application.MacWindowCollectionBehaviorIgnoresCycle,
		},

		Windows: application.WindowsWindow{
			HiddenOnTaskbar:                   true,
			DisableFramelessWindowDecorations: true,
		},
	})
}

func registerPetContextMenu(app *application.App, manager *PetManager, profile petProfile, petWindow *application.WebviewWindow) {
	menu := app.ContextMenu.New()

	menu.Add(profile.Title + "：原地待机").OnClick(func(_ *application.Context) {
		manager.setPetRuntimeMode(&petRuntime{profile: profile, window: petWindow, visible: true}, movementModeIdle)
	})

	menu.Add(profile.Title + "：沿窗口边缘巡游").OnClick(func(_ *application.Context) {
		manager.setPetRuntimeMode(&petRuntime{profile: profile, window: petWindow, visible: true}, movementModeEdgeWander)
	})

	menu.Add(profile.Title + "：跟随鼠标").OnClick(func(_ *application.Context) {
		manager.setPetRuntimeMode(&petRuntime{profile: profile, window: petWindow, visible: true}, movementModeFollowMouse)
	})

	menu.AddSeparator()

	menu.Add("设置...").OnClick(func(_ *application.Context) {
		manager.OpenSettingsWindow()
	})

	menu.Add("Exit").OnClick(func(_ *application.Context) {
		app.Quit()
	})

	// 前端通过 CSS 自定义属性 --custom-contextmenu 绑定到对应原生菜单。
	app.ContextMenu.Add(profile.MenuName, menu)
}
