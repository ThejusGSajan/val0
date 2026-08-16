package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorModel displays a full-screen error and waits for q/ctrl+c to quit.
type ErrorModel struct {
	message string
	width   int
	height  int
}

func NewErrorModel(msg string) ErrorModel {
	return ErrorModel{message: msg}
}

func (m ErrorModel) Init() tea.Cmd { return nil }

func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ErrorModel) View() string {
	content := ErrorStyle.Render(
		"✕  " + m.message + "\n\nPress q to exit.",
	)
	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, content)
}
