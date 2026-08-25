package launcher

import (
	"testing"
)

func TestIsStandaloneConhost_EnvOverrides(t *testing.T) {
	// If WT_SESSION is set, it should never be considered standalone conhost
	t.Setenv("WT_SESSION", "some-guid-value")
	if IsStandaloneConhost() {
		t.Errorf("Expected IsStandaloneConhost() = false when WT_SESSION is set")
	}

	// If VAL0_RELAUNCHED is set to 1, it should prevent loops
	t.Setenv("WT_SESSION", "")
	t.Setenv("VAL0_RELAUNCHED", "1")
	if IsStandaloneConhost() {
		t.Errorf("Expected IsStandaloneConhost() = false when VAL0_RELAUNCHED is 1")
	}
}

func TestFindWTPath(t *testing.T) {
	// Just verify FindWTPath does not panic
	path, err := FindWTPath()
	if err == nil && path == "" {
		t.Errorf("FindWTPath returned nil error but empty path")
	}
}
