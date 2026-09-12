package sprite

import (
	"bytes"
	"fmt"
	"image"
	"strings"

	"github.com/mattn/go-sixel"
)

func renderSixel(img image.Image) (string, error) {
	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	enc.Dither = true
	enc.Colors = 256
	if err := enc.Encode(img); err != nil {
		return "", err
	}
	s := buf.String()
	w := img.Bounds().Dx()
	h := img.Bounds().Dy()

	// Fix Sixel DCS header:
	// 1. P2=1: Set background mode to transparent/do not clear background (default P2=0 clears background).
	// 2. "1;1;w;h: Explicitly provide Ph (w) and Pv (h) extents so DEC/ConPTY parsers do not default to wiping to the right terminal edge.
	oldHeader := "\x1bP0;0;8q\"1;1"
	newHeader := fmt.Sprintf("\x1bP0;1;8q\"1;1;%d;%d", w, h)
	s = strings.Replace(s, oldHeader, newHeader, 1)
	return s, nil
}
