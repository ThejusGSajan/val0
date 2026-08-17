package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
)

type ProgressModel struct {
	data     *BattlepassData
	missions []models.Mission
	width    int
	height   int
}

func NewProgressModel(data *BattlepassData, missions []models.Mission) ProgressModel {
	return ProgressModel{data: data, missions: missions}
}

// Alias for backwards compatibility
type BattlepassModel = ProgressModel

func NewBattlepassModel(data *BattlepassData) BattlepassModel {
	return NewProgressModel(data, nil)
}

func (m *ProgressModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ProgressModel) Update(msg tea.Msg) (ProgressModel, tea.Cmd) {
	return m, nil
}

func (m ProgressModel) View() string {
	var sections []string

	// ── Battlepass Section ──────────────────────────────────────────
	if m.data != nil {
		d := m.data
		header := lipgloss.NewStyle().
			Foreground(ColorAccent).
			Bold(true).
			Render(fmt.Sprintf("  ⚔  BATTLEPASS — Tier %d / %d", d.CurrentTier, d.MaxTier))

		barWidth := 50
		if m.width > 60 {
			barWidth = m.width - 20
			if barWidth > 70 {
				barWidth = 70
			}
		}

		overallPct := 0.0
		if d.MaxTier > 0 {
			overallPct = float64(d.CurrentTier) / float64(d.MaxTier)
		}
		overallBar := renderProgressBar(overallPct, barWidth, "Overall Progression")

		tierPct := 0.0
		if d.XPForNextTier > 0 {
			tierPct = float64(d.XPInCurrentTier) / float64(d.XPForNextTier)
		}
		tierBar := renderProgressBar(tierPct, barWidth,
			fmt.Sprintf("Tier %d → %d  (%d / %d XP)",
				d.CurrentTier, d.CurrentTier+1,
				d.XPInCurrentTier, d.XPForNextTier))

		totalXP := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(fmt.Sprintf("  Total XP Earned: %d", d.TotalXP))

		sections = append(sections, header, "", overallBar, "", tierBar, "", totalXP)
	} else {
		sections = append(sections, lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No battlepass data available."))
	}

	// ── Missions Section ────────────────────────────────────────────
	sections = append(sections, "", strings.Repeat("─", 60), "")

	missionHeader := lipgloss.NewStyle().
		Foreground(ColorSelect).
		Bold(true).
		Render("  🎯 ACTIVE MISSIONS")
	sections = append(sections, missionHeader, "")

	if len(m.missions) == 0 {
		sections = append(sections, lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No active missions."))
	} else {
		barWidth := 40
		if m.width > 60 {
			barWidth = m.width - 25
			if barWidth > 60 {
				barWidth = 60
			}
		}

		for idx, ms := range m.missions {
			title := fmt.Sprintf("Mission #%d", idx+1)
			if len(ms.ID) >= 8 {
				title = fmt.Sprintf("Mission %s", ms.ID[:8])
			}

			if ms.Complete {
				completedText := lipgloss.NewStyle().
					Foreground(ColorDeluxe).
					Bold(true).
					Render(fmt.Sprintf("  ✔ %s — Completed", title))
				sections = append(sections, completedText)
			} else {
				// Show objective counts
				for objID, count := range ms.Objectives {
					objLabel := fmt.Sprintf("%s (Progress: %d)", title, count)
					if len(objID) >= 8 {
						objLabel = fmt.Sprintf("%s [%s] — %d completed", title, objID[:8], count)
					}
					// If we have count, render bar; assume standard progress or indeterminate
					sections = append(sections, renderProgressBar(0.5, barWidth, objLabel))
				}
				if len(ms.Objectives) == 0 {
					sections = append(sections, lipgloss.NewStyle().
						Foreground(ColorFg).
						Render(fmt.Sprintf("  • %s — In Progress", title)))
				}
			}
			sections = append(sections, "")
		}
	}

	return strings.Join(sections, "\n")
}

// renderProgressBar creates a terminal progress bar with label.
func renderProgressBar(pct float64, width int, label string) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 1 {
		pct = 1
	}
	if width < 10 {
		width = 10
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
