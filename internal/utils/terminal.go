package utils

import (
	"fmt"
	"io"
	"os"
	gosruntime "runtime"
	"strings"
	"sync"
	"time"

	"github.com/JadlionHD/Enty/internal/config"
	"github.com/aymanbagabas/go-pty"
)

// TerminalSession manages a single PTY terminal session
type TerminalSession struct {
	pty          pty.Pty
	cmd          *pty.Cmd
	mutex        sync.RWMutex // Changed to RWMutex for better performance
	isRunning    bool
	sessionID    string
	terminalType string
	// serviceName removed: PATH is now managed globally based on config
	lastActivity time.Time
	timeoutTimer *time.Timer
	onTimeout    func(sessionID string)
	readBuffer   chan []byte   // Buffered channel for efficient data streaming
	writeBuffer  chan []byte   // Buffered channel for write operations
	stopChannel  chan struct{} // Channel for graceful shutdown
}

// NewTerminalSession creates a new terminal session
func NewTerminalSession() *TerminalSession {
	return &TerminalSession{
		isRunning: false,
	}
}

// TerminalSessionOptions holds options for creating a terminal session
type TerminalSessionOptions struct {
	SessionID    string
	TerminalType string
}

// NewTerminalSessionWithOptions creates a new terminal session with the specified options
// This unified function replaces NewTerminalSessionWithID and NewTerminalSessionWithService
func NewTerminalSessionWithOptions(opts TerminalSessionOptions) *TerminalSession {
	return &TerminalSession{
		isRunning:    false,
		sessionID:    opts.SessionID,
		terminalType: opts.TerminalType,
		lastActivity: time.Now(),
		readBuffer:   make(chan []byte, 100), // Buffered channel for performance
		writeBuffer:  make(chan []byte, 50),  // Buffered channel for writes
		stopChannel:  make(chan struct{}, 1), // Channel for clean shutdown
	}
}

// SetTimeoutCallback sets the callback function for session timeout
func (ts *TerminalSession) SetTimeoutCallback(callback func(sessionID string)) {
	ts.onTimeout = callback
}

// Start initializes and starts a PTY terminal session
func (ts *TerminalSession) Start() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.isRunning {
		return fmt.Errorf("terminal is already running")
	}

	// Determine shell based on terminal type or platform default
	shell, args := ts.getShellCommand()

	// Create PTY with optimized settings
	ptyInstance, err := pty.New()
	if err != nil {
		return fmt.Errorf("failed to create PTY: %w", err)
	}

	// Create command using PTY's Command method with optimizations
	cmd := ptyInstance.Command(shell, args...)

	// Set up isolated environment if service is specified
	// This does NOT tamper with global environment - only affects this specific session
	// Always build environment based on config, ignore serviceName
	cmd.Env = BuildIsolatedEnvForService(ts.terminalType, "")

	// Start the command
	err = cmd.Start()
	if err != nil {
		ptyInstance.Close()
		return fmt.Errorf("failed to start command: %w", err)
	}

	ts.pty = ptyInstance
	ts.cmd = cmd
	ts.isRunning = true
	ts.lastActivity = time.Now()

	// Start timeout timer (60 minutes)
	ts.startTimeoutTimer()

	// Start optimized I/O goroutines
	go ts.writeHandler() // Handle buffered writes
	go ts.readHandler()  // Handle buffered reads

	return nil
}

// Write sends input to the terminal (optimized with buffering)
func (ts *TerminalSession) Write(input string) error {
	ts.mutex.RLock()
	if !ts.isRunning || ts.pty == nil {
		ts.mutex.RUnlock()
		return fmt.Errorf("terminal is not running")
	}
	ts.mutex.RUnlock()

	// Update last activity time
	ts.mutex.Lock()
	ts.lastActivity = time.Now()
	ts.resetTimeoutTimer()
	ts.mutex.Unlock()

	// Use buffered write for better performance
	data := []byte(input)
	select {
	case ts.writeBuffer <- data:
		return nil
	default:
		// If buffer is full, write directly (fallback)
		_, err := ts.pty.Write(data)
		return err
	}
}

// Stop closes the PTY terminal session
func (ts *TerminalSession) Stop() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning {
		return nil
	}

	// Signal shutdown to goroutines
	close(ts.stopChannel)

	// Stop timeout timer
	if ts.timeoutTimer != nil {
		ts.timeoutTimer.Stop()
		ts.timeoutTimer = nil
	}

	if ts.pty != nil {
		ts.pty.Close()
	}

	if ts.cmd != nil && ts.cmd.Process != nil {
		ts.cmd.Process.Kill()
	}

	ts.isRunning = false
	ts.pty = nil
	ts.cmd = nil

	// Close channels
	close(ts.readBuffer)
	close(ts.writeBuffer)

	return nil
}

// IsRunning returns the terminal status (optimized with RLock)
func (ts *TerminalSession) IsRunning() bool {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.isRunning
}

// Resize resizes the PTY terminal (optimized with RLock)
func (ts *TerminalSession) Resize(cols, rows int) error {
	ts.mutex.RLock()
	pty := ts.pty
	isRunning := ts.isRunning
	ts.mutex.RUnlock()

	if !isRunning || pty == nil {
		return fmt.Errorf("terminal is not running")
	}

	return pty.Resize(cols, rows)
}

// Read reads data from the terminal (blocking)
func (ts *TerminalSession) Read(buf []byte) (int, error) {
	ts.mutex.RLock()
	pty := ts.pty
	isRunning := ts.isRunning
	ts.mutex.RUnlock()

	if !isRunning || pty == nil {
		return 0, fmt.Errorf("terminal is not running")
	}

	return pty.Read(buf)
}

// writeHandler handles buffered writes to PTY (performance optimization)
func (ts *TerminalSession) writeHandler() {
	for {
		select {
		case data := <-ts.writeBuffer:
			ts.mutex.RLock()
			pty := ts.pty
			ts.mutex.RUnlock()

			if pty != nil {
				pty.Write(data)
			}
		case <-ts.stopChannel:
			return
		}
	}
}

// readHandler handles buffered reads from PTY (performance optimization)
func (ts *TerminalSession) readHandler() {
	buf := make([]byte, 4096)

	for {
		select {
		case <-ts.stopChannel:
			return
		default:
			n, err := ts.Read(buf)
			if err != nil {
				if err == io.EOF {
					return
				}
				ts.mutex.RLock()
				running := ts.isRunning
				ts.mutex.RUnlock()

				if !running {
					return
				}
				continue
			}

			if n > 0 {
				// Send to read buffer channel
				data := make([]byte, n)
				copy(data, buf[:n])

				select {
				case ts.readBuffer <- data:
					// Update activity time
					ts.mutex.Lock()
					ts.lastActivity = time.Now()
					ts.resetTimeoutTimer()
					ts.mutex.Unlock()
				case <-ts.stopChannel:
					return
				}
			}
		}
	}
}

// GetPlatformShell returns the appropriate shell command and arguments for the current platform
func GetPlatformShell() (string, []string) {
	switch gosruntime.GOOS {
	case "windows":
		return "powershell.exe", []string{}
	case "darwin", "linux":
		return "/bin/bash", []string{}
	default:
		// Fallback to sh for unknown platforms
		return "/bin/sh", []string{}
	}
}

// TerminalReadCallback defines the function signature for handling terminal output
type TerminalReadCallback func(data string)

// TerminalExitCallback defines the function signature for handling terminal exit
type TerminalExitCallback func(message string)

// StartReadLoop starts a goroutine to continuously read from the terminal and call the provided callback
func (ts *TerminalSession) StartReadLoop(onData TerminalReadCallback, onExit TerminalExitCallback) {
	go func() {
		for {
			select {
			case data := <-ts.readBuffer:
				if onData != nil && len(data) > 0 {
					onData(string(data))
				}
			case <-ts.stopChannel:
				// Terminal process ended, clean up
				ts.mutex.Lock()
				ts.isRunning = false
				ts.mutex.Unlock()

				if onExit != nil {
					onExit("Terminal session ended")
				}
				return
			}
		}
	}()
}

// getShellCommand returns the appropriate shell command based on terminal type
func (ts *TerminalSession) getShellCommand() (string, []string) {
	switch ts.terminalType {
	case "cmd":
		return "cmd.exe", []string{}
	case "powershell":
		return "powershell.exe", []string{}
	case "bash":
		return "/bin/bash", []string{}
	default:
		// Use platform default
		return GetPlatformShell()
	}
}

func BuildIsolatedEnvForService(shellType, serviceName string) (envSlice []string) {
	defer func() {
		if r := recover(); r != nil {
			envSlice = os.Environ()
		}
	}()
	baseEnv := make(map[string]string)
	for _, env := range os.Environ() {
		parts := strings.SplitN(env, "=", 2)
		if len(parts) == 2 {
			baseEnv[parts[0]] = parts[1]
		}
	}

	pathsConfig := config.LivePathsConfigManager()
	if pathsConfig == nil {
		return os.Environ()
	}

	var pathComponents []string

	if serviceName == "" {
		// Prepend all valid service paths
		for _, servicePath := range pathsConfig.GetAllServicePaths() {
			if stat, err := os.Stat(servicePath); err == nil && stat.IsDir() {
				pathComponents = append(pathComponents, servicePath)
			}
		}
	} else {
		// Add only the requested service path if valid
		if servicePath, exists := pathsConfig.GetServicePath(serviceName); exists {
			if stat, err := os.Stat(servicePath); err == nil && stat.IsDir() {
				pathComponents = append(pathComponents, servicePath)
			}
		}
	}

	// Add default paths if valid
	for _, p := range pathsConfig.GetDefaultPaths() {
		if stat, err := os.Stat(p); err == nil && stat.IsDir() {
			pathComponents = append(pathComponents, p)
		}
	}

	// Add standard Unix paths if not on Windows
	if gosruntime.GOOS != "windows" {
		for _, p := range pathsConfig.GetStandardUnixPaths() {
			if stat, err := os.Stat(p); err == nil && stat.IsDir() {
				pathComponents = append(pathComponents, p)
			}
		}
	}

	sep := ":"
	if gosruntime.GOOS == "windows" {
		sep = ";"
	}
	pathStr := strings.Join(pathComponents, sep)
	if pathStr == "" {
		return os.Environ()
	}
	baseEnv["PATH"] = pathStr

	for key, value := range baseEnv {
		envSlice = append(envSlice, key+"="+value)
	}
	return
}

// startTimeoutTimer starts or resets the 60-minute inactivity timeout timer
func (ts *TerminalSession) startTimeoutTimer() {
	if ts.timeoutTimer != nil {
		ts.timeoutTimer.Stop()
	}

	ts.timeoutTimer = time.AfterFunc(60*time.Minute, func() {
		ts.mutex.Lock()
		sessionID := ts.sessionID
		ts.mutex.Unlock()

		// Stop the session
		ts.Stop()

		// Notify callback if set
		if ts.onTimeout != nil {
			ts.onTimeout(sessionID)
		}
	})
}

// resetTimeoutTimer resets the timeout timer on activity
func (ts *TerminalSession) resetTimeoutTimer() {
	ts.startTimeoutTimer()
}

func (ts *TerminalSession) GetSessionInfo() (sessionID, terminalType string, lastActivity time.Time) {
	ts.mutex.RLock()
	defer ts.mutex.RUnlock()
	return ts.sessionID, ts.terminalType, ts.lastActivity
}

// TerminalManager manages multiple terminal sessions
type TerminalManager struct {
	sessions    map[string]*TerminalSession
	sessionPool map[string][]*TerminalSession // Pool for reusing sessions
	mutex       sync.RWMutex
}

// NewTerminalManager creates a new terminal manager
func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		sessions:    make(map[string]*TerminalSession),
		sessionPool: make(map[string][]*TerminalSession),
	}
}

// CreateSessionOptions holds options for creating a terminal session via TerminalManager
type CreateSessionOptions struct {
	SessionID    string
	TerminalType string
}

// CreateSession creates a new terminal session with the specified options
// This unified method replaces the old CreateSession and CreateSessionWithService methods
func (tm *TerminalManager) CreateSession(opts CreateSessionOptions) (*TerminalSession, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.sessions[opts.SessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", opts.SessionID)
	}

	session := NewTerminalSessionWithOptions(TerminalSessionOptions(opts))

	session.SetTimeoutCallback(func(id string) {
		tm.removeSession(id)
	})

	tm.sessions[opts.SessionID] = session
	return session, nil
}

// GetSession retrieves an existing session
func (tm *TerminalManager) GetSession(sessionID string) (*TerminalSession, error) {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	session, exists := tm.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session %s not found", sessionID)
	}

	return session, nil
}

// RemoveSession removes and stops a session
func (tm *TerminalManager) RemoveSession(sessionID string) error {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	return tm.removeSession(sessionID)
}

// removeSession removes a session (internal, assumes lock is held)
func (tm *TerminalManager) removeSession(sessionID string) error {
	session, exists := tm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session %s not found", sessionID)
	}

	session.Stop()
	delete(tm.sessions, sessionID)
	return nil
}

// ListSessions returns all active session IDs
func (tm *TerminalManager) ListSessions() []string {
	tm.mutex.RLock()
	defer tm.mutex.RUnlock()

	sessionIDs := make([]string, 0, len(tm.sessions))
	for id := range tm.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	return sessionIDs
}

// CleanupAll stops and removes all sessions
func (tm *TerminalManager) CleanupAll() {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	for id := range tm.sessions {
		tm.removeSession(id)
	}
}
