package main

import (
	"errors"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/tui"
)

func main() {
	// 1. Attempt to build a session from the local lockfile
	session, err := auth.BuildSession()
	if err != nil {
		var model tea.Model
		if errors.Is(err, auth.ErrLockfileNotFound) {
			model = tui.NewErrorModel("Please start the Riot Client first.\n(Lockfile was not found)")
		} else if errors.Is(err, auth.ErrLockfileStale) {
			model = tui.NewErrorModel("Please restart the Riot Client.\n(Lockfile is older than 1 hour)")
		} else {
			model = tui.NewErrorModel(fmt.Sprintf("Authentication failed:\n%s", err.Error()))
		}

		p := tea.NewProgram(model, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Fprintf(os.Stderr, "Error running TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// 2. Fetch the current client version (for X-Riot-ClientVersion header and cache validation)
	ver, err := cache.FetchRemoteVersion()
	if err != nil {
		// Fallback to a standard release version string if remote lookup fails
		session.ClientVersion = "release-13.02-shipping-17-5277781"
	} else {
		session.ClientVersion = ver.RiotClientVersion
	}

	// 3. Launch the MainModel
	mainModel := tui.NewMainModel(session)
	p := tea.NewProgram(mainModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
