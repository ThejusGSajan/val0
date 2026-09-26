package models

import (
	"fmt"
	"sort"
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

// IsGauntletMode returns true if the queue or game mode represents Gauntlet (Ability Draft Arena).
func IsGauntletMode(queueID, gameMode string) bool {
	q := strings.ToLower(strings.TrimSpace(queueID))
	gm := strings.ToLower(strings.TrimSpace(gameMode))
	return q == "abilitydraftarena" || q == "abilitydraft" || q == "gauntlet" ||
		strings.Contains(gm, "abilitydraftarena") || strings.Contains(gm, "abilitydraft") || strings.Contains(gm, "gauntlet")
}

// IsGauntlet returns true if this match is a Gauntlet / Ability Draft Arena match.
func (m *MatchDetails) IsGauntlet() bool {
	if m == nil {
		return false
	}
	return IsGauntletMode(m.MatchInfo.QueueID, m.MatchInfo.GameMode)
}

// IsAbilityDraft is an alias for IsGauntlet.
func (m *MatchDetails) IsAbilityDraft() bool {
	return m.IsGauntlet()
}

// GauntletTeamEntry represents a consolidated team entry in a Gauntlet match with aggregated stats.
type GauntletTeamEntry struct {
	Team         MatchTeam
	Players      []MatchPlayer
	Rank         int
	TotalScore   int
	TotalKills   int
	TotalDeaths  int
	TotalAssists int
	IsMyTeam     bool
}

// GetGauntletLeaderboard calculates and returns all teams sorted 1st through 8th according to Gauntlet ranking rules.
func (m *MatchDetails) GetGauntletLeaderboard(myPUUID string) []GauntletTeamEntry {
	if m == nil || len(m.Teams) == 0 {
		return nil
	}

	teamMap := make(map[string]*GauntletTeamEntry)
	for _, t := range m.Teams {
		teamMap[strings.ToUpper(t.TeamID)] = &GauntletTeamEntry{
			Team:    t,
			Players: make([]MatchPlayer, 0, 2),
		}
	}

	for _, p := range m.Players {
		tID := strings.ToUpper(p.TeamID)
		entry, ok := teamMap[tID]
		if !ok {
			entry = &GauntletTeamEntry{
				Team:    MatchTeam{TeamID: p.TeamID},
				Players: make([]MatchPlayer, 0, 2),
			}
			teamMap[tID] = entry
		}
		entry.Players = append(entry.Players, p)
		entry.TotalScore += p.Stats.Score
		entry.TotalKills += p.Stats.Kills
		entry.TotalDeaths += p.Stats.Deaths
		entry.TotalAssists += p.Stats.Assists
		if p.Subject == myPUUID && myPUUID != "" {
			entry.IsMyTeam = true
		}
	}

	entries := make([]GauntletTeamEntry, 0, len(teamMap))
	for _, entry := range teamMap {
		entries = append(entries, *entry)
	}

	sort.SliceStable(entries, func(i, j int) bool {
		// 1. Match Won flag
		if entries[i].Team.Won != entries[j].Team.Won {
			return entries[i].Team.Won
		}
		// 2. Rounds Won
		if entries[i].Team.RoundsWon != entries[j].Team.RoundsWon {
			return entries[i].Team.RoundsWon > entries[j].Team.RoundsWon
		}
		// 3. Rounds Played
		if entries[i].Team.RoundsPlayed != entries[j].Team.RoundsPlayed {
			return entries[i].Team.RoundsPlayed > entries[j].Team.RoundsPlayed
		}
		// 4. Total Combat Score
		if entries[i].TotalScore != entries[j].TotalScore {
			return entries[i].TotalScore > entries[j].TotalScore
		}
		// 5. Total Kills
		if entries[i].TotalKills != entries[j].TotalKills {
			return entries[i].TotalKills > entries[j].TotalKills
		}
		// 6. Total Deaths (ascending)
		if entries[i].TotalDeaths != entries[j].TotalDeaths {
			return entries[i].TotalDeaths < entries[j].TotalDeaths
		}
		// 7. Total Assists
		if entries[i].TotalAssists != entries[j].TotalAssists {
			return entries[i].TotalAssists > entries[j].TotalAssists
		}
		// 8. Deterministic fallback: TeamID
		return entries[i].Team.TeamID < entries[j].Team.TeamID
	})

	for i := range entries {
		entries[i].Rank = i + 1
	}

	return entries
}

// GetGauntletRank returns the 1-based team placement (1 to 8) of a player in Gauntlet.
func (m *MatchDetails) GetGauntletRank(puuid string) int {
	if m == nil || puuid == "" {
		return 0
	}
	leaderboard := m.GetGauntletLeaderboard(puuid)
	for _, entry := range leaderboard {
		for _, p := range entry.Players {
			if p.Subject == puuid {
				return entry.Rank
			}
		}
	}
	return 0
}

// IsDeathmatch returns true if the match is a Free-For-All Deathmatch (excluding Team Deathmatch / Hurm).
func (m *MatchDetails) IsDeathmatch() bool {
	if m == nil {
		return false
	}
	queue := strings.ToLower(m.MatchInfo.QueueID)
	mode := strings.ToLower(m.MatchInfo.GameMode)
	if strings.Contains(queue, "hurm") || strings.Contains(mode, "hurm") {
		return false
	}
	return queue == "deathmatch" || strings.Contains(mode, "deathmatch")
}

// IsCasualMode returns true if the match is a casual or non-structured mode
// (Free-For-All Deathmatch, Team Deathmatch / Hurm, Escalation / GGTeam, Skirmish, Snowball Fight, or Gauntlet)
// which should not contribute to tactical/competitive agent performance metrics.
func (m *MatchDetails) IsCasualMode() bool {
	if m == nil {
		return false
	}
	if m.IsDeathmatch() || m.IsGauntlet() {
		return true
	}
	queue := strings.ToLower(m.MatchInfo.QueueID)
	mode := strings.ToLower(m.MatchInfo.GameMode)

	casualIdentifiers := []string{"deathmatch", "hurm", "ggteam", "skirmish", "snowball", "abilitydraftarena", "abilitydraft", "gauntlet"}
	for _, id := range casualIdentifiers {
		if queue == id || strings.Contains(mode, id) {
			return true
		}
	}
	return false
}


// FormatOrdinal returns the 1-based ordinal representation of a rank (e.g. 1 -> "1st", 2 -> "2nd", 3 -> "3rd", 11 -> "11th").
func FormatOrdinal(n int) string {
	if n <= 0 {
		return ""
	}
	switch n % 100 {
	case 11, 12, 13:
		return fmt.Sprintf("%dth", n)
	}
	switch n % 10 {
	case 1:
		return fmt.Sprintf("%dst", n)
	case 2:
		return fmt.Sprintf("%dnd", n)
	case 3:
		return fmt.Sprintf("%drd", n)
	default:
		return fmt.Sprintf("%dth", n)
	}
}

// DeathmatchPlayerEntry pairs a MatchPlayer with their computed ACS for Deathmatch leaderboard ranking.
type DeathmatchPlayerEntry struct {
	Player MatchPlayer
	ACS    int
}

// GetDeathmatchLeaderboard returns all players sorted according to Deathmatch ranking rules:
// 1. Descending Kills
// 2. Descending ACS (Combat Score / roundsPlayed)
// 3. Ascending Deaths (fewer deaths ranks higher)
// 4. Descending Assists
func (m *MatchDetails) GetDeathmatchLeaderboard() []DeathmatchPlayerEntry {
	if m == nil || len(m.Players) == 0 {
		return nil
	}
	entries := make([]DeathmatchPlayerEntry, len(m.Players))
	for i, p := range m.Players {
		rounds := p.Stats.RoundsPlayed
		if rounds <= 0 {
			rounds = 1
		}
		acs := p.Stats.Score / rounds
		entries[i] = DeathmatchPlayerEntry{
			Player: p,
			ACS:    acs,
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Player.Stats.Kills != entries[j].Player.Stats.Kills {
			return entries[i].Player.Stats.Kills > entries[j].Player.Stats.Kills
		}
		if entries[i].ACS != entries[j].ACS {
			return entries[i].ACS > entries[j].ACS
		}
		if entries[i].Player.Stats.Deaths != entries[j].Player.Stats.Deaths {
			return entries[i].Player.Stats.Deaths < entries[j].Player.Stats.Deaths
		}
		return entries[i].Player.Stats.Assists > entries[j].Player.Stats.Assists
	})

	return entries
}

// GetDeathmatchRank returns the 1-based placement (1 for 1st, 2 for 2nd, etc.) of a player in a Deathmatch.
// Returns 0 if the player is not found or match has no players.
func (m *MatchDetails) GetDeathmatchRank(puuid string) int {
	if m == nil || puuid == "" {
		return 0
	}
	leaderboard := m.GetDeathmatchLeaderboard()
	for idx, entry := range leaderboard {
		if entry.Player.Subject == puuid {
			return idx + 1
		}
	}
	return 0
}

// GetMatchOutcome returns "WIN", "LOSS", or "DRAW" for the given player PUUID.
func (m *MatchDetails) GetMatchOutcome(puuid string) string {
	if m.IsDeathmatch() {
		rank := m.GetDeathmatchRank(puuid)
		if rank == 1 {
			return "WIN"
		}
		if rank > 1 {
			return "LOSS"
		}
		return "DRAW"
	}

	if m.IsGauntlet() {
		rank := m.GetGauntletRank(puuid)
		if rank == 1 {
			return "WIN"
		}
		if rank > 1 {
			return "LOSS"
		}
		return "DRAW"
	}

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
	if m.IsDeathmatch() {
		rank := m.GetDeathmatchRank(puuid)
		if rank > 0 {
			return FormatOrdinal(rank)
		}
		return "0-0"
	}

	if m.IsGauntlet() {
		rank := m.GetGauntletRank(puuid)
		if rank > 0 {
			return FormatOrdinal(rank)
		}
		return "0-0"
	}

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
