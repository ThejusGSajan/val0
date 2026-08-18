package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/val-tracker/val-tracker/internal/tui"
)

func main() {
	p := tea.NewProgram(tui.NewRootModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running Val-Tracker: %v\n", err)
		os.Exit(1)
	}
}
