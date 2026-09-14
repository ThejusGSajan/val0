package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ThejusGSajan/val0/internal/models"
)

type MatchViewMode int

const (
	MatchViewList MatchViewMode = iota
	MatchViewDetail
)

type MatchItem struct {
	MatchID     string
	MapName     string
	QueueName   string
	AgentName   string
	Outcome     string // "WIN", "LOSS", "DRAW"
	Score       string // "13-7"
	Kills       int
	Deaths      int
	Assists     int
	RREarned    int
	HasRR       bool
	GameTime    time.Time
	Details     *models.MatchDetails
	PlayerPUUID string
}

type MatchesModel struct {
	items       []MatchItem
	cursor      int
	viewMode    MatchViewMode
	width       int
	height      int
	playerPUUID string
	agentsMap   map[string]string
	mapsMap     map[string]string
}

func NewMatchesModel(items []MatchItem, playerPUUID string, agents map[string]string, maps map[string]string) MatchesModel {
	return MatchesModel{
		items:       items,
		playerPUUID: playerPUUID,
		agentsMap:   agents,
		mapsMap:     maps,
		viewMode:    MatchViewList,
	}
}

func (m *MatchesModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m MatchesModel) Update(msg tea.Msg) (MatchesModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "w":
			if m.viewMode == MatchViewList && m.cursor > 0 {
				m.cursor--
			}
		case "down", "j", "s":
			if m.viewMode == MatchViewList && m.cursor < len(m.items)-1 {
				m.cursor++
			}
		case "enter":
			if m.viewMode == MatchViewList && len(m.items) > 0 {
				m.viewMode = MatchViewDetail
			}
		case "esc", "backspace", "left":
			if m.viewMode == MatchViewDetail {
				m.viewMode = MatchViewList
			}
		}
	}
	return m, nil
}

func (m MatchesModel) View() string {
	if len(m.items) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No match history available.")
	}

	if m.viewMode == MatchViewDetail {
		return m.renderDetailView()
	}
	return m.renderListView()
}

// ── List View ───────────────────────────────────────────────────────

func (m MatchesModel) renderListView() string {
	var sb strings.Builder

	header := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  ⚔  MATCH HISTORY                                         Last %d matches", len(m.items)))
	sb.WriteString(header + "\n")
	sb.WriteString("  " + strings.Repeat("─", max(m.width-4, 70)) + "\n")

	for i, item := range m.items {
		cursor := "  "
		if i == m.cursor {
			cursor = "▸ "
		}

		outcome := item.Outcome
		score := item.Score
		if item.Details != nil && item.Details.IsDeathmatch() {
			if score == "" || score == "0-0" {
				score = item.Details.ScoreString(m.playerPUUID)
			}
			if outcome == "" || outcome == "DRAW" {
				outcome = item.Details.GetMatchOutcome(m.playerPUUID)
			}
		}

		// Outcome styling
		outcomeColor := ColorDraw
		outcomeBox := "■"
		switch outcome {
		case "WIN":
			outcomeColor = ColorWin
		case "LOSS":
			outcomeColor = ColorLoss
		}

		outcomeBadge := lipgloss.NewStyle().
			Foreground(outcomeColor).
			Bold(true).
			Render(fmt.Sprintf("%s %-4s", outcomeBox, outcome))

		// Map & Mode
		mapStr := lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true).
			Render(fmt.Sprintf("%-10s", truncate(item.MapName, 10)))

		queueStr := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(fmt.Sprintf("%-12s", truncate(item.QueueName, 12)))

		// Agent
		agentStr := lipgloss.NewStyle().
			Foreground(ColorSelect).
			Render(fmt.Sprintf("%-9s", truncate(item.AgentName, 9)))

		// Score
		scoreStr := lipgloss.NewStyle().
			Foreground(ColorFg).
			Bold(true).
			Render(fmt.Sprintf("%-6s", score))

		// KDA
		kdaStr := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(fmt.Sprintf("K %2d D %2d A %2d", item.Kills, item.Deaths, item.Assists))

		// RR delta
		var rrStr string
		if item.HasRR {
			sign := "+"
			rrColor := ColorWin
			if item.RREarned < 0 {
				sign = ""
				rrColor = ColorLoss
			} else if item.RREarned == 0 {
				rrColor = ColorDraw
			}
			rawRR := fmt.Sprintf("%s%2d RR", sign, item.RREarned)
			paddedRR := fmt.Sprintf("%-7s", rawRR)
			rrStr = lipgloss.NewStyle().
				Foreground(rrColor).
				Bold(true).
				Render(paddedRR)
		} else {
			rrStr = "       "
		}

		// Relative Time
		timeStr := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render(formatTimeAgo(item.GameTime))

		row := fmt.Sprintf("%s%s  %s  %s  %s  %s  %s   %s  %s",
			cursor, outcomeBadge, mapStr, queueStr, agentStr, scoreStr, kdaStr, rrStr, timeStr)

		if i == m.cursor {
			row = lipgloss.NewStyle().
				Background(lipgloss.Color("#1F2430")).
				Render(row)
		}

		sb.WriteString(row + "\n")
	}

	sb.WriteString("\n  " + RenderKeyLegends(
		[2]string{"↑/↓", "select"},
		[2]string{"enter", "view match detail"},
	))

	return sb.String()
}

// ── Detail View ─────────────────────────────────────────────────────

func (m MatchesModel) renderDetailView() string {
	if m.cursor >= len(m.items) {
		return "Invalid match selection."
	}
	item := m.items[m.cursor]
	if item.Details == nil {
		return fmt.Sprintf("  Loading match detail for %s...\n\n  %s", item.MatchID, RenderKeyItem("esc", "go back"))
	}

	d := item.Details
	var sb strings.Builder

	isDeathmatch := strings.EqualFold(item.QueueName, "Deathmatch") || (d != nil && d.IsDeathmatch())

	outcome := item.Outcome
	score := item.Score
	if isDeathmatch && d != nil {
		if score == "" || score == "0-0" {
			score = d.ScoreString(m.playerPUUID)
		}
		if outcome == "" || outcome == "DRAW" {
			outcome = d.GetMatchOutcome(m.playerPUUID)
		}
	}

	// Header
	outcomeColor := ColorDraw
	if outcome == "WIN" {
		outcomeColor = ColorWin
	} else if outcome == "LOSS" {
		outcomeColor = ColorLoss
	}

	headerOutcome := lipgloss.NewStyle().
		Foreground(outcomeColor).
		Bold(true).
		Render(fmt.Sprintf("%s  %s", outcome, score))

	header := fmt.Sprintf("  MATCH DETAIL — %s (%s)", item.MapName, item.QueueName)
	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(ColorFg).Render(header))
	sb.WriteString("    " + headerOutcome + "\n")
	sb.WriteString("  " + strings.Repeat("─", max(m.width-4, 70)) + "\n")

	if isDeathmatch {
		sb.WriteString(m.renderDeathmatchTable(d))
	} else {
		myTeam := d.GetPlayerTeam(m.playerPUUID)
		myTeamID := "Blue"
		if myTeam != nil {
			myTeamID = myTeam.TeamID
		}

		// Render Friendly Team
		sb.WriteString(m.renderTeamTable(d, myTeamID, true))
		sb.WriteString("\n")

		// Render Opponent Team
		oppTeam := d.GetOpponentTeam(myTeamID)
		oppTeamID := "Red"
		if oppTeam != nil {
			oppTeamID = oppTeam.TeamID
		}
		sb.WriteString(m.renderTeamTable(d, oppTeamID, false))
		sb.WriteString("\n")

		// Render Round Timeline
		sb.WriteString(m.renderRoundTimeline(d, myTeamID))
	}

	sb.WriteString("\n\n  " + RenderKeyItem("esc", "return to match list"))

	return sb.String()
}

func (m MatchesModel) renderTeamTable(d *models.MatchDetails, teamID string, isMyTeam bool) string {
	var sb strings.Builder

	roundsWon := 0
	for _, t := range d.Teams {
		if strings.EqualFold(t.TeamID, teamID) {
			roundsWon = t.RoundsWon
			break
		}
	}

	teamLabel := fmt.Sprintf("  %s TEAM", strings.ToUpper(teamID))
	teamColor := lipgloss.Color("#5A9FE2")
	if strings.EqualFold(teamID, "Red") {
		teamColor = lipgloss.Color("#EF4444")
	}
	if isMyTeam {
		teamLabel += " (YOUR TEAM)"
	}

	teamHeader := lipgloss.NewStyle().Foreground(teamColor).Bold(true).Render(teamLabel) +
		lipgloss.NewStyle().Foreground(ColorMuted).Render(fmt.Sprintf(" — %d rounds", roundsWon))
	sb.WriteString(teamHeader + "\n")
	headerRow := fmt.Sprintf("  %-18s  %-9s  %3s  %3s %3s %3s   %4s   %4s   %4s",
		"Player", "Agent", "ACS", "K", "D", "A", "HS%", "ADR", "Econ")
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(headerRow) + "\n")
	sb.WriteString("  " + strings.Repeat("─", max(m.width-4, 70)) + "\n")

	type teamPlayerRow struct {
		player models.MatchPlayer
		acs    int
		adr    int
		hsPct  int
		econ   int
		name   string
		agent  string
	}

	var rows []teamPlayerRow
	for _, p := range d.Players {
		if strings.EqualFold(p.TeamID, teamID) {
			name := ""
			if p.Subject == m.playerPUUID {
				name = "▸ You"
			} else if p.GameName != "" {
				if p.TagLine != "" {
					name = fmt.Sprintf("%s#%s", p.GameName, p.TagLine)
				} else {
					name = p.GameName
				}
			} else {
				// Name Service returned nothing → player has hidden their name
				name = "<Hidden>"
			}

			agentName := "Agent"
			if m.agentsMap != nil {
				if a, ok := m.agentsMap[strings.ToLower(p.CharacterID)]; ok {
					agentName = a
				}
			}

			acs, adr, hsPct, econ := d.ComputePlayerAdvancedStats(p.Subject)
			rows = append(rows, teamPlayerRow{
				player: p,
				acs:    acs,
				adr:    adr,
				hsPct:  hsPct,
				econ:   econ,
				name:   name,
				agent:  agentName,
			})
		}
	}

	// Sort descending by ACS (highest ACS first); tie-breaker kills descending
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].acs != rows[j].acs {
			return rows[i].acs > rows[j].acs
		}
		return rows[i].player.Stats.Kills > rows[j].player.Stats.Kills
	})

	for _, row := range rows {
		p := row.player
		playerRow := fmt.Sprintf("  %s  %s  %3d  %3d %3d %3d   %3d%%   %4d   %4d",
			fitWidth(row.name, 18),
			fitWidth(row.agent, 9),
			row.acs,
			p.Stats.Kills, p.Stats.Deaths, p.Stats.Assists,
			row.hsPct,
			row.adr,
			row.econ,
		)

		if p.Subject == m.playerPUUID {
			playerRow = lipgloss.NewStyle().Foreground(ColorUltra).Bold(true).Render(playerRow)
		}
		sb.WriteString(playerRow + "\n")
	}
	return sb.String()
}

func (m MatchesModel) renderDeathmatchTable(d *models.MatchDetails) string {
	var sb strings.Builder

	headerRow := fmt.Sprintf("  %-18s  %-9s  %3s %3s %3s",
		"Player", "Agent", "K", "D", "A")
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render(headerRow) + "\n")
	sb.WriteString("  " + strings.Repeat("─", max(m.width-4, 70)) + "\n")

	leaderboard := d.GetDeathmatchLeaderboard()
	for _, entry := range leaderboard {
		p := entry.Player
		name := ""
		if p.Subject == m.playerPUUID {
			name = "▸ You"
		} else if p.GameName != "" {
			if p.TagLine != "" {
				name = fmt.Sprintf("%s#%s", p.GameName, p.TagLine)
			} else {
				name = p.GameName
			}
		} else {
			name = "<Hidden>"
		}

		agentName := "Agent"
		if m.agentsMap != nil {
			if a, ok := m.agentsMap[strings.ToLower(p.CharacterID)]; ok {
				agentName = a
			}
		}

		playerRow := fmt.Sprintf("  %s  %s  %3d %3d %3d",
			fitWidth(name, 18),
			fitWidth(agentName, 9),
			p.Stats.Kills, p.Stats.Deaths, p.Stats.Assists,
		)

		if p.Subject == m.playerPUUID {
			playerRow = lipgloss.NewStyle().Foreground(ColorUltra).Bold(true).Render(playerRow)
		}
		sb.WriteString(playerRow + "\n")
	}
	return sb.String()
}

func (m MatchesModel) renderRoundTimeline(d *models.MatchDetails, myTeamID string) string {
	if len(d.RoundResults) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorFg).Bold(true).Render("  ROUND TIMELINE") + "\n")
	sb.WriteString("  " + strings.Repeat("─", max(m.width-4, 70)) + "\n  ")

	for i, r := range d.RoundResults {
		won := strings.EqualFold(r.WinningTeam, myTeamID)
		symbol := "□"
		symColor := ColorLoss
		if won {
			symbol = "■"
			symColor = ColorWin
		}

		roundStr := lipgloss.NewStyle().
			Foreground(symColor).
			Render(fmt.Sprintf("R%-2d %s", i+1, symbol))

		sb.WriteString(roundStr + "  ")
		if (i+1)%10 == 0 && i < len(d.RoundResults)-1 {
			sb.WriteString("\n  ")
		}
	}
	return sb.String()
}

// ── Helpers ─────────────────────────────────────────────────────────

func formatTimeAgo(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	d := time.Since(t)
	if d < time.Minute {
		return "just now"
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm ago", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(d.Hours()))
	}
	return fmt.Sprintf("%dd ago", int(d.Hours()/24))
}

func padRight(s string, width int) string {
	w := lipgloss.Width(s)
	if w >= width {
		return s
	}
	return s + strings.Repeat(" ", width-w)
}

func fitWidth(s string, width int) string {
	return padRight(truncate(s, width), width)
}

func truncate(s string, maxLen int) string {
	if lipgloss.Width(s) <= maxLen {
		return s
	}
	if maxLen <= 2 {
		var sb strings.Builder
		for _, r := range s {
			if lipgloss.Width(sb.String()+string(r)) > maxLen {
				break
			}
			sb.WriteRune(r)
		}
		return sb.String()
	}
	targetWidth := maxLen - 2
	var sb strings.Builder
	for _, r := range s {
		if lipgloss.Width(sb.String()+string(r)) > targetWidth {
			break
		}
		sb.WriteRune(r)
	}
	return sb.String() + ".."
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
