package sprite

import (
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

const (
	// Target sprite dimensions in terminal "pixels" (columns × half-rows).
	// Each cell uses a Unicode half-block so 1 char = 2 vertical pixels.
	SpriteWidth  = 40 // columns
	SpriteHeight = 10 // character rows (= 20 real pixel rows)
)

// spriteCache prevents re-downloading the same icon within a session.
var (
	spriteMap   = make(map[string]string)
	spriteMutex sync.Mutex
)

// Render downloads the PNG at iconURL and converts it into an ANSI
// block-character string suitable for terminal display.
//
// Uses the Unicode UPPER HALF BLOCK (▀) with foreground = top pixel,
// background = bottom pixel, achieving 2× vertical resolution.
//
// Returns empty string on any error (graceful degradation).
func Render(iconURL string) string {
	if iconURL == "" {
		return ""
	}

	spriteMutex.Lock()
	if cached, ok := spriteMap[iconURL]; ok {
		spriteMutex.Unlock()
		return cached
	}
	spriteMutex.Unlock()

	img, err := fetchImage(iconURL)
	if err != nil {
		return ""
	}

	// Resize: SpriteWidth columns, SpriteHeight*2 actual pixel rows
	// (each terminal row encodes 2 pixel rows via half-blocks).
	resized := resize.Resize(
		uint(SpriteWidth),
		uint(SpriteHeight*2),
		img,
		resize.Lanczos3,
	)

	result := renderHalfBlocks(resized)

	spriteMutex.Lock()
	spriteMap[iconURL] = result
	spriteMutex.Unlock()

	return result
}

func fetchImage(url string) (image.Image, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to fetch image: status %d", resp.StatusCode)
	}

	img, _, err := image.Decode(resp.Body)
	return img, err
}

// renderHalfBlocks converts an image into a string of ANSI-colored
// half-block characters. Each output row merges two pixel rows:
//   - Top pixel → foreground color (\x1b[38;2;R;G;Bm)
//   - Bottom pixel → background color (\x1b[48;2;R;G;Bm)
//   - Character: ▀ (upper half block)
func renderHalfBlocks(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	var sb strings.Builder
	for y := 0; y < height-1; y += 2 {
		for x := 0; x < width; x++ {
			r1, g1, b1, a1 := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r2, g2, b2, a2 := img.At(bounds.Min.X+x, bounds.Min.Y+y+1).RGBA()

			// Convert from 16-bit to 8-bit color
			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			if a1>>8 < 32 && a2>>8 < 32 {
				// Both pixels transparent → space
				sb.WriteString(" ")
			} else if a1>>8 < 32 {
				// Top transparent, bottom visible → lower half block with fg
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▄\x1b[0m", r2, g2, b2))
			} else if a2>>8 < 32 {
				// Top visible, bottom transparent → upper half block with fg
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1))
			} else {
				// Both visible → upper half block, fg=top, bg=bottom
				sb.WriteString(fmt.Sprintf(
					"\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m",
					r1, g1, b1, r2, g2, b2,
				))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

// ClearCache resets the in-memory sprite cache (used on `r` refresh).
func ClearCache() {
	spriteMutex.Lock()
	spriteMap = make(map[string]string)
	spriteMutex.Unlock()
}
