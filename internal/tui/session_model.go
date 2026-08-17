package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
)

type SessionModel struct {
	startTime      time.Time
	initialMatchIDs map[string]bool
	sessionMatches []*models.MatchDetails
	playerPUUID    string
	width, height  int
}

func NewSessionModel(playerPUUID string) SessionModel {
	return SessionModel{
		startTime:       time.Now(),
		initialMatchIDs: make(map[string]bool),
		playerPUUID:     playerPUUID,
	}
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
			// If not in initial snapshot, it's a new match played during this session
			if !m.initialMatchIDs[d.MatchInfo.MatchID] {
				newMatches = append(newMatches, d)
			}
		}
	}
	m.sessionMatches = newMatches
}

func (m *SessionModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m SessionModel) Update(msg tea.Msg) (SessionModel, tea.Cmd) {
	return m, nil
}

func (m SessionModel) View() string {
	var sb strings.Builder

	lineWidth := max(m.width-4, 70)
	duration := time.Since(m.startTime)
	durationStr := formatDuration(duration)

	header := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  ⏱  SESSION TRACKER                                       Started %s ago", durationStr))
	sb.WriteString(header + "\n")
	sb.WriteString("  " + strings.Repeat("─", lineWidth) + "\n\n")

	// Calculate session aggregates
	gamesPlayed := len(m.sessionMatches)
	wins := 0
	losses := 0
	draws := 0
	totalKills := 0
	totalDeaths := 0
	totalAssists := 0
	totalScore := 0
	totalRounds := 0
	totalDamage := 0
	totalHS := 0
	totalShots := 0

	for _, md := range m.sessionMatches {
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

	avgK := 0.0
	avgD := 0.0
	avgA := 0.0
	kd := 0.0
	avgACS := 0
	avgADR := 0
	avgHS := 0

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

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(1, 2).
		Width(min(lineWidth, 68))

	var cardContent strings.Builder

	cardContent.WriteString(fmt.Sprintf("Games Played:     %d\n", gamesPlayed))
	cardContent.WriteString(fmt.Sprintf("Record:           %s\n",
		lipgloss.NewStyle().Foreground(recordColor).Bold(true).Render(fmt.Sprintf("%dW - %dL - %dD (%d%% WR)", wins, losses, draws, winRate))))

	cardContent.WriteString(strings.Repeat("─", 40) + "\n")
	cardContent.WriteString(fmt.Sprintf("Avg K / D / A:    %.1f / %.1f / %.1f  (K/D: %.2f)\n", avgK, avgD, avgA, kd))
	cardContent.WriteString(fmt.Sprintf("Avg ACS:          %d\n", avgACS))
	cardContent.WriteString(fmt.Sprintf("Avg HS%%:          %d%%\n", avgHS))
	cardContent.WriteString(fmt.Sprintf("Avg ADR:          %d\n", avgADR))
	cardContent.WriteString(strings.Repeat("─", 40) + "\n")
	cardContent.WriteString(fmt.Sprintf("Session Duration: %s\n", durationStr))

	sb.WriteString(cardStyle.Render(cardContent.String()))
	sb.WriteString("\n\n" + lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  Session auto-tracks new matches played while Val-Tracker is running."))

	return sb.String()
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
