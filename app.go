package main

import (
	"context"
	"fmt"
	"log"
	"syscall"

	"github.com/lxn/win"
)

// App struct
type App struct {
	ctx          context.Context
	topTimer     TimerD
	bottomTimer  TimerD
	pausedTimers bool
}

// NewApp creates a new App application struct
func NewApp() *App {
	return &App{}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	log.SetPrefix("App: ")
	log.SetFlags(0)
	a.ctx = ctx
	ptr, err := syscall.UTF16PtrFromString("Grandmodoro")
	if err != nil {
		log.Fatalf("syscall %v", err)
	}
	hwnd := win.FindWindow(nil, ptr)
	win.SetWindowLong(hwnd, win.GWL_EXSTYLE, win.GetWindowLong(hwnd, win.GWL_EXSTYLE)|win.WS_EX_LAYERED)
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}
