package cache

import (
	"os"
	"testing"
	"time"
)

func TestSessionSaveAndLoad(t *testing.T) {
	now := time.Now()
	testData := &SessionData{
		PlayerPUUID:        "test-player-puuid",
		SessionStartMillis: now.Add(-1 * time.Hour).UnixMilli(),
		LastActiveMillis:   now.UnixMilli(),
		InitialMatchIDs:    []string{"match-1", "match-2"},
		FilterMode:         1,
	}

	if err := SaveSession(testData); err != nil {
		t.Fatalf("SaveSession failed: %v", err)
	}

	loaded, err := LoadSession()
	if err != nil {
		t.Fatalf("LoadSession failed: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected loaded session, got nil")
	}
	if loaded.PlayerPUUID != testData.PlayerPUUID {
		t.Errorf("PUUID = %s, want %s", loaded.PlayerPUUID, testData.PlayerPUUID)
	}
	if loaded.FilterMode != 1 {
		t.Errorf("FilterMode = %d, want 1", loaded.FilterMode)
	}
	if len(loaded.InitialMatchIDs) != 2 {
		t.Errorf("InitialMatchIDs len = %d, want 2", len(loaded.InitialMatchIDs))
	}
}

func TestSessionCorruptedFile(t *testing.T) {
	path, err := sessionFilePath()
	if err != nil {
		t.Fatalf("sessionFilePath failed: %v", err)
	}
	_ = os.WriteFile(path, []byte("{invalid json corrupt"), 0o644)

	loaded, err := LoadSession()
	if err == nil && loaded != nil {
		t.Errorf("expected error or nil for corrupted session file")
	}
}
