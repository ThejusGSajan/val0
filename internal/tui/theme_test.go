package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestRarityStylesAndUUIDs(t *testing.T) {
	// Verify that all 5 Valorant content tiers are mapped correctly
	expectedTiers := map[string]string{
		"12683d76-48d7-84a3-4e09-6985794f0445": "Select",
		"0cebb8be-46d7-c12a-d306-e9907bfc5a25": "Deluxe",
		"60bca009-4182-7998-dee7-b8a2558dc369": "Premium",
		"e046854e-406c-37f4-6607-19a9ba8426fc": "Exclusive",
		"411e4a55-4e59-7757-41f0-86a53f101bb5": "Ultra",
	}

	for uuid, expectedName := range expectedTiers {
		name, ok := RarityNameMap[uuid]
		if !ok {
			t.Errorf("UUID %s missing from RarityNameMap", uuid)
		}
		if name != expectedName {
			t.Errorf("UUID %s: expected %s, got %s", uuid, expectedName, name)
		}

		if _, ok := RarityColorMap[uuid]; !ok {
			t.Errorf("UUID %s missing from RarityColorMap", uuid)
		}

		style := RarityStyle(uuid)
		rendered := style.Render("Test")
		if rendered == "" {
			t.Errorf("failed to render styled text for tier %s", expectedName)
		}
	}
}

func TestProgressBar(t *testing.T) {
	bar := renderProgressBar(0.5, 20, "Test Bar")
	if bar == "" {
		t.Fatal("expected non-empty progress bar")
	}
}

func TestKeyLegendRenderers(t *testing.T) {
	single := RenderKeyItem("esc", "Close")
	if !strings.Contains(single, "esc") || !strings.Contains(single, "Close") {
		t.Fatalf("RenderKeyItem failed: %q", single)
	}

	joined := RenderKeyLegends(
		[2]string{"↑/↓", "Scroll"},
		[2]string{"esc", "Close"},
	)
	if !strings.Contains(joined, " · ") {
		t.Fatalf("expected middle dot separator ' · ', got: %q", joined)
	}
}

func TestColorPaletteConstraints(t *testing.T) {
	if ColorAccent != lipgloss.Color("#EC666D") {
		t.Fatalf("expected ColorAccent to be #EC666D, got: %v", ColorAccent)
	}
	if ColorTabActive != lipgloss.Color("#EC666D") {
		t.Fatalf("expected ColorTabActive to be #EC666D, got: %v", ColorTabActive)
	}
	if ColorLoss != lipgloss.Color("#EF4444") {
		t.Fatalf("expected ColorLoss (defeat red) to remain #EF4444, got: %v", ColorLoss)
	}
}

