package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ThejusGSajan/val0/internal/auth"
	"github.com/ThejusGSajan/val0/internal/cache"
	"github.com/ThejusGSajan/val0/internal/models"
)

func TestShopModelView(t *testing.T) {
	skins := []models.ResolvedSkin{
		{
			UUID:        "skin-1",
			DisplayName: "Prime Vandal",
			Rarity:      "Premium",
			CostVP:      1775,
			Sprite:      " [SPRITE] ",
		},
		{
			UUID:        "skin-2",
			DisplayName: "Reaver Phantom",
			Rarity:      "Premium",
			CostVP:      1775,
		},
	}

	model := NewShopModel(skins, 3600)
	model.SetSize(100, 30)
	view := model.View()

	if !strings.Contains(view, "Prime Vandal") {
		t.Errorf("expected view to contain 'Prime Vandal', got:\n%s", view)
	}
	if !strings.Contains(view, "VP 1775") {
		t.Errorf("expected view to contain 'VP 1775', got:\n%s", view)
	}
	if !strings.Contains(view, "Resets in 1h 0m") {
		t.Errorf("expected view to contain timer, got:\n%s", view)
	}
}

func TestNightMarketModelView(t *testing.T) {
	skins := []models.ResolvedSkin{
		{
			UUID:        "skin-nm",
			DisplayName: "Magepunk Ghost",
			Rarity:      "Premium",
			CostVP:      1065,
		},
	}
	discounts := []int{40}

	model := NewNightMarketModel(skins, discounts)
	model.SetSize(100, 30)
	view := model.View()

	if !strings.Contains(view, "NIGHT MARKET") {
		t.Errorf("expected view to contain 'NIGHT MARKET', got:\n%s", view)
	}
	if !strings.Contains(view, "-40%") {
		t.Errorf("expected view to contain '-40%%', got:\n%s", view)
	}

	// Empty case
	emptyModel := NewNightMarketModel(nil, nil)
	emptyModel.SetSize(100, 30)
	emptyView := emptyModel.View()
	if !strings.Contains(emptyView, "not currently active") {
		t.Errorf("expected inactive notice for empty Night Market, got:\n%s", emptyView)
	}
}

func TestShopModel_ResponsiveSetSize(t *testing.T) {
	skins := []models.ResolvedSkin{
		{UUID: "s1", DisplayName: "Vandal 1", CostVP: 1775},
		{UUID: "s2", DisplayName: "Vandal 2", CostVP: 1775},
	}
	m := NewShopModel(skins, 1800)
	m.SetSize(120, 30)
	view := m.View()
	if !strings.Contains(view, "Vandal 1") || !strings.Contains(view, "Vandal 2") {
		t.Errorf("expected skins in responsive view, got:\n%s", view)
	}
}

func TestNightMarketModel_ResponsiveSetSize(t *testing.T) {
	skins := []models.ResolvedSkin{
		{UUID: "nm1", DisplayName: "Phantom 1", CostVP: 1500},
		{UUID: "nm2", DisplayName: "Phantom 2", CostVP: 1200},
	}
	m := NewNightMarketModel(skins, []int{20, 30})
	m.SetSize(160, 40) // width >= 140 triggers 3-column layout
	view := m.View()
	if !strings.Contains(view, "Phantom 1") || !strings.Contains(view, "Phantom 2") {
		t.Errorf("expected skins in wide Night Market view, got:\n%s", view)
	}
}

func TestBattlepassModelView(t *testing.T) {
	data := &BattlepassData{
		CurrentTier:     30,
		MaxTier:         55,
		XPInCurrentTier: 12000,
		XPForNextTier:   24000,
		TotalXP:         450000,
	}

	model := NewBattlepassModel(data)
	view := model.View()

	if !strings.Contains(view, "Tier 30 / 55") {
		t.Errorf("expected view to contain 'Tier 30 / 55', got:\n%s", view)
	}
	if !strings.Contains(view, "450,000") {
		t.Errorf("expected view to contain formatted total XP '450,000', got:\n%s", view)
	}
	if !strings.Contains(view, "Tier 29 → 30") || !strings.Contains(view, "12,000 / 24,000 XP") {
		t.Errorf("expected view to contain 'Tier 29 → 30' and formatted XP, got:\n%s", view)
	}

	// Completed battlepass check
	completedData := &BattlepassData{
		CurrentTier:     55,
		MaxTier:         55,
		XPInCurrentTier: 36500,
		XPForNextTier:   36500,
		TotalXP:         1200000,
	}
	completedModel := NewBattlepassModel(completedData)
	compView := completedModel.View()
	if !strings.Contains(compView, "Battlepass Completed! (55 / 55)") {
		t.Errorf("expected completed banner 'Battlepass Completed! (55 / 55)', got:\n%s", compView)
	}
	if !strings.Contains(compView, "100%") {
		t.Errorf("expected 100%% for completed pass, got:\n%s", compView)
	}
}

func TestErrorModelView(t *testing.T) {
	errModel := NewErrorModel("Please start the Riot Client first.")
	errModel.SetSize(80, 24)
	view := errModel.View()

	if !strings.Contains(view, "Please start the Riot Client first.") {
		t.Errorf("expected error message in view, got:\n%s", view)
	}

	// 0-dimension check
	errModel0 := NewErrorModel("Zero dimensions test")
	v0 := errModel0.View()
	if !strings.Contains(v0, "Zero dimensions test") {
		t.Errorf("expected error message in 0-dim view, got:\n%s", v0)
	}
}

func TestRegionSelectModelNavigation(t *testing.T) {
	m := NewRegionSelectModel()

	// Initial cursor at 0 ("na")
	if m.cursor != 0 {
		t.Errorf("expected cursor at 0, got %d", m.cursor)
	}

	// Move down
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(RegionSelectModel)
	if m.cursor != 1 {
		t.Errorf("expected cursor at 1 after down key, got %d", m.cursor)
	}

	// Press Enter to select
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected command on Enter")
	}

	msg := cmd()
	selectedMsg, ok := msg.(RegionSelectedMsg)
	if !ok || selectedMsg.Region != "eu" {
		t.Errorf("expected RegionSelectedMsg with 'eu', got: %+v", msg)
	}
}

func TestMainModelTabNavigation(t *testing.T) {
	session := &models.Session{
		Region:        "na",
		Shard:         "na",
		ClientVersion: "13.02",
	}

	m := NewMainModel(session)
	if m.activeTab != TabStore {
		t.Errorf("expected initial activeTab TabStore, got %v", m.activeTab)
	}

	// Next tab -> Matches
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(MainModel)
	if m.activeTab != TabMatches {
		t.Errorf("expected activeTab TabMatches, got %v", m.activeTab)
	}

	// Previous tab -> Store
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(MainModel)
	if m.activeTab != TabStore {
		t.Errorf("expected activeTab TabStore, got %v", m.activeTab)
	}
}

func TestProgressModelWithMissions(t *testing.T) {
	bpData := &BattlepassData{
		CurrentTier:     15,
		MaxTier:         55,
		XPInCurrentTier: 5000,
		XPForNextTier:   10000,
		TotalXP:         150000,
	}

	missions := []models.Mission{
		{
			ID: "5163c5f2-4c28-9844-3d07-2ebdb22f98e6", // DefaultMissions "Get Headshots"
			Objectives: map[string]int{
				"04ff6167-4d76-8051-789a-dc853b05f23d": 3,
			},
			Complete: false,
		},
		{
			ID:       "f3e5cfb8-4682-1402-995b-2bb548483f81", // DefaultMissions "Deal Damage"
			Complete: true,
		},
	}

	pm := NewProgressModel(bpData, missions, cache.DefaultMissions)
	pm.SetSize(80, 24)
	view := pm.View()

	if !strings.Contains(view, "BATTLEPASS — Tier 15 / 55") {
		t.Errorf("expected battlepass tier in view, got:\n%s", view)
	}
	if !strings.Contains(view, "ACTIVE MISSIONS") {
		t.Errorf("expected active missions in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Get Headshots — 3 / 5") {
		t.Errorf("expected 'Get Headshots — 3 / 5' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Deal Damage — Completed") {
		t.Errorf("expected 'Deal Damage — Completed' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "+2,000 XP") {
		t.Errorf("expected '+2,000 XP' in view, got:\n%s", view)
	}
}

func TestProgressModel_ResolvedMissions(t *testing.T) {
	missions := []models.Mission{
		{
			ID: "custom-mission-uuid",
			Objectives: map[string]int{
				"obj-1": 4,
			},
			Complete: false,
		},
	}

	customMap := map[string]cache.MissionInfo{
		"custom-mission-uuid": {
			UUID:               "custom-mission-uuid",
			Title:              "Play Swiftplay Games",
			XPGrant:            3500,
			ProgressToComplete: 5,
		},
	}

	pm := NewProgressModel(nil, missions, customMap)
	pm.SetSize(80, 24)
	view := pm.View()

	if !strings.Contains(view, "Play Swiftplay Games — 4 / 5") {
		t.Errorf("expected custom mission title and progress in view, got:\n%s", view)
	}
	if !strings.Contains(view, "+3,500 XP") {
		t.Errorf("expected '+3,500 XP' in view, got:\n%s", view)
	}
	if !strings.Contains(view, "80%") {
		t.Errorf("expected 80%% progress in view, got:\n%s", view)
	}
}

func TestMainModelHeaderWithWalletAndRank(t *testing.T) {
	session := &models.Session{
		Region:        "ap",
		Shard:         "ap",
		ClientVersion: "13.02",
	}

	m := NewMainModel(session)
	m.wallet = &models.WalletResponse{
		Balances: map[string]int{
			models.VPUUID: 4350,
			models.RPUUID: 85,
		},
	}
	m.rankName = "Ascendant 1"
	m.mmr = &models.MMRResponse{
		LatestCompetitiveUpdate: models.CompetitiveUpdate{
			TierAfterUpdate:         21,
			RankedRatingAfterUpdate: 52,
		},
	}
	m.loading = false

	view := m.View()
	if !strings.Contains(view, "Ascendant 1") {
		t.Errorf("expected rank name 'Ascendant 1' in header, got:\n%s", view)
	}
	if !strings.Contains(view, "4,350 VP") {
		t.Errorf("expected '4,350 VP' in header, got:\n%s", view)
	}
	if !strings.Contains(view, "85 RP") {
		t.Errorf("expected '85 RP' in header, got:\n%s", view)
	}
}

func TestMatchesModelListAndDetail(t *testing.T) {
	puuid := "player-me"
	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "match-123",
			MapID:    "/Game/Maps/Ascent/Ascent",
			QueueID:  "competitive",
			IsRanked: true,
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puuid,
				GameName:    "Player",
				TagLine:     "1234",
				TeamID:      "Blue",
				CharacterID: "add6443a-4814-a636-2241-60a3a2777160",
				Stats: models.PlayerStats{
					Kills:        24,
					Deaths:       12,
					Assists:      4,
					Score:        5500,
					RoundsPlayed: 20,
				},
			},
			{
				Subject:     "enemy-1",
				GameName:    "Enemy",
				TagLine:     "9999",
				TeamID:      "Red",
				CharacterID: "a3bfb80f-4041-f0a0-a540-49b4e40b00a5",
				Stats: models.PlayerStats{
					Kills:        12,
					Deaths:       18,
					Assists:      2,
					Score:        2800,
					RoundsPlayed: 20,
				},
			},
		},
		Teams: []models.MatchTeam{
			{TeamID: "Blue", Won: true, RoundsWon: 13, RoundsPlayed: 20},
			{TeamID: "Red", Won: false, RoundsWon: 7, RoundsPlayed: 20},
		},
		RoundResults: []models.RoundResult{
			{
				RoundNum:    1,
				WinningTeam: "Blue",
				PlayerStats: []models.RoundPlayerStat{
					{
						Subject: puuid,
						Damage: []models.RoundDamage{
							{Receiver: "enemy-1", Damage: 150, Headshots: 1},
						},
						Kills: []models.RoundKill{
							{
								Killer: puuid,
								Victim: "enemy-1",
								FinishingDamage: models.FinishingDamage{
									DamageItem: "9c82e19d-4575-0200-1a81-3eacf00cf872",
								},
							},
						},
					},
				},
			},
		},
	}

	items := []MatchItem{
		{
			MatchID:     "match-123",
			MapName:     "Ascent",
			QueueName:   "Competitive",
			AgentName:   "Jett",
			Outcome:     "WIN",
			Score:       "13-7",
			Kills:       24,
			Deaths:      12,
			Assists:     4,
			RREarned:    21,
			HasRR:       true,
			Details:     details,
			PlayerPUUID: puuid,
		},
	}

	mm := NewMatchesModel(items, puuid, nil, nil)
	mm.SetSize(100, 30)

	// Test List View
	listView := mm.View()
	if !strings.Contains(listView, "MATCH HISTORY") {
		t.Errorf("expected MATCH HISTORY header, got:\n%s", listView)
	}
	if !strings.Contains(listView, "WIN") || !strings.Contains(listView, "Ascent") {
		t.Errorf("expected WIN and Ascent in list, got:\n%s", listView)
	}

	// Press Enter to drill down into detail
	updated, _ := mm.Update(tea.KeyMsg{Type: tea.KeyEnter})
	mm = updated

	detailView := mm.View()
	if !strings.Contains(detailView, "MATCH DETAIL") {
		t.Errorf("expected MATCH DETAIL header, got:\n%s", detailView)
	}
	if !strings.Contains(detailView, "BLUE TEAM") {
		t.Errorf("expected BLUE TEAM in detail view, got:\n%s", detailView)
	}
	if !strings.Contains(detailView, "ROUND TIMELINE") {
		t.Errorf("expected ROUND TIMELINE in detail view, got:\n%s", detailView)
	}

	// Press Esc to return to list
	updated, _ = mm.Update(tea.KeyMsg{Type: tea.KeyEsc})
	mm = updated
	backView := mm.View()
	if !strings.Contains(backView, "MATCH HISTORY") {
		t.Errorf("expected return to MATCH HISTORY list view, got:\n%s", backView)
	}
}

func TestStatsModelView(t *testing.T) {
	puuid := "player-me"
	details := []*models.MatchDetails{
		{
			MatchInfo: models.MatchInfo{MatchID: "m1", MapID: "Ascent"},
			Players: []models.MatchPlayer{
				{
					Subject:     puuid,
					CharacterID: "add6443a-4814-a636-2241-60a3a2777160", // Jett
					Stats: models.PlayerStats{
						Kills:        20,
						Deaths:       10,
						Score:        4000,
						RoundsPlayed: 16,
					},
				},
			},
			Teams: []models.MatchTeam{
				{TeamID: "Blue", Won: true, RoundsWon: 13},
				{TeamID: "Red", Won: false, RoundsWon: 3},
			},
			RoundResults: []models.RoundResult{
				{
					RoundNum: 1,
					PlayerStats: []models.RoundPlayerStat{
						{
							Subject: puuid,
							Damage: []models.RoundDamage{
								{Damage: 160, Headshots: 1, Bodyshots: 1},
							},
							Kills: []models.RoundKill{
								{
									Killer: puuid,
									FinishingDamage: models.FinishingDamage{
										DamageItem: "9c82e19d-4575-0200-1a81-3eacf00cf872", // Vandal
									},
								},
							},
						},
					},
				},
			},
		},
	}

	compUpdates := &models.CompetitiveUpdatesResponse{
		Matches: []models.CompetitiveUpdateMatch{
			{MatchID: "m1", RankedRatingEarned: 24},
		},
	}

	sm := NewStatsModel(details, puuid, compUpdates, nil, nil, "Ascendant 1")
	sm.SetSize(100, 35)
	view := sm.View()

	if !strings.Contains(view, "AGENT PERFORMANCE") {
		t.Errorf("expected AGENT PERFORMANCE in view, got:\n%s", view)
	}
	if !strings.Contains(view, "WEAPON STATS") {
		t.Errorf("expected WEAPON STATS in view, got:\n%s", view)
	}
	if strings.Contains(view, "AIM ANALYSIS") {
		t.Errorf("expected AIM ANALYSIS to be removed from view, got:\n%s", view)
	}
	if !strings.Contains(view, "RANK RATING") {
		t.Errorf("expected RANK RATING in view, got:\n%s", view)
	}
}

func TestMainModel5TabSwitching(t *testing.T) {
	session := &models.Session{Region: "ap", Shard: "ap"}
	m := NewMainModel(session)

	// Tab 1: Store
	if m.activeTab != TabStore {
		t.Errorf("expected TabStore, got %v", m.activeTab)
	}

	// Press '2' -> Matches
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'2'}})
	m = updated.(MainModel)
	if m.activeTab != TabMatches {
		t.Errorf("expected TabMatches, got %v", m.activeTab)
	}

	// Press '3' -> Stats
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'3'}})
	m = updated.(MainModel)
	if m.activeTab != TabStats {
		t.Errorf("expected TabStats, got %v", m.activeTab)
	}

	// Press '4' -> Progress
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'4'}})
	m = updated.(MainModel)
	if m.activeTab != TabProgress {
		t.Errorf("expected TabProgress, got %v", m.activeTab)
	}

	// Press '5' -> Session
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'5'}})
	m = updated.(MainModel)
	if m.activeTab != TabSession {
		t.Errorf("expected TabSession, got %v", m.activeTab)
	}
}

func TestWishlistModel(t *testing.T) {
	wm := NewWishlistModel()
	wm.entries = []cache.WishlistEntry{
		{UUID: "skin-wl-1", Name: "Reaver Vandal", Rarity: "Premium", CostVP: 1775},
		{UUID: "skin-wl-2", Name: "Prime Phantom", Rarity: "Premium", CostVP: 1775},
	}
	tier := "60bca009-4182-7998-dee7-b8a2558dc369"
	wm.SetAllSkins([]models.SkinAsset{
		{UUID: "skin-wl-1", DisplayName: "Reaver Vandal", ContentTierUUID: &tier},
		{UUID: "skin-wl-2", DisplayName: "Prime Phantom", ContentTierUUID: &tier},
		{UUID: "skin-wl-3", DisplayName: "Prime Vandal", ContentTierUUID: &tier},
	})
	wm.SetSize(80, 24)

	view := wm.View()
	if !strings.Contains(view, "SKIN WISHLIST") {
		t.Errorf("expected SKIN WISHLIST header, got:\n%s", view)
	}
	if !strings.Contains(view, "Reaver Vandal") || !strings.Contains(view, "Prime Phantom") {
		t.Errorf("expected wishlist entries in view, got:\n%s", view)
	}
	if !strings.Contains(view, "BROWSE & ADD SKINS") {
		t.Errorf("expected BROWSE & ADD SKINS section, got:\n%s", view)
	}

	// Move cursor down in wishlist section
	updated, _ := wm.Update(tea.KeyMsg{Type: tea.KeyDown})
	wm = updated
	if wm.wlCursor != 1 {
		t.Errorf("expected wlCursor 1 after down key, got %d", wm.wlCursor)
	}

	// Switch to browse section via Tab
	updated, _ = wm.Update(tea.KeyMsg{Type: tea.KeyTab})
	wm = updated
	if wm.focusSection != 1 {
		t.Errorf("expected focusSection 1 after Tab, got %d", wm.focusSection)
	}

	// Type search query "vandal"
	for _, r := range "vandal" {
		updated, _ = wm.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		wm = updated
	}
	if wm.searchInput != "vandal" {
		t.Errorf("expected searchInput 'vandal', got '%s'", wm.searchInput)
	}
	if len(wm.filtered) != 2 {
		t.Errorf("expected 2 filtered skins for 'vandal', got %d", len(wm.filtered))
	}

	// Backspace in search
	updated, _ = wm.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	wm = updated
	if wm.searchInput != "vanda" {
		t.Errorf("expected searchInput 'vanda' after backspace, got '%s'", wm.searchInput)
	}
}

func TestSessionModel(t *testing.T) {
	puuid := "player-me"
	sm := NewSessionModel(puuid)
	sm.SetInitialSnapshot([]string{"old-match-1", "old-match-2"})

	// Add a new session match
	newMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{MatchID: "new-match-3", MapID: "Ascent", QueueID: "competitive"},
		Players: []models.MatchPlayer{
			{
				Subject: puuid,
				TeamID:  "Blue",
				Stats: models.PlayerStats{
					Kills:        18,
					Deaths:       12,
					Assists:      5,
					Score:        3600,
					RoundsPlayed: 18,
				},
			},
		},
		Teams: []models.MatchTeam{
			{TeamID: "Blue", Won: true, RoundsWon: 13},
			{TeamID: "Red", Won: false, RoundsWon: 5},
		},
	}

	sm.UpdateSessionMatches([]*models.MatchDetails{newMatch})
	sm.SetSize(80, 24)
	view := sm.View()

	if !strings.Contains(view, "SESSION TRACKER") {
		t.Errorf("expected SESSION TRACKER header, got:\n%s", view)
	}
	if !strings.Contains(view, "Games Played:     1") {
		t.Errorf("expected 1 game played in session, got:\n%s", view)
	}
	if !strings.Contains(view, "1W - 0L - 0D") {
		t.Errorf("expected 1W - 0L in session, got:\n%s", view)
	}
}

func TestShopModelWishlistRendering(t *testing.T) {
	skin := models.ResolvedSkin{
		UUID:        "toggle-vandal-uuid",
		DisplayName: "Glitchpop Vandal",
		Rarity:      "Exclusive",
		CostVP:      2175,
	}

	_ = cache.AddToWishlist(cache.ConvertResolvedSkinToWishlist(skin))
	defer cache.RemoveFromWishlist(skin.UUID)

	sm := NewShopModel([]models.ResolvedSkin{skin}, 3600)
	sm.SetSize(80, 24)

	view := sm.View()
	if !strings.Contains(view, "WISHLIST MATCH") {
		t.Errorf("expected WISHLIST MATCH banner in view, got:\n%s", view)
	}
	if !strings.Contains(view, "enter wishlist tab") {
		t.Errorf("expected 'enter wishlist tab' in view, got:\n%s", view)
	}
}

func TestGetQueueDisplayName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"competitive", "Competitive"},
		{"unrated", "Unrated"},
		{"deathmatch", "Deathmatch"},
		{"spikerush", "Spike Rush"},
		{"swiftplay", "Swiftplay"},
		{"hurm", "Team Deathmatch"},
		{"ggteam", "Escalation"},
		{"onefa", "Replication"},
		{"snowball", "Snowball Fight"},
		{"newmap", "New Map"},
		{"custom", "Custom"},
		{"", "Unknown"},
		{"premier", "Premier"},
	}

	for _, tt := range tests {
		got := GetQueueDisplayName(tt.input)
		if got != tt.expected {
			t.Errorf("GetQueueDisplayName(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestMatchesModelPlayerNames(t *testing.T) {
	puuid := "my-puuid"
	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{MatchID: "m-names", MapID: "Ascent"},
		Players: []models.MatchPlayer{
			{Subject: puuid, TeamID: "Blue", GameName: "MyName", TagLine: "123"},
			{Subject: "other-puuid-1", TeamID: "Blue", GameName: "Friend", TagLine: "TAG"},
			{Subject: "other-puuid-2", TeamID: "Blue", GameName: "", TagLine: ""}, // hidden
		},
		Teams: []models.MatchTeam{
			{TeamID: "Blue", Won: true, RoundsWon: 13},
			{TeamID: "Red", Won: false, RoundsWon: 5},
		},
	}

	items := []MatchItem{
		{
			MatchID:     "m-names",
			Details:     details,
			PlayerPUUID: puuid,
		},
	}

	mm := NewMatchesModel(items, puuid, nil, nil)
	mm.SetSize(100, 30)
	mm.viewMode = MatchViewDetail

	view := mm.View()
	if !strings.Contains(view, "▸ You") {
		t.Errorf("expected '▸ You' for player's own row, got:\n%s", view)
	}
	if !strings.Contains(view, "Friend#TAG") {
		t.Errorf("expected 'Friend#TAG' for resolved player name, got:\n%s", view)
	}
	if !strings.Contains(view, "<Hidden>") {
		t.Errorf("expected '<Hidden>' for streamer mode / unresolved name, got:\n%s", view)
	}
	if strings.Contains(view, "other-puuid-2") {
		t.Errorf("UUID should never be displayed in view, got:\n%s", view)
	}
}

func TestRootModel_AuthFlow(t *testing.T) {
	root := NewRootModel()
	if root.state != StateAuthenticating {
		t.Fatalf("expected initial state StateAuthenticating, got %v", root.state)
	}

	cmd := root.Init()
	if cmd == nil {
		t.Fatal("expected Init() to return authenticateCmd")
	}

	// Test 0-dim View in StateAuthenticating
	v0 := root.View()
	if !strings.Contains(v0, "Connecting to Riot Client") {
		t.Errorf("expected connecting message in View, got:\n%s", v0)
	}

	// Test window resize during authenticating
	updated, _ := root.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	root = updated.(RootModel)
	if root.width != 100 || root.height != 30 {
		t.Errorf("expected root dimensions 100x30, got %dx%d", root.width, root.height)
	}
	vPlaced := root.View()
	if !strings.Contains(vPlaced, "Connecting to Riot Client") {
		t.Errorf("expected connecting message in Placed View, got:\n%s", vPlaced)
	}

	// Test AuthFailedMsg
	testErr := errors.New("connection refused")
	updated, _ = root.Update(AuthFailedMsg{Err: testErr})
	root = updated.(RootModel)
	if root.state != StateAuthError {
		t.Fatalf("expected state StateAuthError, got %v", root.state)
	}
	errView := root.View()
	if !strings.Contains(errView, "connection refused") {
		t.Errorf("expected error message in View, got:\n%s", errView)
	}

	// Test Lockfile not found special message
	updated, _ = root.Update(AuthFailedMsg{Err: auth.ErrLockfileNotFound})
	root = updated.(RootModel)
	if !strings.Contains(root.View(), "Please start the Riot Client first") {
		t.Errorf("expected Riot Client lockfile message, got:\n%s", root.View())
	}

	// Test AuthSuccessMsg
	session := &models.Session{
		PUUID:         "test-puuid",
		Region:        "na",
		Shard:         "na",
		ClientVersion: "release-13.02",
	}
	updated, initCmd := root.Update(AuthSuccessMsg{Session: session})
	root = updated.(RootModel)
	if root.state != StateMain {
		t.Fatalf("expected state StateMain, got %v", root.state)
	}
	if initCmd == nil {
		t.Error("expected initCmd from MainModel upon AuthSuccessMsg")
	}
	mainView := root.View()
	if mainView == "" {
		t.Error("expected non-empty MainModel view")
	}
}

func TestRootModel_Retry(t *testing.T) {
	root := NewRootModel()
	updated, _ := root.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	root = updated.(RootModel)

	// Move to error state
	updated, _ = root.Update(AuthFailedMsg{Err: errors.New("auth failure")})
	root = updated.(RootModel)
	if root.state != StateAuthError {
		t.Fatalf("expected StateAuthError, got %v", root.state)
	}

	// Press 'r' to retry
	updated, cmd := root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	root = updated.(RootModel)
	if root.state != StateAuthenticating {
		t.Fatalf("expected StateAuthenticating after pressing 'r', got %v", root.state)
	}
	if cmd == nil {
		t.Fatal("expected retry authenticateCmd on pressing 'r'")
	}

	// Move to error state again
	updated, _ = root.Update(AuthFailedMsg{Err: errors.New("auth failure 2")})
	root = updated.(RootModel)

	// Send RetryAuthMsg
	updated, cmd = root.Update(RetryAuthMsg{})
	root = updated.(RootModel)
	if root.state != StateAuthenticating {
		t.Fatalf("expected StateAuthenticating on RetryAuthMsg, got %v", root.state)
	}
	if cmd == nil {
		t.Fatal("expected authenticateCmd on RetryAuthMsg")
	}

	// Test 'q' quits in error state
	updated, cmd = root.Update(AuthFailedMsg{Err: errors.New("quit test")})
	root = updated.(RootModel)
	_, cmd = root.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd on 'q'")
	}
}

func TestErrorModel_DynamicSizing(t *testing.T) {
	errModel := NewErrorModel("Test Error Message")

	// 0-dimension safety check
	v0 := errModel.View()
	if !strings.Contains(v0, "Test Error Message") || !strings.Contains(v0, "retry") {
		t.Errorf("expected compact error message on 0 dimensions, got:\n%s", v0)
	}

	// Narrow terminal (< 40)
	errModel.SetSize(35, 15)
	vNarrow := errModel.View()
	if !strings.Contains(vNarrow, "Test Error Message") {
		t.Errorf("expected error message on narrow view, got:\n%s", vNarrow)
	}

	// Standard terminal (80x24)
	errModel.SetSize(80, 24)
	vStandard := errModel.View()
	if !strings.Contains(vStandard, "Test Error Message") {
		t.Errorf("expected error message on standard view, got:\n%s", vStandard)
	}

	// Key handling: 'r' returns retry message
	_, cmd := errModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'r'}})
	if cmd == nil {
		t.Fatal("expected cmd on 'r'")
	}
	msg := cmd()
	if _, ok := msg.(RetryAuthMsg); !ok {
		t.Errorf("expected RetryAuthMsg on 'r', got %+v", msg)
	}

	// Key handling: 'q' returns quit
	_, cmd = errModel.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Fatal("expected quit cmd on 'q'")
	}
}

func TestShopModel_WishlistMatching(t *testing.T) {
	// Add test skin to wishlist
	targetUUID := "prime-vandal-uuid-123"
	cache.AddToWishlist(cache.WishlistEntry{
		UUID:   targetUUID,
		Name:   "Prime Vandal",
		Rarity: "Premium",
		CostVP: 1775,
	})
	defer cache.RemoveFromWishlist(targetUUID)

	skins := []models.ResolvedSkin{
		{
			UUID:        targetUUID,
			DisplayName: "Prime Vandal",
			Rarity:      "Premium",
			CostVP:      1775,
		},
		{
			UUID:        "other-skin-uuid",
			DisplayName: "Reaver Phantom",
			Rarity:      "Premium",
			CostVP:      1775,
		},
	}

	model := NewShopModel(skins, 3600)
	model.SetSize(100, 30)
	view := model.View()

	if !strings.Contains(view, "WISHLIST MATCH!") {
		t.Errorf("expected wishlist match banner in shop view, got:\n%s", view)
	}
	if !strings.Contains(view, "WISHLIST ITEM") {
		t.Errorf("expected wishlist item tag on skin card, got:\n%s", view)
	}
}

func TestStatsModel_ViewportBounds(t *testing.T) {
	// Create mock match details with multiple rounds and kills to generate long sections
	playerPUUID := "player-1"
	var matches []*models.MatchDetails
	for i := 0; i < 10; i++ {
		md := &models.MatchDetails{
			MatchInfo: models.MatchInfo{MatchID: fmt.Sprintf("match-%d", i), MapID: "Ascent"},
			Players: []models.MatchPlayer{
				{
					Subject:     playerPUUID,
					CharacterID: "agent-jett",
					TeamID:      "Blue",
					Stats:       models.PlayerStats{Kills: 15, Deaths: 10, Assists: 5, Score: 3500, RoundsPlayed: 20},
				},
			},
			RoundResults: []models.RoundResult{
				{
					PlayerStats: []models.RoundPlayerStat{
						{
							Subject: playerPUUID,
							Damage: []models.RoundDamage{
								{Headshots: 2, Bodyshots: 3, Legshots: 0, Damage: 180},
							},
							Kills: []models.RoundKill{
								{Killer: playerPUUID, FinishingDamage: models.FinishingDamage{DamageItem: "vandal"}},
							},
						},
					},
				},
			},
		}
		matches = append(matches, md)
	}

	compUpdates := &models.CompetitiveUpdatesResponse{
		Matches: []models.CompetitiveUpdateMatch{
			{RankedRatingEarned: 20},
			{RankedRatingEarned: -15},
		},
	}

	sm := NewStatsModel(matches, playerPUUID, compUpdates, nil, nil, "Gold 2")
	targetHeight := 10
	sm.SetSize(80, targetHeight)

	view := sm.View()
	lines := strings.Split(view, "\n")
	if len(lines) > targetHeight {
		t.Errorf("expected rendered line count <= %d, got %d", targetHeight, len(lines))
	}

	// Test scroll down
	updated, _ := sm.Update(tea.KeyMsg{Type: tea.KeyDown})
	sm = updated
	viewDown := sm.View()
	linesDown := strings.Split(viewDown, "\n")
	if len(linesDown) > targetHeight {
		t.Errorf("expected rendered line count <= %d after scroll down, got %d", targetHeight, len(linesDown))
	}

	// Test scroll up
	updated, _ = sm.Update(tea.KeyMsg{Type: tea.KeyUp})
	sm = updated
	viewUp := sm.View()
	linesUp := strings.Split(viewUp, "\n")
	if len(linesUp) > targetHeight {
		t.Errorf("expected rendered line count <= %d after scroll up, got %d", targetHeight, len(linesUp))
	}
}

func TestMainModel_ViewportBudgetingAndTruncation(t *testing.T) {
	session := &models.Session{
		PUUID:         "test-puuid",
		Region:        "na",
		Shard:         "na",
		ClientVersion: "release-13.02",
	}

	m := NewMainModel(session)
	m.loading = false

	// Window resize to small terminal (80x24)
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = updated.(MainModel)

	if m.width != 80 || m.height != 24 {
		t.Fatalf("expected dimensions 80x24, got %dx%d", m.width, m.height)
	}

	// Verify stats model size budgeted correctly (24 - 8 = 16)
	if m.statsModel.height != 16 {
		t.Errorf("expected statsModel height 16 (24-8), got %d", m.statsModel.height)
	}

	view := m.View()
	lines := strings.Split(view, "\n")
	if len(lines) > 24 {
		t.Errorf("expected MainModel.View() output <= 24 lines, got %d", len(lines))
	}
}

func TestStoreSubTabNavigationKeys(t *testing.T) {
	session := &models.Session{Region: "na", Shard: "na"}
	m := NewMainModel(session)
	m.loading = false

	// Initial store sub-tab is Shop
	if m.storeSubTab != SubTabShop {
		t.Fatalf("expected initial SubTabShop, got %v", m.storeSubTab)
	}

	// Press 'w' -> Wishlist
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'w'}})
	m = updated.(MainModel)
	if m.storeSubTab != SubTabWishlist {
		t.Errorf("expected SubTabWishlist after pressing 'w', got %v", m.storeSubTab)
	}

	// Press 's' -> Shop
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	m = updated.(MainModel)
	if m.storeSubTab != SubTabShop {
		t.Errorf("expected SubTabShop after pressing 's', got %v", m.storeSubTab)
	}

	// Without night market data, pressing 'n' should stay on current sub-tab
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(MainModel)
	if m.storeSubTab != SubTabShop {
		t.Errorf("expected to stay on SubTabShop when NightMarket is empty, got %v", m.storeSubTab)
	}

	// Enable night market data
	m.nightModel = NewNightMarketModel([]models.ResolvedSkin{
		{UUID: "nm-1", DisplayName: "Prime Vandal", CostVP: 1200},
	}, []int{30})

	// Now press 'n' -> NightMarket
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	m = updated.(MainModel)
	if m.storeSubTab != SubTabNightMarket {
		t.Errorf("expected SubTabNightMarket after pressing 'n' with active NM, got %v", m.storeSubTab)
	}
}

func TestStoreSubTabContextualLegends(t *testing.T) {
	// 1. Shop without Night Market
	shopModel := NewShopModel([]models.ResolvedSkin{
		{UUID: "s1", DisplayName: "Prime Vandal", CostVP: 1775},
	}, 3600)
	shopModel.SetSize(80, 24)
	shopView := shopModel.View()
	if !strings.Contains(shopView, "w") || !strings.Contains(shopView, "enter wishlist tab") {
		t.Errorf("expected 'w enter wishlist tab' in shop, got:\n%s", shopView)
	}
	if strings.Contains(shopView, "enter nightmarket tab") {
		t.Errorf("expected NO 'enter nightmarket tab' in shop without NM, got:\n%s", shopView)
	}

	// 2. Shop with Night Market
	shopModel.SetNightMarketActive(true)
	shopViewNM := shopModel.View()
	if !strings.Contains(shopViewNM, "enter wishlist tab") || !strings.Contains(shopViewNM, "enter nightmarket tab") {
		t.Errorf("expected both wishlist and nightmarket legends in shop with NM, got:\n%s", shopViewNM)
	}

	// 3. Wishlist without Night Market
	wm := NewWishlistModel()
	wm.SetSize(80, 24)
	wView := wm.View()
	if !strings.Contains(wView, "enter shop tab") {
		t.Errorf("expected 'enter shop tab' in wishlist, got:\n%s", wView)
	}
	if strings.Contains(wView, "enter nightmarket tab") {
		t.Errorf("expected NO 'enter nightmarket tab' in wishlist without NM, got:\n%s", wView)
	}

	// 4. Wishlist with Night Market
	wm.SetNightMarketActive(true)
	wViewNM := wm.View()
	if !strings.Contains(wViewNM, "enter shop tab") || !strings.Contains(wViewNM, "enter nightmarket tab") {
		t.Errorf("expected both shop and nightmarket legends in wishlist with NM, got:\n%s", wViewNM)
	}

	// 5. Night Market legends
	nm := NewNightMarketModel([]models.ResolvedSkin{
		{UUID: "nm1", DisplayName: "Magepunk Ghost", CostVP: 1000},
	}, []int{20})
	nm.SetSize(80, 24)
	nmView := nm.View()
	if !strings.Contains(nmView, "enter shop tab") || !strings.Contains(nmView, "enter wishlist tab") {
		t.Errorf("expected 'enter shop tab' and 'enter wishlist tab' in Night Market view, got:\n%s", nmView)
	}
}

func TestGlobalKeyLegendStandardization(t *testing.T) {
	// 1. Header bar and MainModel status bar
	session := &models.Session{Region: "na", Shard: "na"}
	mainModel := NewMainModel(session)
	mainModel.loading = false
	mainView := mainModel.View()
	if !strings.Contains(mainView, "val0") {
		t.Errorf("MainModel header missing 'val0', got:\n%s", mainView)
	}
	expectedMainLegend := RenderKeyLegends(
		[2]string{"1-5", "switch tabs"},
		[2]string{"←/→", "prev/next"},
		[2]string{"r", "refresh"},
		[2]string{"q", "quit"},
	)
	if !strings.Contains(mainView, expectedMainLegend) {
		t.Errorf("MainModel status bar missing legend %q, got:\n%s", expectedMainLegend, mainView)
	}

	// 2. MainModel error rendering
	mainModel.err = errors.New("network error")
	errView := mainModel.View()
	expectedErrLegend := RenderKeyLegends([2]string{"r", "retry"}, [2]string{"q", "quit"})
	if !strings.Contains(errView, expectedErrLegend) {
		t.Errorf("MainModel renderError missing keys, got:\n%s", errView)
	}

	// 3. MatchesModel list & detail views
	matchesModel := NewMatchesModel([]MatchItem{
		{MatchID: "m1", MapName: "Haven", QueueName: "Competitive"},
	}, "p1", nil, nil)
	matchesModel.SetSize(80, 24)
	listMatchesView := matchesModel.View()
	expectedMatchesListLegend := RenderKeyLegends(
		[2]string{"↑/↓", "select"},
		[2]string{"enter", "view match detail"},
	)
	if !strings.Contains(listMatchesView, expectedMatchesListLegend) {
		t.Errorf("MatchesModel list view missing legend, got:\n%s", listMatchesView)
	}

	matchesModel.viewMode = MatchViewDetail
	detailMatchesView := matchesModel.View()
	if !strings.Contains(detailMatchesView, RenderKeyItem("esc", "return to match list")) && !strings.Contains(detailMatchesView, RenderKeyItem("esc", "go back")) {
		t.Errorf("MatchesModel detail view missing legend, got:\n%s", detailMatchesView)
	}

	// 4. StatsModel scroll view
	statsModel := NewStatsModel(nil, "p1", nil, nil, nil, "Silver 1")
	statsModel.SetSize(80, 24)
	statsView := statsModel.View()
	if !strings.Contains(statsView, RenderKeyItem("↑/↓", "scroll stats view")) {
		t.Errorf("StatsModel missing scroll legend, got:\n%s", statsView)
	}

	// 5. RegionSelectModel
	regionModel := NewRegionSelectModel()
	regionModel.SetSize(80, 24)
	regionView := regionModel.View()
	expectedRegionLegend := RenderKeyLegends(
		[2]string{"↑/↓", "move"},
		[2]string{"enter", "select"},
	)
	if !strings.Contains(regionView, expectedRegionLegend) {
		t.Errorf("RegionSelectModel missing legend, got:\n%s", regionView)
	}

	// 6. ErrorModel
	errM := NewErrorModel("Auth Error")
	errM.SetSize(80, 24)
	expectedErrModelLegend := RenderKeyLegends(
		[2]string{"r", "retry"},
		[2]string{"q", "exit"},
	)
	if !strings.Contains(errM.View(), expectedErrModelLegend) {
		t.Errorf("ErrorModel missing legend, got:\n%s", errM.View())
	}

	// 7. WishlistModel preserves "Type to search" while modernizing keys
	wlModel := NewWishlistModel()
	wlModel.SetSize(80, 24)
	wlView := wlModel.View()
	if !strings.Contains(wlView, "Type to search") {
		t.Errorf("WishlistModel should preserve 'Type to search', got:\n%s", wlView)
	}
	expectedWlLegend := RenderKeyLegends(
		[2]string{"s", "enter shop tab"},
		[2]string{"tab", "switch section"},
		[2]string{"↑/↓", "navigate"},
		[2]string{"", "Type to search"},
		[2]string{"enter", "add to wishlist"},
		[2]string{"x", "remove"},
	)
	if !strings.Contains(wlView, expectedWlLegend) {
		t.Errorf("WishlistModel missing modernized key legends, got:\n%s", wlView)
	}
}

func TestResolveQueueDisplayName(t *testing.T) {
	tests := []struct {
		queueID  string
		gameMode string
		expected string
	}{
		{"competitive", "", "Competitive"},
		{"unrated", "", "Unrated"},
		{"deathmatch", "", "Deathmatch"},
		{"skirmish", "", "Skirmish"},
		{"", "/Game/GameModes/Skirmish/SkirmishGameMode.SkirmishGameMode_C", "Skirmish"},
		{"", "/Game/GameModes/Deathmatch/DeathmatchGameMode.DeathmatchGameMode_C", "Deathmatch"},
		{"", "", "Unknown"},
	}

	for _, tc := range tests {
		got := ResolveQueueDisplayName(tc.queueID, tc.gameMode)
		if got != tc.expected {
			t.Errorf("ResolveQueueDisplayName(%q, %q) = %q, want %q", tc.queueID, tc.gameMode, got, tc.expected)
		}
	}
}

func TestFitWidthWideCharacters(t *testing.T) {
	// Katakana wide character 'ツ'
	nameWithWideRune := "Luffyツ #Doofy"
	fitted := fitWidth(nameWithWideRune, 18)
	if lipgloss.Width(fitted) != 18 {
		t.Errorf("expected fitWidth visual width 18, got %d for %q", lipgloss.Width(fitted), fitted)
	}

	asciiName := "NormalPlayer#123"
	fittedAscii := fitWidth(asciiName, 18)
	if lipgloss.Width(fittedAscii) != 18 {
		t.Errorf("expected fitWidth visual width 18, got %d for %q", lipgloss.Width(fittedAscii), fittedAscii)
	}
}

func TestMatchListRelativeTimeAlignment(t *testing.T) {
	now := time.Now().Add(-2 * time.Hour)
	items := []MatchItem{
		{
			MatchID:     "m1",
			MapName:     "Ascent",
			QueueName:   "Competitive",
			AgentName:   "Jett",
			Score:       "13-11",
			Outcome:     "WIN",
			HasRR:       true,
			RREarned:    14,
			GameTime:    now,
			PlayerPUUID: "p1",
		},
		{
			MatchID:     "m2",
			MapName:     "Bind",
			QueueName:   "Deathmatch",
			AgentName:   "Reyna",
			Score:       "40-35",
			Outcome:     "WIN",
			HasRR:       false,
			GameTime:    now,
			PlayerPUUID: "p1",
		},
	}

	m := NewMatchesModel(items, "p1", nil, nil)
	m.SetSize(120, 30)
	view := m.View()

	lines := strings.Split(view, "\n")
	var timeColIdx []int
	for _, l := range lines {
		if idx := strings.Index(l, "2h ago"); idx != -1 {
			timeColIdx = append(timeColIdx, lipgloss.Width(l[:idx]))
		}
	}

	if len(timeColIdx) < 2 {
		t.Fatalf("expected at least 2 match lines with time, found %d", len(timeColIdx))
	}
	if timeColIdx[0] != timeColIdx[1] {
		t.Errorf("time column index mismatch: row 0 is at col %d, row 1 is at col %d", timeColIdx[0], timeColIdx[1])
	}
}

func TestRenderTeamTableDescendingACS(t *testing.T) {
	details := &models.MatchDetails{
		Players: []models.MatchPlayer{
			{
				Subject:     "p1",
				TeamID:      "Blue",
				GameName:    "HighScorer",
				CharacterID: "c1",
				Stats: models.PlayerStats{
					Score:   6000,
					RoundsPlayed: 20,
					Kills:   25,
					Deaths:  10,
					Assists: 5,
				},
			},
			{
				Subject:     "p2",
				TeamID:      "Blue",
				GameName:    "LowScorer",
				CharacterID: "c2",
				Stats: models.PlayerStats{
					Score:   2000,
					RoundsPlayed: 20,
					Kills:   8,
					Deaths:  15,
					Assists: 2,
				},
			},
			{
				Subject:     "p3",
				TeamID:      "Blue",
				GameName:    "MidScorer",
				CharacterID: "c3",
				Stats: models.PlayerStats{
					Score:   4000,
					RoundsPlayed: 20,
					Kills:   15,
					Deaths:  12,
					Assists: 4,
				},
			},
		},
		Teams: []models.MatchTeam{
			{TeamID: "Blue", Won: true, RoundsWon: 13},
		},
	}

	m := NewMatchesModel(nil, "spectator-puuid", nil, nil)
	m.SetSize(100, 30)
	table := m.renderTeamTable(details, "Blue", true)

	lowIdx := strings.Index(table, "LowScorer")
	midIdx := strings.Index(table, "MidScorer")
	highIdx := strings.Index(table, "HighScorer")

	if lowIdx == -1 || midIdx == -1 || highIdx == -1 {
		t.Fatalf("players missing from table:\n%s", table)
	}

	if !(highIdx < midIdx && midIdx < lowIdx) {
		t.Errorf("players not sorted in descending order of ACS! indices: High=%d, Mid=%d, Low=%d\n%s", highIdx, midIdx, lowIdx, table)
	}
}

func TestMatchesModelDeathmatchDetail(t *testing.T) {
	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			QueueID:  "deathmatch",
			GameMode: "/Game/GameModes/Deathmatch/DeathmatchGameMode.DeathmatchGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     "p-tied-lower-acs",
				GameName:    "TiedLowerACS",
				CharacterID: "c1",
				Stats: models.PlayerStats{
					Kills:        35,
					Deaths:       15,
					Assists:      2,
					Score:        3500,
					RoundsPlayed: 1,
				},
			},
			{
				Subject:     "p-tied-higher-acs",
				GameName:    "TiedHigherACS",
				CharacterID: "c2",
				Stats: models.PlayerStats{
					Kills:        35,
					Deaths:       12,
					Assists:      3,
					Score:        4200,
					RoundsPlayed: 1,
				},
			},
			{
				Subject:     "p-top-fragger",
				GameName:    "TopFragger",
				CharacterID: "c3",
				Stats: models.PlayerStats{
					Kills:        40,
					Deaths:       10,
					Assists:      5,
					Score:        3800,
					RoundsPlayed: 1,
				},
			},
			{
				Subject:     "p-high-acs-low-kills",
				GameName:    "HighAcsLowKills",
				CharacterID: "c4",
				Stats: models.PlayerStats{
					Kills:        20,
					Deaths:       20,
					Assists:      1,
					Score:        5000,
					RoundsPlayed: 1,
				},
			},
			{
				Subject:     "my-puuid",
				GameName:    "SelfPlayer",
				CharacterID: "c5",
				Stats: models.PlayerStats{
					Kills:        10,
					Deaths:       25,
					Assists:      0,
					Score:        1000,
					RoundsPlayed: 1,
				},
			},
		},
		RoundResults: []models.RoundResult{
			{
				RoundNum:    0,
				WinningTeam: "Blue",
			},
		},
	}

	items := []MatchItem{
		{
			MatchID:     "dm-match-1",
			MapName:     "Ascent",
			QueueName:   "Deathmatch",
			Details:     details,
			PlayerPUUID: "my-puuid",
		},
	}

	m := NewMatchesModel(items, "my-puuid", nil, nil)
	m.SetSize(100, 30)
	m.viewMode = MatchViewDetail

	output := m.renderDetailView()

	// Header outcome and placement checks (local player placed 5th)
	if !strings.Contains(output, "LOSS  5th") {
		t.Errorf("expected Deathmatch detail header to contain 'LOSS  5th', got:\n%s", output)
	}

	// 1. Structural checks: FFA layout, no team headers
	if strings.Contains(output, "BLUE TEAM") || strings.Contains(output, "RED TEAM") || strings.Contains(output, "YOUR TEAM") {
		t.Errorf("expected Deathmatch detail view to not contain team headers, got:\n%s", output)
	}

	// 2. Timeline omitted
	if strings.Contains(output, "ROUND TIMELINE") {
		t.Errorf("expected Deathmatch detail view to omit ROUND TIMELINE, got:\n%s", output)
	}

	// 3. Stats columns: Player, Agent, K, D, A present; ACS, HS%, ADR, Econ omitted
	if !strings.Contains(output, "Player") || !strings.Contains(output, "Agent") || !strings.Contains(output, "K") || !strings.Contains(output, "D") || !strings.Contains(output, "A") {
		t.Errorf("expected columns Player, Agent, K, D, A in Deathmatch table, got:\n%s", output)
	}
	if strings.Contains(output, " ACS ") || strings.Contains(output, "HS%") || strings.Contains(output, "ADR") || strings.Contains(output, "Econ") {
		t.Errorf("expected ACS, HS%%, ADR, Econ to be omitted from Deathmatch table, got:\n%s", output)
	}

	// 4. Local player representation
	if !strings.Contains(output, "▸ You") {
		t.Errorf("expected local player to be displayed as '▸ You', got:\n%s", output)
	}

	// 5. Ranking & Sorting checks
	topIdx := strings.Index(output, "TopFragger")
	tiedHighIdx := strings.Index(output, "TiedHigherACS")
	tiedLowIdx := strings.Index(output, "TiedLowerACS")
	highAcsLowKillsIdx := strings.Index(output, "HighAcsLowKills")
	selfIdx := strings.Index(output, "▸ You")

	if topIdx == -1 || tiedHighIdx == -1 || tiedLowIdx == -1 || highAcsLowKillsIdx == -1 || selfIdx == -1 {
		t.Fatalf("one or more players missing from Deathmatch table:\n%s", output)
	}

	if !(topIdx < tiedHighIdx) {
		t.Errorf("expected TopFragger (40 kills) before TiedHigherACS (35 kills), got topIdx=%d, tiedHighIdx=%d", topIdx, tiedHighIdx)
	}
	if !(tiedHighIdx < tiedLowIdx) {
		t.Errorf("expected TiedHigherACS (ACS 4200) before TiedLowerACS (ACS 3500) on kill tie-breaker, got tiedHighIdx=%d, tiedLowIdx=%d", tiedHighIdx, tiedLowIdx)
	}
	if !(tiedLowIdx < highAcsLowKillsIdx) {
		t.Errorf("expected TiedLowerACS (35 kills) before HighAcsLowKills (20 kills), got tiedLowIdx=%d, highAcsLowKillsIdx=%d", tiedLowIdx, highAcsLowKillsIdx)
	}
	if !(highAcsLowKillsIdx < selfIdx) {
		t.Errorf("expected HighAcsLowKills (20 kills) before self (10 kills), got highAcsLowKillsIdx=%d, selfIdx=%d", highAcsLowKillsIdx, selfIdx)
	}

	// 6. Check winner detail view rendering (top fragger)
	mWinner := NewMatchesModel(items, "p-top-fragger", nil, nil)
	mWinner.SetSize(100, 30)
	mWinner.viewMode = MatchViewDetail
	winnerOutput := mWinner.renderDetailView()
	if !strings.Contains(winnerOutput, "WIN  1st") {
		t.Errorf("expected winner detail header to contain 'WIN  1st', got:\n%s", winnerOutput)
	}

	// 7. Check list view rendering (score column shows '5th', not '0-0')
	mList := NewMatchesModel(items, "my-puuid", nil, nil)
	mList.SetSize(100, 30)
	listOutput := mList.renderListView()
	if strings.Contains(listOutput, "0-0") || !strings.Contains(listOutput, "5th") {
		t.Errorf("expected list view to render '5th' and not '0-0', got:\n%s", listOutput)
	}
}

func TestMatchesModelRoundTimelineUnderline(t *testing.T) {
	details := &models.MatchDetails{
		RoundResults: []models.RoundResult{
			{RoundNum: 0, WinningTeam: "Blue"},
		},
	}

	m := NewMatchesModel(nil, "spectator-puuid", nil, nil)
	m.SetSize(80, 30)

	timeline := m.renderRoundTimeline(details, "Blue")

	lines := strings.Split(timeline, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected timeline to contain multiple lines, got:\n%s", timeline)
	}
	if !strings.Contains(lines[0], "ROUND TIMELINE") {
		t.Errorf("expected first line to contain ROUND TIMELINE, got: %q", lines[0])
	}
	if !strings.HasPrefix(lines[1], "  ─") {
		t.Errorf("expected second line to start with '  ─', got: %q", lines[1])
	}
}

func TestStatsModel_CasualModesFiltered(t *testing.T) {
	puUID := "test-player-puuid"
	agentsMap := map[string]string{
		"agent-jett":  "Jett",
		"agent-sova":  "Sova",
		"agent-gekko": "Gekko",
		"agent-reyna": "Reyna",
		"agent-clove": "Clove",
		"agent-raze":  "Raze",
	}
	weaponsMap := map[string]string{
		"w-vandal":  "Vandal",
		"w-phantom": "Phantom",
	}

	// 1. Tactical Match 1: Competitive (Jett)
	compMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-comp",
			QueueID:  "competitive",
			GameMode: "/Game/GameModes/Bomb/BombGameMode.BombGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				TeamID:      "Blue",
				CharacterID: "agent-jett",
				Stats: models.PlayerStats{
					Score:        4500,
					RoundsPlayed: 20,
					Kills:        20,
					Deaths:       10,
					Assists:      5,
				},
			},
		},
		Teams: []models.MatchTeam{
			{TeamID: "Blue", Won: true, RoundsWon: 13},
			{TeamID: "Red", Won: false, RoundsWon: 7},
		},
	}

	// 2. Tactical Match 2: Unrated (Sova)
	unratedMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-unrated",
			QueueID:  "unrated",
			GameMode: "/Game/GameModes/Bomb/BombGameMode.BombGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				TeamID:      "Red",
				CharacterID: "agent-sova",
				Stats: models.PlayerStats{
					Score:        3000,
					RoundsPlayed: 18,
					Kills:        15,
					Deaths:       12,
					Assists:      8,
				},
			},
		},
		Teams: []models.MatchTeam{
			{TeamID: "Red", Won: false, RoundsWon: 5},
			{TeamID: "Blue", Won: true, RoundsWon: 13},
		},
	}

	// 3. FFA Deathmatch (Empty QueueID, GameMode asset path) — THE BUG SCENARIO (Gekko)
	dmEmptyQueueMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-dm-empty-queue",
			QueueID:  "",
			GameMode: "/Game/GameModes/Deathmatch/DeathmatchGameMode.DeathmatchGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				CharacterID: "agent-gekko",
				Stats: models.PlayerStats{
					Score:        6800,
					RoundsPlayed: 1,
					Kills:        32,
					Deaths:       25,
					Assists:      4,
				},
			},
		},
	}

	// 4. FFA Deathmatch (Explicit QueueID "deathmatch") (Reyna)
	dmExplicitMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-dm-explicit",
			QueueID:  "deathmatch",
			GameMode: "Deathmatch",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				CharacterID: "agent-reyna",
				Stats: models.PlayerStats{
					Score:        8000,
					RoundsPlayed: 1,
					Kills:        40,
					Deaths:       20,
					Assists:      2,
				},
			},
		},
	}

	// 5. Team Deathmatch / Hurm (Empty QueueID, Hurm GameMode) (Clove)
	tdmMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-tdm",
			QueueID:  "",
			GameMode: "/Game/GameModes/Hurm/HurmGameMode.HurmGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				CharacterID: "agent-clove",
				Stats: models.PlayerStats{
					Score:        5000,
					RoundsPlayed: 1,
					Kills:        25,
					Deaths:       18,
					Assists:      6,
				},
			},
		},
	}

	// 6. Escalation / GGTeam (Explicit QueueID "ggteam") (Raze)
	escalationMatch := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID:  "m-escalation",
			QueueID:  "ggteam",
			GameMode: "/Game/GameModes/GunGame/GGTeamGameMode.GGTeamGameMode_C",
		},
		Players: []models.MatchPlayer{
			{
				Subject:     puUID,
				CharacterID: "agent-raze",
				Stats: models.PlayerStats{
					Score:        3500,
					RoundsPlayed: 1,
					Kills:        18,
					Deaths:       14,
					Assists:      3,
				},
			},
		},
	}

	allMatches := []*models.MatchDetails{
		compMatch,
		dmEmptyQueueMatch,
		unratedMatch,
		dmExplicitMatch,
		tdmMatch,
		escalationMatch,
	}

	sm := NewStatsModel(allMatches, puUID, nil, agentsMap, weaponsMap, "Diamond 2")
	sm.SetSize(120, 35)
	view := sm.View()

	// Verification 1: Only 2 tactical matches counted in hsHistory
	if len(sm.hsHistory) != 2 {
		t.Errorf("expected hsHistory length 2 (tactical matches only), got %d", len(sm.hsHistory))
	}

	// Verification 2: View header reflects "Last 2 matches"
	if !strings.Contains(view, "Last 2 matches") {
		t.Errorf("expected view to contain 'Last 2 matches', got:\n%s", view)
	}

	// Verification 3: Only 2 agents in agentStats (Jett, Sova)
	if len(sm.agentStats) != 2 {
		t.Errorf("expected 2 agent stats, got %d", len(sm.agentStats))
	}

	// Verification 4: Tactical agents exist in View
	if !strings.Contains(view, "Jett") {
		t.Errorf("expected view to contain Jett")
	}
	if !strings.Contains(view, "Sova") {
		t.Errorf("expected view to contain Sova")
	}

	// Verification 5: Casual agents MUST NOT exist in agentStats or View
	excludedAgents := []string{"Gekko", "Reyna", "Clove", "Raze"}
	for _, agent := range sm.agentStats {
		for _, excl := range excludedAgents {
			if agent.AgentName == excl {
				t.Errorf("excluded agent %s found in agentStats", excl)
			}
		}
	}
	for _, excl := range excludedAgents {
		if strings.Contains(view, excl) {
			t.Errorf("excluded agent %s found in View() output", excl)
		}
	}

	// Verification 6: Jett stats not distorted (1 match, 20 rounds, ACS = 4500/20 = 225)
	var jettStat *AgentStat
	for i := range sm.agentStats {
		if sm.agentStats[i].AgentName == "Jett" {
			jettStat = &sm.agentStats[i]
			break
		}
	}
	if jettStat == nil {
		t.Fatalf("jettStat not found")
	}
	if jettStat.Matches != 1 || jettStat.Wins != 1 || jettStat.TotalRounds != 20 {
		t.Errorf("unexpected Jett stats: Matches=%d Wins=%d Rounds=%d",
			jettStat.Matches, jettStat.Wins, jettStat.TotalRounds)
	}
}

func TestNightMarketModel_2Col_Alignment(t *testing.T) {
	skins := []models.ResolvedSkin{
		{UUID: "nm-1", DisplayName: "Prime Vandal", Rarity: "Select", CostVP: 1775},
		{UUID: "nm-2", DisplayName: "Reaver Phantom", Rarity: "Premium", CostVP: 1775},
		{UUID: "nm-3", DisplayName: "Sovereign Sword", Rarity: "Exclusive", CostVP: 3550},
		{UUID: "nm-4", DisplayName: "Ion Sheriff", Rarity: "Premium", CostVP: 1775},
		{UUID: "nm-5", DisplayName: "Glitchpop Dagger", Rarity: "Exclusive", CostVP: 4350},
		{UUID: "nm-6", DisplayName: "Magepunk Ghost", Rarity: "Premium", CostVP: 1775},
	}
	discounts := []int{10, 20, 30, 40, 45, 50}

	m := NewNightMarketModel(skins, discounts)
	m.SetSize(90, 40) // width 90 < 98 triggers 2-column layout
	view := m.View()

	// Verify all skin names and discounts are present
	for _, skin := range skins {
		if !strings.Contains(view, skin.DisplayName) {
			t.Errorf("expected view to contain %q", skin.DisplayName)
		}
	}
	for _, disc := range discounts {
		discStr := fmt.Sprintf("-%d%%", disc)
		if !strings.Contains(view, discStr) {
			t.Errorf("expected view to contain discount %q", discStr)
		}
	}

	// In Lipgloss, Width(cardContentWidth) includes padding (0,1), plus 2 border columns.
	// Outer card width = cardContentWidth + 2 = 42.
	// Row width with 1 spacer = 2 * 42 + 1 = 85.
	expectedRowWidth := 2*(40+2) + 1
	lines := strings.Split(view, "\n")
	matchingLineCount := 0
	for _, line := range lines {
		w := lipgloss.Width(line)
		if w == expectedRowWidth {
			matchingLineCount++
		}
	}
	// 3 rows * 11 lines per card = 33 lines
	if matchingLineCount != 33 {
		t.Errorf("expected 33 grid lines with width %d (3 rows of 2 cards), got %d", expectedRowWidth, matchingLineCount)
	}
}

func TestNightMarketModel_3Col_Alignment(t *testing.T) {
	skins := []models.ResolvedSkin{
		{UUID: "nm-1", DisplayName: "Prime Vandal", Rarity: "Select", CostVP: 1775},
		{UUID: "nm-2", DisplayName: "Reaver Phantom", Rarity: "Premium", CostVP: 1775},
		{UUID: "nm-3", DisplayName: "Sovereign Sword", Rarity: "Exclusive", CostVP: 3550},
		{UUID: "nm-4", DisplayName: "Ion Sheriff", Rarity: "Premium", CostVP: 1775},
		{UUID: "nm-5", DisplayName: "Glitchpop Dagger", Rarity: "Exclusive", CostVP: 4350},
		{UUID: "nm-6", DisplayName: "Magepunk Ghost", Rarity: "Premium", CostVP: 1775},
	}
	discounts := []int{10, 20, 30, 40, 45, 50}

	m := NewNightMarketModel(skins, discounts)
	m.SetSize(160, 40) // width 160 >= 140 triggers 3-column layout
	view := m.View()

	// Verify all 6 skins render properly without line overflow
	for _, skin := range skins {
		if !strings.Contains(view, skin.DisplayName) {
			t.Errorf("expected view to contain %q", skin.DisplayName)
		}
	}
	for _, disc := range discounts {
		discStr := fmt.Sprintf("-%d%%", disc)
		if !strings.Contains(view, discStr) {
			t.Errorf("expected view to contain discount %q", discStr)
		}
	}

	// In 3-col mode at width 160: cardContentWidth = 40.
	// Outer card width = 40 + 2 = 42.
	// Row width with 2 spacers = 3 * 42 + 2 = 128.
	expectedRowWidth := 3*(40+2) + 2
	lines := strings.Split(view, "\n")
	matchingLineCount := 0
	for _, line := range lines {
		w := lipgloss.Width(line)
		if w == expectedRowWidth {
			matchingLineCount++
		}
	}
	// 2 rows * 11 lines per card = 22 lines
	if matchingLineCount != 22 {
		t.Errorf("expected 22 grid lines with width %d (2 rows of 3 cards), got %d", expectedRowWidth, matchingLineCount)
	}
}

func TestNightMarketModel_WishlistBadge(t *testing.T) {
	skin := models.ResolvedSkin{
		UUID:        "nm-wishlist-uuid",
		DisplayName: "Araxys Vandal",
		Rarity:      "Exclusive",
		CostVP:      2175,
	}
	_ = cache.AddToWishlist(cache.ConvertResolvedSkinToWishlist(skin))
	defer cache.RemoveFromWishlist(skin.UUID)

	m := NewNightMarketModel([]models.ResolvedSkin{skin}, []int{30})
	m.SetSize(120, 30)
	view := m.View()

	if !strings.Contains(view, "WISHLIST ITEM") {
		t.Errorf("expected view to contain 'WISHLIST ITEM' badge, got:\n%s", view)
	}
}

func TestNightMarketModel_BreakpointThreshold(t *testing.T) {
	skins := []models.ResolvedSkin{
		{UUID: "nm-1", DisplayName: "Prime Vandal", CostVP: 1775},
		{UUID: "nm-2", DisplayName: "Reaver Phantom", CostVP: 1775},
		{UUID: "nm-3", DisplayName: "Sovereign Sword", CostVP: 3550},
		{UUID: "nm-4", DisplayName: "Ion Sheriff", CostVP: 1775},
		{UUID: "nm-5", DisplayName: "Glitchpop Dagger", CostVP: 4350},
		{UUID: "nm-6", DisplayName: "Magepunk Ghost", CostVP: 1775},
	}
	discounts := []int{10, 20, 30, 40, 45, 50}

	// Width 97 (< 98) -> 2 columns (3 rows of 2 cards)
	m97 := NewNightMarketModel(skins, discounts)
	m97.SetSize(97, 40)
	view97 := m97.View()
	// At width 97, cardContentWidth = (97 - 5) / 2 = 46 -> clamped to 40
	// Card width = 42, row width = 2 * 42 + 1 = 85
	width97Count := 0
	for _, line := range strings.Split(view97, "\n") {
		if lipgloss.Width(line) == 85 {
			width97Count++
		}
	}
	if width97Count != 33 {
		t.Errorf("expected 33 lines of width 85 for 2-column layout at width 97, got %d", width97Count)
	}

	// Width 98 (>= 98) -> 3 columns (2 rows of 3 cards)
	m98 := NewNightMarketModel(skins, discounts)
	m98.SetSize(98, 40)
	view98 := m98.View()
	// At width 98, cardContentWidth = (98 - 8) / 3 = 30
	// Card width = 30 + 2 = 32, row width = 3 * 32 + 2 = 98
	expectedRowWidth98 := 3*(30+2) + 2
	width98Count := 0
	for _, line := range strings.Split(view98, "\n") {
		if lipgloss.Width(line) == expectedRowWidth98 {
			width98Count++
		}
	}
	if width98Count != 22 {
		t.Errorf("expected 22 lines of width %d for 3-column layout at width 98, got %d", expectedRowWidth98, width98Count)
	}

	// Width 120 (14-inch laptop viewport) -> 3 columns (2 rows of 3 cards)
	m120 := NewNightMarketModel(skins, discounts)
	m120.SetSize(120, 40)
	view120 := m120.View()
	// At width 120, cardContentWidth = (120 - 8) / 3 = 37
	// Card width = 37 + 2 = 39, row width = 3 * 39 + 2 = 119
	expectedRowWidth120 := 3*(37+2) + 2
	width120Count := 0
	for _, line := range strings.Split(view120, "\n") {
		if lipgloss.Width(line) == expectedRowWidth120 {
			width120Count++
		}
	}
	if width120Count != 22 {
		t.Errorf("expected 22 lines of width %d for 3-column layout at width 120, got %d", expectedRowWidth120, width120Count)
	}
}

func TestResolveQueueDisplayName_Gauntlet(t *testing.T) {
	if name := ResolveQueueDisplayName("abilitydraftarena", ""); name != "Gauntlet" {
		t.Errorf("ResolveQueueDisplayName(abilitydraftarena) = %q, want 'Gauntlet'", name)
	}
	if name := ResolveQueueDisplayName("abilitydraft", ""); name != "Gauntlet" {
		t.Errorf("ResolveQueueDisplayName(abilitydraft) = %q, want 'Gauntlet'", name)
	}
	if name := ResolveQueueDisplayName("gauntlet", ""); name != "Gauntlet" {
		t.Errorf("ResolveQueueDisplayName(gauntlet) = %q, want 'Gauntlet'", name)
	}
}

func TestResolveAgentDisplayName_Gauntlet(t *testing.T) {
	md := &models.MatchDetails{
		MatchInfo: models.MatchInfo{QueueID: "abilitydraftarena"},
	}
	agentsMap := map[string]string{"dade69b4-4f5a-8528-247b-219e5a1facd6": "Fade"}
	if name := ResolveAgentDisplayName("dade69b4-4f5a-8528-247b-219e5a1facd6", agentsMap, md); name != "-" {
		t.Errorf("expected '-', got %q", name)
	}
}

func TestMatchesModel_GauntletDetailView(t *testing.T) {
	teams := make([]models.MatchTeam, 8)
	players := make([]models.MatchPlayer, 16)
	for i := 0; i < 8; i++ {
		tID := fmt.Sprintf("Team_%d", i+1)
		won := (i == 3)
		rounds := 0
		if won {
			rounds = 8
		}
		teams[i] = models.MatchTeam{TeamID: tID, Won: won, RoundsWon: rounds}
		players[i*2] = models.MatchPlayer{
			Subject:  fmt.Sprintf("player-%d-a", i+1),
			TeamID:   tID,
			GameName: fmt.Sprintf("Player%d", i+1),
		}
		players[i*2+1] = models.MatchPlayer{
			Subject:  fmt.Sprintf("player-%d-b", i+1),
			TeamID:   tID,
			GameName: fmt.Sprintf("Partner%d", i+1),
		}
	}

	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID: "gauntlet-1",
			QueueID: "abilitydraftarena",
			MapID:   "/Game/Maps/Gauntlet/Gauntlet",
		},
		Teams:   teams,
		Players: players,
	}

	item := MatchItem{
		MatchID:     "gauntlet-1",
		MapName:     "Gauntlet",
		QueueName:   ResolveQueueDisplayName("abilitydraftarena", ""),
		Outcome:     details.GetMatchOutcome("player-6-a"),
		Score:       details.ScoreString("player-6-a"),
		Details:     details,
		PlayerPUUID: "player-6-a",
	}

	m := NewMatchesModel([]MatchItem{item}, "player-6-a", nil, nil)
	m.SetSize(120, 60)

	// Enter detail view
	detailModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	view := detailModel.View()

	// Verify Header does NOT contain 0-0 and contains LOSS
	if strings.Contains(view, "LOSS  0-0") {
		t.Errorf("header must NOT contain 'LOSS  0-0', got:\n%s", view)
	}
	if !strings.Contains(view, "MATCH DETAIL — Gauntlet (Gauntlet)") {
		t.Errorf("header missing title, got:\n%s", view)
	}

	// Verify all 8 teams are rendered
	for i := 1; i <= 8; i++ {
		expectedTeam := fmt.Sprintf("TEAM_%d TEAM", i)
		if !strings.Contains(view, expectedTeam) {
			t.Errorf("expected view to contain %s, got:\n%s", expectedTeam, view)
		}
	}

	// Verify Agent and Econ are NOT in the table
	if strings.Contains(view, "Agent") {
		t.Errorf("detail scoreboard should NOT display 'Agent', got:\n%s", view)
	}
	if strings.Contains(view, "Econ") {
		t.Errorf("detail scoreboard should NOT display 'Econ', got:\n%s", view)
	}

	// Verify table header contains streamlined columns
	if !strings.Contains(view, "Player") || !strings.Contains(view, "ACS") || !strings.Contains(view, "HS%") || !strings.Contains(view, "ADR") {
		t.Errorf("detail scoreboard missing streamlined columns, got:\n%s", view)
	}

	// Verify rounds won suffix is completely omitted
	if strings.Contains(view, "rounds won") || strings.Contains(view, "round won") {
		t.Errorf("expected rounds won suffix to be omitted from Gauntlet header, got:\n%s", view)
	}
}

func TestMatchesModel_DetailViewScrolling(t *testing.T) {
	teams := make([]models.MatchTeam, 8)
	players := make([]models.MatchPlayer, 16)
	for i := 0; i < 8; i++ {
		tID := fmt.Sprintf("Team_%d", i+1)
		teams[i] = models.MatchTeam{TeamID: tID, Won: i == 0, RoundsWon: 8 - i}
		players[i*2] = models.MatchPlayer{Subject: fmt.Sprintf("p%d-a", i+1), TeamID: tID, GameName: fmt.Sprintf("Player%d", i+1)}
		players[i*2+1] = models.MatchPlayer{Subject: fmt.Sprintf("p%d-b", i+1), TeamID: tID, GameName: fmt.Sprintf("Partner%d", i+1)}
	}

	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{MatchID: "gauntlet-scroll", QueueID: "abilitydraftarena"},
		Teams:     teams,
		Players:   players,
	}

	item := MatchItem{
		MatchID:     "gauntlet-scroll",
		MapName:     "Gauntlet",
		QueueName:   "Gauntlet",
		Details:     details,
		PlayerPUUID: "p1-a",
	}

	m := NewMatchesModel([]MatchItem{item}, "p1-a", nil, nil)
	// Small terminal height to force scrolling
	m.SetSize(120, 20)

	// Enter detail view
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if m.detailScrollOffset != 0 {
		t.Errorf("expected initial detailScrollOffset=0, got %d", m.detailScrollOffset)
	}

	view0 := m.View()
	if !strings.Contains(view0, "[1-15/") {
		t.Errorf("expected scroll indicator [1-15/...], got:\n%s", view0)
	}
	if !strings.Contains(view0, "TEAM_1 TEAM (YOUR TEAM)") {
		t.Errorf("expected top team in initial scroll view, got:\n%s", view0)
	}

	// Scroll down with j
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.detailScrollOffset != 1 {
		t.Errorf("expected detailScrollOffset=1 after 'j', got %d", m.detailScrollOffset)
	}

	// Scroll down with down arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.detailScrollOffset != 2 {
		t.Errorf("expected detailScrollOffset=2 after 'down', got %d", m.detailScrollOffset)
	}

	// Scroll pgdown (+8)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	if m.detailScrollOffset != 10 {
		t.Errorf("expected detailScrollOffset=10 after 'pgdown', got %d", m.detailScrollOffset)
	}

	// Scroll end
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	if m.detailScrollOffset == 0 {
		t.Errorf("expected clamped max scroll offset on 'end', got 0")
	}

	// Scroll home
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyHome})
	if m.detailScrollOffset != 0 {
		t.Errorf("expected detailScrollOffset=0 on 'home', got %d", m.detailScrollOffset)
	}

	// Scroll up with k
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.detailScrollOffset != 0 {
		t.Errorf("expected detailScrollOffset clamped at 0 on 'k', got %d", m.detailScrollOffset)
	}

	// Return to list with esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.viewMode != MatchViewList {
		t.Errorf("expected viewMode=MatchViewList on 'esc', got %v", m.viewMode)
	}
	if m.detailScrollOffset != 0 {
		t.Errorf("expected detailScrollOffset reset to 0 on 'esc', got %d", m.detailScrollOffset)
	}
}

func TestMatchesModel_GauntletListView(t *testing.T) {
	details := &models.MatchDetails{
		MatchInfo: models.MatchInfo{
			MatchID: "gauntlet-1",
			QueueID: "abilitydraftarena",
			MapID:   "/Game/Maps/Gauntlet/Gauntlet",
		},
		Teams: []models.MatchTeam{
			{TeamID: "Team_1", Won: true, RoundsWon: 8},
			{TeamID: "Team_6", Won: false, RoundsWon: 0},
		},
		Players: []models.MatchPlayer{
			{Subject: "player-6-a", TeamID: "Team_6", Stats: models.PlayerStats{Kills: 5, Deaths: 6, Assists: 2}},
			{Subject: "player-1-a", TeamID: "Team_1", Stats: models.PlayerStats{Kills: 15, Deaths: 2, Assists: 4}},
		},
	}

	item := MatchItem{
		MatchID:     "gauntlet-1",
		MapName:     "Gauntlet",
		QueueName:   ResolveQueueDisplayName("abilitydraftarena", ""),
		AgentName:   ResolveAgentDisplayName("non-agent-uuid", nil, details),
		Outcome:     details.GetMatchOutcome("player-6-a"),
		Score:       details.ScoreString("player-6-a"),
		Details:     details,
		PlayerPUUID: "player-6-a",
		Kills:       5,
		Deaths:      6,
		Assists:     2,
	}

	m := NewMatchesModel([]MatchItem{item}, "player-6-a", nil, nil)
	m.SetSize(120, 40)
	listView := m.renderListView()

	if !strings.Contains(listView, "Gauntlet") {
		t.Errorf("expected list view to contain 'Gauntlet', got:\n%s", listView)
	}
	if strings.Contains(listView, "Abilitydra..") {
		t.Errorf("list view should not truncate queue to 'Abilitydra..', got:\n%s", listView)
	}
	if strings.Contains(listView, "Agent") {
		t.Errorf("list view should not display 'Agent', got:\n%s", listView)
	}
	if !strings.Contains(listView, "2nd") {
		t.Errorf("expected list view to contain '2nd' finish, got:\n%s", listView)
	}
}



