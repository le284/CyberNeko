package main

import "time"

// PetService 暴露给前端设置窗口调用。
// 前端通过 Wails runtime 的 Call.ByName 调用这些方法，不需要手写 IPC。
type PetService struct {
	manager *PetManager
}

func (s *PetService) GetSettings() AppSettings {
	if s.manager == nil {
		return defaultAppSettingsValue()
	}
	return s.manager.Settings()
}

func (s *PetService) SetPetCount(count int) (AppSettings, error) {
	if s.manager == nil {
		return defaultAppSettingsValue(), nil
	}
	return s.manager.SetPetCount(count)
}

func (s *PetService) NotifyVisualSettingsChanged() {
	if s.manager == nil {
		return
	}
	s.manager.BroadcastVisualsChanged(time.Now().UnixNano())
}
