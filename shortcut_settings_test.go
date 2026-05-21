package main

import (
	"runtime"
	"testing"
)

func TestNormalizeShortcutSettingsDefaultsInvalidAndDuplicates(t *testing.T) {
	settings := normalizeShortcutSettings(ShortcutSettings{
		Idle:        "ctrl+alt+1",
		EdgeWander:  "Ctrl+Alt+1",
		FollowMouse: "not-a-shortcut",
		Cycle:       "Shift+F2",
	})

	if settings.Idle != "Ctrl+Alt+1" {
		t.Fatalf("Idle = %q, want Ctrl+Alt+1", settings.Idle)
	}
	if settings.EdgeWander != defaultShortcutSettings().EdgeWander {
		t.Fatalf("EdgeWander = %q, want default", settings.EdgeWander)
	}
	if settings.FollowMouse != defaultShortcutSettings().FollowMouse {
		t.Fatalf("FollowMouse = %q, want default", settings.FollowMouse)
	}
	if settings.Cycle != "Shift+F2" {
		t.Fatalf("Cycle = %q, want Shift+F2", settings.Cycle)
	}
}

func TestNormalizeShortcutSettingsRejectsBareLetterButAllowsFunctionKey(t *testing.T) {
	settings := normalizeShortcutSettings(ShortcutSettings{
		Idle:        "A",
		EdgeWander:  "F8",
		FollowMouse: "Ctrl+Alt+3",
		Cycle:       "Ctrl+Alt+Space",
	})

	if settings.Idle != defaultShortcutSettings().Idle {
		t.Fatalf("Idle = %q, want default", settings.Idle)
	}
	if settings.EdgeWander != "F8" {
		t.Fatalf("EdgeWander = %q, want F8", settings.EdgeWander)
	}
}

func TestShortcutBindingKeyUsesPlatformModifierNames(t *testing.T) {
	binding, err := shortcutBindingKey("Ctrl+Alt+Space")
	if err != nil {
		t.Fatalf("shortcutBindingKey returned error: %v", err)
	}

	expected := "Alt+Ctrl+SPACE"
	if runtime.GOOS == "darwin" {
		expected = "Ctrl+Option+SPACE"
	}
	if binding != expected {
		t.Fatalf("binding = %q, want %q", binding, expected)
	}
}
