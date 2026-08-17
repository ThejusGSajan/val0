package models

import (
	"fmt"
	"strings"
)

// MatchHistoryResponse is returned by GET /match-history/v1/history/{puuid}
type MatchHistoryResponse struct {
	Subject    string         `json:"Subject"`
	BeginIndex int            `json:"BeginIndex"`
	EndIndex   int            `json:"EndIndex"`
	Total      int            `json:"Total"`
	History    []MatchSummary `json:"History"`
}

type MatchSummary struct {
	MatchID       string `json:"MatchID"`
	GameStartTime int64  `json:"GameStartTime"` // epoch ms
	QueueID       string `json:"QueueID"`
}

// MatchDetails is returned by GET /match-details/v1/matches/{matchId}
type MatchDetails struct {
	MatchInfo    MatchInfo     `json:"matchInfo"`
	Players      []MatchPlayer `json:"players"`
	Teams        []MatchTeam   `json:"teams"`
	RoundResults []RoundResult `json:"roundResults"`
}

type MatchInfo struct {
	MatchID          string `json:"matchId"`
	MapID            string `json:"mapId"` // "/Game/Maps/Ascent/Ascent"
	GameLengthMillis int    `json:"gameLengthMillis"`
	GameStartMillis  int64  `json:"gameStartMillis"`
	IsRanked         bool   `json:"isRanked"`
	QueueID          string `json:"queueID"`
	GameMode         string `json:"gameMode"`
	SeasonID         string `json:"seasonId"`
}

type MatchPlayer struct {
	Subject         string      `json:"subject"`
	GameName        string      `json:"gameName"`
	TagLine         string      `json:"tagLine"`
	TeamID          string      `json:"teamId"`
	CharacterID     string      `json:"characterId"` // agent UUID
	Stats           PlayerStats `json:"stats"`
	CompetitiveTier int         `json:"competitiveTier"`
	AccountLevel    int         `json:"accountLevel"`
}

type PlayerStats struct {
	Score          int          `json:"score"`
	RoundsPlayed   int          `json:"roundsPlayed"`
	Kills          int          `json:"kills"`
	Deaths         int          `json:"deaths"`
	Assists        int          `json:"assists"`
	PlaytimeMillis int          `json:"playtimeMillis"`
	AbilityCasts   AbilityCasts `json:"abilityCasts"`
}

type AbilityCasts struct {
	GrenadeCasts  int `json:"grenadeCasts"`
	Ability1Casts int `json:"ability1Casts"`
	Ability2Casts int `json:"ability2Casts"`
	UltimateCasts int `json:"ultimateCasts"`
}

type MatchTeam struct {
	TeamID       string `json:"teamId"` // "Blue", "Red"
	Won          bool   `json:"won"`
	RoundsPlayed int    `json:"roundsPlayed"`
	RoundsWon    int    `json:"roundsWon"`
}

type RoundResult struct {
	RoundNum    int               `json:"roundNum"`
	RoundResult string            `json:"roundResult"` // "Eliminated", "Bomb defused", "Bomb detonated"
	WinningTeam string            `json:"winningTeam"`
	BombPlanter string            `json:"bombPlanter"`
	BombDefuser string            `json:"bombDefuser"`
	PlantSite   string            `json:"plantSite"`
	PlayerStats []RoundPlayerStat `json:"playerStats"`
}

type RoundPlayerStat struct {
	Subject string        `json:"subject"`
	Kills   []RoundKill   `json:"kills"`
	Damage  []RoundDamage `json:"damage"`
	Score   int           `json:"score"`
	Economy RoundEconomy  `json:"economy"`
}

type RoundKill struct {
	Killer          string          `json:"killer"`
	Victim          string          `json:"victim"`
	FinishingDamage FinishingDamage `json:"finishingDamage"`
}

type FinishingDamage struct {
	DamageType string `json:"damageType"`
	DamageItem string `json:"damageItem"` // weapon UUID
}

type RoundDamage struct {
	Receiver  string `json:"receiver"`
	Damage    int    `json:"damage"`
	Legshots  int    `json:"legshots"`
	Bodyshots int    `json:"bodyshots"`
	Headshots int    `json:"headshots"`
}

type RoundEconomy struct {
	LoadoutValue int    `json:"loadoutValue"`
	Weapon       string `json:"weapon"`
	Armor        string `json:"armor"`
	Remaining    int    `json:"remaining"`
	Spent        int    `json:"spent"`
}

type PlayerNameResponse struct {
	DisplayName string `json:"DisplayName"`
	Subject     string `json:"Subject"`
	GameName    string `json:"GameName"`
	TagLine     string `json:"TagLine"`
}

// GetPlayer finds player by subject PUUID.
func (m *MatchDetails) GetPlayer(puuid string) *MatchPlayer {
	if m == nil {
		return nil
	}
	for i := range m.Players {
		if m.Players[i].Subject == puuid {
			return &m.Players[i]
		}
	}
	return nil
}

// GetPlayerTeam returns the team for the given player PUUID.
func (m *MatchDetails) GetPlayerTeam(puuid string) *MatchTeam {
	p := m.GetPlayer(puuid)
	if p == nil {
		return nil
	}
	for i := range m.Teams {
		if strings.EqualFold(m.Teams[i].TeamID, p.TeamID) {
			return &m.Teams[i]
		}
	}
	return nil
}

// GetOpponentTeam returns the other team.
func (m *MatchDetails) GetOpponentTeam(playerTeamID string) *MatchTeam {
	for i := range m.Teams {
		if !strings.EqualFold(m.Teams[i].TeamID, playerTeamID) {
			return &m.Teams[i]
		}
	}
	return nil
}

// GetMatchOutcome returns "WIN", "LOSS", or "DRAW" for the given player PUUID.
func (m *MatchDetails) GetMatchOutcome(puuid string) string {
	myTeam := m.GetPlayerTeam(puuid)
	if myTeam == nil {
		return "DRAW"
	}
	if len(m.Teams) == 2 {
		oppTeam := m.GetOpponentTeam(myTeam.TeamID)
		if oppTeam != nil {
			if myTeam.RoundsWon > oppTeam.RoundsWon {
				return "WIN"
			} else if myTeam.RoundsWon < oppTeam.RoundsWon {
				return "LOSS"
			}
			return "DRAW"
		}
	}
	if myTeam.Won {
		return "WIN"
	}
	return "LOSS"
}

// ScoreString returns e.g. "13-7" (player team rounds - enemy team rounds).
func (m *MatchDetails) ScoreString(puuid string) string {
	myTeam := m.GetPlayerTeam(puuid)
	if myTeam == nil {
		if len(m.Teams) >= 2 {
			return fmt.Sprintf("%d-%d", m.Teams[0].RoundsWon, m.Teams[1].RoundsWon)
		}
		return "0-0"
	}
	oppTeam := m.GetOpponentTeam(myTeam.TeamID)
	oppWon := 0
	if oppTeam != nil {
		oppWon = oppTeam.RoundsWon
	}
	return fmt.Sprintf("%d-%d", myTeam.RoundsWon, oppWon)
}

// ComputePlayerAdvancedStats calculates ACS, ADR, HS%, and Avg Econ for a player PUUID.
func (m *MatchDetails) ComputePlayerAdvancedStats(puuid string) (acs int, adr int, hsPct int, econ int) {
	player := m.GetPlayer(puuid)
	if player == nil {
		return 0, 0, 0, 0
	}

	rounds := player.Stats.RoundsPlayed
	if rounds <= 0 {
		rounds = 1
	}

	// ACS = score / roundsPlayed
	acs = player.Stats.Score / rounds

	// Calculate ADR, HS%, and Econ across RoundResults
	totalDamage := 0
	headshots := 0
	bodyshots := 0
	legshots := 0
	totalLoadout := 0
	economyRounds := 0

	for _, rr := range m.RoundResults {
		for _, ps := range rr.PlayerStats {
			if ps.Subject == puuid {
				if ps.Economy.LoadoutValue > 0 {
					totalLoadout += ps.Economy.LoadoutValue
					economyRounds++
				}
				for _, dmg := range ps.Damage {
					totalDamage += dmg.Damage
					headshots += dmg.Headshots
					bodyshots += dmg.Bodyshots
					legshots += dmg.Legshots
				}
			}
		}
	}

	adr = totalDamage / rounds
	totalShots := headshots + bodyshots + legshots
	if totalShots > 0 {
		hsPct = (headshots * 100) / totalShots
	}
	if economyRounds > 0 {
		econ = totalLoadout / economyRounds
	}

	return acs, adr, hsPct, econ
}
