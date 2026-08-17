package tui

import (
	"fmt"
	"strings"
	"sync"

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
	Wallet        *models.WalletResponse
	MMR           *models.MMRResponse
	RankName      string
	Missions      []models.Mission
	RanksMap      map[int]string
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
	progressModel ProgressModel
	bpModel       ProgressModel
	session       *models.Session
	regionModel   RegionSelectModel
	needsRegion   bool
	loading       bool
	err           error
	width, height int
	wallet        *models.WalletResponse
	mmr           *models.MMRResponse
	rankName      string
	missions      []models.Mission
	ranksMap      map[int]string
}

func NewMainModel(session *models.Session) MainModel {
	tabs := []Tab{TabShop, TabBattlepass}
	tabNames := map[Tab]string{
		TabShop:        "  Daily Shop",
		TabNightMarket: " ✦ Night Market",
		TabBattlepass:  "  Progress",
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
		ranksMap:    cache.DefaultRankNames,
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
		headerHeight := 5
		contentHeight := msg.Height - headerHeight
		if contentHeight < 0 {
			contentHeight = 0
		}
		m.shopModel.SetSize(msg.Width, contentHeight)
		m.nightModel.SetSize(msg.Width, contentHeight)
		m.progressModel.SetSize(msg.Width, contentHeight)
		m.bpModel = m.progressModel
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

		m.wallet = msg.Wallet
		m.mmr = msg.MMR
		m.rankName = msg.RankName
		m.missions = msg.Missions
		m.ranksMap = msg.RanksMap

		// Update sub-models
		headerHeight := 5
		contentHeight := m.height - headerHeight
		if contentHeight < 0 {
			contentHeight = 0
		}

		m.shopModel = NewShopModel(msg.ShopSkins, msg.TimeRemaining)
		m.shopModel.SetSize(m.width, contentHeight)

		if msg.NightMarket != nil && len(msg.NightMarket) > 0 {
			m.nightModel = NewNightMarketModel(msg.NightMarket, msg.NMDiscounts)
			m.nightModel.SetSize(m.width, contentHeight)
			if !m.hasNightMarket() {
				m.tabs = []Tab{TabShop, TabNightMarket, TabBattlepass}
			}
		} else {
			m.tabs = []Tab{TabShop, TabBattlepass}
			if m.activeTab == TabNightMarket {
				m.activeTab = TabShop
			}
		}

		m.progressModel = NewProgressModel(msg.Battlepass, msg.Missions)
		m.progressModel.SetSize(m.width, contentHeight)
		m.bpModel = m.progressModel
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
		m.progressModel, cmd = m.progressModel.Update(msg)
		m.bpModel = m.progressModel
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
	var headerParts []string
	headerParts = append(headerParts, TitleStyle.Render("⚡ VAL-TRACKER TUI"))
	regionTag := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render(fmt.Sprintf(" [%s]", strings.ToUpper(m.session.Region)))
	headerParts = append(headerParts, regionTag)

	if m.rankName != "" {
		tier, rr := 0, 0
		if m.mmr != nil {
			tier, rr = m.mmr.GetCurrentCompetitiveInfo()
		}
		rankStr := fmt.Sprintf(" ◆ %s (%d RR)", m.rankName, rr)
		if tier == 0 {
			rankStr = " ◆ " + m.rankName
		}
		rankBadge := lipgloss.NewStyle().
			Foreground(ColorExclusive).
			Bold(true).
			Render(" " + rankStr)
		headerParts = append(headerParts, rankBadge)
	}

	if m.wallet != nil {
		vp := m.wallet.VP()
		rp := m.wallet.RP()
		kc := m.wallet.KC()
		walletStr := fmt.Sprintf("💰 %s VP", formatNumber(vp))
		if rp > 0 {
			walletStr += fmt.Sprintf("  %d RP", rp)
		}
		if kc > 0 {
			walletStr += fmt.Sprintf("  %d KC", kc)
		}
		walletBadge := lipgloss.NewStyle().
			Foreground(ColorUltra).
			Bold(true).
			Render("  " + walletStr)
		headerParts = append(headerParts, walletBadge)
	}

	headerLine := lipgloss.JoinHorizontal(lipgloss.Center, headerParts...)
	sb.WriteString(headerLine + "\n\n")

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
		sb.WriteString(m.progressModel.View())
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

func formatNumber(n int) string {
	in := fmt.Sprintf("%d", n)
	if len(in) <= 3 {
		return in
	}
	var res []byte
	offset := len(in) % 3
	if offset > 0 {
		res = append(res, in[:offset]...)
		if len(in) > offset {
			res = append(res, ',')
		}
	}
	for i := offset; i < len(in); i += 3 {
		res = append(res, in[i:i+3]...)
		if i+3 < len(in) {
			res = append(res, ',')
		}
	}
	return string(res)
}

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

	var (
		sf        *models.StorefrontResponse
		sfErr     error
		skins     []models.SkinAsset
		skinsErr  error
		wallet    *models.WalletResponse
		mmr       *models.MMRResponse
		contracts *models.ContractsResponse
		content   *models.ContentResponse
		ranksMap  map[int]string
		wg        sync.WaitGroup
	)

	wg.Add(5)

	// 1. Fetch storefront
	go func() {
		defer wg.Done()
		sf, sfErr = client.FetchStorefront()
	}()

	// 2. Load skin assets (cached)
	go func() {
		defer wg.Done()
		skins, skinsErr = cache.LoadOrFetchSkins(m.session.ClientVersion)
	}()

	// 3. Fetch wallet
	go func() {
		defer wg.Done()
		wallet, _ = client.FetchWallet()
	}()

	// 4. Fetch MMR
	go func() {
		defer wg.Done()
		mmr, _ = client.FetchMMR()
	}()

	// 5. Fetch contracts and content
	go func() {
		defer wg.Done()
		contracts, _ = client.FetchContracts()
		content, _ = client.FetchContent()
	}()

	// Load ranks (cached)
	ranksMap, _ = cache.LoadOrFetchRanks(m.session.ClientVersion)

	wg.Wait()

	if sfErr != nil {
		return DataLoadedMsg{Err: fmt.Errorf("storefront: %w", sfErr)}
	}
	if skinsErr != nil {
		return DataLoadedMsg{Err: fmt.Errorf("skin cache: %w", skinsErr)}
	}

	lookup := cache.BuildSkinLookup(skins)

	// VP currency UUID
	const vpUUID = "85ad13f7-3d1b-5128-9eb2-7cd8ee0b5741"

	// Resolve daily shop skins
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
			Sprite:      sprite.Render(iconURL, 40),
		})
	}

	// Resolve night market (if active)
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
				Sprite:      sprite.Render(iconURL, 40),
			})
			nmDiscounts = append(nmDiscounts, bo.DiscountPercent)
		}
	}

	// Resolve battlepass progression & missions
	var bpData *BattlepassData
	var missions []models.Mission
	if contracts != nil {
		missions = contracts.Missions
		if content != nil {
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

	// Determine rank name
	rankName := "Unranked"
	if mmr != nil {
		tier, _ := mmr.GetCurrentCompetitiveInfo()
		rankName = cache.GetRankName(tier, ranksMap)
	}

	return DataLoadedMsg{
		ShopSkins:     shopSkins,
		NightMarket:   nightSkins,
		NMDiscounts:   nmDiscounts,
		Battlepass:    bpData,
		TimeRemaining: sf.SkinsPanelLayout.SingleItemOffersRemainingDurationInSeconds,
		Wallet:        wallet,
		MMR:           mmr,
		RankName:      rankName,
		Missions:      missions,
		RanksMap:      ranksMap,
	}
}
