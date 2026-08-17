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

// spriteCache prevents re-downloading the same icon within a session.
var (
	spriteMap   = make(map[string]string)
	spriteMutex sync.Mutex
)

// Render downloads the PNG at iconURL and converts it into an ANSI
// block-character string suitable for terminal display.
//
// The width parameter controls columns; height is computed proportionally
// with a 0.45 aspect ratio (weapon icons are landscape-oriented, ~3:1 aspect).
//
// Width is clamped to a maximum of 55. If width < 20, Render returns an empty string.
//
// Uses the Unicode UPPER HALF BLOCK (▀) with foreground = top pixel,
// background = bottom pixel, achieving 2× vertical resolution.
//
// Returns empty string on any error (graceful degradation).
func Render(iconURL string, width int) string {
	if iconURL == "" {
		return ""
	}

	if width > 55 {
		width = 55
	}
	if width < 20 {
		return ""
	}

	cacheKey := fmt.Sprintf("%s:%d", iconURL, width)

	spriteMutex.Lock()
	if cached, ok := spriteMap[cacheKey]; ok {
		spriteMutex.Unlock()
		return cached
	}
	spriteMutex.Unlock()

	img, err := fetchImage(iconURL)
	if err != nil {
		return ""
	}

	// Resize the image while preserving aspect ratio
	// By passing 0 for height, nfnt/resize automatically calculates the height
	// to maintain the original image's proportions.
	// Since each terminal cell renders two vertical pixels (using half-blocks)
	// and terminal cells are approximately 1:2 aspect ratio, a half-block is a perfect 1:1 square.
	// Thus, resizing the image in normal pixels translates perfectly to terminal cells.
	resized := resize.Resize(
		uint(width),
		0, // Preserve aspect ratio
		img,
		resize.Lanczos3,
	)

	result := renderHalfBlocks(resized)

	spriteMutex.Lock()
	spriteMap[cacheKey] = result
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
