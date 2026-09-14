package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
)

type ProgressModel struct {
	data        *BattlepassData
	missions    []models.Mission
	missionsMap map[string]cache.MissionInfo
	width       int
	height      int
}

func NewProgressModel(data *BattlepassData, missions []models.Mission, missionsMap map[string]cache.MissionInfo) ProgressModel {
	return ProgressModel{
		data:        data,
		missions:    missions,
		missionsMap: missionsMap,
	}
}

// Alias for backwards compatibility
type BattlepassModel = ProgressModel

func NewBattlepassModel(data *BattlepassData) BattlepassModel {
	return NewProgressModel(data, nil, nil)
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
			tierFraction := 0.0
			if d.XPForNextTier > 0 {
				tierFraction = float64(d.XPInCurrentTier) / float64(d.XPForNextTier)
			}
			overallPct = (float64(d.CurrentTier-1) + tierFraction) / float64(d.MaxTier)
		}
		overallBar := renderProgressBar(overallPct, barWidth, "Overall Progression")

		var tierBar string
		if d.CurrentTier >= d.MaxTier && d.XPInCurrentTier >= d.XPForNextTier {
			tierBar = renderProgressBar(1.0, barWidth, fmt.Sprintf("Battlepass Completed! (%d / %d)", d.MaxTier, d.MaxTier))
		} else {
			tierPct := 0.0
			if d.XPForNextTier > 0 {
				tierPct = float64(d.XPInCurrentTier) / float64(d.XPForNextTier)
			}
			tierBar = renderProgressBar(tierPct, barWidth,
				fmt.Sprintf("Tier %d → %d  (%s / %s XP)",
					d.CurrentTier-1, d.CurrentTier,
					formatNumber(d.XPInCurrentTier), formatNumber(d.XPForNextTier)))
		}

		totalXP := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(fmt.Sprintf("  Total XP Earned: %s", formatNumber(d.TotalXP)))

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
			info, found := m.lookupMissionInfo(ms.ID)
			if found {
				target := info.ProgressToComplete
				if target <= 0 {
					target = 1
				}

				currentProgress := 0
				for _, count := range ms.Objectives {
					currentProgress += count
				}
				if currentProgress > target {
					currentProgress = target
				}

				if ms.Complete {
					label := fmt.Sprintf("  ✔ %s — Completed", info.Title)
					if info.XPGrant > 0 {
						label = fmt.Sprintf("  ✔ %s — Completed  (+%s XP)", info.Title, formatNumber(info.XPGrant))
					}
					completedText := lipgloss.NewStyle().
						Foreground(ColorDeluxe).
						Bold(true).
						Render(label)
					sections = append(sections, completedText)
				} else {
					pct := float64(currentProgress) / float64(target)
					label := fmt.Sprintf("%s — %d / %d", info.Title, currentProgress, target)
					if info.XPGrant > 0 {
						label = fmt.Sprintf("%s — %d / %d  (+%s XP)", info.Title, currentProgress, target, formatNumber(info.XPGrant))
					}
					sections = append(sections, renderProgressBar(pct, barWidth, label))
				}
			} else {
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
					count := 0
					for _, c := range ms.Objectives {
						count += c
					}
					label := fmt.Sprintf("%s — %d completed", title, count)
					sections = append(sections, renderProgressBar(0.0, barWidth, label))
				}
			}
			sections = append(sections, "")
		}
	}

	return strings.Join(sections, "\n")
}

func (m ProgressModel) lookupMissionInfo(id string) (cache.MissionInfo, bool) {
	idLower := strings.ToLower(id)
	if m.missionsMap != nil {
		if info, ok := m.missionsMap[idLower]; ok {
			return info, true
		}
	}
	if info, ok := cache.DefaultMissions[idLower]; ok {
		return info, true
	}
	return cache.MissionInfo{}, false
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
