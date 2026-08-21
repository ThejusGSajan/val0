package tui

import (
	"fmt"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
)

type AgentStat struct {
	AgentName    string
	Matches      int
	Wins         int
	Kills        int
	Deaths       int
	Assists      int
	TotalScore   int
	TotalRounds  int
	TotalDamage  int
	Headshots    int
	TotalShots   int
}

type WeaponStat struct {
	WeaponName string
	Kills      int
	Headshots  int
	Bodyshots  int
	Legshots   int
}

type StatsModel struct {
	agentStats     []AgentStat
	weaponStats    []WeaponStat
	hsHistory      []int // HS% per match chronologically
	rrHistory      []int // RR earned per match
	currentRank    string
	netRR          int
	scrollOffset   int
	width          int
	height         int
}

func NewStatsModel(
	matchDetails []*models.MatchDetails,
	playerPUUID string,
	compUpdates *models.CompetitiveUpdatesResponse,
	agentsMap map[string]string,
	weaponsMap map[string]string,
	currentRank string,
) StatsModel {
	agentMap := make(map[string]*AgentStat)
	weaponMap := make(map[string]*WeaponStat)
	var hsHistory []int

	for _, md := range matchDetails {
		if md == nil {
			continue
		}
		player := md.GetPlayer(playerPUUID)
		if player == nil {
			continue
		}

		agentName := cache.GetAgentName(player.CharacterID, agentsMap)
		aStat, ok := agentMap[agentName]
		if !ok {
			aStat = &AgentStat{AgentName: agentName}
			agentMap[agentName] = aStat
		}

		aStat.Matches++
		if md.GetMatchOutcome(playerPUUID) == "WIN" {
			aStat.Wins++
		}
		aStat.Kills += player.Stats.Kills
		aStat.Deaths += player.Stats.Deaths
		aStat.Assists += player.Stats.Assists
		aStat.TotalScore += player.Stats.Score
		aStat.TotalRounds += max(player.Stats.RoundsPlayed, 1)

		// Aggregate round damage and weapon kills
		matchHS := 0
		matchShots := 0
		for _, rr := range md.RoundResults {
			for _, ps := range rr.PlayerStats {
				if ps.Subject == playerPUUID {
					for _, dmg := range ps.Damage {
						aStat.TotalDamage += dmg.Damage
						aStat.Headshots += dmg.Headshots
						shots := dmg.Headshots + dmg.Bodyshots + dmg.Legshots
						aStat.TotalShots += shots
						matchHS += dmg.Headshots
						matchShots += shots
					}
					for _, k := range ps.Kills {
						if k.Killer == playerPUUID {
							wName := cache.GetWeaponName(k.FinishingDamage.DamageItem, weaponsMap)
							if wName == "" || wName == "Weapon" {
								wName = "Gun"
							}
							ws, ok := weaponMap[wName]
							if !ok {
								ws = &WeaponStat{WeaponName: wName}
								weaponMap[wName] = ws
							}
							ws.Kills++
						}
					}
				}
			}
		}

		if matchShots > 0 {
			hsHistory = append(hsHistory, (matchHS*100)/matchShots)
		} else {
			hsHistory = append(hsHistory, 0)
		}
	}

	// Sort agents by matches played
	var agentList []AgentStat
	for _, as := range agentMap {
		agentList = append(agentList, *as)
	}
	sort.Slice(agentList, func(i, j int) bool {
		return agentList[i].Matches > agentList[j].Matches
	})

	// Sort weapons by kills
	var weaponList []WeaponStat
	for _, ws := range weaponMap {
		weaponList = append(weaponList, *ws)
	}
	sort.Slice(weaponList, func(i, j int) bool {
		return weaponList[i].Kills > weaponList[j].Kills
	})

	// RR history from CompetitiveUpdates
	var rrHistory []int
	netRR := 0
	if compUpdates != nil {
		for _, m := range compUpdates.Matches {
			rrHistory = append(rrHistory, m.RankedRatingEarned)
			netRR += m.RankedRatingEarned
		}
	}

	return StatsModel{
		agentStats:   agentList,
		weaponStats:  weaponList,
		hsHistory:    hsHistory,
		rrHistory:    rrHistory,
		currentRank:  currentRank,
		netRR:        netRR,
	}
}

func (m *StatsModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m StatsModel) Update(msg tea.Msg) (StatsModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "w":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
		case "down", "j", "s":
			m.scrollOffset++
		}
	}
	return m, nil
}

func (m StatsModel) View() string {
	var sections []string

	lineWidth := max(m.width-4, 70)

	// ── Section 1: Agent Performance ──────────────────────────────
	sec1Header := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  🎯 AGENT PERFORMANCE                                    Last %d matches", len(m.hsHistory)))
	sections = append(sections, sec1Header)
	sections = append(sections, "  "+strings.Repeat("─", lineWidth))
	agentHeaderRow := fmt.Sprintf("  %-10s  %5s    %4s   %4s   %4s   %4s   %4s",
		"Agent", "Games", "Win%", "K/D", "ACS", "ADR", "HS%")
	sections = append(sections, lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(agentHeaderRow))

	if len(m.agentStats) == 0 {
		sections = append(sections, "  No agent data available.")
	} else {
		for _, as := range m.agentStats {
			winPct := 0
			if as.Matches > 0 {
				winPct = (as.Wins * 100) / as.Matches
			}
			kd := float64(as.Kills)
			if as.Deaths > 0 {
				kd = float64(as.Kills) / float64(as.Deaths)
			}
			rounds := max(as.TotalRounds, 1)
			acs := as.TotalScore / rounds
			adr := as.TotalDamage / rounds
			hsPct := 0
			if as.TotalShots > 0 {
				hsPct = (as.Headshots * 100) / as.TotalShots
			}

			row := fmt.Sprintf("  %-10s  %5d    %3d%%   %4.2f   %4d   %4d   %3d%%",
				truncate(as.AgentName, 10),
				as.Matches,
				winPct,
				kd,
				acs,
				adr,
				hsPct,
			)
			sections = append(sections, row)
		}
	}

	sections = append(sections, "")

	// ── Section 2: Weapon Stats ────────────────────────────────────
	sec2Header := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render("  🔫 WEAPON STATS")
	sections = append(sections, sec2Header)
	sections = append(sections, "  "+strings.Repeat("─", lineWidth))
	weaponHeaderRow := fmt.Sprintf("  %-13s  %s", "Weapon", "Kills")
	sections = append(sections, lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(weaponHeaderRow))

	if len(m.weaponStats) == 0 {
		sections = append(sections, "  No weapon data available.")
	} else {
		for i, ws := range m.weaponStats {
			if i >= 6 {
				break
			}
			sections = append(sections, fmt.Sprintf("  %-13s  %4d kills", truncate(ws.WeaponName, 13), ws.Kills))
		}
	}

	sections = append(sections, "")

	// ── Section 3: Aim Analysis (HS% Sparkline) ────────────────────
	sec3Header := lipgloss.NewStyle().
		Foreground(ColorSelect).
		Bold(true).
		Render("  📈 AIM ANALYSIS — Headshot % Trend")
	sections = append(sections, sec3Header)
	sections = append(sections, "  "+strings.Repeat("─", lineWidth))

	if len(m.hsHistory) > 0 {
		sparkline := renderSparkline(m.hsHistory, 40)
		sections = append(sections, sparkline)

		currentHS := m.hsHistory[0]
		avgHS := 0
		sum := 0
		for _, h := range m.hsHistory {
			sum += h
		}
		if len(m.hsHistory) > 0 {
			avgHS = sum / len(m.hsHistory)
		}
		trendStr := "→ Flat"
		if currentHS > avgHS {
			trendStr = fmt.Sprintf("↑ +%d%%", currentHS-avgHS)
		} else if currentHS < avgHS {
			trendStr = fmt.Sprintf("↓ -%d%%", avgHS-currentHS)
		}

		summary := lipgloss.NewStyle().
			Foreground(ColorFg).
			Render(fmt.Sprintf("  Current: %d%%   Avg: %d%%   Trend: %s", currentHS, avgHS, trendStr))
		sections = append(sections, summary)
	} else {
		sections = append(sections, "  No HS% trend data available.")
	}

	sections = append(sections, "")

	// ── Section 4: Rank Rating History ─────────────────────────────
	sec4Header := lipgloss.NewStyle().
		Foreground(ColorDeluxe).
		Bold(true).
		Render("  💎 RANK RATING (RR) HISTORY")
	sections = append(sections, sec4Header)
	sections = append(sections, "  "+strings.Repeat("─", lineWidth))

	if len(m.rrHistory) > 0 {
		rrSparkline := renderSignedSparkline(m.rrHistory, 40)
		sections = append(sections, rrSparkline)

		sign := "+"
		if m.netRR < 0 {
			sign = ""
		}
		summary := lipgloss.NewStyle().
			Foreground(ColorFg).
			Render(fmt.Sprintf("  Current Rank: %s   Net: %s%d RR over %d games",
				m.currentRank, sign, m.netRR, len(m.rrHistory)))
		sections = append(sections, summary)
	} else {
		sections = append(sections, "  No competitive updates available.")
	}

	sections = append(sections, "")
	sections = append(sections, lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  ↑/↓ scroll stats view"))

	allLines := strings.Split(strings.Join(sections, "\n"), "\n")

	contentHeight := m.height
	if contentHeight <= 0 {
		contentHeight = 20
	}
	maxScroll := max(0, len(allLines)-contentHeight)
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}
	if m.scrollOffset > maxScroll {
		m.scrollOffset = maxScroll
	}

	endLine := min(m.scrollOffset+contentHeight, len(allLines))
	visibleLines := allLines[m.scrollOffset:endLine]
	return strings.Join(visibleLines, "\n")
}

// ── Sparkline Renderers ─────────────────────────────────────────────

func renderSparkline(values []int, maxBars int) string {
	if len(values) == 0 {
		return ""
	}
	bars := []rune{' ', '▂', '▃', '▄', '▅', '▆', '▇', '█'}

	minVal := 100
	maxVal := 0
	for _, v := range values {
		if v < minVal {
			minVal = v
		}
		if v > maxVal {
			maxVal = v
		}
	}
	if maxVal == minVal {
		maxVal = minVal + 1
	}

	var sb strings.Builder
	sb.WriteString("  [Recent] ")
	n := min(len(values), maxBars)
	for i := n - 1; i >= 0; i-- {
		val := values[i]
		idx := (val - minVal) * (len(bars) - 1) / (maxVal - minVal)
		if idx < 0 {
			idx = 0
		}
		if idx >= len(bars) {
			idx = len(bars) - 1
		}
		sb.WriteRune(bars[idx])
	}
	sb.WriteString(" [Oldest]\n")
	return sb.String()
}

func renderSignedSparkline(values []int, maxBars int) string {
	if len(values) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString("  [Recent] ")
	n := min(len(values), maxBars)
	for i := n - 1; i >= 0; i-- {
		v := values[i]
		if v > 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(ColorWin).Render("▲"))
		} else if v < 0 {
			sb.WriteString(lipgloss.NewStyle().Foreground(ColorLoss).Render("▼"))
		} else {
			sb.WriteString(lipgloss.NewStyle().Foreground(ColorDraw).Render("─"))
		}
		sb.WriteString(" ")
	}
	sb.WriteString("[Oldest]\n")
	return sb.String()
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
