package tui

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
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
	if !strings.Contains(view, "450000") {
		t.Errorf("expected view to contain total XP, got:\n%s", view)
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
			ID: "daily-mission-1",
			Objectives: map[string]int{
				"obj-1": 10,
			},
			Complete: false,
		},
		{
			ID:       "daily-mission-2",
			Complete: true,
		},
	}

	pm := NewProgressModel(bpData, missions)
	pm.SetSize(80, 24)
	view := pm.View()

	if !strings.Contains(view, "BATTLEPASS — Tier 15 / 55") {
		t.Errorf("expected battlepass tier in view, got:\n%s", view)
	}
	if !strings.Contains(view, "ACTIVE MISSIONS") {
		t.Errorf("expected active missions in view, got:\n%s", view)
	}
	if !strings.Contains(view, "Completed") {
		t.Errorf("expected completed mission in view, got:\n%s", view)
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
	if !strings.Contains(view, "AIM ANALYSIS") {
		t.Errorf("expected AIM ANALYSIS in view, got:\n%s", view)
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
		MatchInfo: models.MatchInfo{MatchID: "new-match-3", MapID: "Ascent"},
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
	if !strings.Contains(view, "Press 's' to enter wishlist tab") {
		t.Errorf("expected help text in view, got:\n%s", view)
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
	if !strings.Contains(v0, "Test Error Message") || !strings.Contains(v0, "Press 'r' to retry") {
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
