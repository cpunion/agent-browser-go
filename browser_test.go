package agentbrowser_test

import (
	"testing"

	agentbrowser "github.com/cpunion/agent-browser-go"
)

// TestBrowserManager_LaunchAndClose tests browser lifecycle
func TestBrowserManager_LaunchAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()

	// Test launch
	err := browser.Launch(agentbrowser.LaunchOptions{
		Headless: true,
	})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	// Test IsLaunched
	if !browser.IsLaunched() {
		t.Error("expected browser to be launched")
	}

	// Test close
	err = browser.Close()
	if err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	// Test IsLaunched after close
	if browser.IsLaunched() {
		t.Error("expected browser to be closed")
	}
}

// TestBrowserManager_Navigate tests navigation
func TestBrowserManager_Navigate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	// Navigate to example.com
	url, title, err := browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	if url != "https://example.com/" {
		t.Errorf("expected URL https://example.com/, got %s", url)
	}

	if title != "Example Domain" {
		t.Errorf("expected title 'Example Domain', got %s", title)
	}
}

// TestBrowserManager_GetText tests text extraction
func TestBrowserManager_GetText(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Get h1 text
	text, err := browser.GetText("h1")
	if err != nil {
		t.Fatalf("GetText() error = %v", err)
	}

	if text != "Example Domain" {
		t.Errorf("expected 'Example Domain', got %s", text)
	}
}

// TestBrowserManager_IsVisible tests visibility check
func TestBrowserManager_IsVisible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Check h1 visibility
	visible, err := browser.IsVisible("h1")
	if err != nil {
		t.Fatalf("IsVisible() error = %v", err)
	}

	if !visible {
		t.Error("expected h1 to be visible")
	}
}

// TestBrowserManager_Count tests element counting
func TestBrowserManager_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Count paragraphs
	count, err := browser.Count("p")
	if err != nil {
		t.Fatalf("Count() error = %v", err)
	}

	if count == 0 {
		t.Error("expected at least one paragraph")
	}
}

// TestBrowserManager_Screenshot tests screenshot functionality
func TestBrowserManager_Screenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Take screenshot
	buf, err := browser.Screenshot(false, "", 80)
	if err != nil {
		t.Fatalf("Screenshot() error = %v", err)
	}

	if len(buf) == 0 {
		t.Error("expected screenshot buffer to have data")
	}
}

// TestBrowserManager_Evaluate tests JavaScript evaluation
func TestBrowserManager_Evaluate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Evaluate JavaScript
	result, err := browser.Evaluate("document.title")
	if err != nil {
		t.Fatalf("Evaluate() error = %v", err)
	}

	if result != "Example Domain" {
		t.Errorf("expected 'Example Domain', got %v", result)
	}
}

// TestBrowserManager_Tabs tests tab management
func TestBrowserManager_Tabs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	// Create new tab
	index, err := browser.NewTab("")
	if err != nil {
		t.Fatalf("NewTab() error = %v", err)
	}

	if index != 1 {
		t.Errorf("expected new tab index 1, got %d", index)
	}

	// List tabs
	tabs, err := browser.ListTabs()
	if err != nil {
		t.Fatalf("ListTabs() error = %v", err)
	}

	if len(tabs) != 2 {
		t.Errorf("expected 2 tabs, got %d", len(tabs))
	}

	// Close tab
	err = browser.CloseTab(1)
	if err != nil {
		t.Fatalf("CloseTab() error = %v", err)
	}

	tabs, err = browser.ListTabs()
	if err != nil {
		t.Fatalf("ListTabs() error = %v", err)
	}

	if len(tabs) != 1 {
		t.Errorf("expected 1 tab after close, got %d", len(tabs))
	}
}

// TestBrowserManager_Snapshot tests snapshot generation
func TestBrowserManager_Snapshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	_, _, err = browser.Navigate("https://example.com", "load")
	if err != nil {
		t.Fatalf("Navigate() error = %v", err)
	}

	// Get snapshot
	snapshot, err := browser.GetSnapshot(agentbrowser.SnapshotOptions{})
	if err != nil {
		t.Fatalf("GetSnapshot() error = %v", err)
	}

	if snapshot.Tree == "" {
		t.Error("expected snapshot tree to have content")
	}

	if len(snapshot.Refs) == 0 {
		t.Error("expected snapshot to have refs")
	}
}

// TestBrowserManager_SetViewport tests viewport setting
func TestBrowserManager_SetViewport(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	browser := agentbrowser.NewBrowserManager()
	defer browser.Close()

	err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
	if err != nil {
		t.Fatalf("Launch() error = %v", err)
	}

	// Set viewport
	err = browser.SetViewport(1920, 1080)
	if err != nil {
		t.Fatalf("SetViewport() error = %v", err)
	}
}
