package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
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
	cardContentWidth := 40
	if m.width > 0 {
		cardContentWidth = (m.width / cols) - 8
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
		disc := 0
		if i < len(m.discounts) {
			disc = m.discounts[i]
		}
		rendered, overlay := renderSkinCard(skin, disc, cardContentWidth)
		cardsWithOverlay = append(cardsWithOverlay, cardWithOverlay{
			rendered: rendered,
			overlay:  overlay,
			colIndex: i % cols,
		})
	}

	// Night market has 6 items — cols x rows grid
	var rows []string
	for i := 0; i < len(cardsWithOverlay); i += cols {
		end := i + cols
		if end > len(cardsWithOverlay) {
			end = len(cardsWithOverlay)
		}
		rowCards := cardsWithOverlay[i:end]

		var renderedPieces []string
		for _, c := range rowCards {
			renderedPieces = append(renderedPieces, c.rendered)
		}
		gridRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedPieces...)

		gridRowLines := strings.Split(gridRow, "\n")
		totalLines := len(gridRowLines)
		linesUp := totalLines - 3

		for colIdx, c := range rowCards {
			if c.overlay != nil {
				colOffset := colIdx*(cardContentWidth+5) + 2
				wipe := buildWipeSeq(colOffset, c.overlay.ContentWidth, c.overlay.SpriteRows)
				payload := "\x1b7" +
					fmt.Sprintf("\x1b[%dA", linesUp) +
					fmt.Sprintf("\x1b[%dG", colOffset+1) +
					wipe + c.overlay.Payload + "\x1b8"
				placeholder := sprite.RegisterPayload(payload)
				gridRowLines[totalLines-1] += placeholder
			}
		}

		gridRow = strings.Join(gridRowLines, "\n")
		rows = append(rows, gridRow)
	}

	grid := lipgloss.JoinVertical(lipgloss.Left, rows...)
	helpBar := "  " + RenderKeyLegends(
		[2]string{"s", "enter shop tab"},
		[2]string{"w", "enter wishlist tab"},
	)

	return header + "\n" + grid + "\n\n" + helpBar
}

func (m NightMarketModel) HasData() bool {
	return len(m.skins) > 0
}
