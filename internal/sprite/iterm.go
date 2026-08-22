package sprite

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"image"
	"image/png"
)

func renderITerm2(img image.Image, widthCols int) (string, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", err
	}
	encoded := base64.StdEncoding.EncodeToString(buf.Bytes())
	return fmt.Sprintf("\x1b]1337;File=inline=1;width=%d:%s\x07", widthCols, encoded), nil
}
