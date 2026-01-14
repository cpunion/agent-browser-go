package agentbrowser_test

import (
	"encoding/json"
	"testing"

	agentbrowser "github.com/cpunion/agent-browser-go"
)

// TestParseCommand_Navigation tests navigation command parsing
func TestParseCommand_Navigation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
		check   func(*testing.T, agentbrowser.Command)
	}{
		{
			name:    "navigate with URL",
			input:   `{"id":"1","action":"navigate","url":"https://example.com"}`,
			wantErr: false,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				navCmd, ok := cmd.(*agentbrowser.NavigateCommand)
				if !ok {
					t.Fatal("expected NavigateCommand")
				}
				if navCmd.URL != "https://example.com" {
					t.Errorf("expected URL https://example.com, got %s", navCmd.URL)
				}
			},
		},
		{
			name:    "navigate without URL",
			input:   `{"id":"1","action":"navigate"}`,
			wantErr: true,
		},
		{
			name:    "back command",
			input:   `{"id":"1","action":"back"}`,
			wantErr: false,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				_, ok := cmd.(*agentbrowser.BackCommand)
				if !ok {
					t.Fatal("expected BackCommand")
				}
			},
		},
		{
			name:    "forward command",
			input:   `{"id":"1","action":"forward"}`,
			wantErr: false,
		},
		{
			name:    "reload command",
			input:   `{"id":"1","action":"reload"}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := agentbrowser.ParseCommand([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && tt.check != nil {
				tt.check(t, cmd)
			}
		})
	}
}

// TestParseCommand_Click tests click command parsing
func TestParseCommand_Click(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantErr  bool
		selector string
	}{
		{
			name:     "click with selector",
			input:    `{"id":"1","action":"click","selector":"#btn"}`,
			wantErr:  false,
			selector: "#btn",
		},
		{
			name:    "click without selector",
			input:   `{"id":"1","action":"click"}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := agentbrowser.ParseCommand([]byte(tt.input))
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCommand() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil {
				clickCmd, ok := cmd.(*agentbrowser.ClickCommand)
				if !ok {
					t.Fatal("expected ClickCommand")
				}
				if clickCmd.Selector != tt.selector {
					t.Errorf("expected selector %s, got %s", tt.selector, clickCmd.Selector)
				}
			}
		})
	}
}

// TestParseCommand_Type tests type command parsing
func TestParseCommand_Type(t *testing.T) {
	input := `{"id":"1","action":"type","selector":"#input","text":"hello"}`
	cmd, err := agentbrowser.ParseCommand([]byte(input))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	typeCmd, ok := cmd.(*agentbrowser.TypeCommand)
	if !ok {
		t.Fatal("expected TypeCommand")
	}
	if typeCmd.Selector != "#input" {
		t.Errorf("expected selector #input, got %s", typeCmd.Selector)
	}
	if typeCmd.Text != "hello" {
		t.Errorf("expected text hello, got %s", typeCmd.Text)
	}
}

// TestParseCommand_Fill tests fill command parsing
func TestParseCommand_Fill(t *testing.T) {
	input := `{"id":"1","action":"fill","selector":"#input","value":"hello"}`
	cmd, err := agentbrowser.ParseCommand([]byte(input))
	if err != nil {
		t.Fatalf("ParseCommand() error = %v", err)
	}

	fillCmd, ok := cmd.(*agentbrowser.FillCommand)
	if !ok {
		t.Fatal("expected FillCommand")
	}
	if fillCmd.Value != "hello" {
		t.Errorf("expected value hello, got %s", fillCmd.Value)
	}
}

// TestParseCommand_Screenshot tests screenshot command parsing
func TestParseCommand_Screenshot(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		fullPage bool
	}{
		{
			name:     "screenshot with path",
			input:    `{"id":"1","action":"screenshot","path":"test.png"}`,
			fullPage: false,
		},
		{
			name:     "screenshot with fullPage",
			input:    `{"id":"1","action":"screenshot","fullPage":true}`,
			fullPage: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := agentbrowser.ParseCommand([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseCommand() error = %v", err)
			}

			screenshotCmd, ok := cmd.(*agentbrowser.ScreenshotCommand)
			if !ok {
				t.Fatal("expected ScreenshotCommand")
			}
			if screenshotCmd.FullPage != tt.fullPage {
				t.Errorf("expected fullPage %v, got %v", tt.fullPage, screenshotCmd.FullPage)
			}
		})
	}
}

// TestParseCommand_Snapshot tests snapshot command parsing
func TestParseCommand_Snapshot(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		interactive bool
		compact     bool
		maxDepth    int
	}{
		{
			name:        "basic snapshot",
			input:       `{"id":"1","action":"snapshot"}`,
			interactive: false,
			compact:     false,
			maxDepth:    0,
		},
		{
			name:        "interactive snapshot",
			input:       `{"id":"1","action":"snapshot","interactive":true}`,
			interactive: true,
		},
		{
			name:     "snapshot with maxDepth",
			input:    `{"id":"1","action":"snapshot","maxDepth":3}`,
			maxDepth: 3,
		},
		{
			name:        "snapshot with all options",
			input:       `{"id":"1","action":"snapshot","interactive":true,"compact":true,"maxDepth":5}`,
			interactive: true,
			compact:     true,
			maxDepth:    5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := agentbrowser.ParseCommand([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseCommand() error = %v", err)
			}

			snapshotCmd, ok := cmd.(*agentbrowser.SnapshotCommand)
			if !ok {
				t.Fatal("expected SnapshotCommand")
			}
			if snapshotCmd.Interactive != tt.interactive {
				t.Errorf("expected interactive %v, got %v", tt.interactive, snapshotCmd.Interactive)
			}
			if snapshotCmd.Compact != tt.compact {
				t.Errorf("expected compact %v, got %v", tt.compact, snapshotCmd.Compact)
			}
			if snapshotCmd.MaxDepth != tt.maxDepth {
				t.Errorf("expected maxDepth %d, got %d", tt.maxDepth, snapshotCmd.MaxDepth)
			}
		})
	}
}

// TestParseCommand_Tabs tests tab command parsing
func TestParseCommand_Tabs(t *testing.T) {
	tests := []struct {
		name  string
		input string
		check func(*testing.T, agentbrowser.Command)
	}{
		{
			name:  "tab_new",
			input: `{"id":"1","action":"tab_new"}`,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				_, ok := cmd.(*agentbrowser.TabNewCommand)
				if !ok {
					t.Fatal("expected TabNewCommand")
				}
			},
		},
		{
			name:  "tab_list",
			input: `{"id":"1","action":"tab_list"}`,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				_, ok := cmd.(*agentbrowser.TabListCommand)
				if !ok {
					t.Fatal("expected TabListCommand")
				}
			},
		},
		{
			name:  "tab_switch",
			input: `{"id":"1","action":"tab_switch","index":0}`,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				switchCmd, ok := cmd.(*agentbrowser.TabSwitchCommand)
				if !ok {
					t.Fatal("expected TabSwitchCommand")
				}
				if switchCmd.Index != 0 {
					t.Errorf("expected index 0, got %d", switchCmd.Index)
				}
			},
		},
		{
			name:  "tab_close",
			input: `{"id":"1","action":"tab_close"}`,
			check: func(t *testing.T, cmd agentbrowser.Command) {
				_, ok := cmd.(*agentbrowser.TabCloseCommand)
				if !ok {
					t.Fatal("expected TabCloseCommand")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, err := agentbrowser.ParseCommand([]byte(tt.input))
			if err != nil {
				t.Fatalf("ParseCommand() error = %v", err)
			}
			tt.check(t, cmd)
		})
	}
}

// TestSerializeResponse tests response serialization
func TestSerializeResponse(t *testing.T) {
	tests := []struct {
		name     string
		response agentbrowser.Response
		check    func(*testing.T, []byte)
	}{
		{
			name:     "success response",
			response: agentbrowser.SuccessResponse("1", map[string]string{"url": "https://example.com"}),
			check: func(t *testing.T, data []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(data, &resp); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if resp["success"] != true {
					t.Error("expected success to be true")
				}
				if resp["id"] != "1" {
					t.Errorf("expected id 1, got %v", resp["id"])
				}
			},
		},
		{
			name:     "error response",
			response: agentbrowser.ErrorResponse("2", "test error"),
			check: func(t *testing.T, data []byte) {
				var resp map[string]interface{}
				if err := json.Unmarshal(data, &resp); err != nil {
					t.Fatalf("failed to unmarshal: %v", err)
				}
				if resp["success"] != false {
					t.Error("expected success to be false")
				}
				if resp["error"] == nil {
					t.Error("expected error field")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := agentbrowser.SerializeResponse(tt.response)
			if err != nil {
				t.Fatalf("SerializeResponse() error = %v", err)
			}
			tt.check(t, data)
		})
	}
}
