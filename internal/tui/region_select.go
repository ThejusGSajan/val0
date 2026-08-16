package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var RegionOptions = []struct {
	Code string
	Name string
}{
	{"na", "North America"},
	{"eu", "Europe"},
	{"ap", "Asia Pacific"},
	{"kr", "Korea"},
	{"latam", "Latin America"},
	{"br", "Brazil"},
}

// RegionSelectedMsg carries the user's chosen region back to main.
type RegionSelectedMsg struct{ Region string }

type RegionSelectModel struct {
	cursor int
	width  int
	height int
}

func NewRegionSelectModel() RegionSelectModel {
	return RegionSelectModel{}
}

func (m RegionSelectModel) Init() tea.Cmd { return nil }

func (m RegionSelectModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(RegionOptions)-1 {
				m.cursor++
			}
		case "enter":
			return m, func() tea.Msg {
				return RegionSelectedMsg{Region: RegionOptions[m.cursor].Code}
			}
		}
	}
	return m, nil
}

func (m RegionSelectModel) View() string {
	title := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render("Region auto-detection failed.")

	subtitle := lipgloss.NewStyle().
		Foreground(ColorFg).
		Render("Select your region:")

	var items strings.Builder
	for i, r := range RegionOptions {
		cursor := "  "
		style := lipgloss.NewStyle().Foreground(ColorMuted)
		if i == m.cursor {
			cursor = "▸ "
			style = lipgloss.NewStyle().Foreground(ColorAccent).Bold(true)
		}
		items.WriteString(style.Render(fmt.Sprintf("%s%s  (%s)", cursor, r.Name, r.Code)))
		items.WriteString("\n")
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 3).
		Render(title + "\n" + subtitle + "\n\n" + items.String() + "\n↑/↓ to move, Enter to select")

	return lipgloss.Place(m.width, m.height,
		lipgloss.Center, lipgloss.Center, box)
}
