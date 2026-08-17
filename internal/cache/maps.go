package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const mapsURL = "https://valorant-api.com/v1/maps"

// DefaultMapNames provides fallback map names by map URL or asset name.
var DefaultMapNames = map[string]string{
	"/game/maps/ascent/ascent":       "Ascent",
	"/game/maps/duality/duality":     "Bind",
	"/game/maps/bonsai/bonsai":       "Split",
	"/game/maps/triad/triad":         "Haven",
	"/game/maps/port/port":           "Icebox",
	"/game/maps/foxtrot/foxtrot":     "Breeze",
	"/game/maps/canyon/canyon":       "Fracture",
	"/game/maps/pitt/pitt":           "Pearl",
	"/game/maps/jam/jam":             "Lotus",
	"/game/maps/juliett/juliett":     "Sunset",
	"/game/maps/hurm_yard/hurm_yard": "District",
	"/game/maps/infinity/infinity":   "Abyss",
	"/game/maps/kasbah/kasbah":       "Kasbah",
	"/game/maps/drift/drift":         "Drift",
	"/game/maps/glitch/glitch":       "Glitch",
	"ascent":                         "Ascent",
	"bind":                           "Bind",
	"split":                          "Split",
	"haven":                          "Haven",
	"icebox":                         "Icebox",
	"breeze":                         "Breeze",
	"fracture":                       "Fracture",
	"pearl":                          "Pearl",
	"lotus":                          "Lotus",
	"sunset":                         "Sunset",
	"abyss":                          "Abyss",
}

type mapsCache struct {
	Version  string            `json:"version"`
	Maps     map[string]string `json:"maps"`
	CachedAt time.Time         `json:"cachedAt"`
}

type mapsAPIResponse struct {
	Status int `json:"status"`
	Data   []struct {
		UUID        string `json:"uuid"`
		DisplayName string `json:"displayName"`
		MapURL      string `json:"mapUrl"`
	} `json:"data"`
}

// LoadOrFetchMaps returns map from mapURL/UUID -> DisplayName.
func LoadOrFetchMaps(currentVersion string) (map[string]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return DefaultMapNames, nil
	}
	path := filepath.Join(dir, "maps.json")

	if data, err := os.ReadFile(path); err == nil {
		var mc mapsCache
		if json.Unmarshal(data, &mc) == nil && mc.Maps != nil && len(mc.Maps) > 0 {
			if currentVersion == "" || mc.Version == currentVersion {
				return mc.Maps, nil
			}
		}
	}

	maps, err := fetchMaps()
	if err != nil {
		return DefaultMapNames, nil
	}

	mc := mapsCache{Version: currentVersion, Maps: maps, CachedAt: time.Now()}
	if data, err := json.MarshalIndent(mc, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}

	return maps, nil
}

func fetchMaps() (map[string]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(mapsURL)
	if err != nil {
		return nil, fmt.Errorf("fetch maps: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var mr mapsAPIResponse
	if err := json.Unmarshal(body, &mr); err != nil {
		return nil, err
	}

	maps := make(map[string]string, len(mr.Data)*2)
	for _, m := range mr.Data {
		maps[strings.ToLower(m.MapURL)] = m.DisplayName
		maps[strings.ToLower(m.UUID)] = m.DisplayName
		// Also store simple name
		parts := strings.Split(m.MapURL, "/")
		if len(parts) > 0 {
			last := strings.ToLower(parts[len(parts)-1])
			maps[last] = m.DisplayName
		}
	}

	if len(maps) == 0 {
		return DefaultMapNames, nil
	}
	return maps, nil
}

// GetMapName resolves a map path or ID to a human-friendly name.
func GetMapName(mapID string, mapsMap map[string]string) string {
	clean := strings.ToLower(strings.TrimSpace(mapID))
	if mapsMap != nil {
		if name, ok := mapsMap[clean]; ok && name != "" {
			return name
		}
	}
	if name, ok := DefaultMapNames[clean]; ok {
		return name
	}
	// Try parsing base name from "/Game/Maps/Ascent/Ascent"
	parts := strings.Split(clean, "/")
	if len(parts) > 0 {
		last := parts[len(parts)-1]
		if name, ok := DefaultMapNames[last]; ok {
			return name
		}
		if len(last) > 0 {
			return strings.Title(last)
		}
	}
	return "Unknown Map"
}
