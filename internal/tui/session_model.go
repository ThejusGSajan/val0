package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
)

// SessionFilterMode represents the active game mode filter for the session.
type SessionFilterMode int

const (
	FilterModeCompetitive SessionFilterMode = 0 // Competitive Only
	FilterModeAllAllowed  SessionFilterMode = 1 // Comp + Unrated + Swiftplay + Spike Rush
)

// IsEligibleSessionMode checks if a match qualifies under the given filter mode.
// Mode 0: strictly competitive.
// Mode 1: competitive, unrated, swiftplay, spikerush.
// All other game modes (Deathmatch, Hurm/TDM, Escalation, Replication, Custom, etc.) return false in BOTH modes.
func IsEligibleSessionMode(queueID, gameMode string, mode SessionFilterMode) bool {
	q := strings.ToLower(strings.TrimSpace(queueID))
	if q == "" && gameMode != "" {
		gm := strings.ToLower(gameMode)
		switch {
		case strings.Contains(gm, "competitive"):
			q = "competitive"
		case strings.Contains(gm, "swiftplay"):
			q = "swiftplay"
		case strings.Contains(gm, "spikerush"):
			q = "spikerush"
		case strings.Contains(gm, "unrated"):
			q = "unrated"
		}
	}

	if mode == FilterModeCompetitive {
		return q == "competitive"
	}

	// FilterModeAllAllowed: only comp + unrated + swiftplay + spikerush
	return q == "competitive" || q == "unrated" || q == "swiftplay" || q == "spikerush"
}

type SessionModel struct {
	startTime         time.Time
	initialMatchIDs   map[string]bool
	allSessionMatches []*models.MatchDetails
	filterMode        SessionFilterMode
	playerPUUID       string
	rrMap             map[string]int
	currentRank       string
	currentTier       int
	currentRR         int
	agentsMap         map[string]string
	mapsMap           map[string]string
	width, height     int
	scrollOffset      int
	flashMsg          string
}

func NewSessionModel(playerPUUID string) SessionModel {
	return SessionModel{
		startTime:       time.Now(),
		initialMatchIDs: make(map[string]bool),
		playerPUUID:     playerPUUID,
		filterMode:      FilterModeCompetitive,
		rrMap:           make(map[string]int),
		agentsMap:       cache.DefaultAgentNames,
		mapsMap:         cache.DefaultMapNames,
	}
}

func (m *SessionModel) SetSessionAnchor(startTime time.Time, initialIDs []string, filterMode SessionFilterMode) {
	m.startTime = startTime
	m.initialMatchIDs = make(map[string]bool)
	for _, id := range initialIDs {
		m.initialMatchIDs[id] = true
	}
	m.filterMode = filterMode
}

func (m *SessionModel) SetInitialSnapshot(matchIDs []string) {
	if len(m.initialMatchIDs) == 0 && len(matchIDs) > 0 {
		for _, id := range matchIDs {
			m.initialMatchIDs[id] = true
		}
	}
}

func (m *SessionModel) UpdateSessionMatches(currentDetails []*models.MatchDetails) {
	var newMatches []*models.MatchDetails
	for _, d := range currentDetails {
		if d != nil {
			if !m.initialMatchIDs[d.MatchInfo.MatchID] {
				newMatches = append(newMatches, d)
			}
		}
	}
	m.allSessionMatches = newMatches
}

func (m *SessionModel) SetCompetitiveUpdates(compUpdates *models.CompetitiveUpdatesResponse) {
	if compUpdates == nil {
		return
	}
	if m.rrMap == nil {
		m.rrMap = make(map[string]int)
	}
	for _, cu := range compUpdates.Matches {
		m.rrMap[cu.MatchID] = cu.RankedRatingEarned
	}
}

func (m *SessionModel) SetRankInfo(rankName string, tier, rr int) {
	m.currentRank = rankName
	m.currentTier = tier
	m.currentRR = rr
}

func (m *SessionModel) SetMetadataMaps(agents, maps map[string]string) {
	if agents != nil {
		m.agentsMap = agents
	}
	if maps != nil {
		m.mapsMap = maps
	}
}

func (m *SessionModel) ToggleFilterMode() SessionFilterMode {
	if m.filterMode == FilterModeCompetitive {
		m.filterMode = FilterModeAllAllowed
		m.flashMsg = "Filter: Comp + Unrated + Swiftplay + Spike Rush"
	} else {
		m.filterMode = FilterModeCompetitive
		m.flashMsg = "Filter: Competitive Only"
	}
	m.scrollOffset = 0
	return m.filterMode
}

func (m *SessionModel) SetFilterMode(mode SessionFilterMode) {
	m.filterMode = mode
	m.scrollOffset = 0
}

func (m *SessionModel) GetFilterMode() SessionFilterMode {
	return m.filterMode
}

func (m *SessionModel) ResetSession(currentMatchIDs []string) {
	m.startTime = time.Now()
	m.initialMatchIDs = make(map[string]bool)
	for _, id := range currentMatchIDs {
		m.initialMatchIDs[id] = true
	}
	m.allSessionMatches = nil
	m.scrollOffset = 0
	m.flashMsg = "Session reset!"
}

func (m *SessionModel) FilteredMatches() []*models.MatchDetails {
	var filtered []*models.MatchDetails
	for _, d := range m.allSessionMatches {
		if d != nil && IsEligibleSessionMode(d.MatchInfo.QueueID, d.MatchInfo.GameMode, m.filterMode) {
			filtered = append(filtered, d)
		}
	}
	return filtered
}

func (m *SessionModel) NetRR() int {
	net := 0
	filtered := m.FilteredMatches()
	for _, md := range filtered {
		if strings.EqualFold(md.MatchInfo.QueueID, "competitive") {
			if rr, ok := m.rrMap[md.MatchInfo.MatchID]; ok {
				net += rr
			}
		}
	}
	return net
}

func (m *SessionModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SessionModel) Update(msg tea.Msg) (SessionModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.scrollOffset > 0 {
				m.scrollOffset--
			}
			return m, nil
		case "down", "j":
			filtered := m.FilteredMatches()
			if m.scrollOffset < len(filtered)-1 {
				m.scrollOffset++
			}
			return m, nil
		}
	}
	return m, nil
}

func (m SessionModel) View() string {
	var sb strings.Builder

	lineWidth := max(m.width-4, 76)
	duration := time.Since(m.startTime)
	durationStr := formatDuration(duration)

	modeLabel := "[ Competitive Only ]"
	if m.filterMode == FilterModeAllAllowed {
		modeLabel = "[ Comp + Unrated + Swiftplay + Spike Rush ]"
	}

	modeBadge := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render(modeLabel)

	header := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  ⏱  SESSION TRACKER    Started %s ago    Mode: %s [m]    Reset [x]", durationStr, modeBadge))
	sb.WriteString(header + "\n")
	sb.WriteString("  " + strings.Repeat("─", lineWidth) + "\n\n")

	filtered := m.FilteredMatches()
	gamesPlayed := len(filtered)
	wins, losses, draws := 0, 0, 0
	totalKills, totalDeaths, totalAssists := 0, 0, 0
	totalScore, totalRounds, totalDamage := 0, 0, 0
	totalHS, totalShots := 0, 0

	for _, md := range filtered {
		outcome := md.GetMatchOutcome(m.playerPUUID)
		switch outcome {
		case "WIN":
			wins++
		case "LOSS":
			losses++
		case "DRAW":
			draws++
		}

		p := md.GetPlayer(m.playerPUUID)
		if p != nil {
			totalKills += p.Stats.Kills
			totalDeaths += p.Stats.Deaths
			totalAssists += p.Stats.Assists
			totalScore += p.Stats.Score
			totalRounds += max(p.Stats.RoundsPlayed, 1)

			for _, rr := range md.RoundResults {
				for _, ps := range rr.PlayerStats {
					if ps.Subject == m.playerPUUID {
						for _, dmg := range ps.Damage {
							totalDamage += dmg.Damage
							totalHS += dmg.Headshots
							totalShots += (dmg.Headshots + dmg.Bodyshots + dmg.Legshots)
						}
					}
				}
			}
		}
	}

	winRate := 0
	if gamesPlayed > 0 {
		winRate = (wins * 100) / gamesPlayed
	}

	avgK, avgD, avgA, kd := 0.0, 0.0, 0.0, 0.0
	avgACS, avgADR, avgHS := 0, 0, 0

	if gamesPlayed > 0 {
		avgK = float64(totalKills) / float64(gamesPlayed)
		avgD = float64(totalDeaths) / float64(gamesPlayed)
		avgA = float64(totalAssists) / float64(gamesPlayed)
		if totalDeaths > 0 {
			kd = float64(totalKills) / float64(totalDeaths)
		} else {
			kd = float64(totalKills)
		}
		if totalRounds > 0 {
			avgACS = totalScore / totalRounds
			avgADR = totalDamage / totalRounds
		}
		if totalShots > 0 {
			avgHS = (totalHS * 100) / totalShots
		}
	}

	recordColor := ColorFg
	if wins > losses {
		recordColor = ColorWin
	} else if losses > wins {
		recordColor = ColorLoss
	}

	// Net RR styling
	netRR := m.NetRR()
	netRRStr := fmt.Sprintf("%+d RR", netRR)
	netRRStyle := lipgloss.NewStyle().Bold(true)
	if netRR > 0 {
		netRRStyle = netRRStyle.Foreground(ColorWin)
	} else if netRR < 0 {
		netRRStyle = netRRStyle.Foreground(ColorLoss)
	} else {
		netRRStyle = netRRStyle.Foreground(ColorMuted)
		netRRStr = "0 RR"
	}

	rankInfoStr := m.currentRank
	if m.currentRR > 0 || m.currentTier > 0 {
		rankInfoStr = fmt.Sprintf("%s (%d RR)", m.currentRank, m.currentRR)
	}
	if rankInfoStr == "" {
		rankInfoStr = "Unranked"
	}

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2).
		Width(min(lineWidth, 72))

	var cardContent strings.Builder
	cardContent.WriteString(fmt.Sprintf("Games Played:     %d\n", gamesPlayed))
	cardContent.WriteString(fmt.Sprintf("Record:           %s\n",
		lipgloss.NewStyle().Foreground(recordColor).Bold(true).Render(fmt.Sprintf("%dW - %dL - %dD (%d%% WR)", wins, losses, draws, winRate))))
	cardContent.WriteString(fmt.Sprintf("Net RR:           %s\n", netRRStyle.Render(netRRStr)))
	cardContent.WriteString(fmt.Sprintf("Current Rank:     %s\n", lipgloss.NewStyle().Foreground(ColorExclusive).Bold(true).Render(rankInfoStr)))
	cardContent.WriteString(strings.Repeat("─", 44) + "\n")
	cardContent.WriteString(fmt.Sprintf("Avg K / D / A:    %.1f / %.1f / %.1f  (K/D: %.2f)\n", avgK, avgD, avgA, kd))
	cardContent.WriteString(fmt.Sprintf("Avg ACS:          %d\n", avgACS))
	cardContent.WriteString(fmt.Sprintf("Avg HS%%:          %d%%\n", avgHS))
	cardContent.WriteString(fmt.Sprintf("Avg ADR:          %d\n", avgADR))
	cardContent.WriteString(strings.Repeat("─", 44) + "\n")
	cardContent.WriteString(fmt.Sprintf("Session Duration: %s\n", durationStr))

	sb.WriteString(cardStyle.Render(cardContent.String()))
	sb.WriteString("\n\n")

	// ── Match Breakdown Table ──────────────────────────────────────────
	tableHeader := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Bold(true).
		Render("  OUTCOME   MAP          MODE          SCORE   AGENT       K/D/A       RR        TIME")
	sb.WriteString(fmt.Sprintf("  ── Session Matches (%d) %s\n", gamesPlayed, strings.Repeat("─", max(lineWidth-24-len(fmt.Sprintf("%d", gamesPlayed)), 10))))
	sb.WriteString(tableHeader + "\n")

	if gamesPlayed == 0 {
		sb.WriteString("\n" + lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("    No matches recorded yet in this session.\n    Play a match or press [m] to switch mode filter.") + "\n")
	} else {
		// Available vertical space for matches
		maxVisibleMatches := 8
		if m.height > 30 {
			maxVisibleMatches = m.height - 22
		}

		startIndex := m.scrollOffset
		if startIndex > gamesPlayed-1 {
			startIndex = 0
		}
		endIndex := min(startIndex+maxVisibleMatches, gamesPlayed)

		for i := startIndex; i < endIndex; i++ {
			md := filtered[i]
			outcome := md.GetMatchOutcome(m.playerPUUID)
			var outcomeStyled string
			switch outcome {
			case "WIN":
				outcomeStyled = lipgloss.NewStyle().Foreground(ColorWin).Bold(true).Render(" WIN ")
			case "LOSS":
				outcomeStyled = lipgloss.NewStyle().Foreground(ColorLoss).Bold(true).Render(" LOSS")
			default:
				outcomeStyled = lipgloss.NewStyle().Foreground(ColorDraw).Bold(true).Render(" DRAW")
			}

			mapName := cache.GetMapName(md.MatchInfo.MapID, m.mapsMap)
			if mapName == "" {
				mapName = "Unknown"
			}
			modeName := ResolveQueueDisplayName(md.MatchInfo.QueueID, md.MatchInfo.GameMode)
			scoreStr := md.ScoreString(m.playerPUUID)

			agentName := "Unknown"
			kdaStr := "0/0/0"
			p := md.GetPlayer(m.playerPUUID)
			if p != nil {
				agentName = cache.GetAgentName(p.CharacterID, m.agentsMap)
				kdaStr = fmt.Sprintf("%d/%d/%d", p.Stats.Kills, p.Stats.Deaths, p.Stats.Assists)
			}

			rrCol := "--"
			if strings.EqualFold(md.MatchInfo.QueueID, "competitive") {
				if rVal, ok := m.rrMap[md.MatchInfo.MatchID]; ok {
					if rVal > 0 {
						rrCol = lipgloss.NewStyle().Foreground(ColorWin).Render(fmt.Sprintf("+%d RR", rVal))
					} else if rVal < 0 {
						rrCol = lipgloss.NewStyle().Foreground(ColorLoss).Render(fmt.Sprintf("%d RR", rVal))
					} else {
						rrCol = "0 RR"
					}
				}
			}

			timeStr := ""
			if md.MatchInfo.GameStartMillis > 0 {
				matchStart := time.UnixMilli(md.MatchInfo.GameStartMillis)
				timeStr = formatRelativeTime(matchStart)
			}

			row := fmt.Sprintf("  %-7s  %-11s  %-12s  %-6s  %-10s  %-10s  %-8s  %s",
				outcomeStyled,
				truncateString(mapName, 11),
				truncateString(modeName, 12),
				truncateString(scoreStr, 6),
				truncateString(agentName, 10),
				truncateString(kdaStr, 10),
				rrCol,
				timeStr,
			)
			sb.WriteString(row + "\n")
		}
	}

	if m.flashMsg != "" {
		sb.WriteString("\n  " + lipgloss.NewStyle().Foreground(ColorAccent).Italic(true).Render(m.flashMsg) + "\n")
	}

	sb.WriteString("\n" + lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  [m] toggle mode  •  [x] reset session  •  [↑/↓] scroll matches  •  auto-polls every 30s"))

	return sb.String()
}

func formatRelativeTime(t time.Time) string {
	elapsed := time.Since(t)
	if elapsed < time.Minute {
		return "just now"
	}
	if elapsed < time.Hour {
		return fmt.Sprintf("%dm ago", int(elapsed.Minutes()))
	}
	if elapsed < 24*time.Hour {
		return fmt.Sprintf("%dh ago", int(elapsed.Hours()))
	}
	return t.Format("01/02")
}

func truncateString(s string, maxLen int) string {
	if len(s) > maxLen {
		return s[:maxLen-1] + "…"
	}
	return s
}

func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm", int(d.Minutes()))
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	return fmt.Sprintf("%dh %dm", h, m)
}
