# CyberNeko

CyberNeko 是一个基于 Wails 3 的跨平台桌面宠物基础框架。当前版本支持通过设置窗口调整启动宠物数量，每只宠物都可以独立切换待机、沿窗口边缘巡游、跟随鼠标等行为，并支持为不同宠物上传自定义图片。

## 目录结构

```text
CyberNeko/
├── main.go                    # Wails 入口；创建多只透明无边框桌宠窗口和独立原生右键菜单
├── go.mod                     # Go 模块与 Wails 依赖
├── frontend/
│   ├── index.html             # 透明舞台、猫咪主体 DOM 和设置窗口 DOM
│   ├── package.json           # Vite 前端脚本
│   ├── public/logo.svg        # 前端 favicon 与设置页 logo
│   ├── public/style.css       # 猫咪占位造型、设置窗口、自定义图片和拖拽命中区样式
│   └── src/main.js            # 150ms 动画循环、宠物皮肤、设置、图片上传和气泡文案
├── app_settings.go            # 宠物数量、行为快捷键等设置的落盘读写
├── pet_manager.go             # 多宠物窗口管理、设置窗口、快捷键和右键菜单
├── pet_service.go             # 暴露给前端设置窗口的 Wails 服务
└── build/                     # Wails 平台构建配置与图标资源
```

## 当前能力

- `Frameless`: 去除系统标题栏和窗口边框。
- `BackgroundTypeTransparent`: 让窗口背景透明，只显示前端绘制出来的猫咪主体。
- `AlwaysOnTop`: 默认置顶显示。
- `Windows.HiddenOnTaskbar`: Windows 下隐藏任务栏按钮，让窗口更像工具组件。
- 手动 Pointer Events 拖拽：左键按住猫咪主体时调用 Wails `Window.SetPosition` 移动对应窗口。
- `--custom-contextmenu`: 每只宠物绑定自己的 Wails 原生菜单，可单独切换原地待机、沿窗口边缘巡游、跟随鼠标或退出。
- 边缘巡游会切换为“趴边探头”姿态，并在桌宠窗口内绘制前景边缘挡板，让猫咪像从当前窗口后面伸出头、把爪子搭在窗口边缘上；上方空间不足时会自动改趴到下边缘。
- `?pet=neko` / `?pet=momo`: 同一套前端资源按 URL 参数切换不同宠物皮肤、台词和菜单。
- 右键菜单中的“设置...”会打开设置窗口，可修改启动宠物数量。当前上限为 6 只，设置会保存到系统用户配置目录。
- 设置窗口支持录制行为快捷键：默认 `Ctrl+Alt+1` 待机、`Ctrl+Alt+2` 沿边缘巡游、`Ctrl+Alt+3` 跟随鼠标、`Ctrl+Alt+Space` 循环切换。焦点在某只宠物窗口时只切换这只，焦点在设置窗口时会切换所有已显示宠物。
- 设置窗口支持为每个宠物槽位上传 PNG、JPG、WebP 或 GIF；非 GIF 图片会在前端压缩为 WebP 后写入本地 `localStorage`。

Logo 与图标说明：`frontend/public/logo.svg` 是前端 favicon 与设置页 logo，`build/appicon.png`、`build/windows/icon.ico`、`build/darwin/icons.icns` 是 Wails 打包时使用的平台图标。

透明区域穿透说明：当前版本已让 DOM 空白区域 `pointer-events: none`，不会被前端元素拦截。但操作系统级“按透明像素穿透到后方应用”需要下一阶段接入平台命中测试；不能直接全局启用 `IgnoreMouseEvents`，否则猫咪主体也无法拖拽和右键。

## 运行

需要先安装：

- Go 1.25+，并确保 `wails3` 调用到的 `go` 也是这个版本。
- Node.js 18+、20+ 或 22+。
- Wails 3 CLI。

开发模式：

```bash
cd /Users/fanwenbin/github/desk-mochi/CyberNeko
wails3 dev
```

如果机器上同时存在多个 Go 版本，可以临时把 Go 1.25 放到 PATH 前面：

```bash
export PATH="/opt/homebrew/bin:$PATH"
go version
wails3 dev
```

生产构建：

```bash
cd /Users/fanwenbin/github/desk-mochi/CyberNeko
wails3 build
```

只验证前端构建：

```bash
cd /Users/fanwenbin/github/desk-mochi/CyberNeko/frontend
npm run build
```

## GitHub Actions 自动打包

仓库内置 `.github/workflows/build-packages.yml`：

- push 到 `main` 或手动运行 workflow 时，会自动构建 Windows amd64、macOS arm64、macOS amd64 三份 zip，并上传为 Actions artifact。
- 创建 `v*` tag（例如 `v0.1.0`）时，会额外把这些 zip 汇总上传到 GitHub Release。
- macOS 包当前使用 ad-hoc 签名，尚未接入 Apple Developer ID 公证；首次打开时如遇 Gatekeeper 提示，可右键选择“打开”，或在终端执行 `xattr -dr com.apple.quarantine CyberNeko.app` 后再启动。

发布一个带 Release 包的版本：

```bash
git tag v0.1.0
git push origin main --tags
```
