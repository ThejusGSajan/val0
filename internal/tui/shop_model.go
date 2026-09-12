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

// SpriteOverlay holds a native graphics payload and metadata for grid-level injection.
type SpriteOverlay struct {
	Payload      string // The raw Sixel/Kitty/iTerm2 payload (without cursor jumps)
	ContentWidth int    // Width of the sprite content area (columns)
	SpriteRows   int    // Number of terminal rows the sprite occupies
	OffsetX      int    // Horizontal shift in terminal columns (e.g. -2 for right column)
}

// buildWipeSeq creates an in-band cell wipe sequence with absolute column positioning.
func buildWipeSeq(colOffset, contentWidth, spriteRows int) string {
	bgSpaces := fmt.Sprintf("\x1b[48;2;15;17;23m%s\x1b[0m", strings.Repeat(" ", contentWidth))
	var sb strings.Builder
	for i := 0; i < spriteRows; i++ {
		sb.WriteString(fmt.Sprintf("\x1b[%dG", colOffset+1)) // Absolute column (1-indexed)
		sb.WriteString(bgSpaces)
		if i < spriteRows-1 {
			sb.WriteString("\x1b[1B") // Move down 1 line
		}
	}
	// Move back up to sprite row 0
	if spriteRows > 1 {
		sb.WriteString(fmt.Sprintf("\x1b[%dA", spriteRows-1))
		sb.WriteString(fmt.Sprintf("\x1b[%dG", colOffset+1)) // Reset column after final move-up
	}
	return sb.String()
}

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

	// Dynamic sizing calculation - cardContentWidth is exact inner content width
	cardContentWidth := 40
	if m.width > 0 {
		cardContentWidth = (m.width / 2) - 8
	}
	if cardContentWidth > 40 {
		cardContentWidth = 40
	}
	if cardContentWidth < 20 {
		cardContentWidth = 20
	}

	type cardWithOverlay struct {
		rendered string
		overlay  *SpriteOverlay
		colIndex int
	}

	var cardsWithOverlay []cardWithOverlay
	for i, skin := range m.skins {
		inWishlist := cache.IsInWishlist(skin.UUID)
		offsetX := 0
		if i%2 == 1 {
			offsetX = -2
		}
		rendered, overlay := renderSkinCardWithWishlist(skin, -1, cardContentWidth, inWishlist, offsetX)
		cardsWithOverlay = append(cardsWithOverlay, cardWithOverlay{
			rendered: rendered,
			overlay:  overlay,
			colIndex: i % 2,
		})
	}

	// Layout: 2 × 2 grid if we have 4 skins (the standard daily shop)
	var rows []string
	for i := 0; i < len(cardsWithOverlay); i += 2 {
		leftCard := cardsWithOverlay[i]
		var rightCard *cardWithOverlay
		if i+1 < len(cardsWithOverlay) {
			rc := cardsWithOverlay[i+1]
			rightCard = &rc
		}

		var gridRow string
		if rightCard != nil {
			gridRow = lipgloss.JoinHorizontal(lipgloss.Top, leftCard.rendered, rightCard.rendered)
		} else {
			gridRow = leftCard.rendered
		}

		gridRowLines := strings.Split(gridRow, "\n")
		totalLines := len(gridRowLines)

		// The card structure: top border (1) + wishlist header (1) + sprite rows (5) + name (1) + rarity (1) + price (1) + bottom border (1) = 11 lines
		// Sprite row 0 starts at line index 2.
		// linesUp from last line (totalLines - 1) to sprite row 0 (line 2) = totalLines - 3
		linesUp := totalLines - 3

		// Inject left card overlay
		if leftCard.overlay != nil {
			spriteMargin := (cardContentWidth - leftCard.overlay.ContentWidth) / 2
			colOffset := 2 + spriteMargin + leftCard.overlay.OffsetX // left border(1) + left pad(1) + margin + offset
			wipe := buildWipeSeq(colOffset, leftCard.overlay.ContentWidth, leftCard.overlay.SpriteRows)
			payload := "\x1b7" +
				fmt.Sprintf("\x1b[%dA", linesUp) +
				fmt.Sprintf("\x1b[%dG", colOffset+1) +
				wipe + leftCard.overlay.Payload + "\x1b8"
			placeholder := sprite.RegisterPayload(payload)
			gridRowLines[totalLines-1] += placeholder
		}

		// Inject right card overlay
		if rightCard != nil && rightCard.overlay != nil {
			// Right card starts after left card total width (cardContentWidth + 5)
			rightCardStart := cardContentWidth + 5 // 1 border + 1 pad + ccw + 1 pad + 1 border + 1 margin
			spriteMarginR := (cardContentWidth - rightCard.overlay.ContentWidth) / 2
			colOffset := rightCardStart + 2 + spriteMarginR + rightCard.overlay.OffsetX // + right card's left border(1) + left pad(1) + margin + offset
			wipe := buildWipeSeq(colOffset, rightCard.overlay.ContentWidth, rightCard.overlay.SpriteRows)
			payload := "\x1b7" +
				fmt.Sprintf("\x1b[%dA", linesUp) +
				fmt.Sprintf("\x1b[%dG", colOffset+1) +
				wipe + rightCard.overlay.Payload + "\x1b8"
			placeholder := sprite.RegisterPayload(payload)
			gridRowLines[totalLines-1] += placeholder
		}

		gridRow = strings.Join(gridRowLines, "\n")
		rows = append(rows, gridRow)
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
func renderSkinCard(skin models.ResolvedSkin, discountPct int, cardContentWidth int) (string, *SpriteOverlay) {
	return renderSkinCardWithWishlist(skin, discountPct, cardContentWidth, false, 0)
}

func renderSkinCardWithWishlist(skin models.ResolvedSkin, discountPct int, cardContentWidth int, inWishlist bool, offsetX int) (string, *SpriteOverlay) {
	if cardContentWidth <= 0 {
		cardContentWidth = 40
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
	const spriteTargetRows = 5 // fixed height for all sprite containers
	spriteRenderWidth := cardContentWidth - 4
	if spriteRenderWidth < 16 {
		spriteRenderWidth = 16
	}
	if renderTargetURL != "" && cardContentWidth >= 20 {
		spr = sprite.Render(renderTargetURL, spriteRenderWidth, spriteTargetRows)
	} else if skin.Sprite != "" {
		spr = skin.Sprite
	}

	var overlay *SpriteOverlay
	isNative := strings.Contains(spr, "\x1bP") || strings.Contains(spr, "\x1b_G") || strings.Contains(spr, "\x1b]1337")
	if isNative {
		// Post-Border Overlay: Output 5 clean background-styled lines to Lipgloss
		emptySpriteLine := lipgloss.NewStyle().Background(ColorBg).Render(strings.Repeat(" ", cardContentWidth))
		for i := 0; i < spriteTargetRows; i++ {
			content.WriteString(emptySpriteLine + "\n")
		}
		if spr != "" {
			overlay = &SpriteOverlay{
				Payload:      spr,
				ContentWidth: spriteRenderWidth,
				SpriteRows:   spriteTargetRows,
				OffsetX:      offsetX,
			}
		}
	} else {
		// Half-block fallback
		spr = padSpriteToHeightWithOffset(spr, spriteTargetRows, cardContentWidth, string(ColorBg), spriteRenderWidth, offsetX)
		content.WriteString(spr + "\n")
	}

	// Line 6: Skin name
	content.WriteString(nameStyle.Render(truncate(skin.DisplayName, cardContentWidth)))
	content.WriteString("\n")

	// Line 7: Rarity tag
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

	// Render Card: Width(cardContentWidth) ensures zero trailing space padding inside the card
	cardBox := CardStyle.
		BorderForeground(borderCol).
		Width(cardContentWidth).
		Render(content.String())

	return cardBox, overlay
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

	// Half-block only (native graphics are handled at grid level)
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

// padSpriteToHeightWithOffset pads sprite rows vertically and horizontally with a column offset.
// Total line width is guaranteed to equal width, preserving Lipgloss border invariants.
func padSpriteToHeightWithOffset(spr string, targetRows int, width int, bgHex string, spriteWidth int, offsetX int) string {
	bgStyle := lipgloss.NewStyle().Background(lipgloss.Color(bgHex))
	emptyLine := bgStyle.Render(strings.Repeat(" ", width))

	if spr == "" {
		lines := make([]string, targetRows)
		for i := range lines {
			lines[i] = emptyLine
		}
		return strings.Join(lines, "\n")
	}

	// Half-block only (native graphics are handled at grid level)
	lines := strings.Split(strings.TrimRight(spr, "\n"), "\n")
	if len(lines) > targetRows {
		lines = lines[:targetRows]
	}

	leftPad := ((width - spriteWidth) / 2) + offsetX
	if leftPad < 0 {
		leftPad = 0
	}
	if leftPad > width-spriteWidth {
		leftPad = width - spriteWidth
	}
	rightPad := width - spriteWidth - leftPad
	if rightPad < 0 {
		rightPad = 0
	}

	leftPadStr := bgStyle.Render(strings.Repeat(" ", leftPad))
	rightPadStr := bgStyle.Render(strings.Repeat(" ", rightPad))

	var paddedLines []string
	for _, l := range lines {
		paddedLines = append(paddedLines, leftPadStr+l+rightPadStr)
	}

	// Center vertically
	topPad := (targetRows - len(paddedLines)) / 2
	bottomPad := targetRows - len(paddedLines) - topPad
	var result []string
	for i := 0; i < topPad; i++ {
		result = append(result, emptyLine)
	}
	result = append(result, paddedLines...)
	for i := 0; i < bottomPad; i++ {
		result = append(result, emptyLine)
	}
	return strings.Join(result, "\n")
}
