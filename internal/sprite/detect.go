package sprite

import (
	"os"
	"strings"
	"sync"
)

type GraphicsProtocol int

const (
	ProtocolHalfBlock GraphicsProtocol = iota // Standard ANSI Unicode ▀/▄ fallback
	ProtocolSixel                             // Sixel Graphics (Windows Terminal v1.22+, WezTerm, Foot)
	ProtocolKitty                             // Kitty Graphics Protocol (Ghostty, Kitty, WezTerm)
	ProtocolITerm2                            // iTerm2 Inline Images (iTerm2, VS Code Terminal)
)

func (p GraphicsProtocol) String() string {
	switch p {
	case ProtocolSixel:
		return "Sixel"
	case ProtocolKitty:
		return "Kitty"
	case ProtocolITerm2:
		return "iTerm2"
	default:
		return "HalfBlock"
	}
}

var (
	detectedProtocol GraphicsProtocol
	detectOnce       sync.Once
)

// DetectTerminalProtocol returns the best graphics protocol supported by the current terminal.
func DetectTerminalProtocol() GraphicsProtocol {
	detectOnce.Do(func() {
		detectedProtocol = probeEnvironment()
	})
	return detectedProtocol
}

// OverrideProtocol sets the protocol manually (useful for testing or overrides).
func OverrideProtocol(p GraphicsProtocol) {
	detectedProtocol = p
}

func probeEnvironment() GraphicsProtocol {
	// 1. Manual user override via VAL0_GRAPHICS
	override := strings.ToLower(os.Getenv("VAL0_GRAPHICS"))
	switch override {
	case "sixel":
		return ProtocolSixel
	case "kitty":
		return ProtocolKitty
	case "iterm2":
		return ProtocolITerm2
	case "none", "halfblock":
		return ProtocolHalfBlock
	}

	// 2. TMUX check: without passthrough, multiplexers strip graphics sequences
	if os.Getenv("TMUX") != "" {
		return ProtocolHalfBlock
	}

	// 3. Kitty Protocol (Ghostty, Kitty)
	if os.Getenv("KITTY_WINDOW_ID") != "" || os.Getenv("GHOSTTY_RESOURCES_DIR") != "" {
		return ProtocolKitty
	}

	// 4. WezTerm (Supports Kitty, Sixel, iTerm2 — Kitty offers best quality)
	termProgram := os.Getenv("TERM_PROGRAM")
	if termProgram == "WezTerm" {
		return ProtocolKitty
	}

	// 5. iTerm2 / VS Code Integrated Terminal
	if termProgram == "iTerm.app" || termProgram == "vscode" {
		return ProtocolITerm2
	}

	// 6. Windows Terminal: Default to ANSI Half-Blocks because ConPTY's cell buffer desynchronizes
	// with Sixel DCS cursor jumping, causing horizontal raster clearing black bars.
	// Users can still explicitly opt-in to Sixel via VAL0_GRAPHICS=sixel.
	if os.Getenv("WT_SESSION") != "" {
		return ProtocolHalfBlock
	}

	// 7. Sixel-capable TERM names
	term := strings.ToLower(os.Getenv("TERM"))
	if strings.Contains(term, "sixel") || strings.Contains(term, "foot") || strings.Contains(term, "mlterm") {
		return ProtocolSixel
	}

	return ProtocolHalfBlock
}
