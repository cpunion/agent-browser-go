// Command agent-browser-go provides a CLI for browser automation.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	agentbrowser "github.com/cpunion/agent-browser-go"
	"github.com/sevlyar/go-daemon"
)

var version = "0.1.0"

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		printHelp()
		os.Exit(0)
	}

	// Parse global flags
	session := "default"
	jsonMode := false
	headed := false
	backend := "chromedp"
	backendSpecified := false
	userDataDir := os.Getenv("AGENT_BROWSER_USER_DATA_DIR") // Default from env
	locale := os.Getenv("AGENT_BROWSER_LOCALE")             // Default from env
	var remainingArgs []string

	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--session" || arg == "-s":
			if i+1 < len(args) {
				session = args[i+1]
				i++
			}
		case arg == "--json":
			jsonMode = true
		case arg == "--headed" || arg == "--head":
			headed = true
		case arg == "--backend" || arg == "-b":
			if i+1 < len(args) {
				backend = args[i+1]
				backendSpecified = true
				i++
			}
		case arg == "--user-data-dir" || arg == "--profile":
			if i+1 < len(args) {
				userDataDir = args[i+1]
				i++
			}
		case arg == "--locale" || arg == "-l":
			if i+1 < len(args) {
				locale = args[i+1]
				i++
			}
		case arg == "--help" || arg == "-h":
			if len(remainingArgs) == 0 {
				printHelp()
			} else {
				printCommandHelp(remainingArgs[0])
			}
			return
		case arg == "--version" || arg == "-v":
			fmt.Println(version)
			return
		case strings.HasPrefix(arg, "-"):
			remainingArgs = append(remainingArgs, arg)
		default:
			remainingArgs = append(remainingArgs, arg)
		}
	}

	// Check for session from env
	if envSession := os.Getenv("AGENT_BROWSER_SESSION"); envSession != "" && session == "default" {
		session = envSession
	}

	// Check for backend from env (only if not set via CLI)
	if !backendSpecified && os.Getenv("AGENT_BROWSER_BACKEND") != "" {
		backend = os.Getenv("AGENT_BROWSER_BACKEND")
		backendSpecified = true
	}

	// Only load saved backend if user didn't specify one
	if !backendSpecified {
		savedBackend := agentbrowser.GetSessionBackend(session)
		if savedBackend != "" {
			backend = savedBackend
		}
	}

	if len(remainingArgs) == 0 {
		printHelp()
		os.Exit(0)
	}

	// Handle commands
	command := remainingArgs[0]
	cmdArgs := remainingArgs[1:]

	switch command {
	case "install":
		handleInstall(cmdArgs)
		return
	case "session":
		handleSession(cmdArgs, session)
		return
	case "daemon":
		if len(cmdArgs) > 0 && cmdArgs[0] == "stop" {
			handleDaemonStop(cmdArgs[1:], session)
			return
		}
		handleDaemon(session, backend, userDataDir, locale)
		return
	case "help":
		if len(cmdArgs) > 0 {
			printCommandHelp(cmdArgs[0])
		} else {
			printHelp()
		}
		return
	}

	// Check if we need to restart daemon (only for certain parameter changes)
	if agentbrowser.IsDaemonRunning(session) {
		needsRestart := false
		savedBackend := agentbrowser.GetSessionBackend(session)
		savedUserDataDir := agentbrowser.GetSessionUserDataDir(session)
		if backendSpecified && savedBackend != backend {
			needsRestart = true
		}
		if userDataDir != "" && savedUserDataDir != userDataDir {
			needsRestart = true
		}

		// Only check headed mode change for open/launch commands
		// Other commands (snapshot, click, etc.) should ignore --headed flag
		isLaunchCommand := command == "open" || command == "launch"
		if isLaunchCommand {
			savedHeaded := agentbrowser.GetSessionHeaded(session)
			if headed != savedHeaded {
				needsRestart = true
			}
		}

		if needsRestart {
			_ = agentbrowser.StopDaemon(session) // Ignore error, just try to start new daemon
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Ensure daemon is running
	if !agentbrowser.IsDaemonRunning(session) {
		// Save backend, headed preference, and userDataDir for this session
		if err := agentbrowser.SaveSessionBackend(session, backend); err != nil {
			printError(jsonMode, "Failed to save backend: "+err.Error())
		}
		if err := agentbrowser.SaveSessionHeaded(session, headed); err != nil {
			printError(jsonMode, "Failed to save headed preference: "+err.Error())
		}
		if err := agentbrowser.SaveSessionUserDataDir(session, userDataDir); err != nil {
			printError(jsonMode, "Failed to save userDataDir: "+err.Error())
		}
		if err := startDaemon(session, backend, userDataDir, locale); err != nil {
			printError(jsonMode, "Failed to start daemon: "+err.Error())
			os.Exit(1)
		}
		// Wait a moment for daemon to start
		time.Sleep(200 * time.Millisecond)
	}

	// Connect to daemon
	client := agentbrowser.NewClient(session)
	if err := client.Connect(); err != nil {
		printError(jsonMode, "Failed to connect to daemon: "+err.Error())
		os.Exit(1)
	}
	defer client.Close()

	// Special handling for open command - just navigate, daemon will auto-launch browser
	if command == "open" || command == "goto" {
		if len(cmdArgs) < 1 {
			printError(jsonMode, "open requires a URL")
			os.Exit(1)
		}
		url := cmdArgs[0]

		// Send navigate command - daemon will auto-launch browser with correct settings
		navCmd := &agentbrowser.NavigateCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: genID(), Action: "navigate"},
			URL:         url,
		}
		resp, err := client.Send(navCmd)
		if err != nil {
			printError(jsonMode, "Failed to navigate: "+err.Error())
			os.Exit(1)
		}
		printResponse(resp, jsonMode)
		if !resp.Success {
			os.Exit(1)
		}
		return
	}

	// Build command
	cmd, err := buildCommand(command, cmdArgs, headed)
	if err != nil {
		printError(jsonMode, err.Error())
		os.Exit(1)
	}

	// Send command
	resp, err := client.Send(cmd)
	if err != nil {
		printError(jsonMode, "Failed to send command: "+err.Error())
		os.Exit(1)
	}

	// Print response
	printResponse(resp, jsonMode)

	if !resp.Success {
		os.Exit(1)
	}
}

func buildCommand(command string, args []string, headed bool) (agentbrowser.Command, error) {
	id := genID()

	switch command {
	// Navigate command (when called directly, not via open)
	case "navigate":
		if len(args) < 1 {
			return nil, fmt.Errorf("navigate requires a URL")
		}
		return &agentbrowser.NavigateCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "navigate"},
			URL:         args[0],
		}, nil

	case "click":
		if len(args) < 1 {
			return nil, fmt.Errorf("click requires a selector")
		}
		return &agentbrowser.ClickCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "click"},
			Selector:    args[0],
		}, nil

	case "dblclick":
		if len(args) < 1 {
			return nil, fmt.Errorf("dblclick requires a selector")
		}
		return &agentbrowser.DoubleClickCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "dblclick"},
			Selector:    args[0],
		}, nil

	case "type":
		if len(args) < 2 {
			return nil, fmt.Errorf("type requires selector and text")
		}
		return &agentbrowser.TypeCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "type"},
			Selector:    args[0],
			Text:        args[1],
		}, nil

	case "fill":
		if len(args) < 2 {
			return nil, fmt.Errorf("fill requires selector and value")
		}
		return &agentbrowser.FillCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "fill"},
			Selector:    args[0],
			Value:       args[1],
		}, nil

	case "press", "key":
		if len(args) < 1 {
			return nil, fmt.Errorf("press requires a key")
		}
		var selector string
		if len(args) > 1 {
			selector = args[1]
		}
		return &agentbrowser.PressCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "press"},
			Key:         args[0],
			Selector:    selector,
		}, nil

	case "hover":
		if len(args) < 1 {
			return nil, fmt.Errorf("hover requires a selector")
		}
		return &agentbrowser.HoverCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "hover"},
			Selector:    args[0],
		}, nil

	case "focus":
		if len(args) < 1 {
			return nil, fmt.Errorf("focus requires a selector")
		}
		return &agentbrowser.FocusCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "focus"},
			Selector:    args[0],
		}, nil

	case "check":
		if len(args) < 1 {
			return nil, fmt.Errorf("check requires a selector")
		}
		return &agentbrowser.CheckCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "check"},
			Selector:    args[0],
		}, nil

	case "uncheck":
		if len(args) < 1 {
			return nil, fmt.Errorf("uncheck requires a selector")
		}
		return &agentbrowser.UncheckCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "uncheck"},
			Selector:    args[0],
		}, nil

	case "screenshot":
		var path string
		fullPage := false
		for i, arg := range args {
			if arg == "--full" || arg == "-f" {
				fullPage = true
			} else if !strings.HasPrefix(arg, "-") && path == "" {
				path = args[i]
			}
		}
		return &agentbrowser.ScreenshotCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "screenshot"},
			Path:        path,
			FullPage:    fullPage,
		}, nil

	case "snapshot":
		interactive := false
		compact := false
		var maxDepth int
		var selector string
		for i := 0; i < len(args); i++ {
			switch args[i] {
			case "-i", "--interactive":
				interactive = true
			case "-c", "--compact":
				compact = true
			case "-d", "--depth":
				if i+1 < len(args) {
					maxDepth, _ = strconv.Atoi(args[i+1])
					i++
				}
			case "-s", "--selector":
				if i+1 < len(args) {
					selector = args[i+1]
					i++
				}
			}
		}
		return &agentbrowser.SnapshotCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "snapshot"},
			Interactive: interactive,
			Compact:     compact,
			MaxDepth:    maxDepth,
			Selector:    selector,
		}, nil

	case "eval":
		if len(args) < 1 {
			return nil, fmt.Errorf("eval requires a script")
		}
		return &agentbrowser.EvaluateCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "evaluate"},
			Script:      args[0],
		}, nil

	case "wait":
		if len(args) < 1 {
			return nil, fmt.Errorf("wait requires a selector or timeout")
		}
		// Check if it's a number (timeout in ms)
		if timeout, err := strconv.Atoi(args[0]); err == nil {
			return &agentbrowser.WaitCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "wait"},
				Timeout:     timeout,
			}, nil
		}
		return &agentbrowser.WaitCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "wait"},
			Selector:    args[0],
		}, nil

	case "scroll":
		direction := "down"
		amount := 100
		if len(args) > 0 {
			direction = args[0]
		}
		if len(args) > 1 {
			amount, _ = strconv.Atoi(args[1])
		}
		return &agentbrowser.ScrollCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "scroll"},
			Direction:   direction,
			Amount:      amount,
		}, nil

	case "scrollintoview", "scrollinto":
		if len(args) < 1 {
			return nil, fmt.Errorf("scrollintoview requires a selector")
		}
		return &agentbrowser.ScrollIntoViewCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "scrollintoview"},
			Selector:    args[0],
		}, nil

	case "back":
		return &agentbrowser.BackCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "back"},
		}, nil

	case "forward":
		return &agentbrowser.ForwardCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "forward"},
		}, nil

	case "reload":
		return &agentbrowser.ReloadCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "reload"},
		}, nil

	case "close", "quit", "exit":
		return &agentbrowser.CloseCommand{
			BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "close"},
		}, nil

	// Get subcommands
	case "get":
		if len(args) < 1 {
			return nil, fmt.Errorf("get requires a subcommand (text, html, value, attr, title, url, count, box)")
		}
		subcmd := args[0]
		subArgs := args[1:]

		switch subcmd {
		case "text":
			if len(subArgs) < 1 {
				return nil, fmt.Errorf("get text requires a selector")
			}
			return &agentbrowser.GetTextCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "gettext"},
				Selector:    subArgs[0],
			}, nil
		case "html":
			if len(subArgs) < 1 {
				return nil, fmt.Errorf("get html requires a selector")
			}
			return &agentbrowser.InnerHTMLCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "innerhtml"},
				Selector:    subArgs[0],
			}, nil
		case "value":
			if len(subArgs) < 1 {
				return nil, fmt.Errorf("get value requires a selector")
			}
			return &agentbrowser.InputValueCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "inputvalue"},
				Selector:    subArgs[0],
			}, nil
		case "attr":
			if len(subArgs) < 2 {
				return nil, fmt.Errorf("get attr requires selector and attribute name")
			}
			return &agentbrowser.GetAttributeCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "getattribute"},
				Selector:    subArgs[0],
				Attribute:   subArgs[1],
			}, nil
		case "title":
			return &agentbrowser.TitleCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "title"},
			}, nil
		case "url":
			return &agentbrowser.URLCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "url"},
			}, nil
		case "count":
			if len(subArgs) < 1 {
				return nil, fmt.Errorf("get count requires a selector")
			}
			return &agentbrowser.CountCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "count"},
				Selector:    subArgs[0],
			}, nil
		case "box":
			if len(subArgs) < 1 {
				return nil, fmt.Errorf("get box requires a selector")
			}
			return &agentbrowser.BoundingBoxCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "boundingbox"},
				Selector:    subArgs[0],
			}, nil
		default:
			return nil, fmt.Errorf("unknown get subcommand: %s", subcmd)
		}

	// Is subcommands
	case "is":
		if len(args) < 2 {
			return nil, fmt.Errorf("is requires subcommand and selector")
		}
		subcmd := args[0]
		selector := args[1]

		switch subcmd {
		case "visible":
			return &agentbrowser.IsVisibleCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "isvisible"},
				Selector:    selector,
			}, nil
		case "enabled":
			return &agentbrowser.IsEnabledCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "isenabled"},
				Selector:    selector,
			}, nil
		case "checked":
			return &agentbrowser.IsCheckedCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "ischecked"},
				Selector:    selector,
			}, nil
		default:
			return nil, fmt.Errorf("unknown is subcommand: %s", subcmd)
		}

	// Tab commands
	case "tab":
		if len(args) == 0 {
			return &agentbrowser.TabListCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "tab_list"},
			}, nil
		}

		subcmd := args[0]
		switch subcmd {
		case "new":
			var url string
			if len(args) > 1 {
				url = args[1]
			}
			return &agentbrowser.TabNewCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "tab_new"},
				URL:         url,
			}, nil
		case "close":
			var index *int
			if len(args) > 1 {
				i, _ := strconv.Atoi(args[1])
				index = &i
			}
			return &agentbrowser.TabCloseCommand{
				BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "tab_close"},
				Index:       index,
			}, nil
		default:
			// Try as tab index
			if i, err := strconv.Atoi(subcmd); err == nil {
				return &agentbrowser.TabSwitchCommand{
					BaseCommand: agentbrowser.BaseCommand{ID: id, Action: "tab_switch"},
					Index:       i,
				}, nil
			}
			return nil, fmt.Errorf("unknown tab subcommand: %s", subcmd)
		}

	default:
		return nil, fmt.Errorf("unknown command: %s", command)
	}
}

func genID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func printError(jsonMode bool, msg string) {
	if jsonMode {
		resp := agentbrowser.ErrorResponse("", msg)
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
	} else {
		fmt.Fprintf(os.Stderr, "Error: %s\n", msg)
	}
}

func printResponse(resp agentbrowser.Response, jsonMode bool) {
	if jsonMode {
		data, _ := json.Marshal(resp)
		fmt.Println(string(data))
		return
	}

	if !resp.Success {
		fmt.Fprintf(os.Stderr, "Error: %s\n", resp.Error)
		return
	}

	if len(resp.Data) == 0 || string(resp.Data) == "null" {
		fmt.Println("OK")
		return
	}

	// Try to pretty print the data
	var data interface{}
	if err := json.Unmarshal(resp.Data, &data); err == nil {
		switch v := data.(type) {
		case map[string]interface{}:
			// Handle specific response types
			if snapshot, ok := v["snapshot"]; ok {
				fmt.Println(snapshot)
				return
			}
			if text, ok := v["text"]; ok {
				fmt.Println(text)
				return
			}
			if html, ok := v["html"]; ok {
				fmt.Println(html)
				return
			}
			if value, ok := v["value"]; ok {
				fmt.Println(value)
				return
			}
			if url, ok := v["url"]; ok {
				fmt.Println(url)
				return
			}
			if title, ok := v["title"]; ok {
				fmt.Println(title)
				return
			}
			// Default: print as JSON
			prettyData, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(prettyData))
		case bool:
			if v {
				fmt.Println("true")
			} else {
				fmt.Println("false")
			}
		default:
			prettyData, _ := json.MarshalIndent(data, "", "  ")
			fmt.Println(string(prettyData))
		}
	} else {
		fmt.Println(string(resp.Data))
	}
}

func startDaemon(session string, backend string, userDataDir string, locale string) error {
	// Get executable path
	exe, err := os.Executable()
	if err != nil {
		return err
	}

	// Build daemon command with arguments
	args := []string{"daemon", "--session", session, "--backend", backend}
	if userDataDir != "" {
		args = append(args, "--user-data-dir", userDataDir)
	}
	if locale != "" {
		args = append(args, "--locale", locale)
	}

	// Start daemon in background
	cmd := exec.Command(exe, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	if err := cmd.Start(); err != nil {
		return err
	}

	// Detach from parent
	if err := cmd.Process.Release(); err != nil {
		return err
	}

	return nil
}

func handleDaemon(session string, backend string, userDataDir string, locale string) {
	// Use go-daemon library for proper daemonization
	// Note: LogFileName is required for stdout/stderr to work properly
	// Without it, chromedp headed mode fails because Chrome's output is lost
	logFile := agentbrowser.GetLogFile(session)
	ctx := &daemon.Context{
		PidFileName: agentbrowser.GetPIDFile(session),
		PidFilePerm: 0644,
		LogFileName: logFile,
		LogFilePerm: 0640,
		Umask:       027,
		Args:        os.Args, // Explicitly pass command line args to child
	}

	// Reborn creates a child process and returns in the child
	child, err := ctx.Reborn()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to daemonize: %v\n", err)
		os.Exit(1)
	}

	if child != nil {
		// Parent process - just exit
		return
	}
	defer func() { _ = ctx.Release() }()

	// Child process - re-parse args since we're a new process
	// The function parameters are from parent, not from our actual os.Args
	childSession := session
	childBackend := backend
	childUserDataDir := userDataDir
	childLocale := locale

	for i := 1; i < len(os.Args); i++ {
		arg := os.Args[i]
		switch {
		case arg == "--session" || arg == "-s":
			if i+1 < len(os.Args) {
				childSession = os.Args[i+1]
				i++
			}
		case arg == "--backend" || arg == "-b":
			if i+1 < len(os.Args) {
				childBackend = os.Args[i+1]
				i++
			}
		case arg == "--user-data-dir" || arg == "--profile":
			if i+1 < len(os.Args) {
				childUserDataDir = os.Args[i+1]
				i++
			}
		case arg == "--locale" || arg == "-l":
			if i+1 < len(os.Args) {
				childLocale = os.Args[i+1]
				i++
			}
		}
	}

	// Child process - run the daemon
	d := agentbrowser.NewDaemonFull(childSession, childBackend, childUserDataDir, childLocale)
	if err := d.Start(); err != nil {
		// Can't write to stderr in daemon, so just exit
		os.Exit(1)
	}
	d.Wait()
}

func handleDaemonStop(args []string, currentSession string) {
	stopAll := false
	var targetSession string

	// Parse args
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--all", "-a":
			stopAll = true
		case "--session", "-s":
			if i+1 < len(args) {
				targetSession = args[i+1]
				i++
			}
		}
	}

	if stopAll {
		// Stop all daemons
		sessions, err := agentbrowser.ListRunningSessions()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to list sessions: %v\n", err)
			os.Exit(1)
		}

		if len(sessions) == 0 {
			fmt.Println("No running daemons found")
			return
		}

		fmt.Printf("Stopping %d daemon(s)...\n", len(sessions))
		for _, session := range sessions {
			fmt.Printf("  Stopping %s...", session)
			if err := agentbrowser.StopDaemon(session); err != nil {
				fmt.Printf(" failed: %v\n", err)
			} else {
				fmt.Println(" done")
			}
		}
	} else {
		// Stop specific session
		if targetSession == "" {
			targetSession = currentSession
		}

		if !agentbrowser.IsDaemonRunning(targetSession) {
			fmt.Printf("Daemon not running for session: %s\n", targetSession)
			os.Exit(1)
		}

		fmt.Printf("Stopping daemon for session: %s...", targetSession)
		if err := agentbrowser.StopDaemon(targetSession); err != nil {
			fmt.Printf(" failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(" done")
	}
}

func handleSession(args []string, session string) {
	if len(args) == 0 {
		fmt.Println(session)
		return
	}

	switch args[0] {
	case "list":
		// List all sessions by finding socket/port files
		dir := filepath.Join(os.TempDir(), "agent-browser-go")
		entries, err := os.ReadDir(dir)
		if err != nil {
			fmt.Println("No active sessions")
			return
		}

		fmt.Println("Active sessions:")
		for _, entry := range entries {
			name := entry.Name()
			if strings.HasSuffix(name, ".pid") {
				sessionName := strings.TrimSuffix(name, ".pid")
				if agentbrowser.IsDaemonRunning(sessionName) {
					marker := "  "
					if sessionName == session {
						marker = "->"
					}
					fmt.Printf("%s %s\n", marker, sessionName)
				}
			}
		}
	default:
		fmt.Printf("Unknown session command: %s\n", args[0])
	}
}

func handleInstall(args []string) {
	// Parse --backend flag
	backend := "all"
	withDeps := false

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--backend", "-b":
			if i+1 < len(args) {
				backend = args[i+1]
				i++
			}
		case "--with-deps":
			withDeps = true
		}
	}

	switch backend {
	case "chromedp":
		installChromedp()
	case "playwright":
		installPlaywright(withDeps)
	case "all":
		installChromedp()
		installPlaywright(withDeps)
	default:
		fmt.Fprintf(os.Stderr, "Unknown backend: %s\n", backend)
		os.Exit(1)
	}
}

func installChromedp() {
	fmt.Println("=== chromedp ===")
	fmt.Println("chromedp uses an existing Chrome/Chromium installation.")
	fmt.Println("Please ensure Chrome or Chromium is installed on your system.")
	fmt.Println("  Chrome: https://www.google.com/chrome/")
	fmt.Println("  Chromium: https://www.chromium.org/getting-involved/download-chromium/")
	fmt.Println("")
}

func installPlaywright(withDeps bool) {
	fmt.Println("=== playwright ===")
	fmt.Println("Installing Playwright browser driver...")

	// Use playwright CLI to install
	args := []string{"install", "chromium"}
	if withDeps {
		args = append(args, "--with-deps")
	}

	cmd := exec.Command("npx", append([]string{"-y", "playwright@latest"}, args...)...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to install playwright: %v\n", err)
		fmt.Println("\nManual installation:")
		fmt.Println("  npx -y playwright@latest install chromium")
		if withDeps {
			fmt.Println("  npx -y playwright@latest install-deps chromium")
		}
		os.Exit(1)
	}

	fmt.Println("Playwright installed successfully!")
	fmt.Println("")
}

func printHelp() {
	fmt.Printf(`agent-browser-go v%s - Headless browser automation CLI for AI agents

Usage: agent-browser-go [options] <command> [arguments]

Options:
  --session, -s <name>  Use isolated session (default: "default")
  --json               JSON output (for agents)
  --headed, --head     Show browser window
  --backend, -b <type> Browser backend: chromedp (default) or playwright
  --help, -h           Show help
  --version, -v        Show version

Environment Variables:
  AGENT_BROWSER_SESSION  Default session name
  AGENT_BROWSER_BACKEND  Default backend (chromedp or playwright)

Core Commands:
  open <url>              Navigate to URL (aliases: goto, navigate)
  click <sel>             Click element
  dblclick <sel>          Double-click element
  type <sel> <text>       Type into element
  fill <sel> <text>       Clear and fill
  press <key>             Press key (Enter, Tab, Control+a)
  hover <sel>             Hover element
  focus <sel>             Focus element
  check <sel>             Check checkbox
  uncheck <sel>           Uncheck checkbox
  screenshot [path]       Take screenshot (--full for full page)
  snapshot                Accessibility tree with refs
  eval <js>               Run JavaScript
  wait <sel|ms>           Wait for element or time
  scroll <dir> [px]       Scroll (up/down/left/right)
  back                    Go back
  forward                 Go forward
  reload                  Reload page
  close                   Close browser (aliases: quit, exit)

Get Info:
  get text <sel>          Get text content
  get html <sel>          Get innerHTML
  get value <sel>         Get input value
  get attr <sel> <name>   Get attribute
  get title               Get page title
  get url                 Get current URL
  get count <sel>         Count matching elements
  get box <sel>           Get bounding box

Check State:
  is visible <sel>        Check if visible
  is enabled <sel>        Check if enabled
  is checked <sel>        Check if checked

Tabs:
  tab                     List tabs
  tab new [url]           New tab
  tab <n>                 Switch to tab n
  tab close [n]           Close tab

Session:
  session                 Show current session
  session list            List active sessions

Selectors:
  @e1, @e2, ...           Ref from snapshot (recommended for AI)
  #id                     CSS ID selector
  .class                  CSS class selector
  text=Submit             Text selector

Examples:
  agent-browser-go open https://example.com
  agent-browser-go snapshot -i
  agent-browser-go click @e2
  agent-browser-go fill @e3 "test@example.com"
  agent-browser-go screenshot page.png
  agent-browser-go close
`, version)
}

func printCommandHelp(command string) {
	switch command {
	case "snapshot":
		fmt.Println(`snapshot - Get accessibility tree with element refs

Usage: agent-browser-go snapshot [options]

Options:
  -i, --interactive    Only show interactive elements
  -c, --compact        Remove empty structural elements
  -d, --depth <n>      Limit tree depth
  -s, --selector <sel> Scope to CSS selector

Output includes refs like [ref=e1] that can be used with other commands.

Examples:
  agent-browser-go snapshot
  agent-browser-go snapshot -i
  agent-browser-go snapshot -i -c -d 3`)
	default:
		fmt.Printf("No detailed help for: %s\n", command)
		fmt.Println("Use 'agent-browser-go --help' for general help.")
	}
}
