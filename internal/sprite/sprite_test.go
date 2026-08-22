package sprite

import (
	"image"
	"image/color"
	"strings"
	"testing"
)

func TestRenderHalfBlocks(t *testing.T) {
	// Create a simple 4x4 image
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	for y := 0; y < 4; y++ {
		for x := 0; x < 4; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	result := renderHalfBlocks(img)
	if result == "" {
		t.Fatal("expected non-empty rendered string")
	}

	// Should contain ANSI escape codes and half blocks
	if !strings.Contains(result, "▀") {
		t.Errorf("expected rendered output to contain '▀', got: %s", result)
	}

	lines := strings.Split(strings.TrimSpace(result), "\n")
	if len(lines) != 2 {
		t.Errorf("expected 2 terminal lines for 4 pixel rows, got %d", len(lines))
	}
}

func TestRender_EmptyURL(t *testing.T) {
	res := Render("", 40)
	if res != "" {
		t.Errorf("expected empty string for empty URL, got: %s", res)
	}
}

func TestRender_WidthTooSmall(t *testing.T) {
	res := Render("https://example.com/icon.png", 5)
	if res != "" {
		t.Errorf("expected empty string for width < 10, got: %s", res)
	}
}

func TestRenderSixelEncoding(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	sixelStr, err := renderSixel(img)
	if err != nil {
		t.Fatalf("renderSixel failed: %v", err)
	}
	if len(sixelStr) == 0 || !strings.Contains(sixelStr, "\x1bP") {
		t.Errorf("expected Sixel DCS escape sequence, got %q", sixelStr)
	}
}

func TestRenderKittyEncoding(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	kittyStr, err := renderKitty(img)
	if err != nil {
		t.Fatalf("renderKitty failed: %v", err)
	}
	if !strings.Contains(kittyStr, "\x1b_G") {
		t.Errorf("expected Kitty escape sequence, got %q", kittyStr)
	}
}

func TestRenderITerm2Encoding(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	itermStr, err := renderITerm2(img, 40)
	if err != nil {
		t.Fatalf("renderITerm2 failed: %v", err)
	}
	if !strings.Contains(itermStr, "\x1b]1337;File=inline=1;width=40:") {
		t.Errorf("expected iTerm2 escape sequence, got %q", itermStr)
	}
}
