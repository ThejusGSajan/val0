package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ErrorModel displays a full-screen error and waits for r/q/ctrl+c.
type ErrorModel struct {
	message string
	width   int
	height  int
}

func NewErrorModel(msg string) ErrorModel {
	return ErrorModel{message: msg}
}

func (m *ErrorModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ErrorModel) Init() tea.Cmd { return nil }

func (m ErrorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "r", "R":
			return m, func() tea.Msg { return RetryAuthMsg{} }
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m ErrorModel) View() string {
	if m.width <= 0 || m.height <= 0 {
		return "\n  ✕  " + m.message + "\n\n  Press 'r' to retry, 'q' to exit.\n"
	}

	boxWidth := m.width - 8
	if boxWidth > 64 {
		boxWidth = 64
	}
	if boxWidth < 28 {
		boxWidth = 28
	}

	padX := 4
	padY := 2
	if m.width < 40 {
		padX = 1
		padY = 1
	}

	content := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#EF4444")).
		Bold(true).
		Border(lipgloss.DoubleBorder()).
		BorderForeground(lipgloss.Color("#EF4444")).
		Padding(padY, padX).
		Width(boxWidth).
		Align(lipgloss.Center).
		Render("✕  " + m.message + "\n\nPress 'r' to retry, 'q' to exit.")

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, content)
}
