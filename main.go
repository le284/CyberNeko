package main

import (
	"embed"
	_ "embed"
	"log"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// Wails uses Go's `embed` package to embed the frontend files into the binary.
// Any files in the frontend/dist folder will be embedded into the binary and
// made available to the frontend.
// See https://pkg.go.dev/embed for more information.

//go:embed all:frontend/dist
var assets embed.FS

func init() {
	// 注册前后端约定好的宠物状态事件。
	// 右键菜单在 Go 侧切换状态后，通过这个事件通知所有 WebView 中的 JS 更新角色外观。
	application.RegisterEvent[string]("pet:state")
	application.RegisterEvent[string]("pet:direction")
}

type petProfile struct {
	ID         string
	WindowName string
	Title      string
	MenuName   string
	StartRatio float64
}

// main 是 CyberNeko 的桌面端入口。
// 第一阶段只处理“透明无边框窗口 + 置顶 + 主体拖拽 + 右键菜单”这条最小闭环。
func main() {
	// 创建 Wails 应用实例。
	// Assets 使用 embed 后的 frontend/dist，因此 `wails3 build` 会把前端静态文件打进最终二进制。
	app := application.New(application.Options{
		Name:        "CyberNeko",
		Description: "CyberNeko desktop pet",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(assets),
		},
		Mac: application.MacOptions{
			// Accessory 应用不会显示在 macOS Dock 中，行为更接近菜单栏工具或桌面组件。
			ActivationPolicy: application.ActivationPolicyAccessory,

			// 关闭最后一个窗口后直接退出应用；桌宠没有传统多窗口生命周期。
			ApplicationShouldTerminateAfterLastWindowClosed: true,
		},
	})

	// 同一个 Wails 进程可以创建多个透明 WebView 窗口。
	// 每个窗口加载同一套前端资源，但用 URL query 传入不同 pet ID，让 JS/CSS 切换外观和台词。
	petProfiles := []petProfile{
		{ID: "neko", WindowName: "pet-neko", Title: "CyberNeko", MenuName: "neko-menu", StartRatio: 0.32},
		{ID: "momo", WindowName: "pet-momo", Title: "CyberMomo", MenuName: "momo-menu", StartRatio: 0.68},
	}

	for _, profile := range petProfiles {
		petWindow := newPetWindow(app, profile)
		setMovementMode(profile.WindowName, movementModeEdgeWander)
		registerPetContextMenu(app, profile, petWindow)
		startTopEdgeWalker(app, petWindow, profile.WindowName, profile.StartRatio)
	}

	// Run 会阻塞直到应用退出。
	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}

func newPetWindow(app *application.App, profile petProfile) *application.WebviewWindow {
	// 创建桌宠窗口。
	// 这里的关键点是：窗口本身透明且无边框，真正可见的内容只由前端绘制出来的猫咪 DOM 决定。
	return app.Window.NewWithOptions(application.WebviewWindowOptions{
		Name:  profile.WindowName,
		Title: profile.Title,

		// 桌宠窗口要尽量贴合角色主体，减少透明矩形区域对桌面交互的影响。
		Width:  petWindowWidth,
		Height: petWindowHeight,

		// Frameless 去掉系统标题栏和边框。
		// 底层含义：不再由操作系统绘制 NSWindow/HWND/X11 的默认 chrome，拖拽区域需要应用自己声明。
		Frameless: true,

		// AlwaysOnTop 让窗口进入高层级 z-order，保持浮在浏览器、IDE 等普通窗口上方。
		AlwaysOnTop: true,

		// 桌宠第一阶段不允许用户拖拽边缘改变窗口大小，避免透明窗口命中区域变得不可预测。
		DisableResize: true,

		// BackgroundTypeTransparent 是透明窗口的核心。
		// 配合 Alpha=0 的 BackgroundColour，让 WebView 背景不再填充实色，只显示前端绘制的猫咪。
		BackgroundType:   application.BackgroundTypeTransparent,
		BackgroundColour: application.NewRGBA(0, 0, 0, 0),

		// 禁用浏览器默认右键菜单，右键交给 Wails 原生 ContextMenu。
		DefaultContextMenuDisabled: true,

		// 不在这里设置 IgnoreMouseEvents=true。
		// 该属性会让整个窗口都忽略鼠标事件，适合“纯展示浮层”，但会同时破坏猫咪主体拖拽和右键菜单。
		// 后续做“透明像素穿透”时，应在平台层按命中测试切换，而不是一刀切忽略整个窗口。
		IgnoreMouseEvents: false,

		// 加载 Vite 构建出的首页，并把宠物 ID 交给前端选择皮肤、台词和可访问性名称。
		URL: "/?pet=" + profile.ID,

		Mac: application.MacWindow{
			// macOS 下设置透明 Backdrop，避免系统给 NSWindow 补一层默认背景。
			Backdrop: application.MacBackdropTransparent,

			// 隐藏标题栏并让内容区域占满整个窗口。
			TitleBar: application.MacTitleBar{
				Hide:            true,
				HideTitle:       true,
				FullSizeContent: true,
			},

			// 关闭系统阴影，避免透明窗口外沿出现一圈矩形阴影。
			DisableShadow: true,

			// Floating 让窗口处于普通窗口之上；配合 AlwaysOnTop 提升桌宠稳定性。
			WindowLevel: application.MacWindowLevelFloating,

			// 允许窗口跨 Spaces 显示，并尽量不参与系统窗口循环，行为更像桌面小组件。
			CollectionBehavior: application.MacWindowCollectionBehaviorCanJoinAllSpaces |
				application.MacWindowCollectionBehaviorFullScreenAuxiliary |
				application.MacWindowCollectionBehaviorIgnoresCycle,
		},

		Windows: application.WindowsWindow{
			// Windows 下隐藏任务栏按钮，Wails 会使用相应的窗口扩展样式让它更接近 Tool Window。
			HiddenOnTaskbar: true,

			// Frameless 模式下去掉系统额外装饰，避免透明窗口周围残留圆角、阴影或边框。
			DisableFramelessWindowDecorations: true,
		},
	})
}

func registerPetContextMenu(app *application.App, profile petProfile, petWindow *application.WebviewWindow) {
	menu := app.ContextMenu.New()

	menu.Add(profile.Title + "：原地待机").OnClick(func(_ *application.Context) {
		setMovementMode(profile.WindowName, movementModeIdle)
		setPetAnimationState(petWindow, "idle")
	})

	menu.Add(profile.Title + "：沿窗口边缘巡游").OnClick(func(_ *application.Context) {
		setMovementMode(profile.WindowName, movementModeEdgeWander)
	})

	menu.Add(profile.Title + "：跟随鼠标").OnClick(func(_ *application.Context) {
		setMovementMode(profile.WindowName, movementModeFollowMouse)
	})

	menu.AddSeparator()

	menu.Add("Exit").OnClick(func(_ *application.Context) {
		app.Quit()
	})

	// 前端通过 CSS 自定义属性 `--custom-contextmenu` 把不同宠物主体绑定到不同原生菜单。
	app.ContextMenu.Add(profile.MenuName, menu)
}
