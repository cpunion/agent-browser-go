package agentbrowser

import (
	"bufio"
	"crypto/md5"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"syscall"
)

// Daemon manages the browser server.
type Daemon struct {
	session     string
	browser     *BrowserManager
	listener    net.Listener
	connections sync.WaitGroup
	shutdown    chan struct{}
	mu          sync.Mutex
}

// NewDaemon creates a new daemon instance.
func NewDaemon(session string) *Daemon {
	return &Daemon{
		session:  session,
		browser:  NewBrowserManager(),
		shutdown: make(chan struct{}),
	}
}

// GetSocketPath returns the socket path for a session.
func GetSocketPath(session string) string {
	if runtime.GOOS == "windows" {
		return "" // Windows uses TCP
	}

	dir := filepath.Join(os.TempDir(), "agent-browser-go")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, fmt.Sprintf("%s.sock", session))
}

// GetPortForSession returns a port number for a session (Windows).
func GetPortForSession(session string) int {
	hash := md5.Sum([]byte(session))
	port := binary.BigEndian.Uint16(hash[:2])
	// Use ports in range 49152-65535
	return 49152 + int(port)%(65535-49152)
}

// GetPIDFile returns the PID file path for a session.
func GetPIDFile(session string) string {
	dir := filepath.Join(os.TempDir(), "agent-browser-go")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, fmt.Sprintf("%s.pid", session))
}

// GetPortFile returns the port file path for a session (Windows).
func GetPortFile(session string) string {
	dir := filepath.Join(os.TempDir(), "agent-browser-go")
	os.MkdirAll(dir, 0755)
	return filepath.Join(dir, fmt.Sprintf("%s.port", session))
}

// IsDaemonRunning checks if a daemon is running for the session.
func IsDaemonRunning(session string) bool {
	pidFile := GetPIDFile(session)
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return false
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return false
	}

	// Check if process exists
	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// On Unix, FindProcess always succeeds, so we need to signal check
	if runtime.GOOS != "windows" {
		err = process.Signal(syscall.Signal(0))
		return err == nil
	}

	return true
}

// Start starts the daemon server.
func (d *Daemon) Start() error {
	var err error

	if runtime.GOOS == "windows" {
		// Use TCP on Windows
		port := GetPortForSession(d.session)
		d.listener, err = net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", port))
		if err != nil {
			return fmt.Errorf("failed to listen on port %d: %w", port, err)
		}

		// Write port file
		portFile := GetPortFile(d.session)
		if err := os.WriteFile(portFile, []byte(strconv.Itoa(port)), 0644); err != nil {
			d.listener.Close()
			return fmt.Errorf("failed to write port file: %w", err)
		}
	} else {
		// Use Unix socket on Unix-like systems
		socketPath := GetSocketPath(d.session)

		// Remove existing socket
		os.Remove(socketPath)

		d.listener, err = net.Listen("unix", socketPath)
		if err != nil {
			return fmt.Errorf("failed to listen on socket %s: %w", socketPath, err)
		}
	}

	// Write PID file
	pidFile := GetPIDFile(d.session)
	if err := os.WriteFile(pidFile, []byte(strconv.Itoa(os.Getpid())), 0644); err != nil {
		d.listener.Close()
		return fmt.Errorf("failed to write PID file: %w", err)
	}

	// Handle shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		d.Stop()
	}()

	// Accept connections
	go d.acceptLoop()

	return nil
}

// acceptLoop accepts incoming connections.
func (d *Daemon) acceptLoop() {
	for {
		select {
		case <-d.shutdown:
			return
		default:
		}

		conn, err := d.listener.Accept()
		if err != nil {
			select {
			case <-d.shutdown:
				return
			default:
				continue
			}
		}

		d.connections.Add(1)
		go d.handleConnection(conn)
	}
}

// handleConnection handles a single connection.
func (d *Daemon) handleConnection(conn net.Conn) {
	defer d.connections.Done()
	defer conn.Close()

	reader := bufio.NewReader(conn)

	for {
		// Read line (command is JSON terminated by newline)
		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err != io.EOF {
				// Log error but don't crash
			}
			return
		}

		// Parse and execute command
		cmd, err := ParseCommand(line)
		if err != nil {
			resp := ErrorResponse("", err.Error())
			d.writeResponse(conn, resp)
			continue
		}

		// Ensure browser is launched for most commands
		action := cmd.GetAction()
		if action != "launch" && action != "close" && !d.browser.IsLaunched() {
			// Auto-launch with defaults
			d.browser.Launch(LaunchOptions{Headless: true})
		}

		// Execute command
		resp := ExecuteCommand(cmd, d.browser)
		d.writeResponse(conn, resp)

		// Close connection after close command
		if action == "close" {
			return
		}
	}
}

// writeResponse writes a response to the connection.
func (d *Daemon) writeResponse(conn net.Conn, resp Response) {
	data, err := SerializeResponse(resp)
	if err != nil {
		data = []byte(fmt.Sprintf(`{"id":"","success":false,"error":"failed to serialize response: %s"}`, err.Error()))
	}
	data = append(data, '\n')
	conn.Write(data)
}

// Stop stops the daemon.
func (d *Daemon) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()

	select {
	case <-d.shutdown:
		// Already stopped
		return
	default:
		close(d.shutdown)
	}

	// Close listener
	if d.listener != nil {
		d.listener.Close()
	}

	// Wait for connections to finish
	d.connections.Wait()

	// Close browser
	d.browser.Close()

	// Cleanup files
	d.cleanup()
}

// cleanup removes socket/port/PID files.
func (d *Daemon) cleanup() {
	os.Remove(GetPIDFile(d.session))

	if runtime.GOOS == "windows" {
		os.Remove(GetPortFile(d.session))
	} else {
		os.Remove(GetSocketPath(d.session))
	}
}

// Wait waits for the daemon to stop.
func (d *Daemon) Wait() {
	<-d.shutdown
	d.connections.Wait()
}

// Client connects to a running daemon.
type Client struct {
	session string
	conn    net.Conn
}

// NewClient creates a new client.
func NewClient(session string) *Client {
	return &Client{session: session}
}

// Connect connects to the daemon.
func (c *Client) Connect() error {
	var err error

	if runtime.GOOS == "windows" {
		// Read port from file
		portFile := GetPortFile(c.session)
		data, err := os.ReadFile(portFile)
		if err != nil {
			return fmt.Errorf("daemon not running (no port file)")
		}
		port, err := strconv.Atoi(string(data))
		if err != nil {
			return fmt.Errorf("invalid port file")
		}
		c.conn, err = net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", port))
	} else {
		socketPath := GetSocketPath(c.session)
		c.conn, err = net.Dial("unix", socketPath)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to daemon: %w", err)
	}

	return nil
}

// Send sends a command and receives the response.
func (c *Client) Send(cmd Command) (Response, error) {
	data, err := SerializeCommand(cmd)
	if err != nil {
		return Response{}, fmt.Errorf("failed to serialize command: %w", err)
	}
	data = append(data, '\n')

	if _, err := c.conn.Write(data); err != nil {
		return Response{}, fmt.Errorf("failed to send command: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	respData, err := reader.ReadBytes('\n')
	if err != nil {
		return Response{}, fmt.Errorf("failed to read response: %w", err)
	}

	var resp Response
	if err := json.Unmarshal(respData, &resp); err != nil {
		return Response{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return resp, nil
}

// SendRaw sends raw JSON and receives raw JSON response.
func (c *Client) SendRaw(data []byte) ([]byte, error) {
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	if _, err := c.conn.Write(data); err != nil {
		return nil, fmt.Errorf("failed to send: %w", err)
	}

	reader := bufio.NewReader(c.conn)
	return reader.ReadBytes('\n')
}

// Close closes the client connection.
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// EnsureDaemon ensures a daemon is running for the session.
func EnsureDaemon(session string) error {
	if IsDaemonRunning(session) {
		return nil
	}

	// Start daemon in background - this would typically be done
	// by the CLI spawning a subprocess
	return fmt.Errorf("daemon not running for session %s", session)
}
