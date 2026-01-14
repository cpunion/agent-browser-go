# Backend 切换指南

## 当前支持的 Backend

1. **chromedp** (默认) - 使用 Chrome DevTools Protocol
2. **playwright** (存根) - 待实现

## 切换方式

### 方式 1: 代码中切换

```go
import agentbrowser "github.com/cpunion/agent-browser-go"

// 默认使用 chromedp
browser := agentbrowser.NewBrowserManager()

// 显式指定 backend
browser := agentbrowser.NewBrowserManagerWithBackend(agentbrowser.BackendChromedp)

// 使用 playwright (需要先实现)
browser := agentbrowser.NewBrowserManagerWithBackend(agentbrowser.BackendPlaywright)
```

### 方式 2: 环境变量

```bash
# 设置默认 backend
export AGENT_BROWSER_BACKEND=chromedp  # 或 playwright

# 运行 CLI
./agent-browser-go open https://example.com
```

### 方式 3: CLI 参数

```bash
# 使用 chromedp (默认)
./agent-browser-go open https://example.com

# 使用 chromedp (显式)
./agent-browser-go --backend chromedp open https://example.com

# 使用 playwright
./agent-browser-go --backend playwright open https://example.com

# 简写形式
./agent-browser-go -b playwright open https://example.com
```

## 实现状态

### ✅ 已完成
- [x] BrowserBackend 接口定义
- [x] ChromeDPBackend 完整实现
- [x] PlaywrightBackend 存根
- [x] 工厂模式支持
- [x] BrowserManager 包装器
- [x] CLI `--backend` 参数
- [x] 环境变量 `AGENT_BROWSER_BACKEND`
- [x] Daemon backend 配置

### 🚧 待实现
- [ ] PlaywrightBackend 完整实现

## Backend 对比

| 特性 | chromedp | playwright |
|------|----------|------------|
| 状态 | ✅ 完整实现 | 🚧 存根 |
| 依赖 | chromedp | playwright-go |
| 浏览器 | Chrome/Chromium | Chrome/Firefox/WebKit |
| 性能 | 快 | 中等 |
| 兼容性 | Chrome only | 多浏览器 |
| 二进制大小 | 小 | 大 |

## 快速开始

当前推荐使用默认的 chromedp backend：

```go
browser := agentbrowser.NewBrowserManager()
browser.Launch(agentbrowser.LaunchOptions{Headless: true})
defer browser.Close()

browser.Navigate("https://example.com", "load")
```

## 添加新 Backend

1. 实现 `BrowserBackend` 接口
2. 在 `browser_factory.go` 中注册
3. 添加 `BackendType` 常量
4. 实现所有必需方法

示例见 `playwright_backend.go`。
