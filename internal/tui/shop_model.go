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

type ShopModel struct {
	skins          []models.ResolvedSkin
	timeRemaining  int // seconds
	hasNightMarket bool
	width          int
	height         int
}

func NewShopModel(skins []models.ResolvedSkin, timeRemaining int) ShopModel {
	return ShopModel{skins: skins, timeRemaining: timeRemaining}
}

func (m *ShopModel) SetNightMarketActive(active bool) {
	m.hasNightMarket = active
}

func (m *ShopModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ShopModel) Update(msg tea.Msg) (ShopModel, tea.Cmd) {
	return m, nil
}

func (m ShopModel) View() string {
	if m.width == 0 {
		return ""
	}
	if len(m.skins) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No shop data available.")
	}

	var sb strings.Builder

	// Check if any shop skin is in wishlist
	var wishlistMatches []string
	for _, skin := range m.skins {
		if cache.IsInWishlist(skin.UUID) {
			wishlistMatches = append(wishlistMatches, skin.DisplayName)
		}
	}

	if len(wishlistMatches) > 0 {
		banner := lipgloss.NewStyle().
			Foreground(ColorUltra).
			Bold(true).
			Background(lipgloss.Color("#2A2410")).
			Padding(0, 2).
			Render(fmt.Sprintf("⭐ WISHLIST MATCH! %s is in your shop today!", strings.Join(wishlistMatches, ", ")))
		sb.WriteString("  " + banner + "\n")
	}

	// Timer
	hours := m.timeRemaining / 3600
	minutes := (m.timeRemaining % 3600) / 60
	timer := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  ⏱  Resets in %dh %dm", hours, minutes))
	sb.WriteString(timer + "\n")

	// Dynamic sizing calculation
	cardWidth := 44
	if m.width > 0 {
		cardWidth = int(float64(m.width/2) * 0.6)
	}
	if cardWidth > 45 {
		cardWidth = 45
	}
	if cardWidth < 25 {
		cardWidth = 25
	}
	spriteW := cardWidth - 4

	// Render each skin as a card
	var cards []string
	for _, skin := range m.skins {
		inWishlist := cache.IsInWishlist(skin.UUID)
		cards = append(cards, renderSkinCardWithWishlist(skin, -1, cardWidth, spriteW, inWishlist))
	}

	// Layout: 2 × 2 grid if we have 4 skins (the standard daily shop)
	var rows []string
	for i := 0; i < len(cards); i += 2 {
		if i+1 < len(cards) {
			row := lipgloss.JoinHorizontal(lipgloss.Top, cards[i], cards[i+1])
			rows = append(rows, row)
		} else {
			rows = append(rows, cards[i])
		}
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	sb.WriteString(grid + "\n")

	var storePairs [][2]string
	storePairs = append(storePairs, [2]string{"w", "enter wishlist tab"})
	if m.hasNightMarket {
		storePairs = append(storePairs, [2]string{"n", "enter nightmarket tab"})
	}
	sb.WriteString("  " + RenderKeyLegends(storePairs...))

	return sb.String()
}

// renderSkinCard creates a single skin display card with sprite + name + price.
func renderSkinCard(skin models.ResolvedSkin, discountPct int, cardWidth int, spriteWidth int) string {
	return renderSkinCardWithWishlist(skin, discountPct, cardWidth, spriteWidth, false)
}

func renderSkinCardWithWishlist(skin models.ResolvedSkin, discountPct int, cardWidth int, spriteWidth int, inWishlist bool) string {
	// Get rarity color
	rarityColor := ColorMuted
	for uuid, name := range RarityNameMap {
		if name == skin.Rarity {
			rarityColor = RarityColorMap[uuid]
			break
		}
	}

	nameStyle := lipgloss.NewStyle().Foreground(rarityColor).Bold(true)
	rarityTagStyle := lipgloss.NewStyle().
		Foreground(rarityColor).
		Italic(true)

	// Build card content
	var content strings.Builder

	// Wishlist indicator tag
	if inWishlist {
		content.WriteString(lipgloss.NewStyle().Foreground(ColorUltra).Bold(true).Render("⭐ WISHLIST ITEM") + "\n")
	}

	// ANSI sprite or Native Graphic
	spr := ""
	renderTargetURL := skin.FullRenderURL
	if renderTargetURL == "" {
		renderTargetURL = skin.IconURL
	}
	if renderTargetURL != "" && spriteWidth >= 20 {
		spr = sprite.Render(renderTargetURL, spriteWidth)
	} else if skin.Sprite != "" {
		spr = skin.Sprite
	}

	const spriteTargetRows = 4 // fixed height for all sprite containers
	spr = padSpriteToHeight(spr, spriteTargetRows, spriteWidth)

	if spr != "" {
		content.WriteString(spr)
		content.WriteString("\n")
	}

	// Skin name
	content.WriteString(nameStyle.Render(skin.DisplayName))
	content.WriteString("\n")

	// Rarity tag
	if skin.Rarity != "" {
		content.WriteString(rarityTagStyle.Render("● " + skin.Rarity + " Edition"))
		content.WriteString("\n")
	}

	// Price
	priceStr := fmt.Sprintf("VP %d", skin.CostVP)
	if discountPct > 0 {
		priceStr = fmt.Sprintf("VP %d  (-%d%%)", skin.CostVP, discountPct)
	}
	content.WriteString(VPBadgeStyle.Render(priceStr))

	if cardWidth <= 0 {
		cardWidth = 44
	}

	borderCol := rarityColor
	if inWishlist {
		borderCol = ColorUltra // Gold border for wishlist matches
	}

	return CardStyle.
		BorderForeground(borderCol).
		Width(cardWidth).
		Render(content.String())
}

// padSpriteToHeight ensures the sprite string occupies exactly `targetRows` lines.
// If the sprite has fewer lines, blank lines are prepended (top-padding) to vertically center it.
// If the sprite has more lines, it is truncated from the bottom.
func padSpriteToHeight(spr string, targetRows int, width int) string {
	if spr == "" {
		// Return empty box of targetRows × width spaces
		emptyLine := strings.Repeat(" ", width)
		lines := make([]string, targetRows)
		for i := range lines {
			lines[i] = emptyLine
		}
		return strings.Join(lines, "\n")
	}

	// Protect Sixel/Kitty/OSC sequences from strings.Split
	if strings.Contains(spr, "\x1bP") || strings.Contains(spr, "\x1b_G") || strings.Contains(spr, "\x1b]1337") {
		// Native graphics physically move the cursor down 4 rows (80px),
		// but Lip Gloss needs placeholder characters to measure string bounds correctly.
		emptyLine := strings.Repeat(" ", width)
		lines := make([]string, targetRows)
		lines[0] = spr + emptyLine
		for i := 1; i < targetRows; i++ {
			lines[i] = emptyLine
		}
		// Wrap the sequence so Lip Gloss allocates a (width x 4) text block over the image area.
		return strings.Join(lines, "\n")
	}

	lines := strings.Split(strings.TrimRight(spr, "\n"), "\n")
	if len(lines) > targetRows {
		lines = lines[:targetRows]
	}
	// Center vertically
	topPad := (targetRows - len(lines)) / 2
	bottomPad := targetRows - len(lines) - topPad
	emptyLine := strings.Repeat(" ", width)
	var result []string
	for i := 0; i < topPad; i++ {
		result = append(result, emptyLine)
	}
	result = append(result, lines...)
	for i := 0; i < bottomPad; i++ {
		result = append(result, emptyLine)
	}
	return strings.Join(result, "\n")
}
