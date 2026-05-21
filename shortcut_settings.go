package main

import (
	"fmt"
	"runtime"
	"sort"
	"strings"
)

const (
	shortcutActionIdle        = "idle"
	shortcutActionEdgeWander  = "edgeWander"
	shortcutActionFollowMouse = "followMouse"
	shortcutActionCycle       = "cycle"
)

// ShortcutSettings 保存用户可配置的行为快捷键。
// 支持 CmdOrCtrl 作为跨平台修饰键；默认快捷键使用 Ctrl+Alt，避免和常见 Cmd 快捷键冲突。
type ShortcutSettings struct {
	Idle        string `json:"idle"`
	EdgeWander  string `json:"edgeWander"`
	FollowMouse string `json:"followMouse"`
	Cycle       string `json:"cycle"`
}

type parsedShortcut struct {
	Modifiers []string
	Key       string
}

func defaultShortcutSettings() ShortcutSettings {
	return ShortcutSettings{
		Idle:        "Ctrl+Alt+1",
		EdgeWander:  "Ctrl+Alt+2",
		FollowMouse: "Ctrl+Alt+3",
		Cycle:       "Ctrl+Alt+Space",
	}
}

func shortcutActionMap(shortcuts ShortcutSettings) map[string]string {
	return map[string]string{
		shortcutActionIdle:        shortcuts.Idle,
		shortcutActionEdgeWander:  shortcuts.EdgeWander,
		shortcutActionFollowMouse: shortcuts.FollowMouse,
		shortcutActionCycle:       shortcuts.Cycle,
	}
}

func normalizeShortcutSettings(shortcuts ShortcutSettings) ShortcutSettings {
	defaults := defaultShortcutSettings()
	items := []struct {
		value        *string
		defaultValue string
	}{
		{&shortcuts.Idle, defaults.Idle},
		{&shortcuts.EdgeWander, defaults.EdgeWander},
		{&shortcuts.FollowMouse, defaults.FollowMouse},
		{&shortcuts.Cycle, defaults.Cycle},
	}

	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		value := item.defaultValue
		if item.value != nil && strings.TrimSpace(*item.value) != "" {
			value = *item.value
		}
		normalized, binding, err := normalizeShortcut(value)
		if err != nil || binding == "" {
			normalized, binding = firstAvailableDefaultShortcut(item.defaultValue, defaults, seen)
		}
		if _, exists := seen[binding]; exists {
			normalized, binding = firstAvailableDefaultShortcut(item.defaultValue, defaults, seen)
		}
		*item.value = normalized
		seen[binding] = struct{}{}
	}
	return shortcuts
}

func firstAvailableDefaultShortcut(preferred string, defaults ShortcutSettings, seen map[string]struct{}) (string, string) {
	candidates := []string{preferred, defaults.Idle, defaults.EdgeWander, defaults.FollowMouse, defaults.Cycle}
	for _, candidate := range candidates {
		normalized, binding, err := normalizeShortcut(candidate)
		if err != nil {
			continue
		}
		if _, exists := seen[binding]; !exists {
			return normalized, binding
		}
	}
	return "", ""
}

func normalizeShortcut(shortcut string) (string, string, error) {
	parsed, err := parseShortcut(shortcut)
	if err != nil {
		return "", "", err
	}
	stored := shortcutToStoredString(parsed)
	binding, err := shortcutToBindingString(parsed)
	if err != nil {
		return "", "", err
	}
	return stored, binding, nil
}

func shortcutBindingKey(shortcut string) (string, error) {
	_, binding, err := normalizeShortcut(shortcut)
	return binding, err
}

func parseShortcut(shortcut string) (parsedShortcut, error) {
	parts := strings.Split(strings.TrimSpace(shortcut), "+")
	if len(parts) == 0 {
		return parsedShortcut{}, fmt.Errorf("shortcut is empty")
	}

	modifiers := make(map[string]struct{})
	for index, rawPart := range parts {
		part := strings.TrimSpace(rawPart)
		if part == "" {
			return parsedShortcut{}, fmt.Errorf("shortcut contains an empty component")
		}

		if index == len(parts)-1 {
			key, ok := normalizeShortcutKey(part)
			if !ok {
				return parsedShortcut{}, fmt.Errorf("%q is not a supported key", part)
			}
			if len(modifiers) == 0 && !isFunctionShortcutKey(key) {
				return parsedShortcut{}, fmt.Errorf("shortcut must include a modifier unless it uses a function key")
			}
			return parsedShortcut{Modifiers: orderedShortcutModifiers(modifiers), Key: key}, nil
		}

		modifier, ok := normalizeShortcutModifier(part)
		if !ok {
			return parsedShortcut{}, fmt.Errorf("%q is not a supported modifier", part)
		}
		modifiers[modifier] = struct{}{}
	}

	return parsedShortcut{}, fmt.Errorf("shortcut is empty")
}

func normalizeShortcutModifier(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "cmdorctrl", "cmd", "command":
		return "CmdOrCtrl", true
	case "ctrl", "control":
		return "Ctrl", true
	case "alt", "option", "optionoralt":
		return "Alt", true
	case "shift":
		return "Shift", true
	case "super", "win", "windows":
		return "Super", true
	default:
		return "", false
	}
}

func orderedShortcutModifiers(modifiers map[string]struct{}) []string {
	order := []string{"CmdOrCtrl", "Ctrl", "Alt", "Shift", "Super"}
	result := make([]string, 0, len(modifiers))
	for _, modifier := range order {
		if _, ok := modifiers[modifier]; ok {
			result = append(result, modifier)
		}
	}
	return result
}

func isFunctionShortcutKey(key string) bool {
	if !strings.HasPrefix(strings.ToUpper(key), "F") {
		return false
	}
	number := strings.TrimPrefix(strings.ToUpper(key), "F")
	if number == "" {
		return false
	}
	value := 0
	for _, char := range number {
		if char < '0' || char > '9' {
			return false
		}
		value = value*10 + int(char-'0')
	}
	return value >= 1 && value <= 35
}

func normalizeShortcutKey(value string) (string, bool) {
	key := strings.TrimSpace(value)
	lower := strings.ToLower(key)
	namedKeys := map[string]string{
		"backspace":  "Backspace",
		"tab":        "Tab",
		"return":     "Return",
		"enter":      "Enter",
		"escape":     "Escape",
		"esc":        "Escape",
		"left":       "Left",
		"arrowleft":  "Left",
		"right":      "Right",
		"arrowright": "Right",
		"up":         "Up",
		"arrowup":    "Up",
		"down":       "Down",
		"arrowdown":  "Down",
		"space":      "Space",
		" ":          "Space",
		"delete":     "Delete",
		"home":       "Home",
		"end":        "End",
		"page up":    "Page Up",
		"pageup":     "Page Up",
		"page down":  "Page Down",
		"pagedown":   "Page Down",
		"numlock":    "Numlock",
	}
	if normalized, ok := namedKeys[lower]; ok {
		return normalized, true
	}
	if strings.HasPrefix(lower, "f") {
		number := strings.TrimPrefix(lower, "f")
		if number != "" {
			value := 0
			for _, char := range number {
				if char < '0' || char > '9' {
					return "", false
				}
				value = value*10 + int(char-'0')
			}
			if value >= 1 && value <= 35 {
				return strings.ToUpper(lower), true
			}
		}
	}
	if len(key) == 1 {
		return strings.ToUpper(key), true
	}
	return "", false
}

func shortcutToStoredString(shortcut parsedShortcut) string {
	parts := append([]string{}, shortcut.Modifiers...)
	parts = append(parts, shortcut.Key)
	return strings.Join(parts, "+")
}

func shortcutToBindingString(shortcut parsedShortcut) (string, error) {
	seenModifiers := make(map[string]struct{}, len(shortcut.Modifiers))
	modifiers := make([]string, 0, len(shortcut.Modifiers))
	for _, modifier := range shortcut.Modifiers {
		bindingName := shortcutModifierBindingName(modifier)
		if _, exists := seenModifiers[bindingName]; exists {
			return "", fmt.Errorf("shortcut contains duplicate physical modifier %q on %s", bindingName, runtime.GOOS)
		}
		seenModifiers[bindingName] = struct{}{}
		modifiers = append(modifiers, bindingName)
	}
	sort.Strings(modifiers)
	modifiers = append(modifiers, strings.ToUpper(shortcut.Key))
	return strings.Join(modifiers, "+"), nil
}

func shortcutModifierBindingName(modifier string) string {
	switch modifier {
	case "CmdOrCtrl":
		if runtime.GOOS == "darwin" {
			return "Cmd"
		}
		return "Ctrl"
	case "Alt":
		if runtime.GOOS == "darwin" {
			return "Option"
		}
		return "Alt"
	case "Super":
		if runtime.GOOS == "darwin" {
			return "Cmd"
		}
		if runtime.GOOS == "windows" {
			return "Win"
		}
		return "Super"
	default:
		return modifier
	}
}
