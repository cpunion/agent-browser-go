# agent-browser-go

Go implementation of [agent-browser](https://github.com/vercel-labs/agent-browser) - headless browser automation for AI agents.

## Table of Contents

- [Installation](#installation)
- [Quick Start](#quick-start)
- [CLI Usage](#cli-usage)
  - [Core Commands](#core-commands)
  - [Sessions](#sessions)
  - [Snapshot Options](#snapshot-options)
  - [Environment Variables](#environment-variables)
  - [CLI Options](#cli-options)
- [Go SDK](#go-sdk)
  - [Basic Usage](#basic-usage)
  - [Backend Selection](#backend-selection)
  - [Advanced Features](#advanced-features)
  - [API Reference](#api-reference)
- [Differences from TypeScript Version](#differences-from-typescript-version)
- [Architecture](#architecture)
- [License](#license)

## Installation

### From Binary

Download the latest release for your platform from [releases](https://github.com/cpunion/agent-browser-go/releases).

### From Source

```bash
git clone https://github.com/cpunion/agent-browser-go
cd agent-browser-go
go build -o agent-browser-go ./cmd/agent-browser-go

# Install Playwright driver (if using playwright backend)
./agent-browser-go install --backend playwright
```

### Linux Dependencies

On Linux, install system dependencies for Chromium:

```bash
# Ubuntu/Debian
sudo apt-get install -y \
  libnss3 libnspr4 libatk1.0-0 libatk-bridge2.0-0 libcups2 \
  libdrm2 libxkbcommon0 libxcomposite1 libxdamage1 libxfixes3 \
  libxrandr2 libgbm1 libasound2
```

## Quick Start

```bash
# Open a page
./agent-browser-go open https://example.com

# Get accessibility tree with refs
./agent-browser-go snapshot

# Click by ref
./agent-browser-go click @e1

# Screenshot
./agent-browser-go screenshot page.png

# Close browser
./agent-browser-go close
```

## CLI Usage

### Core Commands

```bash
# Navigation
agent-browser-go open <url>              # Navigate to URL
agent-browser-go back                    # Go back
agent-browser-go forward                 # Go forward
agent-browser-go reload                  # Reload page

# Interaction
agent-browser-go click <selector>        # Click element
agent-browser-go fill <selector> <text>  # Fill input
agent-browser-go type <selector> <text>  # Type into element
agent-browser-go press <key>             # Press key (Enter, Tab, etc.)
agent-browser-go hover <selector>        # Hover element
agent-browser-go scroll <direction>      # Scroll (up/down/left/right)

# Information
agent-browser-go get text <selector>     # Get text content
agent-browser-go get html <selector>     # Get HTML
agent-browser-go get value <selector>    # Get input value
agent-browser-go get title               # Get page title
agent-browser-go get url                 # Get current URL

# State checks
agent-browser-go is visible <selector>   # Check visibility
agent-browser-go is enabled <selector>   # Check if enabled
agent-browser-go is checked <selector>   # Check if checked

# Snapshot & Screenshot
agent-browser-go snapshot                # Get accessibility tree
agent-browser-go screenshot [path]       # Take screenshot

# Browser control
agent-browser-go close                   # Close browser
```

### Sessions

Run multiple isolated browser instances:

```bash
# Different sessions
agent-browser-go --session agent1 open https://site-a.com
agent-browser-go --session agent2 open https://site-b.com

# Or via environment variable
export AGENT_BROWSER_SESSION=agent1
agent-browser-go click "#btn"

# List sessions
agent-browser-go session list

# Stop specific session
agent-browser-go daemon stop --session agent1

# Stop all sessions
agent-browser-go daemon stop --all
```

Each session has its own:
- Browser instance
- Cookies and storage
- Navigation history
- Configuration (backend, headed mode, user data dir)

### Snapshot Options

```bash
agent-browser-go snapshot                # Full accessibility tree
agent-browser-go snapshot -i             # Interactive elements only
agent-browser-go snapshot -c             # Compact mode
agent-browser-go snapshot -d 3           # Limit depth to 3
agent-browser-go snapshot -s "#main"     # Scope to selector
```

| Option | Description |
|--------|-------------|
| `-i, --interactive` | Only show interactive elements (buttons, links, inputs) |
| `-c, --compact` | Remove empty structural elements |
| `-d, --depth <n>` | Limit tree depth |
| `-s, --selector <sel>` | Scope to CSS selector |

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `AGENT_BROWSER_SESSION` | Default session name | `default` |
| `AGENT_BROWSER_BACKEND` | Default backend (`chromedp` or `playwright`) | `chromedp` |
| `AGENT_BROWSER_USER_DATA_DIR` | User data directory for persistent profiles | - |
| `AGENT_BROWSER_LOCALE` | Browser locale (e.g., `en-US`, `zh-CN`) | - |
| `AGENT_BROWSER_USE_CHROME` | Use system Chrome (Playwright only, set to `1`) | - |

### CLI Options

| Option | Description |
|--------|-------------|
| `--session <name>` | Use isolated session |
| `--backend <name>` | Browser backend (`chromedp` or `playwright`) |
| `--head, --headed` | Show browser window (not headless) |
| `--user-data-dir <path>` | User data directory for persistent profiles |
| `--json` | JSON output |

## Go SDK

### Basic Usage

```go
package main

import (
    "fmt"
    "log"

    agentbrowser "github.com/cpunion/agent-browser-go"
)

func main() {
    // Create backend (chromedp or playwright)
    backend := agentbrowser.NewChromedpBackend()

    // Launch browser
    err := backend.Launch(agentbrowser.LaunchOptions{
        Headless: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    defer backend.Close()

    // Navigate
    err = backend.Navigate("https://example.com", agentbrowser.NavigateOptions{})
    if err != nil {
        log.Fatal(err)
    }

    // Get snapshot
    snapshot, err := backend.GetSnapshot(agentbrowser.SnapshotOptions{})
    if err != nil {
        log.Fatal(err)
    }
    fmt.Println(snapshot.Tree)

    // Click element
    err = backend.Click("button", agentbrowser.ClickOptions{})
    if err != nil {
        log.Fatal(err)
    }

    // Screenshot
    data, err := backend.Screenshot(agentbrowser.ScreenshotOptions{
        FullPage: true,
    })
    if err != nil {
        log.Fatal(err)
    }
    // Save screenshot data...
}
```

### Backend Selection

```go
// Chromedp (default, lightweight)
backend := agentbrowser.NewChromedpBackend()

// Playwright (more features, requires installation)
backend := agentbrowser.NewPlaywrightBackend()
```

**Chromedp vs Playwright:**

| Feature | Chromedp | Playwright |
|---------|----------|------------|
| Installation | No dependencies | Requires driver installation |
| Size | Lightweight | Larger (includes driver) |
| Features | Core automation | Full Playwright features |
| Performance | Faster startup | Slightly slower |
| Recommendation | Default choice | Use if you need Playwright-specific features |

### Advanced Features

#### Persistent Profiles (Login State)

```go
// Step 1: Manual login in headed mode
backend := agentbrowser.NewChromedpBackend()
err := backend.Launch(agentbrowser.LaunchOptions{
    Headless:    false,  // Show browser
    UserDataDir: "./my-profile",
})
// Manually log in to the site...
backend.Close()

// Step 2: Reuse login state in headless mode
backend = agentbrowser.NewChromedpBackend()
err = backend.Launch(agentbrowser.LaunchOptions{
    Headless:    true,   // Headless mode
    UserDataDir: "./my-profile",  // Reuse profile
})
// Now you're logged in!
```

#### Custom Browser Executable

```go
backend := agentbrowser.NewChromedpBackend()
err := backend.Launch(agentbrowser.LaunchOptions{
    ExecutablePath: "/path/to/chrome",
})
```

#### Viewport Configuration

```go
err := backend.Launch(agentbrowser.LaunchOptions{
    Viewport: &agentbrowser.Viewport{
        Width:  1920,
        Height: 1080,
    },
})
```

#### Using Refs from Snapshot

```go
// Get snapshot with refs
snapshot, _ := backend.GetSnapshot(agentbrowser.SnapshotOptions{})

// snapshot.Refs contains ref -> selector mapping
// Example: {"e1": {Selector: "...", Role: "button", Name: "Submit"}}

// Click by ref
err := backend.Click("@e1", agentbrowser.ClickOptions{})
```

### API Reference

#### Backend Interface

```go
type Backend interface {
    // Lifecycle
    Launch(opts LaunchOptions) error
    Close() error

    // Navigation
    Navigate(url string, opts NavigateOptions) error
    Back() error
    Forward() error
    Reload() error

    // Interaction
    Click(selector string, opts ClickOptions) error
    Fill(selector, value string, opts FillOptions) error
    Type(selector, text string, opts TypeOptions) error
    Press(key string) error
    Hover(selector string) error
    Scroll(direction string, pixels int) error

    // Information
    GetText(selector string) (string, error)
    GetHTML(selector string) (string, error)
    GetValue(selector string) (string, error)
    GetTitle() (string, error)
    GetURL() (string, error)

    // State
    IsVisible(selector string) (bool, error)
    IsEnabled(selector string) (bool, error)
    IsChecked(selector string) (bool, error)

    // Snapshot & Screenshot
    GetSnapshot(opts SnapshotOptions) (*EnhancedSnapshot, error)
    Screenshot(opts ScreenshotOptions) ([]byte, error)

    // Tabs
    NewTab() (int, error)
    CloseTab(index int) error
    SwitchTab(index int) error
    ListTabs() ([]TabInfo, error)

    // Cookies & Storage
    GetCookies() ([]Cookie, error)
    SetCookie(cookie Cookie) error
    ClearCookies() error
    GetLocalStorage(key string) (string, error)
    SetLocalStorage(key, value string) error

    // Evaluation
    Evaluate(script string) (interface{}, error)
}
```

#### LaunchOptions

```go
type LaunchOptions struct {
    Headless       bool      // Run in headless mode (default: true)
    UserDataDir    string    // User data directory for persistent profiles
    ExecutablePath string    // Custom browser executable path
    Locale         string    // Browser locale (e.g., "en-US")
    Viewport       *Viewport // Viewport size
}

type Viewport struct {
    Width  int
    Height int
}
```

#### SnapshotOptions

```go
type SnapshotOptions struct {
    Interactive bool   // Only interactive elements
    MaxDepth    int    // Maximum tree depth (0 = unlimited)
    Compact     bool   // Remove empty structural elements
    Selector    string // Scope to CSS selector
}
```

#### ScreenshotOptions

```go
type ScreenshotOptions struct {
    FullPage bool   // Capture full page
    Path     string // Save to file (optional)
}
```

## Differences from TypeScript Version

### Architecture

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Runtime | Node.js daemon | Go daemon |
| Browser Engine | Playwright | Chromedp (default) or Playwright |
| Binary | Rust CLI + Node.js fallback | Single Go binary |
| Installation | npm + Chromium download | Single binary (chromedp) or + driver (playwright) |

### Features

**Implemented:**
- ✅ Core commands (open, click, fill, type, etc.)
- ✅ Snapshot with refs
- ✅ Sessions
- ✅ Headed mode
- ✅ User data dir (persistent profiles)
- ✅ Screenshots
- ✅ Tabs management
- ✅ Get/Is commands
- ✅ Cookies & storage
- ✅ JavaScript evaluation

**Not Yet Implemented:**
- ❌ CDP mode (connect to existing browser)
- ❌ Streaming (WebSocket preview)
- ❌ Network interception
- ❌ Frames
- ❌ Dialogs
- ❌ Trace recording
- ❌ Device emulation
- ❌ Geolocation

### Command Differences

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Backend selection | Playwright only | `--backend chromedp` or `--backend playwright` |
| Default mode | Headless | Headless |
| Installation | `agent-browser install` | `agent-browser-go install --backend playwright` (chromedp needs no install) |
| Session management | `--session` | `--session` (same) |

### Environment Variables

**Go-specific:**
- `AGENT_BROWSER_BACKEND` - Set default backend
- `AGENT_BROWSER_USE_CHROME` - Use system Chrome (Playwright only)

**Shared:**
- `AGENT_BROWSER_SESSION` - Default session
- `AGENT_BROWSER_USER_DATA_DIR` - User data directory
- `AGENT_BROWSER_LOCALE` - Browser locale

## Architecture

agent-browser-go uses a client-daemon architecture:

1. **CLI Client** - Parses commands, communicates with daemon via Unix socket
2. **Daemon** - Manages browser instance (chromedp or playwright)
3. **Backend** - Abstraction layer supporting multiple browser engines

```
┌─────────────┐
│   CLI       │
│  (client)   │
└──────┬──────┘
       │ Unix Socket
┌──────▼──────┐
│   Daemon    │
│  (server)   │
└──────┬──────┘
       │
   ┌───▼────┐
   │Backend │
   └───┬────┘
       │
  ┌────▼────┬──────────┐
  │Chromedp │Playwright│
  └─────────┴──────────┘
```

**Benefits:**
- Fast subsequent commands (daemon stays running)
- Session isolation (multiple daemons)
- Backend flexibility (chromedp or playwright)

## License

Apache-2.0
