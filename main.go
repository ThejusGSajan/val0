package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/val-tracker/val-tracker/internal/launcher"
	"github.com/val-tracker/val-tracker/internal/tui"
)

func main() {
	// Relaunch in Windows Terminal if double-clicked from Explorer
	if launcher.IsStandaloneConhost() {
		launcher.TryRelaunchInWT(os.Args[1:])
	}

	// CLI Flag Parser
	graphicsFlag := flag.String("graphics", "", "Force graphics protocol: sixel, kitty, iterm2, halfblock")
	flag.Parse()

	if *graphicsFlag != "" {
		// Overwrite the environment variable so detect.go naturally picks it up
		os.Setenv("VAL0_GRAPHICS", *graphicsFlag)
	}

	p := tea.NewProgram(tui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running val0: %v\n", err)
		os.Exit(1)
	}
}
