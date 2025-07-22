package main

import (
	"context"
	"fmt"

	"github.com/JadlionHD/Enty/internal/utils"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx      context.Context
	terminal *utils.TerminalSession
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{
		terminal: utils.NewTerminalSession(),
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

// StartTerminal initializes and starts a PTY terminal session
func (a *App) StartTerminal() error {
	err := a.terminal.Start()
	if err != nil {
		return err
	}

	// Start the read loop with callback functions for Wails events
	a.terminal.StartReadLoop(
		// onData callback - emit terminal output to frontend
		func(data string) {
			runtime.EventsEmit(a.ctx, "terminal:data", data)
		},
		// onExit callback - emit terminal exit to frontend  
		func(message string) {
			runtime.EventsEmit(a.ctx, "terminal:exit", message)
		},
	)

	return nil
}

// WriteToTerminal sends input to the terminal
func (a *App) WriteToTerminal(input string) error {
	return a.terminal.Write(input)
}

// StopTerminal closes the PTY terminal session
func (a *App) StopTerminal() error {
	return a.terminal.Stop()
}

// IsTerminalRunning returns the terminal status
func (a *App) IsTerminalRunning() bool {
	return a.terminal.IsRunning()
}

// ResizeTerminal resizes the PTY
func (a *App) ResizeTerminal(cols, rows int) error {
	return a.terminal.Resize(cols, rows)
}
