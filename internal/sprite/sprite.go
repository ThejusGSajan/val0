package sprite

import (
	"fmt"
	"image"
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
func Render(iconURL string, widthCols int, targetRows int) string {
	if iconURL == "" || widthCols < 10 || targetRows <= 0 {
		return ""
	}

	proto := DetectTerminalProtocol()
	cacheKey := fmt.Sprintf("%s:%d:%d:%s", iconURL, widthCols, targetRows, proto.String())

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
		const cellW = 10
		const cellH = 20
		targetPixelWidth := widthCols * cellW
		targetPixelHeight := targetRows * cellH

		// Resize preserving aspect ratio to fit inside bounding box
		resized := resize.Thumbnail(uint(targetPixelWidth), uint(targetPixelHeight), img, resize.Lanczos3)

		// Create Transparent Background Canvas (A=0 everywhere)
		canvas := image.NewRGBA(image.Rect(0, 0, targetPixelWidth, targetPixelHeight))

		// Center the resized weapon onto the canvas
		offsetX := (targetPixelWidth - resized.Bounds().Dx()) / 2
		offsetY := (targetPixelHeight - resized.Bounds().Dy()) / 2
		draw.Draw(canvas, image.Rect(offsetX, offsetY, offsetX+resized.Bounds().Dx(), offsetY+resized.Bounds().Dy()), resized, image.Point{}, draw.Src)

		// Encode the fully normalized canvas
		switch proto {
		case ProtocolSixel:
			if str, err := renderSixel(canvas); err == nil {
				result = str
			}
		case ProtocolKitty:
			if str, err := renderKitty(canvas, widthCols, targetRows); err == nil {
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
		result = renderHalfBlocks(img, clampedWidth, targetRows)
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

func renderHalfBlocks(img image.Image, targetWidth, targetRows int) string {
	maxH := targetRows * 2
	resized := resize.Thumbnail(uint(targetWidth), uint(maxH), img, resize.Lanczos3)
	bounds := resized.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	leftPad := (targetWidth - w) / 2
	rightPad := targetWidth - w - leftPad
	leftPadStr := strings.Repeat(" ", leftPad)
	rightPadStr := strings.Repeat(" ", rightPad)

	var renderedRows []string
	for y := 0; y < h; y += 2 {
		var sb strings.Builder
		sb.WriteString("\x1b[48;2;15;17;23m") // TrueColor ColorBg #0F1117
		sb.WriteString(leftPadStr)
		for x := 0; x < w; x++ {
			r1, g1, b1, a1 := resized.At(bounds.Min.X+x, bounds.Min.Y+y).RGBA()
			r1, g1, b1 = r1>>8, g1>>8, b1>>8

			var r2, g2, b2, a2 uint32
			if y+1 < h {
				r2, g2, b2, a2 = resized.At(bounds.Min.X+x, bounds.Min.Y+y+1).RGBA()
				r2, g2, b2 = r2>>8, g2>>8, b2>>8
			}

			if a1 < 32 && a2 < 32 {
				sb.WriteString("\x1b[48;2;15;17;23m ")
			} else if a1 < 32 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;15;17;23m▄", r2, g2, b2))
			} else if a2 < 32 {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;15;17;23m▀", r1, g1, b1))
			} else {
				sb.WriteString(fmt.Sprintf("\x1b[38;2;%d;%d;%dm\x1b[48;2;%d;%d;%dm▀", r1, g1, b1, r2, g2, b2))
			}
		}
		sb.WriteString(rightPadStr)
		sb.WriteString("\x1b[0m")
		renderedRows = append(renderedRows, sb.String())
	}

	topPad := (targetRows - len(renderedRows)) / 2
	bottomPad := targetRows - len(renderedRows) - topPad
	bgEmptyLine := "\x1b[48;2;15;17;23m" + strings.Repeat(" ", targetWidth) + "\x1b[0m"

	var finalLines []string
	for i := 0; i < topPad; i++ {
		finalLines = append(finalLines, bgEmptyLine)
	}
	finalLines = append(finalLines, renderedRows...)
	for i := 0; i < bottomPad; i++ {
		finalLines = append(finalLines, bgEmptyLine)
	}
	return strings.Join(finalLines, "\n")
}

func ClearCache() {
	spriteMutex.Lock()
	spriteMap = make(map[string]string)
	spriteMutex.Unlock()
}
