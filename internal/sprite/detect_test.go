package sprite

import (
	"os"
	"testing"
)

func TestDetectTerminalProtocol(t *testing.T) {
	// Test Sixel via VAL0_GRAPHICS
	os.Setenv("VAL0_GRAPHICS", "sixel")
	OverrideProtocol(probeEnvironment())
	if DetectTerminalProtocol() != ProtocolSixel {
		t.Errorf("expected ProtocolSixel, got %v", DetectTerminalProtocol())
	}

	// Test Kitty via VAL0_GRAPHICS
	os.Setenv("VAL0_GRAPHICS", "kitty")
	OverrideProtocol(probeEnvironment())
	if DetectTerminalProtocol() != ProtocolKitty {
		t.Errorf("expected ProtocolKitty, got %v", DetectTerminalProtocol())
	}

	// Test HalfBlock fallback
	os.Setenv("VAL0_GRAPHICS", "halfblock")
	OverrideProtocol(probeEnvironment())
	if DetectTerminalProtocol() != ProtocolHalfBlock {
		t.Errorf("expected ProtocolHalfBlock, got %v", DetectTerminalProtocol())
	}

	os.Unsetenv("VAL0_GRAPHICS")
}
