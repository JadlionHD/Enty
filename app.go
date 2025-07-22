package main

import (
	"context"
	"fmt"
	gosruntime "runtime"

	"github.com/JadlionHD/Enty/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx             context.Context
	terminalManager *utils.TerminalManager
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		terminalManager: utils.NewTerminalManager(),
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

// CreateTerminalSession creates a new terminal session
func (a *App) CreateTerminalSession(sessionID, terminalType string) error {
	session, err := a.terminalManager.CreateSession(sessionID, terminalType)
	if err != nil {
		return err
	}

	err = session.Start()
	if err != nil {
		a.terminalManager.RemoveSession(sessionID)
		return err
	}

	// Start the read loop with callback functions for Wails events
	session.StartReadLoop(
		// onData callback - emit terminal output to frontend
		func(data string) {
			runtime.EventsEmit(a.ctx, "terminal:data", map[string]interface{}{
				"sessionID": sessionID,
				"data":      data,
			})
		},
		// onExit callback - emit terminal exit to frontend  
		func(message string) {
			runtime.EventsEmit(a.ctx, "terminal:exit", map[string]interface{}{
				"sessionID": sessionID,
				"message":   message,
			})
			a.terminalManager.RemoveSession(sessionID)
		},
	)

	return nil
}

// WriteToTerminal sends input to a specific terminal session
func (a *App) WriteToTerminal(sessionID, input string) error {
	session, err := a.terminalManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.Write(input)
}

// CloseTerminalSession closes a specific terminal session
func (a *App) CloseTerminalSession(sessionID string) error {
	return a.terminalManager.RemoveSession(sessionID)
}

// IsTerminalSessionRunning returns whether a specific session is running
func (a *App) IsTerminalSessionRunning(sessionID string) bool {
	session, err := a.terminalManager.GetSession(sessionID)
	if err != nil {
		return false
	}
	return session.IsRunning()
}

// ResizeTerminalSession resizes a specific terminal session
func (a *App) ResizeTerminalSession(sessionID string, cols, rows int) error {
	session, err := a.terminalManager.GetSession(sessionID)
	if err != nil {
		return err
	}
	return session.Resize(cols, rows)
}

// ListTerminalSessions returns all active terminal session IDs
func (a *App) ListTerminalSessions() []string {
	return a.terminalManager.ListSessions()
}

// GetAvailableTerminalTypes returns available terminal types for the current platform
func (a *App) GetAvailableTerminalTypes() []string {
	switch gosruntime.GOOS {
	case "windows":
		return []string{"powershell", "cmd"}
	case "darwin", "linux":
		return []string{"bash"}
	default:
		return []string{"bash"}
	}
}
