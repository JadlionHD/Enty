package utils

import (
	"fmt"
	"io"
	gosruntime "runtime"
	"sync"
	"time"

	"github.com/aymanbagabas/go-pty"
)

// TerminalSession manages a single PTY terminal session
type TerminalSession struct {
	pty           pty.Pty
	cmd           *pty.Cmd
	mutex         sync.Mutex
	isRunning     bool
	sessionID     string
	terminalType  string
	lastActivity  time.Time
	timeoutTimer  *time.Timer
	onTimeout     func(sessionID string)
}

// NewTerminalSession creates a new terminal session
func NewTerminalSession() *TerminalSession {
	return &TerminalSession{
		isRunning: false,
	}
}

// NewTerminalSessionWithID creates a new terminal session with specific ID and type
func NewTerminalSessionWithID(sessionID, terminalType string) *TerminalSession {
	return &TerminalSession{
		isRunning:    false,
		sessionID:    sessionID,
		terminalType: terminalType,
		lastActivity: time.Now(),
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

	// Create PTY
	ptyInstance, err := pty.New()
	if err != nil {
		return fmt.Errorf("failed to create PTY: %w", err)
	}

	// Create command using PTY's Command method
	cmd := ptyInstance.Command(shell, args...)
	
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

	return nil
}

// Write sends input to the terminal
func (ts *TerminalSession) Write(input string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning || ts.pty == nil {
		return fmt.Errorf("terminal is not running")
	}

	// Update last activity time
	ts.lastActivity = time.Now()
	ts.resetTimeoutTimer()

	_, err := ts.pty.Write([]byte(input))
	return err
}

// Stop closes the PTY terminal session
func (ts *TerminalSession) Stop() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning {
		return nil
	}

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

	return nil
}

// IsRunning returns the terminal status
func (ts *TerminalSession) IsRunning() bool {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	return ts.isRunning
}

// Resize resizes the PTY terminal
func (ts *TerminalSession) Resize(cols, rows int) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning || ts.pty == nil {
		return fmt.Errorf("terminal is not running")
	}

	return ts.pty.Resize(cols, rows)
}

// Read reads data from the terminal (blocking)
func (ts *TerminalSession) Read(buf []byte) (int, error) {
	ts.mutex.Lock()
	pty := ts.pty
	isRunning := ts.isRunning
	ts.mutex.Unlock()

	if !isRunning || pty == nil {
		return 0, fmt.Errorf("terminal is not running")
	}

	return pty.Read(buf)
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
		buf := make([]byte, 4096)
		
		for {
			n, err := ts.Read(buf)
			if err != nil {
				if err == io.EOF {
					break
				}
				// Continue reading even on error, unless terminal is stopped
				ts.mutex.Lock()
				running := ts.isRunning
				ts.mutex.Unlock()
				
				if !running {
					break
				}
				continue
			}

			if n > 0 && onData != nil {
				data := string(buf[:n])
				onData(data)
				
				// Update activity time
				ts.mutex.Lock()
				ts.lastActivity = time.Now()
				ts.resetTimeoutTimer()
				ts.mutex.Unlock()
			}
		}

		// Terminal process ended, clean up
		ts.mutex.Lock()
		ts.isRunning = false
		ts.mutex.Unlock()
		
		if onExit != nil {
			onExit("Terminal session ended")
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

// GetSessionInfo returns session information
func (ts *TerminalSession) GetSessionInfo() (sessionID, terminalType string, lastActivity time.Time) {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()
	return ts.sessionID, ts.terminalType, ts.lastActivity
}

// TerminalManager manages multiple terminal sessions
type TerminalManager struct {
	sessions map[string]*TerminalSession
	mutex    sync.RWMutex
}

// NewTerminalManager creates a new terminal manager
func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		sessions: make(map[string]*TerminalSession),
	}
}

// CreateSession creates a new terminal session
func (tm *TerminalManager) CreateSession(sessionID, terminalType string) (*TerminalSession, error) {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()

	if _, exists := tm.sessions[sessionID]; exists {
		return nil, fmt.Errorf("session %s already exists", sessionID)
	}

	session := NewTerminalSessionWithID(sessionID, terminalType)
	session.SetTimeoutCallback(func(id string) {
		tm.removeSession(id)
	})
	
	tm.sessions[sessionID] = session
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
