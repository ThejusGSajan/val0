package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

// ── IsEligibleSessionMode ──────────────────────────────────────────────────

func TestIsEligibleSessionMode_CompetitiveOnly(t *testing.T) {
	tests := []struct {
		queueID  string
		gameMode string
		want     bool
	}{
		{"competitive", "", true},
		{"COMPETITIVE", "", true},
		{"unrated", "", false},
		{"swiftplay", "", false},
		{"spikerush", "", false},
		{"deathmatch", "", false},
		{"hurm", "", false},
		{"", "Competitive", true},
		{"", "Unrated", false},
		{"", "Deathmatch FFA", false},
		{"", "", false},
	}
	for _, tc := range tests {
		got := IsEligibleSessionMode(tc.queueID, tc.gameMode, FilterModeCompetitive)
		if got != tc.want {
			t.Errorf("IsEligibleSessionMode(%q, %q, Competitive) = %v, want %v",
				tc.queueID, tc.gameMode, got, tc.want)
		}
	}
}

func TestIsEligibleSessionMode_AllAllowed(t *testing.T) {
	tests := []struct {
		queueID  string
		gameMode string
		want     bool
	}{
		{"competitive", "", true},
		{"unrated", "", true},
		{"swiftplay", "", true},
		{"spikerush", "", true},
		{"deathmatch", "", false},
		{"hurm", "", false},
		{"escalation", "", false},
		{"custom", "", false},
		{"", "Competitive", true},
		{"", "SwiftPlay", true},
		{"", "Spike Rush", false}, // "spike rush" not matching "spikerush" exact substring
		{"", "Spikerush", true},
		{"", "Deathmatch FFA", false},
	}
	for _, tc := range tests {
		got := IsEligibleSessionMode(tc.queueID, tc.gameMode, FilterModeAllAllowed)
		if got != tc.want {
			t.Errorf("IsEligibleSessionMode(%q, %q, AllAllowed) = %v, want %v",
				tc.queueID, tc.gameMode, got, tc.want)
		}
	}
}

// ── Mode Toggling ──────────────────────────────────────────────────────────

func TestToggleFilterMode(t *testing.T) {
	sm := NewSessionModel("test-puuid")
	if sm.GetFilterMode() != FilterModeCompetitive {
		t.Fatalf("expected initial filter mode to be Competitive")
	}

	newMode := sm.ToggleFilterMode()
	if newMode != FilterModeAllAllowed {
		t.Errorf("expected AllAllowed after first toggle, got %d", newMode)
	}
	if !strings.Contains(sm.flashMsg, "Comp + Unrated") {
		t.Errorf("expected flash msg to mention Comp + Unrated, got: %q", sm.flashMsg)
	}

	newMode = sm.ToggleFilterMode()
	if newMode != FilterModeCompetitive {
		t.Errorf("expected Competitive after second toggle, got %d", newMode)
	}
	if !strings.Contains(sm.flashMsg, "Competitive Only") {
		t.Errorf("expected flash msg to mention Competitive Only, got: %q", sm.flashMsg)
	}
}

// ── Net RR Calculation ────────────────────────────────────────────────────

func TestNetRRCalculation(t *testing.T) {
	puuid := "rr-player"
	sm := NewSessionModel(puuid)
	sm.rrMap = map[string]int{
		"match-comp-1": 23,
		"match-comp-2": -14,
		"match-unrated": 5, // should not count even if present
	}

	// Two competitive matches -> net = 23 + (-14) = +9
	matches := []*models.MatchDetails{
		{MatchInfo: models.MatchInfo{MatchID: "match-comp-1", QueueID: "competitive"}},
		{MatchInfo: models.MatchInfo{MatchID: "match-comp-2", QueueID: "competitive"}},
		{MatchInfo: models.MatchInfo{MatchID: "match-unrated", QueueID: "unrated"}},
	}
	sm.allSessionMatches = matches

	netRR := sm.NetRR() // default mode = competitive only; unrated filtered out
	if netRR != 9 {
		t.Errorf("NetRR() = %d, want 9 (comp-only mode)", netRR)
	}

	// Switch to AllAllowed: unrated match now included in FilteredMatches but RR is 0 for non-comp
	sm.filterMode = FilterModeAllAllowed
	netRR = sm.NetRR()
	if netRR != 9 {
		t.Errorf("NetRR() in AllAllowed mode = %d, want 9 (only comp contributes RR)", netRR)
	}
}

// ── Session Reset ─────────────────────────────────────────────────────────

func TestResetSession(t *testing.T) {
	puuid := "reset-player"
	sm := NewSessionModel(puuid)

	// Simulate some session matches
	sm.allSessionMatches = []*models.MatchDetails{
		{MatchInfo: models.MatchInfo{MatchID: "match-1", QueueID: "competitive"}},
	}
	sm.scrollOffset = 3
	sm.flashMsg = "old message"

	beforeReset := time.Now()
	sm.ResetSession([]string{"match-2", "match-3"})

	if sm.allSessionMatches != nil {
		t.Errorf("expected allSessionMatches to be nil after reset, got %v", sm.allSessionMatches)
	}
	if sm.scrollOffset != 0 {
		t.Errorf("expected scrollOffset = 0 after reset, got %d", sm.scrollOffset)
	}
	if !strings.Contains(sm.flashMsg, "reset") {
		t.Errorf("expected flash msg to contain 'reset', got %q", sm.flashMsg)
	}
	if !sm.startTime.After(beforeReset.Add(-time.Second)) {
		t.Errorf("expected startTime to be near now, got %v", sm.startTime)
	}
	if !sm.initialMatchIDs["match-2"] || !sm.initialMatchIDs["match-3"] {
		t.Errorf("expected new initial match IDs to be set after reset")
	}
}

// ── SetSessionAnchor ──────────────────────────────────────────────────────

func TestSetSessionAnchor(t *testing.T) {
	sm := NewSessionModel("anchor-player")
	anchor := time.Now().Add(-45 * time.Minute)
	sm.SetSessionAnchor(anchor, []string{"m1", "m2", "m3"}, FilterModeAllAllowed)

	if sm.startTime != anchor {
		t.Errorf("expected startTime = anchor, got %v", sm.startTime)
	}
	if sm.filterMode != FilterModeAllAllowed {
		t.Errorf("expected filterMode AllAllowed, got %d", sm.filterMode)
	}
	for _, id := range []string{"m1", "m2", "m3"} {
		if !sm.initialMatchIDs[id] {
			t.Errorf("expected initialMatchIDs to contain %q", id)
		}
	}
}

// ── FilteredMatches ───────────────────────────────────────────────────────

func TestFilteredMatches_CompetitiveOnly(t *testing.T) {
	sm := NewSessionModel("filter-player")
	sm.filterMode = FilterModeCompetitive
	sm.allSessionMatches = []*models.MatchDetails{
		{MatchInfo: models.MatchInfo{MatchID: "c1", QueueID: "competitive"}},
		{MatchInfo: models.MatchInfo{MatchID: "u1", QueueID: "unrated"}},
		{MatchInfo: models.MatchInfo{MatchID: "dm1", QueueID: "deathmatch"}},
	}
	filtered := sm.FilteredMatches()
	if len(filtered) != 1 {
		t.Errorf("expected 1 filtered match (competitive only), got %d", len(filtered))
	}
	if filtered[0].MatchInfo.MatchID != "c1" {
		t.Errorf("expected filtered match ID 'c1', got %s", filtered[0].MatchInfo.MatchID)
	}
}

func TestFilteredMatches_AllAllowed(t *testing.T) {
	sm := NewSessionModel("filter-player-all")
	sm.filterMode = FilterModeAllAllowed
	sm.allSessionMatches = []*models.MatchDetails{
		{MatchInfo: models.MatchInfo{MatchID: "c1", QueueID: "competitive"}},
		{MatchInfo: models.MatchInfo{MatchID: "u1", QueueID: "unrated"}},
		{MatchInfo: models.MatchInfo{MatchID: "sp1", QueueID: "swiftplay"}},
		{MatchInfo: models.MatchInfo{MatchID: "sr1", QueueID: "spikerush"}},
		{MatchInfo: models.MatchInfo{MatchID: "dm1", QueueID: "deathmatch"}},
		{MatchInfo: models.MatchInfo{MatchID: "esc1", QueueID: "escalation"}},
	}
	filtered := sm.FilteredMatches()
	if len(filtered) != 4 {
		t.Errorf("expected 4 filtered matches (comp+unrated+swift+spike), got %d", len(filtered))
	}
}

// ── View Rendering ────────────────────────────────────────────────────────

func TestSessionModelView_NoMatches(t *testing.T) {
	sm := NewSessionModel("view-player")
	sm.SetSize(80, 24)
	view := sm.View()

	if !strings.Contains(view, "SESSION TRACKER") {
		t.Errorf("expected SESSION TRACKER in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Games Played:     0") {
		t.Errorf("expected 0 games played, got:\n%s", view)
	}
	if !strings.Contains(view, "No matches recorded") {
		t.Errorf("expected empty state message, got:\n%s", view)
	}
	if !strings.Contains(view, "Competitive Only") {
		t.Errorf("expected mode indicator, got:\n%s", view)
	}
	if !strings.Contains(view, "[m] toggle mode") {
		t.Errorf("expected key hint, got:\n%s", view)
	}
}

func TestSessionModelView_WithMatch(t *testing.T) {
	puuid := "view-player-2"
	sm := NewSessionModel(puuid)
	sm.SetSize(120, 40)
	sm.allSessionMatches = []*models.MatchDetails{
		{
			MatchInfo: models.MatchInfo{
				MatchID: "match-view-1",
				QueueID: "competitive",
				MapID:   "Ascent",
			},
			Players: []models.MatchPlayer{
				{
					Subject: puuid,
					TeamID:  "Blue",
					Stats: models.PlayerStats{
						Kills: 20, Deaths: 10, Assists: 4,
						Score: 4000, RoundsPlayed: 20,
					},
				},
			},
			Teams: []models.MatchTeam{
				{TeamID: "Blue", Won: true, RoundsWon: 13},
				{TeamID: "Red", Won: false, RoundsWon: 7},
			},
		},
	}
	sm.rrMap = map[string]int{"match-view-1": 22}

	view := sm.View()

	if !strings.Contains(view, "Games Played:     1") {
		t.Errorf("expected 1 game played, got:\n%s", view)
	}
	if !strings.Contains(view, "1W - 0L - 0D") {
		t.Errorf("expected win record, got:\n%s", view)
	}
	if !strings.Contains(view, "+22 RR") {
		t.Errorf("expected +22 RR in net RR pill, got:\n%s", view)
	}
}
