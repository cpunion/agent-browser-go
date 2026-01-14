# agent-browser-go

Headless browser automation CLI for AI agents - Go implementation.

## Features

- 🚀 **Single binary** - No external dependencies
- 🎯 **AI-friendly** - Accessibility tree with deterministic refs (`@e1`, `@e2`)
- 🔄 **Session isolation** - Multiple independent browser sessions
- 📡 **JSON protocol** - Easy integration with AI agents
- 🌐 **Cross-platform** - macOS, Linux, Windows

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

## Core Commands

### Navigation
- `open <url>` - Navigate to URL
- `back` - Go back
- `forward` - Go forward
- `reload` - Reload page

### Interaction
- `click <selector>` - Click element
- `type <selector> <text>` - Type into element
- `fill <selector> <value>` - Clear and fill
- `press <key>` - Press key (Enter, Tab, etc.)
- `hover <selector>` - Hover over element

### Inspection
- `snapshot` - Get accessibility tree with refs
- `screenshot [path]` - Take screenshot
- `get text <selector>` - Get element text
- `get html <selector>` - Get element HTML
- `get url` - Get current URL
- `get title` - Get page title

### State Checks
- `is visible <selector>` - Check if visible
- `is enabled <selector>` - Check if enabled
- `is checked <selector>` - Check if checked

## Selectors

### Refs (Recommended for AI)
```bash
agent-browser-go snapshot -i
# Output: - button "Submit" [ref=e1]

agent-browser-go click @e1
```

### CSS Selectors
```bash
agent-browser-go click "#submit-button"
agent-browser-go fill ".email-input" "test@example.com"
```

## Session Management

```bash
# Use isolated sessions
agent-browser-go --session work open site-a.com
agent-browser-go --session personal open site-b.com

# List active sessions
agent-browser-go session list
```

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

## Environment Variables

### Backend Selection
- `AGENT_BROWSER_BACKEND` - Choose browser backend (`chromedp` or `playwright`)
  ```bash
  export AGENT_BROWSER_BACKEND=playwright
  agent-browser-go open https://example.com
  ```

### User Data Directory (Persistent Profiles)
- `AGENT_BROWSER_USER_DATA_DIR` - Path to browser profile directory
  ```bash
  # Use existing Chrome profile (maintains login sessions)
  export AGENT_BROWSER_USER_DATA_DIR="$HOME/Library/Application Support/Google/Chrome/Default"

  # Or create a dedicated profile
  export AGENT_BROWSER_USER_DATA_DIR="$PWD/browser-profile"
  agent-browser-go --head open https://studio.youtube.com
  ```

### Anti-Detection (for sites like YouTube Studio)
- `AGENT_BROWSER_NO_SANDBOX=1` - Disable sandbox (use in containers)
- `AGENT_BROWSER_DISABLE_SHM=1` - Disable shared memory (use in Docker)

**Default anti-detection flags** (always enabled):
- `--disable-blink-features=AutomationControlled` - Hide automation flag
- `--disable-infobars` - Remove automation banner
- `--excludeSwitches=enable-automation` - Remove automation markers

## Architecture

- **Browser Engine**: chromedp (native Go Chrome DevTools Protocol)
- **IPC**: Unix sockets (macOS/Linux), TCP (Windows)
- **Protocol**: JSON-based command/response
- **Sessions**: Isolated via PID files

## Comparison with TypeScript Version

| Feature | TypeScript | Go |
|---------|-----------|-----|
| Runtime | Node.js | Single binary |
| Browser | Playwright | chromedp |
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
```

## License

MIT

## Credits

Based on [agent-browser](https://github.com/vercel-labs/agent-browser) by Vercel Labs.
