package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const (
	defaultPetCount = 1
	maxPetCount     = 6
)

// AppSettings 是会落盘的应用设置。
// 字段使用 json tag，是为了让 Wails 返回给前端时直接得到 camelCase 数据。
type AppSettings struct {
	PetCount int `json:"petCount"`
	MaxPets  int `json:"maxPets"`
}

// AppSettingsChangedEvent 会广播给所有 WebView，让设置窗口和宠物窗口保持一致。
type AppSettingsChangedEvent struct {
	PetCount int `json:"petCount"`
	MaxPets  int `json:"maxPets"`
}

// PetVisualsChangedEvent 用于通知所有宠物窗口重新读取本地图片缓存。
type PetVisualsChangedEvent struct {
	Revision int64 `json:"revision"`
}

func defaultAppSettingsValue() AppSettings {
	return AppSettings{PetCount: defaultPetCount, MaxPets: maxPetCount}
}

func normalizeAppSettings(settings AppSettings) AppSettings {
	settings.MaxPets = maxPetCount
	settings.PetCount = clamp(settings.PetCount, 1, maxPetCount)
	if settings.PetCount == 0 {
		settings.PetCount = defaultPetCount
	}
	return settings
}

func appSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "CyberNeko", "settings.json"), nil
}

func loadAppSettings() (AppSettings, string, error) {
	settingsPath, err := appSettingsPath()
	if err != nil {
		return defaultAppSettingsValue(), "", err
	}

	data, err := os.ReadFile(settingsPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaultAppSettingsValue(), settingsPath, nil
		}
		return defaultAppSettingsValue(), settingsPath, err
	}

	settings := defaultAppSettingsValue()
	if err := json.Unmarshal(data, &settings); err != nil {
		return defaultAppSettingsValue(), settingsPath, err
	}
	return normalizeAppSettings(settings), settingsPath, nil
}

func saveAppSettings(settingsPath string, settings AppSettings) error {
	if settingsPath == "" {
		return errors.New("settings path is empty")
	}

	settings = normalizeAppSettings(settings)
	if err := os.MkdirAll(filepath.Dir(settingsPath), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(settingsPath, data, 0o644)
}
