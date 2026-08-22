package tui

import (
	"fmt"
	"strings"
	"sync"
	"time"

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
	TabStore Tab = iota
	TabMatches
	TabStats
	TabProgress
	TabSession
)

// Backwards compatibility aliases
const (
	TabShop        = TabStore
	TabNightMarket = TabStore
	TabBattlepass  = TabProgress
)

type StoreSubTab int

const (
	SubTabShop StoreSubTab = iota
	SubTabWishlist
	SubTabNightMarket
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
	MissionsMap   map[string]cache.MissionInfo
	MatchItems    []MatchItem
	MatchDetails  []*models.MatchDetails
	CompUpdates   *models.CompetitiveUpdatesResponse
	AgentsMap     map[string]string
	MapsMap       map[string]string
	WeaponsMap    map[string]string
	RanksMap      map[int]string
	AllSkins      []models.SkinAsset
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
	storeSubTab   StoreSubTab
	shopModel     ShopModel
	nightModel    NightMarketModel
	wishlistModel WishlistModel
	matchesModel  MatchesModel
	statsModel    StatsModel
	progressModel ProgressModel
	bpModel       ProgressModel
	sessionModel  SessionModel
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
	missionsMap   map[string]cache.MissionInfo
	ranksMap      map[int]string
	agentsMap     map[string]string
	mapsMap       map[string]string
	weaponsMap    map[string]string
}

func NewMainModel(session *models.Session) MainModel {
	tabs := []Tab{TabStore, TabMatches, TabStats, TabProgress, TabSession}
	tabNames := map[Tab]string{
		TabStore:    "[1] Store",
		TabMatches:  "[2] Matches",
		TabStats:    "[3] Stats",
		TabProgress: "[4] Progress",
		TabSession:  "[5] Session",
	}

	needsRegion := session.Region == ""

	return MainModel{
		activeTab:     TabStore,
		tabs:          tabs,
		tabNames:      tabNames,
		storeSubTab:   SubTabShop,
		session:       session,
		wishlistModel: NewWishlistModel(),
		sessionModel:  NewSessionModel(session.PUUID),
		needsRegion:   needsRegion,
		regionModel:   NewRegionSelectModel(),
		loading:       !needsRegion,
		missionsMap:   cache.DefaultMissions,
		ranksMap:      cache.DefaultRankNames,
		agentsMap:     cache.DefaultAgentNames,
		mapsMap:       cache.DefaultMapNames,
		weaponsMap:    cache.DefaultWeaponNames,
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
		chromeHeight := 8
		contentHeight := msg.Height - chromeHeight
		if contentHeight < 0 {
			contentHeight = 0
		}
		m.shopModel.SetSize(msg.Width, contentHeight)
		m.nightModel.SetSize(msg.Width, contentHeight)
		m.wishlistModel.SetSize(msg.Width, contentHeight)
		m.matchesModel.SetSize(msg.Width, contentHeight)
		m.statsModel.SetSize(msg.Width, contentHeight)
		m.progressModel.SetSize(msg.Width, contentHeight)
		m.bpModel = m.progressModel
		m.sessionModel.SetSize(msg.Width, contentHeight)
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

		if m.activeTab == TabStore && m.storeSubTab == SubTabWishlist && m.wishlistModel.IsSearchFocused() {
			switch msg.String() {
			case "ctrl+c":
				return m, tea.Quit
			case "esc":
				m.wishlistModel.focusSection = 0
				return m, nil
			default:
				var cmd tea.Cmd
				m.wishlistModel, cmd = m.wishlistModel.Update(msg)
				return m, cmd
			}
		}

		// Top-level Global Keys
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
			m.activeTab = TabStore
			return m, nil
		case "2":
			m.activeTab = TabMatches
			return m, nil
		case "3":
			m.activeTab = TabStats
			return m, nil
		case "4":
			m.activeTab = TabProgress
			return m, nil
		case "5":
			m.activeTab = TabSession
			return m, nil
		}

		// Store sub-tab navigation
		if m.activeTab == TabStore {
			switch msg.String() {
			case "s":
				m.storeSubTab = SubTabShop
				return m, nil
			case "w":
				m.storeSubTab = SubTabWishlist
				m.wishlistModel.Refresh()
				return m, nil
			case "n":
				if m.nightModel.HasData() {
					m.storeSubTab = SubTabNightMarket
				}
				return m, nil
			}
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
		m.missionsMap = msg.MissionsMap
		m.ranksMap = msg.RanksMap
		m.agentsMap = msg.AgentsMap
		m.mapsMap = msg.MapsMap
		m.weaponsMap = msg.WeaponsMap

		// Sizing
		chromeHeight := 8
		contentHeight := m.height - chromeHeight
		if contentHeight < 0 {
			contentHeight = 0
		}

		// Update sub-models
		m.shopModel = NewShopModel(msg.ShopSkins, msg.TimeRemaining)
		m.shopModel.SetSize(m.width, contentHeight)

		hasNM := msg.NightMarket != nil && len(msg.NightMarket) > 0
		if hasNM {
			m.nightModel = NewNightMarketModel(msg.NightMarket, msg.NMDiscounts)
			m.nightModel.SetSize(m.width, contentHeight)
		} else {
			m.nightModel = NewNightMarketModel(nil, nil)
			if m.storeSubTab == SubTabNightMarket {
				m.storeSubTab = SubTabShop
			}
		}

		m.shopModel.SetNightMarketActive(hasNM)
		m.wishlistModel.SetNightMarketActive(hasNM)

		m.wishlistModel.SetAllSkins(msg.AllSkins)
		m.wishlistModel.Refresh()
		m.wishlistModel.SetSize(m.width, contentHeight)

		m.matchesModel = NewMatchesModel(msg.MatchItems, m.session.PUUID, msg.AgentsMap, msg.MapsMap)
		m.matchesModel.SetSize(m.width, contentHeight)

		m.statsModel = NewStatsModel(msg.MatchDetails, m.session.PUUID, msg.CompUpdates, msg.AgentsMap, msg.WeaponsMap, msg.RankName)
		m.statsModel.SetSize(m.width, contentHeight)

		m.progressModel = NewProgressModel(msg.Battlepass, msg.Missions, msg.MissionsMap)
		m.progressModel.SetSize(m.width, contentHeight)
		m.bpModel = m.progressModel

		var matchIDs []string
		for _, it := range msg.MatchItems {
			matchIDs = append(matchIDs, it.MatchID)
		}
		m.sessionModel.SetInitialSnapshot(matchIDs)
		m.sessionModel.UpdateSessionMatches(msg.MatchDetails)
		m.sessionModel.SetSize(m.width, contentHeight)
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

	// Route keys to active sub-model
	var cmd tea.Cmd
	switch m.activeTab {
	case TabStore:
		if m.storeSubTab == SubTabNightMarket {
			m.nightModel, cmd = m.nightModel.Update(msg)
		} else if m.storeSubTab == SubTabWishlist {
			m.wishlistModel, cmd = m.wishlistModel.Update(msg)
		} else {
			m.shopModel, cmd = m.shopModel.Update(msg)
		}
	case TabMatches:
		m.matchesModel, cmd = m.matchesModel.Update(msg)
	case TabStats:
		m.statsModel, cmd = m.statsModel.Update(msg)
	case TabProgress:
		m.progressModel, cmd = m.progressModel.Update(msg)
		m.bpModel = m.progressModel
	case TabSession:
		m.sessionModel, cmd = m.sessionModel.Update(msg)
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

	// ── Header Bar ──────────────────────────────────────────────
	var headerParts []string
	headerParts = append(headerParts, TitleStyle.Render("⚡ val0"))
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

	// ── Top-level Tabs ──────────────────────────────────────────
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

	// ── Store Sub-navigation (if Store tab is active) ───────────
	if m.activeTab == TabStore {
		subTabs := []string{}
		if m.storeSubTab == SubTabShop {
			subTabs = append(subTabs, ActiveTabStyle.Render("[s] Shop"))
		} else {
			subTabs = append(subTabs, InactiveTabStyle.Render("[s] Shop"))
		}

		if m.storeSubTab == SubTabWishlist {
			subTabs = append(subTabs, ActiveTabStyle.Render("[w] Wishlist"))
		} else {
			subTabs = append(subTabs, InactiveTabStyle.Render("[w] Wishlist"))
		}

		if m.nightModel.HasData() {
			if m.storeSubTab == SubTabNightMarket {
				subTabs = append(subTabs, ActiveTabStyle.Render("[n] Night Market"))
			} else {
				subTabs = append(subTabs, InactiveTabStyle.Render("[n] Night Market"))
			}
		}
		sb.WriteString("  " + lipgloss.JoinHorizontal(lipgloss.Top, subTabs...) + "\n\n")
	}

	// ── Active Panel ────────────────────────────────────────────
	switch m.activeTab {
	case TabStore:
		if m.storeSubTab == SubTabNightMarket {
			sb.WriteString(m.nightModel.View())
		} else if m.storeSubTab == SubTabWishlist {
			sb.WriteString(m.wishlistModel.View())
		} else {
			sb.WriteString(m.shopModel.View())
		}
	case TabMatches:
		sb.WriteString(m.matchesModel.View())
	case TabStats:
		sb.WriteString(m.statsModel.View())
	case TabProgress:
		sb.WriteString(m.progressModel.View())
	case TabSession:
		sb.WriteString(m.sessionModel.View())
	}

	// ── Status Bar ──────────────────────────────────────────────
	sb.WriteString("\n\n")
	sb.WriteString("  " + RenderKeyLegends(
		[2]string{"1-5", "switch tabs"},
		[2]string{"←/→", "prev/next"},
		[2]string{"r", "refresh"},
		[2]string{"q", "quit"},
	))

	content := sb.String()
	if m.width > 0 && m.height > 0 {
		lines := strings.Split(content, "\n")
		if len(lines) > m.height {
			lines = lines[:m.height]
			content = strings.Join(lines, "\n")
		}
		rendered := AppStyle.Render(content)
		return lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top,
			rendered,
			lipgloss.WithWhitespaceChars(" "),
			lipgloss.WithWhitespaceBackground(ColorBg),
		)
	}
	return AppStyle.Render(content)
}

func resolveSkinImages(asset models.SkinAsset) (iconURL string, fullRenderURL string) {
	if len(asset.Chromas) > 0 {
		if asset.Chromas[0].FullRender != nil && *asset.Chromas[0].FullRender != "" {
			fullRenderURL = *asset.Chromas[0].FullRender
		}
		if asset.Chromas[0].DisplayIcon != nil && *asset.Chromas[0].DisplayIcon != "" {
			iconURL = *asset.Chromas[0].DisplayIcon
		}
	}
	if fullRenderURL == "" && asset.DisplayIcon != nil {
		fullRenderURL = *asset.DisplayIcon
	}
	if iconURL == "" && asset.DisplayIcon != nil {
		iconURL = *asset.DisplayIcon
	}
	if iconURL == "" && len(asset.Levels) > 0 && asset.Levels[0].DisplayIcon != nil {
		iconURL = *asset.Levels[0].DisplayIcon
	}
	if fullRenderURL == "" {
		fullRenderURL = iconURL
	}
	return iconURL, fullRenderURL
}

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
	return m.nightModel.HasData()
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
		Render("⟳  Loading your Valorant tracker data...")

	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return "\n\n  " + content + "\n"
}

func (m MainModel) renderError() string {
	boxWidth := m.width - 8
	if boxWidth > 64 {
		boxWidth = 64
	}
	if boxWidth < 28 {
		boxWidth = 28
	}
	errLegend := RenderKeyLegends([2]string{"r", "retry"}, [2]string{"q", "quit"})
	content := ErrorStyle.Copy().Width(boxWidth).Render(fmt.Sprintf("✕  %s\n\n%s", m.err.Error(), errLegend))
	if m.width > 0 && m.height > 0 {
		return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
	}
	return "\n\n  " + content + "\n"
}

// loadData is the tea.Cmd that fetches all remote data and returns DataLoadedMsg.
func (m MainModel) loadData() tea.Msg {
	client := api.NewClient(m.session)

	var (
		sf          *models.StorefrontResponse
		sfErr       error
		skins       []models.SkinAsset
		skinsErr    error
		wallet      *models.WalletResponse
		mmr         *models.MMRResponse
		contracts   *models.ContractsResponse
		content     *models.ContentResponse
		matchHist   *models.MatchHistoryResponse
		compUpdates *models.CompetitiveUpdatesResponse
		ranksMap    map[int]string
		agentsMap   map[string]string
		mapsMap     map[string]string
		weaponsMap  map[string]string
		missionsMap map[string]cache.MissionInfo
		wg          sync.WaitGroup
	)

	wg.Add(7)

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

	// 6. Fetch Match History
	go func() {
		defer wg.Done()
		matchHist, _ = client.FetchMatchHistory(0, 10, "")
	}()

	// 7. Fetch Competitive Updates
	go func() {
		defer wg.Done()
		compUpdates, _ = client.FetchCompetitiveUpdates(0, 20)
	}()

	// Load static caches
	ranksMap, _ = cache.LoadOrFetchRanks(m.session.ClientVersion)
	agentsMap, _ = cache.LoadOrFetchAgents(m.session.ClientVersion)
	mapsMap, _ = cache.LoadOrFetchMaps(m.session.ClientVersion)
	weaponsMap, _ = cache.LoadOrFetchWeapons(m.session.ClientVersion)
	missionsMap, _ = cache.LoadOrFetchMissions(m.session.ClientVersion)

	wg.Wait()

	if sfErr != nil {
		return DataLoadedMsg{Err: fmt.Errorf("storefront: %w", sfErr)}
	}
	if skinsErr != nil {
		return DataLoadedMsg{Err: fmt.Errorf("skin cache: %w", skinsErr)}
	}

	lookup := cache.BuildSkinLookup(skins)
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

		iconURL, fullRenderURL := resolveSkinImages(asset)

		shopSkins = append(shopSkins, models.ResolvedSkin{
			UUID:          asset.UUID,
			DisplayName:   asset.DisplayName,
			Rarity:        RarityNameMap[tierUUID],
			CostVP:        offer.Cost[vpUUID],
			IconURL:       iconURL,
			FullRenderURL: fullRenderURL,
			Sprite:        "", // Sprites are rendered dynamically by ShopModel.View() using current terminal width
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

			iconURL, fullRenderURL := resolveSkinImages(asset)

			cost := bo.Offer.Cost[vpUUID]
			if len(bo.DiscountCosts) > 0 && bo.DiscountCosts[vpUUID] > 0 {
				cost = bo.DiscountCosts[vpUUID]
			} else if len(bo.DiscountedCost) > 0 && bo.DiscountedCost[vpUUID] > 0 {
				cost = bo.DiscountedCost[vpUUID]
			} else if bo.DiscountPercent > 0 && cost > 0 {
				cost = cost - (cost * bo.DiscountPercent / 100)
			}

			nightSkins = append(nightSkins, models.ResolvedSkin{
				UUID:          asset.UUID,
				DisplayName:   asset.DisplayName,
				Rarity:        RarityNameMap[tierUUID],
				CostVP:        cost,
				IconURL:       iconURL,
				FullRenderURL: fullRenderURL,
				Sprite:        "", // Sprites are rendered dynamically
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

	// Fetch match details for recent matches (up to 10 in parallel)
	var matchDetails []*models.MatchDetails
	var matchItems []MatchItem

	if matchHist != nil && len(matchHist.History) > 0 {
		limit := min(len(matchHist.History), 10)
		detailChan := make(chan *models.MatchDetails, limit)
		var detailWg sync.WaitGroup

		for i := 0; i < limit; i++ {
			matchSummary := matchHist.History[i]
			detailWg.Add(1)
			go func(mID string) {
				defer detailWg.Done()
				d, dErr := client.FetchMatchDetails(mID)
				if dErr == nil && d != nil {
					detailChan <- d
				}
			}(matchSummary.MatchID)
		}

		detailWg.Wait()
		close(detailChan)

		detailsMap := make(map[string]*models.MatchDetails)
		for d := range detailChan {
			detailsMap[d.MatchInfo.MatchID] = d
			matchDetails = append(matchDetails, d)
		}

		// Collect all unique PUUIDs from all match details
		puuidSet := make(map[string]bool)
		for _, d := range matchDetails {
			for _, p := range d.Players {
				puuidSet[p.Subject] = true
			}
		}
		var allPUUIDs []string
		for puuid := range puuidSet {
			allPUUIDs = append(allPUUIDs, puuid)
		}

		// Resolve names
		playerNames, _ := client.FetchPlayerNames(allPUUIDs)

		// Inject resolved names into MatchDetails players
		if playerNames != nil {
			for _, d := range matchDetails {
				for i := range d.Players {
					p := &d.Players[i]
					if p.GameName == "" {
						if resolvedName, ok := playerNames[p.Subject]; ok {
							// resolvedName is "GameName#TagLine"
							parts := strings.SplitN(resolvedName, "#", 2)
							p.GameName = parts[0]
							if len(parts) > 1 {
								p.TagLine = parts[1]
							}
						}
					}
				}
			}
		}

		// Build RR update map by matchID
		rrMap := make(map[string]int)
		if compUpdates != nil {
			for _, cu := range compUpdates.Matches {
				rrMap[cu.MatchID] = cu.RankedRatingEarned
			}
		}

		// Assemble MatchItems preserving history order
		for _, ms := range matchHist.History {
			d, hasDetail := detailsMap[ms.MatchID]
			item := MatchItem{
				MatchID:     ms.MatchID,
				QueueName:   "Unrated",
				Outcome:     "DRAW",
				Score:       "0-0",
				GameTime:    time.UnixMilli(ms.GameStartTime),
				PlayerPUUID: m.session.PUUID,
			}

			if rr, ok := rrMap[ms.MatchID]; ok {
				item.RREarned = rr
				item.HasRR = true
			}

			if hasDetail && d != nil {
				item.Details = d
				item.MapName = cache.GetMapName(d.MatchInfo.MapID, mapsMap)
				item.QueueName = GetQueueDisplayName(d.MatchInfo.QueueID)
				item.Outcome = d.GetMatchOutcome(m.session.PUUID)
				item.Score = d.ScoreString(m.session.PUUID)

				if p := d.GetPlayer(m.session.PUUID); p != nil {
					item.AgentName = cache.GetAgentName(p.CharacterID, agentsMap)
					item.Kills = p.Stats.Kills
					item.Deaths = p.Stats.Deaths
					item.Assists = p.Stats.Assists
				}
			} else {
				item.MapName = "Valorant Match"
			}

			matchItems = append(matchItems, item)
		}
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
		MissionsMap:   missionsMap,
		MatchItems:    matchItems,
		MatchDetails:  matchDetails,
		CompUpdates:   compUpdates,
		AgentsMap:     agentsMap,
		MapsMap:       mapsMap,
		WeaponsMap:    weaponsMap,
		RanksMap:      ranksMap,
		AllSkins:      skins,
	}
}
