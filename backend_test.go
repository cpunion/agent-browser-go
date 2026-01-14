package agentbrowser_test

import (
	"testing"

	agentbrowser "github.com/cpunion/agent-browser-go"
)

// testBackends returns all backends to test
func testBackends() []struct {
	name    string
	backend agentbrowser.BackendType
} {
	return []struct {
		name    string
		backend agentbrowser.BackendType
	}{
		{"chromedp", agentbrowser.BackendChromedp},
		{"playwright", agentbrowser.BackendPlaywright},
	}
}

// TestBackend_LaunchAndClose tests browser lifecycle for all backends
func TestBackend_LaunchAndClose(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			if !browser.IsLaunched() {
				t.Error("expected browser to be launched")
			}

			err = browser.Close()
			if err != nil {
				t.Fatalf("Close() error = %v", err)
			}

			if browser.IsLaunched() {
				t.Error("expected browser to be closed")
			}
		})
	}
}

// TestBackend_Navigate tests navigation for all backends
func TestBackend_Navigate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

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
		})
	}
}

// TestBackend_GetText tests text extraction for all backends
func TestBackend_GetText(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			_, _, err = browser.Navigate("https://example.com", "load")
			if err != nil {
				t.Fatalf("Navigate() error = %v", err)
			}

			text, err := browser.GetText("h1")
			if err != nil {
				t.Fatalf("GetText() error = %v", err)
			}

			if text != "Example Domain" {
				t.Errorf("expected 'Example Domain', got %s", text)
			}
		})
	}
}

// TestBackend_IsVisible tests visibility check for all backends
func TestBackend_IsVisible(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			_, _, err = browser.Navigate("https://example.com", "load")
			if err != nil {
				t.Fatalf("Navigate() error = %v", err)
			}

			visible, err := browser.IsVisible("h1")
			if err != nil {
				t.Fatalf("IsVisible() error = %v", err)
			}

			if !visible {
				t.Error("expected h1 to be visible")
			}
		})
	}
}

// TestBackend_Count tests element counting for all backends
func TestBackend_Count(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			_, _, err = browser.Navigate("https://example.com", "load")
			if err != nil {
				t.Fatalf("Navigate() error = %v", err)
			}

			count, err := browser.Count("p")
			if err != nil {
				t.Fatalf("Count() error = %v", err)
			}

			if count == 0 {
				t.Error("expected at least one paragraph")
			}
		})
	}
}

// TestBackend_Screenshot tests screenshot functionality for all backends
func TestBackend_Screenshot(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			_, _, err = browser.Navigate("https://example.com", "load")
			if err != nil {
				t.Fatalf("Navigate() error = %v", err)
			}

			buf, err := browser.Screenshot(false, "", 80)
			if err != nil {
				t.Fatalf("Screenshot() error = %v", err)
			}

			if len(buf) == 0 {
				t.Error("expected screenshot buffer to have data")
			}
		})
	}
}

// TestBackend_Evaluate tests JavaScript evaluation for all backends
func TestBackend_Evaluate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
			defer browser.Close()

			err := browser.Launch(agentbrowser.LaunchOptions{Headless: true})
			if err != nil {
				t.Fatalf("Launch() error = %v", err)
			}

			_, _, err = browser.Navigate("https://example.com", "load")
			if err != nil {
				t.Fatalf("Navigate() error = %v", err)
			}

			result, err := browser.Evaluate("document.title")
			if err != nil {
				t.Fatalf("Evaluate() error = %v", err)
			}

			if result != "Example Domain" {
				t.Errorf("expected 'Example Domain', got %v", result)
			}
		})
	}
}

// TestBackend_Tabs tests tab management for all backends
func TestBackend_Tabs(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	for _, tt := range testBackends() {
		t.Run(tt.name, func(t *testing.T) {
			browser := agentbrowser.NewBrowserManagerWithBackend(tt.backend)
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
		})
	}
}
