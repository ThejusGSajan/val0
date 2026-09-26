package cache

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

const MaxCacheSize = 25

var (
	memMatchCache   = make(map[string]*models.MatchDetails)
	memCacheOrder   []string // tracks insertion order for FIFO eviction
	memMatchCacheMu sync.RWMutex
)

// matchCacheDir returns the directory path for match detail caches.
func matchCacheDir() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	matchesDir := filepath.Join(dir, "matches")
	if err := os.MkdirAll(matchesDir, 0o755); err != nil {
		return "", err
	}
	return matchesDir, nil
}

// matchCacheFilePath returns the path to a cached match details file.
// On Windows: %AppData%\val-tracker\matches\<matchID>.json
func matchCacheFilePath(matchID string) (string, error) {
	dir, err := matchCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, matchID+".json"), nil
}

// GetCachedMatchDetails retrieves match details from memory or disk cache.
// Returns (details, true) on cache hit, or (nil, false) on miss.
func GetCachedMatchDetails(matchID string) (*models.MatchDetails, bool) {
	if matchID == "" {
		return nil, false
	}

	// 1. Check in-memory cache
	memMatchCacheMu.RLock()
	if d, ok := memMatchCache[matchID]; ok && d != nil {
		memMatchCacheMu.RUnlock()
		return d, true
	}
	memMatchCacheMu.RUnlock()

	// 2. Check disk cache
	path, err := matchCacheFilePath(matchID)
	if err != nil {
		return nil, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}

	var d models.MatchDetails
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, false
	}

	// Populate in-memory cache
	memMatchCacheMu.Lock()
	if _, exists := memMatchCache[matchID]; !exists {
		memCacheOrder = append(memCacheOrder, matchID)
		if len(memCacheOrder) > MaxCacheSize {
			oldest := memCacheOrder[0]
			memCacheOrder = memCacheOrder[1:]
			delete(memMatchCache, oldest)
		}
	}
	memMatchCache[matchID] = &d
	memMatchCacheMu.Unlock()

	return &d, true
}

// SaveCachedMatchDetails writes match details to both in-memory and disk cache.
func SaveCachedMatchDetails(d *models.MatchDetails) error {
	if d == nil || d.MatchInfo.MatchID == "" {
		return nil
	}

	matchID := d.MatchInfo.MatchID

	// 1. Update in-memory cache with FIFO eviction
	memMatchCacheMu.Lock()
	if _, exists := memMatchCache[matchID]; !exists {
		memCacheOrder = append(memCacheOrder, matchID)
		if len(memCacheOrder) > MaxCacheSize {
			oldest := memCacheOrder[0]
			memCacheOrder = memCacheOrder[1:]
			delete(memMatchCache, oldest)
		}
	}
	memMatchCache[matchID] = d
	memMatchCacheMu.Unlock()

	// 2. Persist to disk
	path, err := matchCacheFilePath(matchID)
	if err != nil {
		return err
	}

	data, err := json.Marshal(d)
	if err != nil {
		return err
	}

	if err := os.WriteFile(path, data, 0o644); err != nil {
		return err
	}

	// 3. Trigger async disk cleanup
	go evictOldDiskCaches()
	return nil
}

// evictOldDiskCaches ensures the matches directory doesn't exceed MaxCacheSize.
func evictOldDiskCaches() {
	dir, err := matchCacheDir()
	if err != nil {
		return
	}
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) <= MaxCacheSize {
		return
	}

	type fileInfo struct {
		path    string
		modTime time.Time
	}
	var files []fileInfo
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			info, err := entry.Info()
			if err == nil {
				files = append(files, fileInfo{
					path:    filepath.Join(dir, entry.Name()),
					modTime: info.ModTime(),
				})
			}
		}
	}

	// Sort oldest first
	sort.Slice(files, func(i, j int) bool {
		return files[i].modTime.Before(files[j].modTime)
	})

	// Delete oldest files until we are at MaxCacheSize
	for i := 0; i < len(files)-MaxCacheSize; i++ {
		os.Remove(files[i].path)
	}
}
