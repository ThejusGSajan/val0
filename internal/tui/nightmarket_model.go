package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
)

type NightMarketModel struct {
	skins     []models.ResolvedSkin
	discounts []int // parallel to skins, e.g. [30, 25, 40, ...]
	width     int
	height    int
}

func NewNightMarketModel(skins []models.ResolvedSkin, discounts []int) NightMarketModel {
	return NightMarketModel{skins: skins, discounts: discounts}
}

func (m *NightMarketModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m NightMarketModel) Update(msg tea.Msg) (NightMarketModel, tea.Cmd) {
	return m, nil
}

func (m NightMarketModel) View() string {
	if m.width == 0 {
		return ""
	}
	if len(m.skins) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  Night Market is not currently active.")
	}

	header := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render("  ✦ NIGHT MARKET ✦")

	cols := 2
	if m.width >= 140 {
		cols = 3
	}
	cardWidth := 44
	if m.width > 0 {
		cardWidth = (m.width / cols) - 3
	}
	if cardWidth > 60 {
		cardWidth = 60
	}
	if cardWidth < 30 {
		cardWidth = 30
	}
	spriteW := cardWidth - 4

	var cards []string
	for i, skin := range m.skins {
		disc := 0
		if i < len(m.discounts) {
			disc = m.discounts[i]
		}
		cards = append(cards, renderSkinCard(skin, disc, cardWidth, spriteW))
	}

	// Night market has 6 items — cols x rows grid
	var rows []string
	for i := 0; i < len(cards); i += cols {
		end := i + cols
		if end > len(cards) {
			end = len(cards)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...)
		rows = append(rows, row)
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	helpBar := lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  [s] enter shop tab  •  [w] enter wishlist tab")
	return header + "\n" + grid + "\n\n" + helpBar
}

func (m NightMarketModel) HasData() bool {
	return len(m.skins) > 0
}
