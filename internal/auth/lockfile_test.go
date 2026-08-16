package auth

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseLockfile(t *testing.T) {
	raw := "Riot Client:12345:55555:secretPass:https"
	modTime := time.Now()

	lf, err := parseLockfile(raw, modTime)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if lf.Name != "Riot Client" {
		t.Errorf("expected Name 'Riot Client', got '%s'", lf.Name)
	}
	if lf.PID != "12345" {
		t.Errorf("expected PID '12345', got '%s'", lf.PID)
	}
	if lf.Port != "55555" {
		t.Errorf("expected Port '55555', got '%s'", lf.Port)
	}
	if lf.Password != "secretPass" {
		t.Errorf("expected Password 'secretPass', got '%s'", lf.Password)
	}
	if lf.Protocol != "https" {
		t.Errorf("expected Protocol 'https', got '%s'", lf.Protocol)
	}
}

func TestParseLockfile_Malformed(t *testing.T) {
	raw := "Riot Client:12345:55555"
	_, err := parseLockfile(raw, time.Now())
	if err == nil {
		t.Errorf("expected error for malformed lockfile, got nil")
	}
}

func TestReadLockfile_Stale(t *testing.T) {
	tempDir := t.TempDir()
	// Set LOCALAPPDATA to temp dir to test stale check
	origLocalApp := os.Getenv("LOCALAPPDATA")
	defer os.Setenv("LOCALAPPDATA", origLocalApp)
	os.Setenv("LOCALAPPDATA", tempDir)

	lockDir := filepath.Join(tempDir, "Riot Games", "Riot Client", "Config")
	if err := os.MkdirAll(lockDir, 0o755); err != nil {
		t.Fatal(err)
	}

	lockPath := filepath.Join(lockDir, "lockfile")
	if err := os.WriteFile(lockPath, []byte("Riot Client:123:456:pwd:https"), 0o644); err != nil {
		t.Fatal(err)
	}

	// Change modtime to 2 hours ago
	oldTime := time.Now().Add(-2 * time.Hour)
	if err := os.Chtimes(lockPath, oldTime, oldTime); err != nil {
		t.Fatal(err)
	}

	_, err := ReadLockfile()
	if err != ErrLockfileStale {
		t.Errorf("expected ErrLockfileStale, got: %v", err)
	}
}

func TestReadLockfile_NotFound(t *testing.T) {
	tempDir := t.TempDir()
	origLocalApp := os.Getenv("LOCALAPPDATA")
	defer os.Setenv("LOCALAPPDATA", origLocalApp)
	os.Setenv("LOCALAPPDATA", tempDir)

	_, err := ReadLockfile()
	if err != ErrLockfileNotFound {
		t.Errorf("expected ErrLockfileNotFound, got: %v", err)
	}
}

func TestRegionToShardMapping(t *testing.T) {
	tests := []struct {
		region   string
		expected string
	}{
		{"na", "na"},
		{"eu", "eu"},
		{"ap", "ap"},
		{"kr", "kr"},
		{"latam", "na"},
		{"br", "na"},
	}

	for _, tt := range tests {
		shard, ok := RegionToShard[tt.region]
		if !ok || shard != tt.expected {
			t.Errorf("for region %s: expected %s, got %s (ok=%v)", tt.region, tt.expected, shard, ok)
		}
	}
}

func TestURLBuilders(t *testing.T) {
	if got := PDBaseURL("na"); got != "https://pd.na.a.pvp.net" {
		t.Errorf("unexpected PDBaseURL: %s", got)
	}
	if got := SharedBaseURL("eu"); got != "https://shared.eu.a.pvp.net" {
		t.Errorf("unexpected SharedBaseURL: %s", got)
	}
}
