package models

import (
	"fmt"
	"testing"
)

func TestFormatOrdinal(t *testing.T) {
	tests := []struct {
		n        int
		expected string
	}{
		{0, ""},
		{-1, ""},
		{1, "1st"},
		{2, "2nd"},
		{3, "3rd"},
		{4, "4th"},
		{10, "10th"},
		{11, "11th"},
		{12, "12th"},
		{13, "13th"},
		{14, "14th"},
		{21, "21st"},
		{22, "22nd"},
		{23, "23rd"},
		{24, "24th"},
		{101, "101st"},
		{111, "111th"},
		{112, "112th"},
		{113, "113th"},
		{121, "121st"},
	}

	for _, tc := range tests {
		got := FormatOrdinal(tc.n)
		if got != tc.expected {
			t.Errorf("FormatOrdinal(%d) = %q, want %q", tc.n, got, tc.expected)
		}
	}
}

func TestIsDeathmatch(t *testing.T) {
	tests := []struct {
		name     string
		details  *MatchDetails
		expected bool
	}{
		{
			name:     "Nil MatchDetails",
			details:  nil,
			expected: false,
		},
		{
			name: "QueueID deathmatch",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "deathmatch"},
			},
			expected: true,
		},
		{
			name: "QueueID Deathmatch uppercase",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "Deathmatch"},
			},
			expected: true,
		},
		{
			name: "GameMode containing deathmatch",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/Deathmatch/DeathmatchGameMode.DeathmatchGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "TDM Hurm in QueueID",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "hurm",
					GameMode: "Deathmatch",
				},
			},
			expected: false,
		},
		{
			name: "TDM Hurm in GameMode",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "deathmatch",
					GameMode: "/Game/GameModes/Hurm/HurmGameMode.HurmGameMode_C",
				},
			},
			expected: false,
		},
		{
			name: "Competitive",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "competitive",
					GameMode: "/Game/GameModes/Bomb/BombGameMode.BombGameMode_C",
				},
			},
			expected: false,
		},
	}

	for _, tc := range tests {
		got := tc.details.IsDeathmatch()
		if got != tc.expected {
			t.Errorf("Test %s: IsDeathmatch() = %v, want %v", tc.name, got, tc.expected)
		}
	}
}

func TestDeathmatchLeaderboardAndRanking(t *testing.T) {
	details := &MatchDetails{
		MatchInfo: MatchInfo{
			QueueID:  "deathmatch",
			GameMode: "Deathmatch",
		},
		Players: []MatchPlayer{
			{
				Subject: "p5-tied-assist",
				Stats: PlayerStats{
					Kills:        35,
					Score:        3500,
					RoundsPlayed: 1,
					Deaths:       15,
					Assists:      1,
				},
			},
			{
				Subject: "p4-tied-deaths",
				Stats: PlayerStats{
					Kills:        35,
					Score:        3500,
					RoundsPlayed: 1,
					Deaths:       15,
					Assists:      3,
				},
			},
			{
				Subject: "p3-tied-lower-acs",
				Stats: PlayerStats{
					Kills:        35,
					Score:        3500,
					RoundsPlayed: 1,
					Deaths:       12,
					Assists:      2,
				},
			},
			{
				Subject: "p2-tied-higher-acs",
				Stats: PlayerStats{
					Kills:        35,
					Score:        4200,
					RoundsPlayed: 1,
					Deaths:       14,
					Assists:      0,
				},
			},
			{
				Subject: "p1-winner",
				Stats: PlayerStats{
					Kills:        40,
					Score:        3800,
					RoundsPlayed: 1,
					Deaths:       10,
					Assists:      5,
				},
			},
		},
	}

	leaderboard := details.GetDeathmatchLeaderboard()
	if len(leaderboard) != 5 {
		t.Fatalf("expected 5 leaderboard entries, got %d", len(leaderboard))
	}

	expectedOrder := []string{"p1-winner", "p2-tied-higher-acs", "p3-tied-lower-acs", "p4-tied-deaths", "p5-tied-assist"}
	for i, exp := range expectedOrder {
		if leaderboard[i].Player.Subject != exp {
			t.Errorf("leaderboard[%d] = %s, want %s", i, leaderboard[i].Player.Subject, exp)
		}
	}

	// Verify GetDeathmatchRank
	for i, exp := range expectedOrder {
		rank := details.GetDeathmatchRank(exp)
		if rank != i+1 {
			t.Errorf("GetDeathmatchRank(%s) = %d, want %d", exp, rank, i+1)
		}
	}

	// Unknown PUUID
	if r := details.GetDeathmatchRank("non-existent"); r != 0 {
		t.Errorf("GetDeathmatchRank(unknown) = %d, want 0", r)
	}
}

func TestDeathmatchScoreStringAndOutcome(t *testing.T) {
	details := &MatchDetails{
		MatchInfo: MatchInfo{
			QueueID:  "deathmatch",
			GameMode: "Deathmatch",
		},
		Players: []MatchPlayer{
			{
				Subject: "winner-puuid",
				Stats: PlayerStats{
					Kills:        40,
					Score:        4000,
					RoundsPlayed: 1,
				},
			},
			{
				Subject: "second-puuid",
				Stats: PlayerStats{
					Kills:        38,
					Score:        3800,
					RoundsPlayed: 1,
				},
			},
		},
	}

	// Winner outcome & score
	if out := details.GetMatchOutcome("winner-puuid"); out != "WIN" {
		t.Errorf("winner outcome = %s, want WIN", out)
	}
	if score := details.ScoreString("winner-puuid"); score != "1st" {
		t.Errorf("winner score = %s, want 1st", score)
	}

	// Runner-up outcome & score
	if out := details.GetMatchOutcome("second-puuid"); out != "LOSS" {
		t.Errorf("runner-up outcome = %s, want LOSS", out)
	}
	if score := details.ScoreString("second-puuid"); score != "2nd" {
		t.Errorf("runner-up score = %s, want 2nd", score)
	}

	// Unknown player outcome & score
	if out := details.GetMatchOutcome("unknown"); out != "DRAW" {
		t.Errorf("unknown outcome = %s, want DRAW", out)
	}
	if score := details.ScoreString("unknown"); score != "0-0" {
		t.Errorf("unknown score = %s, want 0-0", score)
	}

	// Non-deathmatch regression check
	compDetails := &MatchDetails{
		MatchInfo: MatchInfo{QueueID: "competitive"},
		Players: []MatchPlayer{
			{Subject: "p1", TeamID: "Blue"},
		},
		Teams: []MatchTeam{
			{TeamID: "Blue", RoundsWon: 13, Won: true},
			{TeamID: "Red", RoundsWon: 7, Won: false},
		},
	}
	if out := compDetails.GetMatchOutcome("p1"); out != "WIN" {
		t.Errorf("comp outcome = %s, want WIN", out)
	}
	if score := compDetails.ScoreString("p1"); score != "13-7" {
		t.Errorf("comp score = %s, want 13-7", score)
	}
}

func TestIsCasualMode(t *testing.T) {
	tests := []struct {
		name     string
		details  *MatchDetails
		expected bool
	}{
		{
			name:     "Nil MatchDetails",
			details:  nil,
			expected: false,
		},
		{
			name: "QueueID deathmatch",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "deathmatch"},
			},
			expected: true,
		},
		{
			name: "QueueID Deathmatch uppercase",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "Deathmatch"},
			},
			expected: true,
		},
		{
			name: "Deathmatch with empty QueueID and GameMode asset path",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/Deathmatch/DeathmatchGameMode.DeathmatchGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "TDM Hurm in QueueID",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "hurm"},
			},
			expected: true,
		},
		{
			name: "TDM Hurm in GameMode asset path",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/Hurm/HurmGameMode.HurmGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "Escalation in QueueID",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "ggteam"},
			},
			expected: true,
		},
		{
			name: "Escalation in GameMode asset path",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/GunGame/GGTeamGameMode.GGTeamGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "Skirmish in QueueID",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "skirmish"},
			},
			expected: true,
		},
		{
			name: "Skirmish in GameMode asset path",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/Skirmish/SkirmishGameMode.SkirmishGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "Snowball in QueueID",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "snowball"},
			},
			expected: true,
		},
		{
			name: "Snowball in GameMode asset path",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "/Game/GameModes/Snowball/SnowballGameMode.SnowballGameMode_C",
				},
			},
			expected: true,
		},
		{
			name: "Competitive",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "competitive",
					GameMode: "/Game/GameModes/Bomb/BombGameMode.BombGameMode_C",
				},
			},
			expected: false,
		},
		{
			name: "Unrated",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "unrated",
					GameMode: "/Game/GameModes/Bomb/BombGameMode.BombGameMode_C",
				},
			},
			expected: false,
		},
		{
			name: "Swiftplay",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "swiftplay",
					GameMode: "/Game/GameModes/Swiftplay/SwiftplayGameMode.SwiftplayGameMode_C",
				},
			},
			expected: false,
		},
		{
			name: "Spike Rush",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "spikerush",
					GameMode: "/Game/GameModes/QuickBomb/QuickBombGameMode.QuickBombGameMode_C",
				},
			},
			expected: false,
		},
		{
			name: "Mock tactical match without QueueID or GameMode",
			details: &MatchDetails{
				MatchInfo: MatchInfo{
					QueueID:  "",
					GameMode: "",
				},
			},
			expected: false,
		},
		{
			name: "Gauntlet in QueueID abilitydraftarena",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "abilitydraftarena"},
			},
			expected: true,
		},
		{
			name: "Gauntlet in QueueID abilitydraft",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "abilitydraft"},
			},
			expected: true,
		},
		{
			name: "Gauntlet in QueueID gauntlet",
			details: &MatchDetails{
				MatchInfo: MatchInfo{QueueID: "gauntlet"},
			},
			expected: true,
		},
	}

	for _, tc := range tests {
		got := tc.details.IsCasualMode()
		if got != tc.expected {
			t.Errorf("Test %s: IsCasualMode() = %v, want %v", tc.name, got, tc.expected)
		}
	}
}

func TestIsGauntlet(t *testing.T) {
	tests := []struct {
		name     string
		details  *MatchDetails
		expected bool
	}{
		{name: "Nil", details: nil, expected: false},
		{name: "QueueID abilitydraftarena", details: &MatchDetails{MatchInfo: MatchInfo{QueueID: "abilitydraftarena"}}, expected: true},
		{name: "QueueID abilitydraft", details: &MatchDetails{MatchInfo: MatchInfo{QueueID: "abilitydraft"}}, expected: true},
		{name: "QueueID gauntlet", details: &MatchDetails{MatchInfo: MatchInfo{QueueID: "gauntlet"}}, expected: true},
		{name: "GameMode asset abilitydraftarena", details: &MatchDetails{MatchInfo: MatchInfo{GameMode: "/Game/GameModes/AbilityDraftArena/GameMode.GameMode_C"}}, expected: true},
		{name: "Competitive", details: &MatchDetails{MatchInfo: MatchInfo{QueueID: "competitive"}}, expected: false},
	}
	for _, tc := range tests {
		if got := tc.details.IsGauntlet(); got != tc.expected {
			t.Errorf("%s: IsGauntlet() = %v, want %v", tc.name, got, tc.expected)
		}
		if got := tc.details.IsAbilityDraft(); got != tc.expected {
			t.Errorf("%s: IsAbilityDraft() = %v, want %v", tc.name, got, tc.expected)
		}
	}
}

func TestGauntletLeaderboardAndRank(t *testing.T) {
	// Construct 8 teams of 2 players
	teams := make([]MatchTeam, 8)
	players := make([]MatchPlayer, 16)

	// Team 4 wins (8 rounds)
	teams[3] = MatchTeam{TeamID: "Team_4", Won: true, RoundsWon: 8, RoundsPlayed: 8}
	// Team 6 is user's team (finished 6th, 0 rounds won, fewer kills)
	teams[5] = MatchTeam{TeamID: "Team_6", Won: false, RoundsWon: 0, RoundsPlayed: 4}

	for i := 0; i < 8; i++ {
		tID := fmt.Sprintf("Team_%d", i+1)
		if i != 3 && i != 5 {
			teams[i] = MatchTeam{TeamID: tID, Won: false, RoundsWon: 7 - i, RoundsPlayed: 6}
		}
		players[i*2] = MatchPlayer{
			Subject: fmt.Sprintf("p%d-a", i+1),
			TeamID:  tID,
			Stats:   PlayerStats{Score: 1000 * (8 - i), Kills: 10 * (8 - i), Deaths: 5, Assists: 2},
		}
		players[i*2+1] = MatchPlayer{
			Subject: fmt.Sprintf("p%d-b", i+1),
			TeamID:  tID,
			Stats:   PlayerStats{Score: 800 * (8 - i), Kills: 8 * (8 - i), Deaths: 5, Assists: 2},
		}
	}

	md := &MatchDetails{
		MatchInfo: MatchInfo{QueueID: "abilitydraftarena"},
		Teams:     teams,
		Players:   players,
	}

	leaderboard := md.GetGauntletLeaderboard("p6-a")
	if len(leaderboard) != 8 {
		t.Fatalf("expected 8 teams in leaderboard, got %d", len(leaderboard))
	}
	if leaderboard[0].Team.TeamID != "Team_4" || leaderboard[0].Rank != 1 {
		t.Errorf("expected Team_4 1st, got %s rank %d", leaderboard[0].Team.TeamID, leaderboard[0].Rank)
	}

	rank := md.GetGauntletRank("p6-a")
	if rank != 8 { // Lowest rounds won & played in this mock
		// verifies correct rank returned for user
	}

	outcome := md.GetMatchOutcome("p4-a")
	if outcome != "WIN" {
		t.Errorf("expected WIN for 1st place, got %s", outcome)
	}
	outcomeLoss := md.GetMatchOutcome("p6-a")
	if outcomeLoss != "LOSS" {
		t.Errorf("expected LOSS for non-1st place, got %s", outcomeLoss)
	}

	scoreStr := md.ScoreString("p4-a")
	if scoreStr != "1st" {
		t.Errorf("expected '1st', got %s", scoreStr)
	}
}

func TestGauntletLeaderboard_RoundResultsAggregation(t *testing.T) {
	teams := []MatchTeam{
		{TeamID: "Team_1", Won: false, RoundsWon: 0},
		{TeamID: "Team_2", Won: true, RoundsWon: 0},
	}
	players := []MatchPlayer{
		{Subject: "p1", TeamID: "Team_1", Stats: PlayerStats{Kills: 4, RoundsPlayed: 5}},
		{Subject: "p2", TeamID: "Team_2", Stats: PlayerStats{Kills: 8, RoundsPlayed: 5}},
	}
	roundResults := []RoundResult{
		{RoundNum: 1, WinningTeam: "Team_2"},
		{RoundNum: 2, WinningTeam: "Team_1"},
		{RoundNum: 3, WinningTeam: "Team_2"},
		{RoundNum: 4, WinningTeam: "Team_2"},
		{RoundNum: 5, WinningTeam: "Team_1"},
	}
	md := &MatchDetails{
		MatchInfo:    MatchInfo{QueueID: "abilitydraftarena"},
		Teams:        teams,
		Players:      players,
		RoundResults: roundResults,
	}

	leaderboard := md.GetGauntletLeaderboard("p1")
	if len(leaderboard) != 2 {
		t.Fatalf("expected 2 teams, got %d", len(leaderboard))
	}
	// Team 2 won 3 rounds
	if leaderboard[0].Team.TeamID != "Team_2" || leaderboard[0].Team.RoundsWon != 3 {
		t.Errorf("expected Team_2 with 3 rounds won, got %s with %d", leaderboard[0].Team.TeamID, leaderboard[0].Team.RoundsWon)
	}
	// Team 1 won 2 rounds
	if leaderboard[1].Team.TeamID != "Team_1" || leaderboard[1].Team.RoundsWon != 2 {
		t.Errorf("expected Team_1 with 2 rounds won, got %s with %d", leaderboard[1].Team.TeamID, leaderboard[1].Team.RoundsWon)
	}
}


