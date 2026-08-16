package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type BattlepassModel struct {
	data *BattlepassData
}

func NewBattlepassModel(data *BattlepassData) BattlepassModel {
	return BattlepassModel{data: data}
}

func (m BattlepassModel) Update(msg tea.Msg) (BattlepassModel, tea.Cmd) {
	return m, nil
}

func (m BattlepassModel) View() string {
	if m.data == nil {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No battlepass data available.")
	}

	d := m.data

	header := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  BATTLEPASS — Tier %d / %d", d.CurrentTier, d.MaxTier))

	// Overall progress bar
	overallPct := 0.0
	if d.MaxTier > 0 {
		overallPct = float64(d.CurrentTier) / float64(d.MaxTier)
	}
	overallBar := renderProgressBar(overallPct, 50, "Overall Progression")

	// Current tier progress
	tierPct := 0.0
	if d.XPForNextTier > 0 {
		tierPct = float64(d.XPInCurrentTier) / float64(d.XPForNextTier)
	}
	tierBar := renderProgressBar(tierPct, 50,
		fmt.Sprintf("Tier %d → %d  (%d / %d XP)",
			d.CurrentTier, d.CurrentTier+1,
			d.XPInCurrentTier, d.XPForNextTier))

	totalXP := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(fmt.Sprintf("  Total XP Earned: %d", d.TotalXP))

	return strings.Join([]string{
		header,
		"",
		overallBar,
		"",
		tierBar,
		"",
		totalXP,
	}, "\n")
}

// renderProgressBar creates a terminal progress bar with label.
func renderProgressBar(pct float64, width int, label string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}

	filled := int(pct * float64(width))
	empty := width - filled

	bar := ProgressFullStyle.Render(strings.Repeat("█", filled)) +
		ProgressEmptyStyle.Render(strings.Repeat("░", empty))

	pctStr := lipgloss.NewStyle().
		Foreground(ColorFg).
		Bold(true).
		Render(fmt.Sprintf(" %d%%", int(pct*100)))

	labelStr := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  " + label)

	return labelStr + "\n  " + bar + pctStr
}
