#!/bin/bash
# Integration test for agent-browser-go
# Tests user-data-dir, headed mode, daemon processes, and snapshot for both backends

# Don't use set -e, we handle errors manually

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

PASS_COUNT=0
FAIL_COUNT=0

log_pass() {
    echo -e "${GREEN}✓ PASS${NC}: $1"
    ((PASS_COUNT++))
}

log_fail() {
    echo -e "${RED}✗ FAIL${NC}: $1"
    ((FAIL_COUNT++))
}

log_info() {
    echo -e "${YELLOW}→${NC} $1"
}

cleanup() {
    log_info "Cleaning up..."
    ./agent-browser-go daemon stop --all 2>/dev/null || true
    pkill -9 -f "chromedp-runner" 2>/dev/null || true
    pkill -9 -f "headless_shell.*user-data-dir.*test-integration" 2>/dev/null || true
    rm -rf test-integration-data 2>/dev/null || true
}

# Cleanup on exit
trap cleanup EXIT

test_backend() {
    local BACKEND="$1"
    local DATA_DIR="test-integration-data/$BACKEND"

    echo ""
    echo "══════════════════════════════════════════════════════════"
    echo " Testing backend: $BACKEND"
    echo "══════════════════════════════════════════════════════════"

    # 1. Stop all daemons and cleanup
    log_info "Stopping all daemons..."
    ./agent-browser-go daemon stop --all 2>/dev/null || true
    sleep 1
    rm -rf "$DATA_DIR"

    # 2. Test: Open URL with user-data-dir
    log_info "Opening URL with --user-data-dir..."
    RESULT=$(./agent-browser-go --backend "$BACKEND" --user-data-dir "$SCRIPT_DIR/$DATA_DIR" open https://example.com/ 2>&1)
    if [[ "$RESULT" == *"https://example.com"* ]]; then
        log_pass "[$BACKEND] Open URL returned correct URL"
    else
        log_fail "[$BACKEND] Open URL failed: $RESULT"
    fi

    sleep 2

    # 3. Test: User data directory was created
    if [ -d "$DATA_DIR" ] && [ "$(ls -A "$DATA_DIR")" ]; then
        log_pass "[$BACKEND] User data directory created and populated"
    else
        log_fail "[$BACKEND] User data directory not created or empty"
    fi

    # 4. Test: Daemon process is running
    if ./agent-browser-go daemon stop --all 2>&1 | grep -q "Stopping"; then
        log_pass "[$BACKEND] Daemon process was running"
    else
        log_fail "[$BACKEND] Daemon process was not running"
    fi

    sleep 1

    # 5. Test: Restart with same data dir, snapshot should work
    log_info "Restarting with same user-data-dir..."
    ./agent-browser-go --backend "$BACKEND" --user-data-dir "$SCRIPT_DIR/$DATA_DIR" open https://example.com/ >/dev/null 2>&1
    sleep 2

    # 6. Test: Snapshot returns content
    log_info "Testing snapshot..."
    SNAPSHOT=$(./agent-browser-go snapshot 2>&1)
    if [[ "$SNAPSHOT" == *"Example Domain"* ]] || [[ "$SNAPSHOT" == *"body"* && "$SNAPSHOT" == *"link"* ]]; then
        log_pass "[$BACKEND] Snapshot returns page content"
    else
        log_fail "[$BACKEND] Snapshot failed or empty: $SNAPSHOT"
    fi

    # 7. Test: Chrome/browser process has correct user-data-dir
    log_info "Checking browser process user-data-dir..."
    if ps ax | grep -v grep | grep "user-data-dir" | grep -q "$DATA_DIR"; then
        log_pass "[$BACKEND] Browser uses correct user-data-dir"
    else
        log_fail "[$BACKEND] Browser not using correct user-data-dir"
    fi

    # 8. Test: Daemon stop actually stops processes
    log_info "Testing daemon stop..."
    ./agent-browser-go daemon stop --all >/dev/null 2>&1
    sleep 2

    DAEMON_COUNT=$(ps ax | grep -v grep | grep "agent-browser-go daemon" | wc -l | tr -d ' ')
    if [ "$DAEMON_COUNT" -eq "0" ]; then
        log_pass "[$BACKEND] Daemon stop works correctly"
    else
        log_fail "[$BACKEND] Daemon processes still running after stop"
    fi

    # 9. Test: Headed mode triggers daemon restart
    log_info "Testing headed mode change restarts daemon..."
    ./agent-browser-go --backend "$BACKEND" --user-data-dir "$SCRIPT_DIR/$DATA_DIR" open https://example.com/ >/dev/null 2>&1
    sleep 1

    # Get current daemon PID
    DAEMON_PID=$(ps ax | grep -v grep | grep "agent-browser-go daemon" | awk '{print $1}' | head -1)

    # Now try with --head (should restart daemon)
    ./agent-browser-go --backend "$BACKEND" --user-data-dir "$SCRIPT_DIR/$DATA_DIR" --head open https://example.com/ >/dev/null 2>&1
    sleep 2

    NEW_DAEMON_PID=$(ps ax | grep -v grep | grep "agent-browser-go daemon" | awk '{print $1}' | head -1)

    if [ "$DAEMON_PID" != "$NEW_DAEMON_PID" ] && [ -n "$NEW_DAEMON_PID" ]; then
        log_pass "[$BACKEND] Headed mode change restarts daemon"
    else
        log_fail "[$BACKEND] Headed mode change did not restart daemon (old=$DAEMON_PID new=$NEW_DAEMON_PID)"
    fi

    # 10. Test: Headed mode actually launches browser without --headless
    log_info "Checking browser is running in headed mode..."
    sleep 1

    # Check that there's a Chrome/browser process for this data dir that does NOT have --headless
    if ps ax | grep -v grep | grep "user-data-dir" | grep "$DATA_DIR" | grep -v "headless" | grep -q .; then
        log_pass "[$BACKEND] Browser running in headed mode (no --headless flag)"
    else
        # Could also be --headless=new or other variants, check more carefully
        BROWSER_CMD=$(ps ax | grep -v grep | grep "user-data-dir" | grep "$DATA_DIR" | head -1)
        if [[ "$BROWSER_CMD" != *"headless"* ]]; then
            log_pass "[$BACKEND] Browser running in headed mode"
        else
            log_fail "[$BACKEND] Browser still has headless flag: $BROWSER_CMD"
        fi
    fi

    # macOS specific: Check if browser window is visible (optional, only on macOS)
    if [[ "$(uname)" == "Darwin" ]] && command -v osascript &> /dev/null; then
        log_info "Checking for visible browser window (macOS)..."
        WINDOW_COUNT=$(osascript -e 'tell app "Google Chrome" to count windows' 2>/dev/null || echo "0")
        if [ "$WINDOW_COUNT" -gt 0 ]; then
            log_pass "[$BACKEND] Chrome has $WINDOW_COUNT visible window(s)"
        else
            log_info "[$BACKEND] Could not detect visible windows (might be using headless_shell)"
        fi
    fi

    # Cleanup for this backend
    ./agent-browser-go daemon stop --all 2>/dev/null || true
    sleep 1
}

# Main
echo "╔══════════════════════════════════════════════════════════╗"
echo "║      Agent-Browser-Go Integration Tests                  ║"
echo "╚══════════════════════════════════════════════════════════╝"

# Build first
log_info "Building agent-browser-go..."
go build -o agent-browser-go ./cmd/agent-browser-go

# Run tests for both backends
test_backend "chromedp"
test_backend "playwright"

# Summary
echo ""
echo "══════════════════════════════════════════════════════════"
echo " Summary"
echo "══════════════════════════════════════════════════════════"
echo -e "  ${GREEN}Passed${NC}: $PASS_COUNT"
echo -e "  ${RED}Failed${NC}: $FAIL_COUNT"
echo ""

if [ "$FAIL_COUNT" -gt 0 ]; then
    echo -e "${RED}Some tests failed!${NC}"
    exit 1
else
    echo -e "${GREEN}All tests passed!${NC}"
    exit 0
fi
