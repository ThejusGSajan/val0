package cache

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

const ranksURL = "https://valorant-api.com/v1/competitivetiers"

// DefaultRankNames provides standard fallback rank names if offline/uncached.
var DefaultRankNames = map[int]string{
	0:  "Unranked",
	3:  "Iron 1",
	4:  "Iron 2",
	5:  "Iron 3",
	6:  "Bronze 1",
	7:  "Bronze 2",
	8:  "Bronze 3",
	9:  "Silver 1",
	10: "Silver 2",
	11: "Silver 3",
	12: "Gold 1",
	13: "Gold 2",
	14: "Gold 3",
	15: "Platinum 1",
	16: "Platinum 2",
	17: "Platinum 3",
	18: "Diamond 1",
	19: "Diamond 2",
	20: "Diamond 3",
	21: "Ascendant 1",
	22: "Ascendant 2",
	23: "Ascendant 3",
	24: "Immortal 1",
	25: "Immortal 2",
	26: "Immortal 3",
	27: "Radiant",
}

type ranksCache struct {
	Version  string         `json:"version"`
	Ranks    map[int]string `json:"ranks"`
	CachedAt time.Time      `json:"cachedAt"`
}

type competitiveTiersResponse struct {
	Status int `json:"status"`
	Data   []struct {
		UUID            string `json:"uuid"`
		AssetObjectName string `json:"assetObjectName"`
		Tiers           []struct {
			Tier        int    `json:"tier"`
			TierName    string `json:"tierName"`
			DivisionName string `json:"divisionName"`
		} `json:"tiers"`
	} `json:"data"`
}

// LoadOrFetchRanks loads competitive tier mapping (tier number -> tier name).
func LoadOrFetchRanks(currentVersion string) (map[int]string, error) {
	dir, err := cacheDir()
	if err != nil {
		return DefaultRankNames, nil
	}
	path := filepath.Join(dir, "ranks.json")

	// Try loading from cache
	if data, err := os.ReadFile(path); err == nil {
		var rc ranksCache
		if json.Unmarshal(data, &rc) == nil && rc.Ranks != nil && len(rc.Ranks) > 0 {
			if currentVersion == "" || rc.Version == currentVersion {
				return rc.Ranks, nil
			}
		}
	}

	// Fetch fresh data
	ranks, err := fetchRanks()
	if err != nil {
		// Graceful fallback to default rank names
		return DefaultRankNames, nil
	}

	rc := ranksCache{Version: currentVersion, Ranks: ranks, CachedAt: time.Now()}
	if data, err := json.MarshalIndent(rc, "", "  "); err == nil {
		_ = os.WriteFile(path, data, 0o644)
	}

	return ranks, nil
}

func fetchRanks() (map[int]string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(ranksURL)
	if err != nil {
		return nil, fmt.Errorf("fetch ranks: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var ctr competitiveTiersResponse
	if err := json.Unmarshal(body, &ctr); err != nil {
		return nil, err
	}

	// Use the latest competitive tier list (last element in Data)
	ranks := make(map[int]string)
	for i := len(ctr.Data) - 1; i >= 0; i-- {
		episode := ctr.Data[i]
		if len(episode.Tiers) > 0 {
			for _, t := range episode.Tiers {
				name := t.TierName
				if name == "" {
					name = t.DivisionName
				}
				ranks[t.Tier] = name
			}
			break
		}
	}

	if len(ranks) == 0 {
		return DefaultRankNames, nil
	}
	return ranks, nil
}

// GetRankName returns the name for a tier or "Unranked".
func GetRankName(tier int, ranksMap map[int]string) string {
	if ranksMap != nil {
		if name, ok := ranksMap[tier]; ok && name != "" {
			return name
		}
	}
	if name, ok := DefaultRankNames[tier]; ok {
		return name
	}
	return "Unranked"
}
