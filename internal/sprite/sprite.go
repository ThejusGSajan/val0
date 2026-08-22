package sprite

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/jpeg"
	_ "image/png"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nfnt/resize"
)

var (
	spriteMap   = make(map[string]string)
	spriteMutex sync.Mutex
)

// Render downloads the image at iconURL and renders it using the best
// available terminal graphics protocol (Sixel, Kitty, iTerm2, or ANSI half-blocks).
func Render(iconURL string, widthCols int) string {
	if iconURL == "" || widthCols < 10 {
		return ""
	}

	proto := DetectTerminalProtocol()
	cacheKey := fmt.Sprintf("%s:%d:%s", iconURL, widthCols, proto.String())

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

	var result string
	if proto != ProtocolHalfBlock {
		// 1B & 2A: Fixed Canvas Normalization
		targetPixelWidth := widthCols * 9
		targetPixelHeight := 80 // Exactly 4 terminal rows (approx 20px per row)

		// Resize preserving aspect ratio to fit inside bounding box
		resized := resize.Thumbnail(uint(targetPixelWidth), uint(targetPixelHeight), img, resize.Lanczos3)

		// Create Solid Background Canvas (#0F1117)
		canvas := image.NewRGBA(image.Rect(0, 0, targetPixelWidth, targetPixelHeight))
		bgColor := color.RGBA{R: 15, G: 17, B: 23, A: 255}
		draw.Draw(canvas, canvas.Bounds(), &image.Uniform{bgColor}, image.Point{}, draw.Src)

		// Center the resized weapon onto the canvas
		offsetX := (targetPixelWidth - resized.Bounds().Dx()) / 2
		offsetY := (targetPixelHeight - resized.Bounds().Dy()) / 2
		draw.Draw(canvas, image.Rect(offsetX, offsetY, offsetX+resized.Bounds().Dx(), offsetY+resized.Bounds().Dy()), resized, image.Point{}, draw.Over)

		// Encode the fully normalized canvas
		switch proto {
		case ProtocolSixel:
			if str, err := renderSixel(canvas); err == nil {
				result = str
			}
		case ProtocolKitty:
			if str, err := renderKitty(canvas); err == nil {
				result = str
			}
		case ProtocolITerm2:
			if str, err := renderITerm2(canvas, widthCols); err == nil {
				result = str
			}
		}
	}

	// Fallback to ANSI Half-Blocks if native protocol failed or unsupported
	if result == "" {
		clampedWidth := widthCols
		if clampedWidth > 55 {
			clampedWidth = 55
		}
		if clampedWidth < 20 {
			clampedWidth = 20
		}
		resized := resize.Resize(uint(clampedWidth), 0, img, resize.Lanczos3)
		result = renderHalfBlocks(resized)
	}

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

func renderHalfBlocks(img image.Image) string {
	bounds := img.Bounds()
	width := bounds.Max.X - bounds.Min.X
	height := bounds.Max.Y - bounds.Min.Y

	var sb strings.Builder
	for y := 0; y < height-1; y += 2 {
		for x := 0; x < width; x++ {
			r1, g1, b1, a1 := img.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r2, g2, b2, a2 := img.At(bounds.Min.X+x, bounds.Min.Y+y+1).RGBA()

			r1, g1, b1 = r1>>8, g1>>8, b1>>8
			r2, g2, b2 = r2>>8, g2>>8, b2>>8

			if a1>>8 < 32 && a2>>8 < 32 {
				sb.WriteString(" ")
			} else if a1>>8 < 32 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▄\x1b[0m", r2, g2, b2))
			} else if a2>>8 < 32 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1))
			} else {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀\x1b[0m", r1, g1, b1, r2, g2, b2))
			}
		}
		sb.WriteString("\n")
	}
	return sb.String()
}

func ClearCache() {
	spriteMutex.Lock()
	spriteMap = make(map[string]string)
	spriteMutex.Unlock()
}
