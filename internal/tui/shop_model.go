package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
)

type ShopModel struct {
	skins         []models.ResolvedSkin
	timeRemaining int // seconds
	width         int
	height        int
}

func NewShopModel(skins []models.ResolvedSkin, timeRemaining int) ShopModel {
	return ShopModel{skins: skins, timeRemaining: timeRemaining}
}

func (m *ShopModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ShopModel) Update(msg tea.Msg) (ShopModel, tea.Cmd) {
	return m, nil
}

func (m ShopModel) View() string {
	if len(m.skins) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  No shop data available.")
	}

	// Timer
	hours := m.timeRemaining / 3600
	minutes := (m.timeRemaining % 3600) / 60
	timer := lipgloss.NewStyle().
		Foreground(ColorAccent).
		Bold(true).
		Render(fmt.Sprintf("  ⏱  Resets in %dh %dm", hours, minutes))

	// Dynamic sizing calculation
	cardWidth := 44
	if m.width > 0 {
		cardWidth = (m.width / 2) - 3
	}
	if cardWidth > 60 {
		cardWidth = 60
	}
	if cardWidth < 30 {
		cardWidth = 30
	}
	spriteW := cardWidth - 4

	// Render each skin as a card
	var cards []string
	for _, skin := range m.skins {
		cards = append(cards, renderSkinCard(skin, -1, cardWidth, spriteW))
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
	return timer + "\n\n" + grid
}

// renderSkinCard creates a single skin display card with sprite + name + price.
// discountPct < 0 means no discount (regular shop). >= 0 means Night Market.
func renderSkinCard(skin models.ResolvedSkin, discountPct int, cardWidth int, spriteWidth int) string {
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

	// ANSI sprite (if available or dynamically rendered)
	spr := ""
	if skin.IconURL != "" && spriteWidth >= 20 {
		spr = sprite.Render(skin.IconURL, spriteWidth)
	} else if skin.Sprite != "" {
		spr = skin.Sprite
	}

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

	return CardStyle.
		BorderForeground(rarityColor).
		Width(cardWidth).
		Render(content.String())
}
