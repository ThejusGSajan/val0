package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/val-tracker/val-tracker/internal/models"
)

const (
	skinsURL   = "https://valorant-api.com/v1/weapons/skins"
	versionURL = "https://valorant-api.com/v1/version"
)

// cacheDir returns the path to our cache directory inside the user's config dir.
// On Windows: %AppData%\val-tracker\
func cacheDir() (string, error) {
	base, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(base, "val-tracker")
	return dir, os.MkdirAll(dir, 0o755)
}

// ── Version Check ───────────────────────────────────────────────────

// FetchRemoteVersion returns the current Valorant client version string
// from valorant-api.com. Used to:
//   - Set X-Riot-ClientVersion on API requests
//   - Detect new patches (invalidate skin cache)
func FetchRemoteVersion() (*models.VersionData, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(versionURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch version failed with status %d", resp.StatusCode)
	}

	var vr models.ValorantAPIVersionResponse
	if err := json.NewDecoder(resp.Body).Decode(&vr); err != nil {
		return nil, err
	}
	return &vr.Data, nil
}

// ── Skin Cache ──────────────────────────────────────────────────────

type skinCache struct {
	Version  string             `json:"version"` // the riotClientVersion when cached
	Skins    []models.SkinAsset `json:"skins"`
	CachedAt time.Time          `json:"cachedAt"`
}

// LoadOrFetchSkins returns the skin asset list.
// It uses a local skins.json cache. The cache is invalidated when:
//   - The file does not exist
//   - The cached version doesn't match the current game version
func LoadOrFetchSkins(currentVersion string) ([]models.SkinAsset, error) {
	dir, err := cacheDir()
	if err != nil {
		return nil, err
	}
	path := filepath.Join(dir, "skins.json")

	// Try loading from cache
	if data, err := os.ReadFile(path); err == nil {
		var sc skinCache
		if json.Unmarshal(data, &sc) == nil && sc.Version == currentVersion {
			// Invalidate cache if cached skins lack AssetPath (migration to v36)
			hasAssetPath := false
			for _, s := range sc.Skins {
				if s.AssetPath != "" {
					hasAssetPath = true
					break
				}
			}
			if len(sc.Skins) > 0 && hasAssetPath {
				return sc.Skins, nil // cache hit
			}
		}
	}

	// Cache miss — fetch fresh data
	skins, err := fetchAllSkins()
	if err != nil {
		return nil, err
	}

	// Write cache
	sc := skinCache{Version: currentVersion, Skins: skins, CachedAt: time.Now()}
	data, _ := json.MarshalIndent(sc, "", "  ")
	_ = os.WriteFile(path, data, 0o644) // best-effort

	return skins, nil
}

func fetchAllSkins() ([]models.SkinAsset, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(skinsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch skins: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var sr models.ValorantAPISkinsResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, err
	}
	return sr.Data, nil
}

// BuildSkinLookup creates a map[uuid] → SkinAsset for O(1) lookups.
func BuildSkinLookup(skins []models.SkinAsset) map[string]models.SkinAsset {
	m := make(map[string]models.SkinAsset, len(skins))
	for _, s := range skins {
		m[s.UUID] = s
		// Also index by level UUID (the storefront returns level-0 UUIDs)
		for _, lvl := range s.Levels {
			m[lvl.UUID] = s
		}
	}
	return m
}
