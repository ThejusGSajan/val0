package tui

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/val-tracker/val-tracker/internal/models"
	"github.com/val-tracker/val-tracker/internal/sprite"
)

func createTestImageServer(t *testing.T) *httptest.Server {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, 32, 32))
	for y := 0; y < 32; y++ {
		for x := 0; x < 32; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: 50, B: 50, A: 255})
		}
	}
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		t.Fatalf("failed to encode test png: %v", err)
	}
	pngBytes := buf.Bytes()

	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(pngBytes)
	}))
}

func TestWishlist_MarginRightZero(t *testing.T) {
	// Validate that previewBox style sets MarginRight(0) and does not leak unstyled trailing margin
	previewContentWidth := 30
	previewContent := strings.Repeat(" ", previewContentWidth)

	// CardStyle without MarginRight(0) inherits MarginRight(1) from theme.go
	defaultCard := CardStyle.
		BorderForeground(ColorBorder).
		Width(previewContentWidth).
		Render(previewContent)

	// CardStyle with MarginRight(0) as configured in wishlist.go
	zeroMarginCard := CardStyle.
		BorderForeground(ColorBorder).
		Width(previewContentWidth).
		MarginRight(0).
		Render(previewContent)

	// defaultCard includes 1-column margin from CardStyle MarginRight(1)
	// zeroMarginCard includes 0-column margin from MarginRight(0)
	if lipgloss.Width(defaultCard)-lipgloss.Width(zeroMarginCard) != 1 {
		t.Errorf("expected defaultCard to be 1 column wider than zeroMarginCard, got default=%d, zero=%d",
			lipgloss.Width(defaultCard), lipgloss.Width(zeroMarginCard))
	}

	// Verify each line of zeroMarginCard ends cleanly with the right border character and no trailing space
	for _, line := range strings.Split(zeroMarginCard, "\n") {
		if strings.HasSuffix(line, " ") {
			t.Errorf("expected zeroMarginCard line to end without trailing margin space: %q", line)
		}
	}
}

func TestWishlist_SixelOverlayConstruction(t *testing.T) {
	ts := createTestImageServer(t)
	defer ts.Close()

	sprite.DetectTerminalProtocol()
	sprite.ClearCache()
	sprite.OverrideProtocol(sprite.ProtocolSixel)
	defer sprite.OverrideProtocol(sprite.ProtocolHalfBlock)

	url := ts.URL + "/skin.png"
	iconURL := url
	m := NewWishlistModel()
	m.SetSize(100, 30)
	m.SetAllSkins([]models.SkinAsset{
		{
			UUID:        "test-skin-1",
			DisplayName: "Prime Vandal",
			DisplayIcon: &iconURL,
		},
	})
	m.focusSection = 1 // focus browse

	rawView := m.View()
	if !strings.Contains(rawView, "\x1b]999;INJECT_") {
		t.Fatalf("expected rawView to contain Sixel payload placeholder: %s", rawView)
	}

	injected := sprite.InjectPayloads(rawView)

	// 1. Verify cursor save (\x1b7) and cursor restore (\x1b8)
	if !strings.Contains(injected, "\x1b7") {
		t.Errorf("expected Sixel payload to contain cursor save \\x1b7")
	}
	if !strings.Contains(injected, "\x1b8") {
		t.Errorf("expected Sixel payload to contain cursor restore \\x1b8")
	}

	// 2. Verify linesUp (\x1b[...A) and colOffset (\x1b[...G)
	if !strings.Contains(injected, "\x1b[") || !strings.Contains(injected, "A") || !strings.Contains(injected, "G") {
		t.Errorf("expected Sixel payload to contain cursor positioning sequences (linesUp/colOffset)")
	}

	// 3. Verify Shop v26 trailingWipe
	if !strings.Contains(injected, "\x1b[48;2;13;15;23m\x1b[K") {
		t.Errorf("expected Sixel payload to contain Shop v26 trailingWipe sequence")
	}

	// 4. Verify invalidator token is attached
	if !strings.Contains(injected, "\x1b]999;wl=") {
		t.Errorf("expected Sixel payload to contain zero-width OSC 999 invalidator token")
	}

	// 5. Verify absence of belowSpriteWipe (after Sixel ST \x1b\\, \x1b8 must follow immediately without \x1b[1B wipes)
	stIdx := strings.LastIndex(injected, "\x1b\\")
	sixelEnd := strings.Index(injected, "\x1b8")
	if stIdx != -1 && sixelEnd != -1 {
		between := injected[stIdx+2 : sixelEnd]
		if strings.Contains(between, "\x1b[1B") {
			t.Errorf("found belowSpriteWipe sequence (\\x1b[1B) between Sixel end and \\x1b8, which was supposed to be eliminated")
		}
		if between != "" {
			t.Errorf("expected immediate \\x1b8 after Sixel payload, got: %q", between)
		}
	} else {
		t.Errorf("expected both Sixel terminator \\x1b\\ and cursor restore \\x1b8 in payload")
	}
}

func TestWishlist_SearchKeystrokeInvalidator(t *testing.T) {
	ts := createTestImageServer(t)
	defer ts.Close()

	sprite.DetectTerminalProtocol()
	sprite.ClearCache()
	sprite.OverrideProtocol(sprite.ProtocolSixel)
	defer sprite.OverrideProtocol(sprite.ProtocolHalfBlock)

	url := ts.URL + "/skin.png"
	iconURL := url
	skins := []models.SkinAsset{
		{
			UUID:        "skin-1",
			DisplayName: "Prime Vandal",
			DisplayIcon: &iconURL,
		},
		{
			UUID:        "skin-2",
			DisplayName: "Prime 2.0 Phantom",
			DisplayIcon: &iconURL,
		},
	}

	m1 := NewWishlistModel()
	m1.SetSize(100, 30)
	m1.SetAllSkins(skins)
	m1.focusSection = 1
	m1.searchInput = "prime"
	m1.filterSkins()

	m2 := NewWishlistModel()
	m2.SetSize(100, 30)
	m2.SetAllSkins(skins)
	m2.focusSection = 1
	m2.searchInput = "prime "
	m2.filterSkins()

	v1 := sprite.InjectPayloads(m1.View())
	v2 := sprite.InjectPayloads(m2.View())

	// Top skin result is "Prime Vandal" in both cases
	// But the invalidator tokens must be distinct
	tok1 := "\x1b]999;wl=prime;"
	tok2 := "\x1b]999;wl=prime ;"

	if !strings.Contains(v1, tok1) {
		t.Errorf("expected view 1 to contain invalidator token %q, got:\n%s", tok1, v1)
	}
	if !strings.Contains(v2, tok2) {
		t.Errorf("expected view 2 to contain invalidator token %q, got:\n%s", tok2, v2)
	}

	// Split views into lines and compare the line holding the payload
	lines1 := strings.Split(v1, "\n")
	lines2 := strings.Split(v2, "\n")

	var payloadLine1, payloadLine2 string
	for _, l := range lines1 {
		if strings.Contains(l, "\x1b]999;wl=") {
			payloadLine1 = l
			break
		}
	}
	for _, l := range lines2 {
		if strings.Contains(l, "\x1b]999;wl=") {
			payloadLine2 = l
			break
		}
	}

	if payloadLine1 == "" || payloadLine2 == "" {
		t.Fatalf("failed to locate payload lines in views")
	}

	if payloadLine1 == payloadLine2 {
		t.Errorf("expected payload lines to be distinct between keystrokes to prevent Bubble Tea line-diff suppression, but they were identical")
	}
}

func TestWishlist_HalfBlockMode(t *testing.T) {
	ts := createTestImageServer(t)
	defer ts.Close()

	sprite.DetectTerminalProtocol()
	sprite.ClearCache()
	sprite.OverrideProtocol(sprite.ProtocolHalfBlock)

	url := ts.URL + "/skin.png"
	iconURL := url
	m := NewWishlistModel()
	m.SetSize(100, 30)
	m.SetAllSkins([]models.SkinAsset{
		{
			UUID:        "test-skin-1",
			DisplayName: "Prime Vandal",
			DisplayIcon: &iconURL,
		},
	})
	m.focusSection = 1

	view := m.View()

	// Half-block mode must not register or inject Sixel overlays
	if strings.Contains(view, "\x1b]999;INJECT_") {
		t.Errorf("expected half-block mode view to not contain Sixel injection placeholder")
	}
	if strings.Contains(view, "\x1b]999;wl=") {
		t.Errorf("expected half-block mode view to not contain invalidator token")
	}
	if strings.Contains(view, "\x1b7") || strings.Contains(view, "\x1b8") {
		t.Errorf("expected half-block mode view to not contain cursor save/restore escapes")
	}
}

func TestWishlist_EmptySearchResults(t *testing.T) {
	m := NewWishlistModel()
	m.SetSize(100, 30)
	m.SetAllSkins([]models.SkinAsset{
		{
			UUID:        "test-skin-1",
			DisplayName: "Prime Vandal",
		},
	})
	m.focusSection = 1
	m.searchInput = "nonexistent-skin-xyz"
	m.filterSkins()

	if len(m.filtered) != 0 {
		t.Fatalf("expected 0 filtered skins, got %d", len(m.filtered))
	}

	// Must render cleanly without panic
	view := m.View()
	if view == "" {
		t.Errorf("expected non-empty view for empty search results")
	}
	if !strings.Contains(view, "BROWSE & ADD SKINS:") {
		t.Errorf("expected view to contain BROWSE & ADD SKINS section header")
	}
	if strings.Contains(view, "\x1b]999;INJECT_") {
		t.Errorf("expected no overlay placeholder when no skin is selected")
	}
}
