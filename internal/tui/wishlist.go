package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/cache"
	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
)

type WishlistModel struct {
	// Wishlist section
	entries  []cache.WishlistEntry
	wlCursor int

	// Browse section
	allSkins    []models.SkinAsset // all skins from cache
	filtered    []models.SkinAsset // filtered by search
	searchInput string
	brCursor    int

	// UI state
	focusSection   int // 0 = wishlist section, 1 = browse section
	hasNightMarket bool
	width          int
	height         int
	flashMsg       string
}

func NewWishlistModel() WishlistModel {
	entries, _ := cache.LoadWishlist()
	return WishlistModel{
		entries:      entries,
		focusSection: 0,
	}
}

func (m *WishlistModel) SetNightMarketActive(active bool) {
	m.hasNightMarket = active
}

func (m *WishlistModel) SetAllSkins(skins []models.SkinAsset) {
	m.allSkins = skins
	m.filterSkins()
}

func (m *WishlistModel) Refresh() {
	entries, _ := cache.LoadWishlist()
	m.entries = entries
	if m.wlCursor >= len(m.entries) && len(m.entries) > 0 {
		m.wlCursor = len(m.entries) - 1
	}
	if m.wlCursor < 0 {
		m.wlCursor = 0
	}
	m.filterSkins()
}

func (m *WishlistModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m WishlistModel) IsSearchFocused() bool {
	return m.focusSection == 1
}

func (m *WishlistModel) filterSkins() {
	if m.searchInput == "" {
		m.filtered = m.allSkins
	} else {
		query := strings.ToLower(m.searchInput)
		m.filtered = nil
		for _, s := range m.allSkins {
			if strings.Contains(strings.ToLower(s.DisplayName), query) {
				m.filtered = append(m.filtered, s)
			}
		}
	}
	if m.brCursor >= len(m.filtered) {
		m.brCursor = max(len(m.filtered)-1, 0)
	}
	if m.brCursor < 0 {
		m.brCursor = 0
	}
}

func (m WishlistModel) Update(msg tea.Msg) (WishlistModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "shift+tab":
			m.focusSection = (m.focusSection + 1) % 2
			m.flashMsg = ""
			return m, nil
		}

		if m.focusSection == 0 {
			// Wishlist section focused
			switch msg.String() {
			case "up", "k":
				if m.wlCursor > 0 {
					m.wlCursor--
					m.flashMsg = ""
				}
			case "down", "j":
				if m.wlCursor < len(m.entries)-1 {
					m.wlCursor++
					m.flashMsg = ""
				}
			case "x", "delete":
				if len(m.entries) > 0 && m.wlCursor < len(m.entries) {
					removed := m.entries[m.wlCursor]
					_ = cache.RemoveFromWishlist(removed.UUID)
					m.Refresh()
					m.flashMsg = fmt.Sprintf("Removed %s from wishlist.", removed.Name)
				}
			}
		} else {
			// Browse section focused
			switch msg.Type {
			case tea.KeyUp:
				if m.brCursor > 0 {
					m.brCursor--
					m.flashMsg = ""
				}
			case tea.KeyDown:
				if m.brCursor < len(m.filtered)-1 {
					m.brCursor++
					m.flashMsg = ""
				}
			case tea.KeyEnter:
				if len(m.filtered) > 0 && m.brCursor < len(m.filtered) {
					selected := m.filtered[m.brCursor]
					tierUUID := ""
					if selected.ContentTierUUID != nil {
						tierUUID = *selected.ContentTierUUID
					}
					iconURL := ""
					if selected.DisplayIcon != nil {
						iconURL = *selected.DisplayIcon
					} else if len(selected.Levels) > 0 && selected.Levels[0].DisplayIcon != nil {
						iconURL = *selected.Levels[0].DisplayIcon
					}
					costVP := 0
					switch RarityNameMap[tierUUID] {
					case "Select":
						costVP = 875
					case "Deluxe":
						costVP = 1275
					case "Premium":
						costVP = 1775
					case "Exclusive":
						costVP = 2175
					case "Ultra":
						costVP = 2475
					}
					entry := cache.WishlistEntry{
						UUID:    selected.UUID,
						Name:    selected.DisplayName,
						Rarity:  RarityNameMap[tierUUID],
						CostVP:  costVP,
						IconURL: iconURL,
					}
					_ = cache.AddToWishlist(entry)
					m.Refresh()
					m.flashMsg = fmt.Sprintf("Added %s to wishlist!", selected.DisplayName)
				}
			case tea.KeyBackspace:
				if len(m.searchInput) > 0 {
					m.searchInput = m.searchInput[:len(m.searchInput)-1]
					m.filterSkins()
					m.flashMsg = ""
				}
			case tea.KeySpace:
				m.searchInput += " "
				m.filterSkins()
				m.flashMsg = ""
			case tea.KeyRunes:
				m.searchInput += string(msg.Runes)
				m.filterSkins()
				m.flashMsg = ""
			}
		}
	}
	return m, nil
}

func (m WishlistModel) View() string {
	if m.width == 0 {
		return ""
	}

	var sb strings.Builder
	lineWidth := max(m.width-4, 70)

	header := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render(fmt.Sprintf("  ⭐ SKIN WISHLIST                                         %d skins wishlisted", len(m.entries)))
	sb.WriteString(header + "\n")
	sb.WriteString("  " + strings.Repeat("─", lineWidth) + "\n")

	if m.flashMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorWin).Bold(true).Render("  ✔ "+m.flashMsg) + "\n\n")
	}

	// 1. Wishlist Section
	wlHeaderStyle := lipgloss.NewStyle().Bold(true)
	if m.focusSection == 0 {
		wlHeaderStyle = wlHeaderStyle.Foreground(ColorAccent)
	} else {
		wlHeaderStyle = wlHeaderStyle.Foreground(ColorFg)
	}
	sb.WriteString("  " + wlHeaderStyle.Render("YOUR WISHLISTED SKINS:") + "\n")

	if len(m.entries) == 0 {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorMuted).Render("    (No skins wishlisted yet. Search and add skins below!)") + "\n")
	} else {
		maxDisplay := 4
		start := 0
		if m.wlCursor >= maxDisplay {
			start = m.wlCursor - maxDisplay + 1
		}
		end := start + maxDisplay
		if end > len(m.entries) {
			end = len(m.entries)
		}
		for i := start; i < end; i++ {
			entry := m.entries[i]
			cursor := "    "
			if m.focusSection == 0 && i == m.wlCursor {
				cursor = "  ▸ "
			}

			rarityColor := ColorMuted
			for uuid, name := range RarityNameMap {
				if name == entry.Rarity {
					rarityColor = RarityColorMap[uuid]
					break
				}
			}

			nameStr := lipgloss.NewStyle().
				Foreground(ColorFg).
				Bold(true).
				Render(fmt.Sprintf("%2d. %-24s", i+1, truncate(entry.Name, 24)))

			rarityStr := lipgloss.NewStyle().
				Foreground(rarityColor).
				Render(fmt.Sprintf("● %-10s", entry.Rarity))

			priceStr := ""
			if entry.CostVP > 0 {
				priceStr = lipgloss.NewStyle().
					Foreground(ColorUltra).
					Bold(true).
					Render(fmt.Sprintf("VP %s", formatNumber(entry.CostVP)))
			}

			row := fmt.Sprintf("%s%s  %s  %s", cursor, nameStr, rarityStr, priceStr)
			if m.focusSection == 0 && i == m.wlCursor {
				row = lipgloss.NewStyle().Background(lipgloss.Color("#1F2430")).Render(row)
			}
			sb.WriteString(row + "\n")
		}
	}

	sb.WriteString("  " + strings.Repeat("─", lineWidth) + "\n\n")

	// 2. Browse Section
	brHeaderStyle := lipgloss.NewStyle().Bold(true)
	if m.focusSection == 1 {
		brHeaderStyle = brHeaderStyle.Foreground(ColorAccent)
	} else {
		brHeaderStyle = brHeaderStyle.Foreground(ColorFg)
	}
	sb.WriteString("  " + brHeaderStyle.Render("BROWSE & ADD SKINS:") + "\n")

	cursorChar := "_"
	if m.focusSection != 1 {
		cursorChar = ""
	}
	searchLine := fmt.Sprintf("  Search: %s%s", m.searchInput, cursorChar)
	sb.WriteString(lipgloss.NewStyle().Foreground(ColorUltra).Bold(true).Render(searchLine) + "\n")

	// Pane sizing
	previewWidth := m.width / 3
	if previewWidth < 20 {
		previewWidth = 20
	}
	if previewWidth > 40 {
		previewWidth = 40
	}
	listWidth := max(m.width-previewWidth-8, 35)

	// Build Browse List (Left Pane)
	var listRows []string
	listRows = append(listRows, "  "+strings.Repeat("─", min(listWidth, lineWidth)))

	if len(m.filtered) == 0 {
		if len(m.allSkins) == 0 {
			listRows = append(listRows, lipgloss.NewStyle().Foreground(ColorMuted).Render("    Loading skins..."))
		} else {
			listRows = append(listRows, lipgloss.NewStyle().Foreground(ColorMuted).Render("    No skins match your search."))
		}
	} else {
		maxBrDisplay := 6
		start := 0
		if m.brCursor >= maxBrDisplay {
			start = m.brCursor - maxBrDisplay + 1
		}
		end := start + maxBrDisplay
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for i := start; i < end; i++ {
			s := m.filtered[i]
			cursor := "    "
			if m.focusSection == 1 && i == m.brCursor {
				cursor = "  ▸ "
			}

			tierUUID := ""
			if s.ContentTierUUID != nil {
				tierUUID = *s.ContentTierUUID
			}
			rarity := RarityNameMap[tierUUID]
			rarityColor := ColorMuted
			if c, ok := RarityColorMap[tierUUID]; ok {
				rarityColor = c
			}

			costVP := 0
			switch rarity {
			case "Select":
				costVP = 875
			case "Deluxe":
				costVP = 1275
			case "Premium":
				costVP = 1775
			case "Exclusive":
				costVP = 2175
			case "Ultra":
				costVP = 2475
			}

			nameStr := lipgloss.NewStyle().
				Foreground(ColorFg).
				Bold(true).
				Render(fmt.Sprintf("%-22s", truncate(s.DisplayName, 22)))

			rarityStr := lipgloss.NewStyle().
				Foreground(rarityColor).
				Render(fmt.Sprintf("● %-9s", rarity))

			priceStr := ""
			if costVP > 0 {
				priceStr = lipgloss.NewStyle().
					Foreground(ColorUltra).
					Bold(true).
					Render(fmt.Sprintf("VP %s", formatNumber(costVP)))
			}

			row := fmt.Sprintf("%s%-24s %-12s %s", cursor, nameStr, rarityStr, priceStr)
			if m.focusSection == 1 && i == m.brCursor {
				row = lipgloss.NewStyle().Background(lipgloss.Color("#1F2430")).Render(row)
			}
			listRows = append(listRows, row)
		}

		// [FIXED]: STATIC ANCHORING WITH TRUECOLOR
		// Pad the list to maxBrDisplay rows, explicitly painting the background color
		// so it doesn't default to the terminal's pure black.
		actualItems := end - start
		for i := actualItems; i < maxBrDisplay; i++ {
			styledEmptyLine := lipgloss.NewStyle().Background(lipgloss.Color("#0F1117")).Render(strings.Repeat(" ", listWidth))
			listRows = append(listRows, styledEmptyLine)
		}
	}

	leftPane := strings.Join(listRows, "\n")

	// Build Sprite Preview (Right Pane)
	previewSpr := ""
	const previewTargetRows = 6
	if len(m.filtered) > 0 && m.brCursor < len(m.filtered) {
		selected := m.filtered[m.brCursor]
		iconURL, fullRenderURL := resolveSkinImages(selected)
		targetURL := fullRenderURL
		if targetURL == "" {
			targetURL = iconURL
		}
		if targetURL != "" && previewWidth >= 20 {
			previewSpr = sprite.Render(targetURL, previewWidth, previewTargetRows)
		}
	}
	previewSpr = padSpriteToHeight(previewSpr, previewTargetRows, previewWidth, string(ColorBg))
	previewBox := CardStyle.
		BorderForeground(ColorBorder).
		Width(previewWidth + 2).
		Render(previewSpr)

	splitView := lipgloss.JoinHorizontal(lipgloss.Top, leftPane, "    ", previewBox)
	sb.WriteString(splitView + "\n\n")

	var wlPairs [][2]string
	wlPairs = append(wlPairs, [2]string{"s", "enter shop tab"})
	if m.hasNightMarket {
		wlPairs = append(wlPairs, [2]string{"n", "enter nightmarket tab"})
	}
	wlPairs = append(wlPairs,
		[2]string{"tab", "switch section"},
		[2]string{"↑/↓", "navigate"},
		[2]string{"", "Type to search"},
		[2]string{"enter", "add to wishlist"},
		[2]string{"x", "remove"},
	)
	sb.WriteString("  " + RenderKeyLegends(wlPairs...))

	return sb.String()
}
