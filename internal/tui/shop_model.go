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
	if cardWidth <= 0 {
		cardWidth = 44
	}
	cardContentWidth := cardWidth - 4
	if cardContentWidth < 10 {
		cardContentWidth = 10
	}

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

	// Build card content - STRICT 8-LINE INVARIANT
	var content strings.Builder

	// Line 0: Wishlist header slot (always 1 line)
	if inWishlist {
		content.WriteString(lipgloss.NewStyle().Foreground(ColorUltra).Bold(true).Render(truncate("⭐ WISHLIST ITEM", cardContentWidth)) + "\n")
	} else {
		content.WriteString(lipgloss.NewStyle().Background(ColorBg).Render(strings.Repeat(" ", cardContentWidth)) + "\n")
	}

	// ANSI sprite or Native Graphic
	spr := ""
	renderTargetURL := skin.FullRenderURL
	if renderTargetURL == "" {
		renderTargetURL = skin.IconURL
	}
	const spriteTargetRows = 4 // fixed height for all sprite containers
	if renderTargetURL != "" && cardContentWidth >= 20 {
		spr = sprite.Render(renderTargetURL, cardContentWidth, spriteTargetRows)
	} else if skin.Sprite != "" {
		spr = skin.Sprite
	}

	isNative := strings.Contains(spr, "\x1bP") || strings.Contains(spr, "\x1b_G") || strings.Contains(spr, "\x1b]1337")
	if isNative {
		// Post-Border Overlay: Output 4 clean background-styled lines to Lipgloss
		emptySpriteLine := lipgloss.NewStyle().Background(ColorBg).Render(strings.Repeat(" ", cardContentWidth))
		for i := 0; i < spriteTargetRows; i++ {
			content.WriteString(emptySpriteLine + "\n")
		}
	} else {
		// Half-block fallback
		spr = padSpriteToHeight(spr, spriteTargetRows, cardContentWidth, string(ColorBg))
		content.WriteString(spr + "\n")
	}

	// Line 5: Skin name
	content.WriteString(nameStyle.Render(truncate(skin.DisplayName, cardContentWidth)))
	content.WriteString("\n")

	// Line 6: Rarity tag
	if skin.Rarity != "" {
		content.WriteString(rarityTagStyle.Render(truncate("● "+skin.Rarity+" Edition", cardContentWidth)))
	} else {
		content.WriteString(lipgloss.NewStyle().Background(ColorBg).Render(strings.Repeat(" ", cardContentWidth)))
	}
	content.WriteString("\n")

	// Line 7: Price
	priceStr := fmt.Sprintf("VP %d", skin.CostVP)
	if discountPct > 0 {
		priceStr = fmt.Sprintf("VP %d  (-%d%%)", skin.CostVP, discountPct)
	}
	content.WriteString(VPBadgeStyle.Render(priceStr))

	borderCol := rarityColor
	if inWishlist {
		borderCol = ColorUltra // Gold border for wishlist matches
	}

	cardBox := CardStyle.
		BorderForeground(borderCol).
		Width(cardWidth).
		Render(content.String())

	// Post-Border Overlay Injection for Native Graphics
	if isNative && spr != "" {
		cardLines := strings.Split(cardBox, "\n")
		// In a 10-line card box:
		// Line 0: Top Border
		// Line 1: Header
		// Line 2..5: Sprite rows (Line 5 is the final sprite row)
		if len(cardLines) >= 6 {
			cursorLeft := fmt.Sprintf("\x1b[%dD", cardWidth-2)
			cursorUp := "\x1b[3A"
			cardLines[5] = cardLines[5] + "\x1b7" + cursorLeft + cursorUp + spr + "\x1b8"
			cardBox = strings.Join(cardLines, "\n")
		}
	}

	return cardBox
}

// padSpriteToHeight ensures the sprite string occupies exactly `targetRows` lines.
// If the sprite has fewer lines, blank lines are prepended (top-padding) to vertically center it.
// If the sprite has more lines, it is truncated from the bottom.
func padSpriteToHeight(spr string, targetRows int, width int, bgHex string) string {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgHex))
	emptyLine := bgStyle.Render(strings.Repeat(" ", width))

	if spr == "" {
		lines := make([]string, targetRows)
		for i := range lines {
			lines[i] = emptyLine
		}
		return strings.Join(lines, "\n")
	}

	// [FIXED]: Sixel Post-Draw Architecture
	// Protect Native Graphics from being overwritten by Lip Gloss background spaces.
	if strings.Contains(spr, "\x1bP") || strings.Contains(spr, "\x1b_G") || strings.Contains(spr, "\x1b]1337") {
		lines := make([]string, targetRows)

		// 1. Give Lip Gloss pure spaces for the first (targetRows - 1) lines.
		// The terminal will print these spaces and paint the background color normally.
		for i := 0; i < targetRows-1; i++ {
			lines[i] = emptyLine
		}

		// 2. On the final line, we print the Lip Gloss spaces, and then physically move
		// the terminal cursor backward over the newly painted spaces, up to the top of
		// the card, and THEN execute the Sixel payload to draw on top of the spaces.

		// \x1b[%dD moves cursor LEFT by `width` columns.
		cursorLeft := fmt.Sprintf("\x1b[%dD", width)

		// \x1b[%dA moves cursor UP by (targetRows - 1) lines.
		cursorUp := fmt.Sprintf("\x1b[%dA", targetRows-1)

		// DEC Save cursor (\x1b7), jump up-left, emit payload, DEC Restore cursor (\x1b8)
		lines[targetRows-1] = emptyLine + "\x1b7" + cursorLeft + cursorUp + spr + "\x1b8"

		return strings.Join(lines, "\n")
	}

	lines := strings.Split(strings.TrimRight(spr, "\n"), "\n")
	if len(lines) > targetRows {
		lines = lines[:targetRows]
	}
	// Center vertically
	topPad := (targetRows - len(lines)) / 2
	bottomPad := targetRows - len(lines) - topPad
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
