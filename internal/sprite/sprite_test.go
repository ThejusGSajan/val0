package sprite

import (
	"fmt"
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

	result := renderHalfBlocks(img, 4, 2)
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
	res := Render("", 40, 4)
	if res != "" {
		t.Errorf("expected empty string for empty URL, got: %s", res)
	}
}

func TestRender_WidthTooSmall(t *testing.T) {
	res := Render("https://example.com/icon.png", 5, 4)
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
	expectedHeader := "\x1bP0;1;8q\"1;1;8;8"
	if !strings.Contains(sixelStr, expectedHeader) {
		t.Errorf("expected Sixel DCS header %q, got: %q", expectedHeader, sixelStr[:min(len(sixelStr), 30)])
	}
}

func TestRenderKittyEncoding(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 4, 4))
	kittyStr, err := renderKitty(img, 4, 4)
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

func TestPayloadRegistry(t *testing.T) {
	payload := "\x1b7\x1b[30D\x1b[3A<SIXEL_BYTES>\x1b8"
	placeholder := RegisterPayload(payload)
	if !strings.HasPrefix(placeholder, "\x1b]999;INJECT_") {
		t.Fatalf("unexpected placeholder prefix: %q", placeholder)
	}

	screen := fmt.Sprintf("Line1\nLine2%s\nLine3", placeholder)
	injected := InjectPayloads(screen)

	if strings.Contains(injected, placeholder) {
		t.Errorf("expected placeholder to be replaced")
	}
	if !strings.Contains(injected, payload) {
		t.Errorf("expected payload to be present in injected output")
	}

	// Secondary injection should have flushed registry
	second := InjectPayloads(screen)
	if second != screen {
		t.Errorf("expected empty registry on subsequent inject, got %q", second)
	}
}

func TestTrimTransparency(t *testing.T) {
	// Create a 10x10 image with transparent border and 4x4 opaque content at (3,3) -> (6,6)
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	for y := 3; y <= 6; y++ {
		for x := 3; x <= 6; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	trimmed := trimTransparency(img)
	if trimmed.Bounds().Dx() != 4 || trimmed.Bounds().Dy() != 4 {
		t.Errorf("expected 4x4 trimmed bounds, got %dx%d (bounds: %+v)", trimmed.Bounds().Dx(), trimmed.Bounds().Dy(), trimmed.Bounds())
	}

	// Fully transparent image should return original image
	emptyImg := image.NewRGBA(image.Rect(0, 0, 10, 10))
	trimmedEmpty := trimTransparency(emptyImg)
	if trimmedEmpty.Bounds().Dx() != 10 || trimmedEmpty.Bounds().Dy() != 10 {
		t.Errorf("expected original bounds for empty image, got %dx%d", trimmedEmpty.Bounds().Dx(), trimmedEmpty.Bounds().Dy())
	}
}

func TestRenderTrimmed_EmptyURL(t *testing.T) {
	res := RenderTrimmed("", 40, 4)
	if res != "" {
		t.Errorf("expected empty string for empty URL, got: %s", res)
	}
}

func TestRenderTrimmed_WidthTooSmall(t *testing.T) {
	res := RenderTrimmed("https://example.com/icon.png", 5, 4)
	if res != "" {
		t.Errorf("expected empty string for width < 10, got: %s", res)
	}
}

