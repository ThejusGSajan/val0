package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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
	view := model.View()

	if !strings.Contains(view, "NIGHT MARKET") {
		t.Errorf("expected view to contain 'NIGHT MARKET', got:\n%s", view)
	}
	if !strings.Contains(view, "-40%") {
		t.Errorf("expected view to contain '-40%%', got:\n%s", view)
	}

	// Empty case
	emptyModel := NewNightMarketModel(nil, nil)
	emptyView := emptyModel.View()
	if !strings.Contains(emptyView, "not currently active") {
		t.Errorf("expected inactive notice for empty Night Market, got:\n%s", emptyView)
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
	errModel.width = 80
	errModel.height = 24
	view := errModel.View()

	if !strings.Contains(view, "Please start the Riot Client first.") {
		t.Errorf("expected error message in view, got:\n%s", view)
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
	if m.activeTab != TabShop {
		t.Errorf("expected initial activeTab TabShop, got %v", m.activeTab)
	}

	// Next tab -> Battlepass
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRight})
	m = updated.(MainModel)
	if m.activeTab != TabBattlepass {
		t.Errorf("expected activeTab TabBattlepass, got %v", m.activeTab)
	}

	// Previous tab -> Shop
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyLeft})
	m = updated.(MainModel)
	if m.activeTab != TabShop {
		t.Errorf("expected activeTab TabShop, got %v", m.activeTab)
	}
}
