package sprite

import (
	"bytes"
	"image"

	"github.com/mattn/go-sixel"
)

func renderSixel(img image.Image) (string, error) {
	var buf bytes.Buffer
	enc := sixel.NewEncoder(&buf)
	enc.Dither = true
	if err := enc.Encode(img); err != nil {
		return "", err
	}
	return buf.String(), nil
}
