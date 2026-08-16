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
	res := Render("")
	if res != "" {
		t.Errorf("expected empty string for empty URL, got: %s", res)
	}
}
