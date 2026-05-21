package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's embed package to embed frontend/dist into the binary.
// The generated app can therefore serve the UI without a separate web server.

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// 这些事件是 Go 与所有 WebView 窗口之间的轻量协议。
	// RegisterEvent 会让 Wails 在运行期校验事件数据类型，并为后续 bindings 生成类型信息。
	application.RegisterEvent[string]("pet:state")
	application.RegisterEvent[string]("pet:direction")
	application.RegisterEvent[AppSettingsChangedEvent]("pet:settings")
	application.RegisterEvent[PetVisualsChangedEvent]("pet:visuals")
}

func main() {
	settings, settingsPath, err := loadAppSettings()
	if err != nil {
		log.Printf("load settings failed, using defaults: %v", err)
	}

	petService := &PetService{}

	app := application.New(application.Options{
		Name:        "CyberNeko",
		Description: "CyberNeko desktop pet",
		Services: []application.Service{
			application.NewService(petService),
		},
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// Accessory 应用不会显示在 macOS Dock 中，更像桌面组件。
			ActivationPolicy: application.ActivationPolicyAccessory,

			// 至少保留一个宠物窗口；所有窗口都关闭时才退出。
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	manager := newPetManager(app, settings, settingsPath)
	petService.manager = manager
	manager.applyShortcuts(manager.Settings().Shortcuts)
	manager.applyInitialPetCount()

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
