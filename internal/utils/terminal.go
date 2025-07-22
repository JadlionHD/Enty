package utils

import (
	"fmt"
	"io"
	gosruntime "runtime"
	"sync"

	"github.com/aymanbagabas/go-pty"
)

// TerminalSession manages a single PTY terminal session
type TerminalSession struct {
	pty       pty.Pty
	cmd       *pty.Cmd
	mutex     sync.Mutex
	isRunning bool
}

// NewTerminalSession creates a new terminal session
func NewTerminalSession() *TerminalSession {
	return &TerminalSession{
		isRunning: false,
	}
}

// Start initializes and starts a PTY terminal session
func (ts *TerminalSession) Start() error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if ts.isRunning {
		return fmt.Errorf("terminal is already running")
	}

	// Determine shell based on OS
	shell, args := GetPlatformShell()

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

	return nil
}

// Write sends input to the terminal
func (ts *TerminalSession) Write(input string) error {
	ts.mutex.Lock()
	defer ts.mutex.Unlock()

	if !ts.isRunning || ts.pty == nil {
		return fmt.Errorf("terminal is not running")
	}

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
