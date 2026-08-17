package tui

import "github.com/charmbracelet/lipgloss"

// ── Rarity Colors ───────────────────────────────────────────────────
// Matched to the exact highlightColor hex from valorant-api.com/v1/contenttiers.

var (
	ColorSelect    = lipgloss.Color("#5A9FE2") // Select — Blue
	ColorDeluxe    = lipgloss.Color("#009587") // Deluxe — Teal/Green
	ColorPremium   = lipgloss.Color("#D1548D") // Premium — Pink/Magenta
	ColorExclusive = lipgloss.Color("#F5955B") // Exclusive — Orange
	ColorUltra     = lipgloss.Color("#FAD663") // Ultra — Gold/Yellow

	ColorMuted     = lipgloss.Color("#6B7280")
	ColorFg        = lipgloss.Color("#E5E7EB")
	ColorBg        = lipgloss.Color("#0F1117")
	ColorAccent    = lipgloss.Color("#FF4655") // Valorant red
	ColorBorder    = lipgloss.Color("#2A2D37")
	ColorTabActive = lipgloss.Color("#FF4655")
	ColorTabInact  = lipgloss.Color("#4B5563")

	// Match outcomes
	ColorWin  = lipgloss.Color("#22C55E") // Green
	ColorLoss = lipgloss.Color("#EF4444") // Red
	ColorDraw = lipgloss.Color("#9CA3AF") // Gray
)

// ContentTierUUID → lipgloss.Color
// UUIDs verified live from https://valorant-api.com/v1/contenttiers.
var RarityColorMap = map[string]lipgloss.Color{
	"12683d76-48d7-84a3-4e09-6985794f0445": ColorSelect,    // Select
	"0cebb8be-46d7-c12a-d306-e9907bfc5a25": ColorDeluxe,    // Deluxe
	"60bca009-4182-7998-dee7-b8a2558dc369": ColorPremium,   // Premium
	"e046854e-406c-37f4-6607-19a9ba8426fc": ColorExclusive, // Exclusive
	"411e4a55-4e59-7757-41f0-86a53f101bb5": ColorUltra,     // Ultra
}

// ContentTierUUID → human-readable name
var RarityNameMap = map[string]string{
	"12683d76-48d7-84a3-4e09-6985794f0445": "Select",
	"0cebb8be-46d7-c12a-d306-e9907bfc5a25": "Deluxe",
	"60bca009-4182-7998-dee7-b8a2558dc369": "Premium",
	"e046854e-406c-37f4-6607-19a9ba8426fc": "Exclusive",
	"411e4a55-4e59-7757-41f0-86a53f101bb5": "Ultra",
}

// ── Reusable Styles ─────────────────────────────────────────────────

var (
	// App frame
	AppStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("#0F1117")).
			Foreground(ColorFg)

	// Header / title bar
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorAccent).
			PaddingLeft(1).
			PaddingRight(1)

	// Active tab
	ActiveTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorTabActive).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(ColorTabActive).
			PaddingLeft(2).
			PaddingRight(2)

	// Inactive tab
	InactiveTabStyle = lipgloss.NewStyle().
				Foreground(ColorTabInact).
				PaddingLeft(2).
				PaddingRight(2)

	// Skin card border
	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorBorder).
			Padding(1).
			MarginRight(1)

	// VP price badge
	VPBadgeStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFFFF")).
			Background(lipgloss.Color("#1F2937")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	// Status bar
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			PaddingTop(1)

	// Error screen
	ErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#EF4444")).
			Bold(true).
			Border(lipgloss.DoubleBorder()).
			BorderForeground(lipgloss.Color("#EF4444")).
			Padding(2, 4).
			Align(lipgloss.Center)

	// Progress bar (battlepass)
	ProgressFullStyle  = lipgloss.NewStyle().Foreground(ColorAccent)
	ProgressEmptyStyle = lipgloss.NewStyle().Foreground(ColorBorder)
)

// RarityStyle returns a lipgloss style that colors text by content tier UUID.
func RarityStyle(contentTierUUID string) lipgloss.Style {
	color, ok := RarityColorMap[contentTierUUID]
	if !ok {
		color = ColorMuted
	}
	return lipgloss.NewStyle().Foreground(color).Bold(true)
}
