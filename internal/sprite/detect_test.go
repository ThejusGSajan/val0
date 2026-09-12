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

	// Test Windows Terminal defaults to Sixel
	os.Setenv("WT_SESSION", "test-guid-123")
	if probeEnvironment() != ProtocolSixel {
		t.Errorf("expected ProtocolSixel for WT_SESSION, got %v", probeEnvironment())
	}

	// Test Windows Terminal with explicit VAL0_GRAPHICS=halfblock override
	os.Setenv("VAL0_GRAPHICS", "halfblock")
	if probeEnvironment() != ProtocolHalfBlock {
		t.Errorf("expected ProtocolHalfBlock for WT_SESSION with VAL0_GRAPHICS=halfblock, got %v", probeEnvironment())
	}
	os.Unsetenv("VAL0_GRAPHICS")
	os.Unsetenv("WT_SESSION")

	// Test fallback when WT_SESSION is not set
	if probeEnvironment() != ProtocolHalfBlock {
		t.Errorf("expected ProtocolHalfBlock when WT_SESSION is unset, got %v", probeEnvironment())
	}
}

