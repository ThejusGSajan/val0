package tui

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
)

type NightMarketModel struct {
	skins     []models.ResolvedSkin
	discounts []int // parallel to skins, e.g. [30, 25, 40, ...]
}

func NewNightMarketModel(skins []models.ResolvedSkin, discounts []int) NightMarketModel {
	return NightMarketModel{skins: skins, discounts: discounts}
}

func (m NightMarketModel) Update(msg tea.Msg) (NightMarketModel, tea.Cmd) {
	return m, nil
}

func (m NightMarketModel) View() string {
	if len(m.skins) == 0 {
		return lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  Night Market is not currently active.")
	}

	header := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render("  ✦ NIGHT MARKET ✦")

	var cards []string
	for i, skin := range m.skins {
		disc := 0
		if i < len(m.discounts) {
			disc = m.discounts[i]
		}
		cards = append(cards, renderSkinCard(skin, disc))
	}

	// Night market has 6 items — 3 × 2 grid or 2x3
	var rows []string
	for i := 0; i < len(cards); i += 2 {
		end := i + 2
		if end > len(cards) {
			end = len(cards)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, cards[i:end]...)
		rows = append(rows, row)
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	return header + "\n\n" + grid
}

func (m NightMarketModel) HasData() bool {
	return len(m.skins) > 0
}
