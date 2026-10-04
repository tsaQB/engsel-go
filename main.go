package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/tsaQB/engsel-go/internal/config"
	"github.com/tsaQB/engsel-go/internal/menu"
	"github.com/tsaQB/engsel-go/internal/service"
	"github.com/tsaQB/engsel-go/internal/tui"
)

func main() {
	// Graceful shutdown on Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		fmt.Print("\033[0m")
		tui.ClearScreen()
		os.Exit(0)
	}()

	// 1. Initialize configuration and environment variables
	config.InitConfig()

	// 2. Check for updates via Atom feed
	if service.CheckForUpdates() {
		tui.TuiPause("Tekan Enter untuk melanjutkan...")
	}

	// 3. Initialize Services
	service.AuthInstance.Init()
	service.BookmarkInstance.Init()

	// 4. Run interactive TUI
	menu.RunMainMenu()
}
