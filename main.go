package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/val-tracker/val-tracker/internal/launcher"
	"github.com/val-tracker/val-tracker/internal/tui"
)

var (
	Version   = "dev"
	Commit    = "none"
	BuildDate = "unknown"
)

func main() {
	// CLI Flag Parser
	graphicsFlag := flag.String("graphics", "", "Force graphics protocol: sixel, kitty, iterm2, halfblock")
	versionFlag := flag.Bool("version", false, "Print version and build information")
	vFlag := flag.Bool("v", false, "Print version and build information (shorthand)")
	flag.Parse()

	if *versionFlag || *vFlag {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Relaunch in Windows Terminal if double-clicked from Explorer
	if launcher.IsStandaloneConhost() {
		launcher.TryRelaunchInWT(os.Args[1:])
	}

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

