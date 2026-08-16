package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/api"
	"github.com/val-tracker/val-tracker/internal/auth"
	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
)

// ── Tab Identifiers ─────────────────────────────────────────────────

type Tab int

const (
	TabShop Tab = iota
	TabNightMarket
	TabBattlepass
)

// ── Messages ────────────────────────────────────────────────────────

// DataLoadedMsg is sent after all API data has been fetched.
type DataLoadedMsg struct {
	ShopSkins     []models.ResolvedSkin
	NightMarket   []models.ResolvedSkin // nil if inactive
	NMDiscounts   []int                 // parallel to NightMarket
	Battlepass    *BattlepassData
	TimeRemaining int // seconds until shop reset
	Err           error
}

type BattlepassData struct {
	CurrentTier     int
	MaxTier         int
	XPInCurrentTier int
	XPForNextTier   int
	TotalXP         int
}

// RefreshMsg triggers a data reload.
type RefreshMsg struct{}

// ── MainModel ───────────────────────────────────────────────────────

type MainModel struct {
	activeTab     Tab
	tabs          []Tab
	tabNames      map[Tab]string
	shopModel     ShopModel
	nightModel    NightMarketModel
	bpModel       BattlepassModel
	session       *models.Session
	regionModel   RegionSelectModel
	needsRegion   bool
	loading       bool
	err           error
	width, height int
}

func NewMainModel(session *models.Session) MainModel {
	tabs := []Tab{TabShop, TabBattlepass}
	tabNames := map[Tab]string{
		TabShop:        "  Daily Shop",
		TabNightMarket: " ✦ Night Market",
		TabBattlepass:  "  Battlepass",
	}

	needsRegion := session.Region == ""

	return MainModel{
		activeTab:   TabShop,
		tabs:        tabs,
		tabNames:    tabNames,
		session:     session,
		needsRegion: needsRegion,
		regionModel: NewRegionSelectModel(),
		loading:     !needsRegion,
	}
}

func (m MainModel) Init() tea.Cmd {
	if m.needsRegion {
		return m.regionModel.Init()
	}
	return func() tea.Msg { return RefreshMsg{} }
}

func (m MainModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.regionModel.width = msg.Width
		m.regionModel.height = msg.Height
		return m, nil

	case RegionSelectedMsg:
		m.session.Region = msg.Region
		if shard, ok := auth.RegionToShard[msg.Region]; ok {
			m.session.Shard = shard
		} else {
			m.session.Shard = msg.Region
		}
		m.needsRegion = false
		m.loading = true
		return m, func() tea.Msg { return RefreshMsg{} }

	case tea.KeyMsg:
		if m.needsRegion {
			var cmd tea.Cmd
			var model tea.Model
			model, cmd = m.regionModel.Update(msg)
			if rm, ok := model.(RegionSelectModel); ok {
				m.regionModel = rm
			}
			return m, cmd
		}

		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "r":
			m.loading = true
			sprite.ClearCache()
			return m, func() tea.Msg { return RefreshMsg{} }
		case "left", "h":
			m.prevTab()
			return m, nil
		case "right", "l":
			m.nextTab()
			return m, nil
		case "1":
			m.activeTab = TabShop
			return m, nil
		case "2":
			if m.hasNightMarket() {
				m.activeTab = TabNightMarket
			} else {
				m.activeTab = TabBattlepass
			}
			return m, nil
		case "3":
			if m.hasNightMarket() {
				m.activeTab = TabBattlepass
			}
			return m, nil
		}

	case RefreshMsg:
		m.loading = true
		return m, m.loadData

	case DataLoadedMsg:
		m.loading = false
		if msg.Err != nil {
			m.err = msg.Err
			return m, nil
		}
		m.err = nil

		// Update sub-models
		m.shopModel = NewShopModel(msg.ShopSkins, msg.TimeRemaining)

		if msg.NightMarket != nil && len(msg.NightMarket) > 0 {
			m.nightModel = NewNightMarketModel(msg.NightMarket, msg.NMDiscounts)
			if !m.hasNightMarket() {
				m.tabs = []Tab{TabShop, TabNightMarket, TabBattlepass}
			}
		} else {
			m.tabs = []Tab{TabShop, TabBattlepass}
			if m.activeTab == TabNightMarket {
				m.activeTab = TabShop
			}
		}

		if msg.Battlepass != nil {
			m.bpModel = NewBattlepassModel(msg.Battlepass)
		}
		return m, nil
	}

	if m.needsRegion {
		var cmd tea.Cmd
		var model tea.Model
		model, cmd = m.regionModel.Update(msg)
		if rm, ok := model.(RegionSelectModel); ok {
			m.regionModel = rm
		}
		return m, cmd
	}

	// Route to active sub-model
	var cmd tea.Cmd
	switch m.activeTab {
	case TabShop:
		m.shopModel, cmd = m.shopModel.Update(msg)
	case TabNightMarket:
		m.nightModel, cmd = m.nightModel.Update(msg)
	case TabBattlepass:
		m.bpModel, cmd = m.bpModel.Update(msg)
	}
	return m, cmd
}

func (m MainModel) View() string {
	if m.needsRegion {
		return m.regionModel.View()
	}
	if m.loading {
		return m.renderLoading()
	}
	if m.err != nil {
		return m.renderError()
	}

	var sb strings.Builder

	// ── Header ──────────────────────────────────────────────────
	header := TitleStyle.Render("⚡ VAL-TRACKER TUI")
	regionTag := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(fmt.Sprintf(" [%s]", strings.ToUpper(m.session.Region)))
	sb.WriteString(header + regionTag + "\n\n")

	// ── Tabs ────────────────────────────────────────────────────
	var tabs []string
	for _, t := range m.tabs {
		name := m.tabNames[t]
		if t == m.activeTab {
			tabs = append(tabs, ActiveTabStyle.Render(name))
		} else {
			tabs = append(tabs, InactiveTabStyle.Render(name))
		}
	}
	sb.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n")
	lineLen := m.width
	if lineLen <= 0 {
		lineLen = 80
	}
	sb.WriteString(strings.Repeat("─", lineLen) + "\n\n")

	// ── Active Panel ────────────────────────────────────────────
	switch m.activeTab {
	case TabShop:
		sb.WriteString(m.shopModel.View())
	case TabNightMarket:
		sb.WriteString(m.nightModel.View())
	case TabBattlepass:
		sb.WriteString(m.bpModel.View())
	}

	// ── Status Bar ──────────────────────────────────────────────
	sb.WriteString("\n\n")
	sb.WriteString(StatusBarStyle.Render(
		"  ←/→ switch tabs  •  r refresh  •  q quit",
	))

	content := sb.String()
	if m.width > 0 && m.height > 0 {
		return AppStyle.Width(m.width).Height(m.height).Render(content)
	}
	return AppStyle.Render(content)
}

// ── Helpers ─────────────────────────────────────────────────────────

func (m MainModel) hasNightMarket() bool {
	for _, t := range m.tabs {
		if t == TabNightMarket {
			return true
		}
	}
	return false
}

func (m *MainModel) nextTab() {
	for i, t := range m.tabs {
		if t == m.activeTab && i < len(m.tabs)-1 {
			m.activeTab = m.tabs[i+1]
			return
		}
	}
}

func (m *MainModel) prevTab() {
	for i, t := range m.tabs {
		if t == m.activeTab && i > 0 {
			m.activeTab = m.tabs[i-1]
			return
		}
	}
}

func (m MainModel) renderLoading() string {
	content := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render("⟳  Loading your Valorant store...")

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return "\n\n  " + content + "\n"
}

func (m MainModel) renderError() string {
	content := ErrorStyle.Render(fmt.Sprintf("✕  %s\n\nPress r to retry, q to quit.", m.err.Error()))
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return "\n\n  " + content + "\n"
}

// loadData is the tea.Cmd that fetches all remote data and returns DataLoadedMsg.
func (m MainModel) loadData() tea.Msg {
	client := api.NewClient(m.session)

	// 1. Fetch storefront
	sf, err := client.FetchStorefront()
	if err != nil {
		return DataLoadedMsg{Err: fmt.Errorf("storefront: %w", err)}
	}

	// 2. Load skin assets (cached)
	skins, err := cache.LoadOrFetchSkins(m.session.ClientVersion)
	if err != nil {
		return DataLoadedMsg{Err: fmt.Errorf("skin cache: %w", err)}
	}
	lookup := cache.BuildSkinLookup(skins)

	// VP currency UUID
	const vpUUID = "85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"

	// 3. Resolve daily shop skins
	var shopSkins []models.ResolvedSkin
	for _, offer := range sf.SkinsPanelLayout.SingleItemStoreOffers {
		skinID := offer.OfferID
		if len(offer.Rewards) > 0 {
			skinID = offer.Rewards[0].ItemID
		}
		asset, ok := lookup[skinID]
		if !ok {
			continue
		}

		tierUUID := ""
		if asset.ContentTierUUID != nil {
			tierUUID = *asset.ContentTierUUID
		}

		iconURL := ""
		if asset.DisplayIcon != nil {
			iconURL = *asset.DisplayIcon
		}
		if iconURL == "" && len(asset.Levels) > 0 && asset.Levels[0].DisplayIcon != nil {
			iconURL = *asset.Levels[0].DisplayIcon
		}

		shopSkins = append(shopSkins, models.ResolvedSkin{
			UUID:        skinID,
			DisplayName: asset.DisplayName,
			Rarity:      RarityNameMap[tierUUID],
			CostVP:      offer.Cost[vpUUID],
			IconURL:     iconURL,
			Sprite:      sprite.Render(iconURL),
		})
	}

	// 4. Resolve night market (if active)
	var nightSkins []models.ResolvedSkin
	var nmDiscounts []int
	if sf.BonusStore != nil && len(sf.BonusStore.BonusStoreOffers) > 0 {
		for _, bo := range sf.BonusStore.BonusStoreOffers {
			skinID := bo.Offer.OfferID
			if len(bo.Offer.Rewards) > 0 {
				skinID = bo.Offer.Rewards[0].ItemID
			}
			asset, ok := lookup[skinID]
			if !ok {
				continue
			}

			tierUUID := ""
			if asset.ContentTierUUID != nil {
				tierUUID = *asset.ContentTierUUID
			}

			iconURL := ""
			if asset.DisplayIcon != nil {
				iconURL = *asset.DisplayIcon
			}
			if iconURL == "" && len(asset.Levels) > 0 && asset.Levels[0].DisplayIcon != nil {
				iconURL = *asset.Levels[0].DisplayIcon
			}

			// Calculate discounted price
			cost := bo.Offer.Cost[vpUUID]
			if len(bo.DiscountCosts) > 0 && bo.DiscountCosts[vpUUID] > 0 {
				cost = bo.DiscountCosts[vpUUID]
			} else if len(bo.DiscountedCost) > 0 && bo.DiscountedCost[vpUUID] > 0 {
				cost = bo.DiscountedCost[vpUUID]
			} else if bo.DiscountPercent > 0 && cost > 0 {
				cost = cost - (cost * bo.DiscountPercent / 100)
			}

			nightSkins = append(nightSkins, models.ResolvedSkin{
				UUID:        skinID,
				DisplayName: asset.DisplayName,
				Rarity:      RarityNameMap[tierUUID],
				CostVP:      cost,
				IconURL:     iconURL,
				Sprite:      sprite.Render(iconURL),
			})
			nmDiscounts = append(nmDiscounts, bo.DiscountPercent)
		}
	}

	// 5. Fetch battlepass progression
	contracts, err := client.FetchContracts()
	var bpData *BattlepassData
	if err == nil {
		content, cerr := client.FetchContent()
		if cerr == nil {
			if bp, found := api.FindActiveBattlepass(contracts, content); found {
				bpData = &BattlepassData{
					CurrentTier:     bp.ProgressionLevelReached,
					MaxTier:         55,
					XPInCurrentTier: bp.ProgressionTowardsNextLevel,
					XPForNextTier:   bp.ContractProgression.TotalProgressionTowardsNextLevel,
					TotalXP:         bp.ContractProgression.TotalProgressionEarned,
				}
			}
		}
	}

	return DataLoadedMsg{
		ShopSkins:     shopSkins,
		NightMarket:   nightSkins,
		NMDiscounts:   nmDiscounts,
		Battlepass:    bpData,
		TimeRemaining: sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds,
	}
}
