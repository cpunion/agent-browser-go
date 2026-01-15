# agent-browser-go

[![CI](https://github.com/cpunion/agent-browser-go/actions/workflows/ci.yml/badge.svg)](https://github.com/cpunion/agent-browser-go/actions/workflows/ci.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/cpunion/agent-browser-go.svg)](https://pkg.go.dev/github.com/cpunion/agent-browser-go)

Headless browser automation CLI for AI agents - Go implementation.

## Features

- 🚀 **Single binary** - No external dependencies, no Node.js required
- 🎯 **AI-friendly** - Accessibility tree with deterministic refs (`@e1`, `@e2`)
- 🔄 **Session isolation** - Multiple independent browser sessions
- 📡 **JSON protocol** - Easy integration with AI agents
- 🌐 **Cross-platform** - macOS, Linux, Windows
- 🔧 **Dual backend** - chromedp (default) or playwright

## Installation

```bash
go install github.com/cpunion/agent-browser-go/cmd/agent-browser-go@latest
```

Or build from source:

```bash
git clone https://github.com/cpunion/agent-browser-go
cd agent-browser-go
go build -o agent-browser-go ./cmd/agent-browser-go
```

## Quick Start

```bash
# Launch browser and navigate
agent-browser-go open https://example.com

# Get accessibility snapshot with refs
agent-browser-go snapshot -i

# Click element by ref
agent-browser-go click @e1

# Fill input
agent-browser-go fill @e3 "test@example.com"

# Take screenshot
agent-browser-go screenshot page.png

# Close browser
agent-browser-go close
```

## Commands

### Navigation

```bash
agent-browser-go open <url>              # Navigate to URL (aliases: goto, navigate)
agent-browser-go back                    # Go back
agent-browser-go forward                 # Go forward
agent-browser-go reload                  # Reload page
```

### Interaction

```bash
agent-browser-go click <selector>        # Click element
agent-browser-go dblclick <selector>     # Double-click element
agent-browser-go type <selector> <text>  # Type into element
agent-browser-go fill <selector> <text>  # Clear and fill
agent-browser-go press <key>             # Press key (Enter, Tab, etc.)
agent-browser-go hover <selector>        # Hover element
agent-browser-go select <selector> <val> # Select dropdown option
agent-browser-go check <selector>        # Check checkbox
agent-browser-go uncheck <selector>      # Uncheck checkbox
agent-browser-go scroll <dir> [px]       # Scroll (up/down/left/right)
agent-browser-go drag <src> <tgt>        # Drag and drop
```

### Inspection

```bash
agent-browser-go snapshot                # Accessibility tree with refs (best for AI)
agent-browser-go snapshot -i             # Interactive elements only
agent-browser-go snapshot -c             # Compact (remove empty elements)
agent-browser-go snapshot -d 3           # Limit depth to 3 levels
agent-browser-go screenshot [path]       # Take screenshot (--full for full page)
agent-browser-go get text <selector>     # Get text content
agent-browser-go get html <selector>     # Get innerHTML
agent-browser-go get value <selector>    # Get input value
agent-browser-go get title               # Get page title
agent-browser-go get url                 # Get current URL
```

### State Checks

```bash
agent-browser-go is visible <selector>   # Check if visible
agent-browser-go is enabled <selector>   # Check if enabled
agent-browser-go is checked <selector>   # Check if checked
```

### Wait

```bash
agent-browser-go wait <selector>         # Wait for element to be visible
agent-browser-go wait <ms>               # Wait for time (milliseconds)
agent-browser-go wait --text "Welcome"   # Wait for text to appear
agent-browser-go wait --url "**/dash"    # Wait for URL pattern
```

### Daemon Management

```bash
agent-browser-go daemon stop             # Stop current session daemon
agent-browser-go daemon stop --all       # Stop all daemons
agent-browser-go session list            # List active sessions
```

## Selectors

### Refs (Recommended for AI)

Refs provide deterministic element selection from snapshots:

```bash
# 1. Get snapshot with refs
agent-browser-go snapshot
# Output:
# - heading "Example Domain" [ref=e1] [level=1]
# - button "Submit" [ref=e2]
# - textbox "Email" [ref=e3]
# - link "Learn more" [ref=e4]

# 2. Use refs to interact
agent-browser-go click @e2                   # Click the button
agent-browser-go fill @e3 "test@example.com" # Fill the textbox
agent-browser-go get text @e1                # Get heading text
```

**Why use refs?**
- **Deterministic**: Ref points to exact element from snapshot
- **Fast**: No DOM re-query needed
- **AI-friendly**: Snapshot + ref workflow is optimal for LLMs

### CSS Selectors

```bash
agent-browser-go click "#id"
agent-browser-go click ".class"
agent-browser-go click "div > button"
```

## Sessions

Run multiple isolated browser instances:

```bash
# Different sessions
agent-browser-go --session agent1 open site-a.com
agent-browser-go --session agent2 open site-b.com

# Or via environment variable
AGENT_BROWSER_SESSION=agent1 agent-browser-go click "#btn"

# List active sessions
agent-browser-go session list
```

Each session has its own:
- Browser instance
- Cookies and storage
- Navigation history
- Authentication state

## JSON Output (for AI agents)

```bash
agent-browser-go --json snapshot -i
```

```json
{
  "id": "1234567890",
  "success": true,
  "data": {
    "snapshot": "- button \"Submit\" [ref=e1]\n- link \"Learn more\" [ref=e2]",
    "refs": {
      "e1": {"role": "button", "name": "Submit"},
      "e2": {"role": "link", "name": "Learn more"}
    }
  }
}
```

## Global Options

| Option | Description |
|--------|-------------|
| `--session <name>` | Use isolated session (or `AGENT_BROWSER_SESSION` env) |
| `--backend <type>` | Browser backend: `chromedp` (default) or `playwright` |
| `--user-data-dir <path>` | Browser profile directory for persistent sessions |
| `--headed` | Show browser window (not headless) |
| `--json` | JSON output (for agents) |

## Environment Variables

| Variable | Description |
|----------|-------------|
| `AGENT_BROWSER_SESSION` | Default session name |
| `AGENT_BROWSER_BACKEND` | Default backend (`chromedp` or `playwright`) |
| `AGENT_BROWSER_USER_DATA_DIR` | Browser profile directory for persistent login |
| `AGENT_BROWSER_NO_SANDBOX` | Set to `1` to disable sandbox (containers) |
| `AGENT_BROWSER_DISABLE_SHM` | Set to `1` to disable shared memory (Docker) |

### Persistent Browser Profile

```bash
# Use existing Chrome profile (maintains login sessions)
agent-browser-go --user-data-dir "$HOME/Library/Application Support/Google/Chrome/Default" --head open https://example.com

# Or create a dedicated profile
agent-browser-go --user-data-dir ./browser-profile --head open https://studio.youtube.com
```

## Architecture

```
┌─────────────────┐     ┌─────────────────┐     ┌─────────────────┐
│   CLI Client    │────▶│     Daemon      │────▶│    Browser      │
│  (agent-browser)│     │   (background)  │     │   (chromedp)    │
└─────────────────┘     └─────────────────┘     └─────────────────┘
        │                       │
        │    Unix Socket        │    Chrome DevTools Protocol
        │    (JSON Protocol)    │
```

- **Browser Engine**: chromedp (native Go Chrome DevTools Protocol) or playwright-go
- **IPC**: Unix sockets (macOS/Linux), TCP (Windows)
- **Protocol**: JSON-based command/response
- **Sessions**: Isolated via PID files

## Comparison with TypeScript Version

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Runtime | Node.js | Single binary |
| Browser | Playwright | chromedp / playwright |
| Startup | ~500ms | ~200ms |
| Memory | ~150MB | ~80MB |
| Distribution | npm | Binary |

## Development

```bash
# Install dependencies
go mod tidy

# Build
go build ./...

# Run tests
go test ./...

# Build CLI
go build -o agent-browser-go ./cmd/agent-browser-go

# Test UserDataDir functionality
go run ./cmd/userdir-test
```

## License

Apache License 2.0

## Credits

This project is a Go implementation inspired by [agent-browser](https://github.com/vercel-labs/agent-browser) by Vercel Labs, which is licensed under the Apache License 2.0.
