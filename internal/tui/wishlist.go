package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/val-tracker/val-tracker/internal/cache"
)

type WishlistModel struct {
	entries  []cache.WishlistEntry
	cursor   int
	width    int
	height   int
	flashMsg string
}

func NewWishlistModel() WishlistModel {
	entries, _ := cache.LoadWishlist()
	return WishlistModel{
		entries: entries,
	}
}

func (m *WishlistModel) Refresh() {
	entries, _ := cache.LoadWishlist()
	m.entries = entries
	if m.cursor >= len(m.entries) && len(m.entries) > 0 {
		m.cursor = len(m.entries) - 1
	}
}

func (m *WishlistModel) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m WishlistModel) Update(msg tea.Msg) (WishlistModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k", "w":
			if m.cursor > 0 {
				m.cursor--
				m.flashMsg = ""
			}
		case "down", "j", "s":
			if m.cursor < len(m.entries)-1 {
				m.cursor++
				m.flashMsg = ""
			}
		case "x", "delete", "backspace":
			if len(m.entries) > 0 && m.cursor < len(m.entries) {
				removed := m.entries[m.cursor]
				_ = cache.RemoveFromWishlist(removed.UUID)
				m.Refresh()
				m.flashMsg = fmt.Sprintf("Removed %s from wishlist.", removed.Name)
			}
		}
	}
	return m, nil
}

func (m WishlistModel) View() string {
	var sb strings.Builder

	lineWidth := max(m.width-4, 70)

	header := lipgloss.NewStyle().
		Foreground(ColorUltra).
		Bold(true).
		Render(fmt.Sprintf("  ⭐ SKIN WISHLIST                                         %d skins tracked", len(m.entries)))
	sb.WriteString(header + "\n")
	sb.WriteString("  " + strings.Repeat("─", lineWidth) + "\n")

	if m.flashMsg != "" {
		sb.WriteString(lipgloss.NewStyle().Foreground(ColorWin).Bold(true).Render("  ✔ "+m.flashMsg) + "\n\n")
	}

	if len(m.entries) == 0 {
		emptyMsg := lipgloss.NewStyle().
			Foreground(ColorMuted).
			Render("  Your wishlist is empty.\n\n  Go to the Shop tab and press 'w' on any skin to track it here!")
		sb.WriteString(emptyMsg + "\n")
	} else {
		for i, entry := range m.entries {
			cursor := "  "
			if i == m.cursor {
				cursor = "▸ "
			}

			// Rarity tag color
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
				Render(fmt.Sprintf("%2d. %-26s", i+1, truncate(entry.Name, 26)))

			rarityStr := lipgloss.NewStyle().
				Foreground(rarityColor).
				Render(fmt.Sprintf("● %-12s", entry.Rarity))

			priceStr := lipgloss.NewStyle().
				Foreground(ColorUltra).
				Bold(true).
				Render(fmt.Sprintf("VP %d", entry.CostVP))

			row := fmt.Sprintf("%s%s  %s  %s", cursor, nameStr, rarityStr, priceStr)

			if i == m.cursor {
				row = lipgloss.NewStyle().
					Background(lipgloss.Color("#1F2430")).
					Render(row)
			}

			sb.WriteString(row + "\n")
		}
	}

	sb.WriteString("\n" + lipgloss.NewStyle().
		Foreground(ColorMuted).
		Render("  Press 'x' to remove selected  •  Add skins from Shop tab with 'w'"))

	return sb.String()
}
