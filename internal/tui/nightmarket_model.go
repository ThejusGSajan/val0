package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ThejusGSajan/val0/internal/cache"
	"github.com/ThejusGSajan/val0/internal/models"
	"github.com/ThejusGSajan/val0/internal/sprite"
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

	minWidthFor3Col := 98
	cols := 3
	if m.width > 0 && m.width < minWidthFor3Col {
		cols = 2
	}
	cardContentWidth := 40
	if m.width > 0 {
		cardContentWidth = (m.width - (3*cols - 1)) / cols
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
		colIdx := i % cols
		offsetX := 0
		inWishlist := cache.IsInWishlist(skin.UUID)
		rendered, overlay := renderSkinCardWithWishlist(skin, disc, cardContentWidth, inWishlist, offsetX)
		cardsWithOverlay = append(cardsWithOverlay, cardWithOverlay{
			rendered: rendered,
			overlay:  overlay,
			colIndex: colIdx,
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
		for idx, c := range rowCards {
			if idx > 0 {
				cardLines := strings.Count(c.rendered, "\n") + 1
				spacerLine := lipgloss.NewStyle().Background(ColorBg).Render(" ")
				spacerLines := make([]string, cardLines)
				for s := range spacerLines {
					spacerLines[s] = spacerLine
				}
				spacer := strings.Join(spacerLines, "\n")
				renderedPieces = append(renderedPieces, spacer)
			}
			renderedPieces = append(renderedPieces, c.rendered)
		}
		gridRow := lipgloss.JoinHorizontal(lipgloss.Top, renderedPieces...)

		gridRowLines := strings.Split(gridRow, "\n")
		totalLines := len(gridRowLines)
		linesUp := totalLines - 3

		numCardsInRow := len(rowCards)
		totalGridRowWidth := numCardsInRow*(cardContentWidth+2) + (numCardsInRow - 1)
		wipeCols := 120
		trailingWipe := fmt.Sprintf("\x1b[48;2;13;15;23m\x1b[K%s\x1b[%dG", strings.Repeat(" ", wipeCols), totalGridRowWidth+1)

		lastOverlayIdx := -1
		for idx, c := range rowCards {
			if c.overlay != nil {
				lastOverlayIdx = idx
			}
		}

		for colIdx, c := range rowCards {
			if c.overlay != nil {
				spriteMarginNM := ((cardContentWidth - 2) - c.overlay.ContentWidth) / 2
				colOffset := colIdx*(cardContentWidth+3) + 2 + spriteMarginNM + c.overlay.OffsetX
				wipe := buildWipeSeq(colOffset, c.overlay.ContentWidth, c.overlay.SpriteRows)
				tw := ""
				if colIdx == lastOverlayIdx {
					tw = trailingWipe
				}
				payload := "\x1b7" +
					fmt.Sprintf("\x1b[%dA", linesUp) +
					fmt.Sprintf("\x1b[%dG", colOffset+1) +
					wipe + c.overlay.Payload + "\x1b8" + tw
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
