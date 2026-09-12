package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// SessionInactivityThreshold defines the inactivity duration after which a new session is started.
const SessionInactivityThreshold = 2 * time.Hour

// SessionData represents the persistent state of a user's tracking session.
type SessionData struct {
	PlayerPUUID        string   `json:"playerPUUID"`
	SessionStartMillis int64    `json:"sessionStartMillis"`
	LastActiveMillis   int64    `json:"lastActiveMillis"`
	InitialMatchIDs    []string `json:"initialMatchIDs"`
	FilterMode         int      `json:"filterMode"` // 0 = Competitive Only, 1 = Comp+Unrated+Swiftplay+SpikeRush
}

var sessionMutex sync.Mutex

func sessionFilePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "session.json"), nil
}

// LoadSession reads session.json from disk.
// Returns nil, nil if the file does not exist.
func LoadSession() (*SessionData, error) {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	path, err := sessionFilePath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var sd SessionData
	if err := json.Unmarshal(data, &sd); err != nil {
		return nil, err
	}
	return &sd, nil
}

// SaveSession writes session data to disk.
func SaveSession(sd *SessionData) error {
	sessionMutex.Lock()
	defer sessionMutex.Unlock()

	path, err := sessionFilePath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(sd, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}
