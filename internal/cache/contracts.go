package cache

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ThejusGSajan/val0/internal/models"
)

const contractsURL = "https://valorant-api.com/v1/contracts"

type contractsCache struct {
	Version   string                 `json:"version"`
	Contracts []models.ContractAsset `json:"contracts"`
	CachedAt  time.Time              `json:"cachedAt"`
}

// KnownBattlepassContracts maps active Act Season UUIDs to their Battlepass Contract Definition UUIDs.
var KnownBattlepassContracts = map[string]string{
	// Season UUID -> Contract Definition UUID
	"8102cd81-43a0-d0d7-bd59-47b8fe9bed1b": "3f04583c-4c7a-6bdf-65ce-d4b6ff53c5e9", // Season 2026 // Act V
}

func contractsCachePath() (string, error) {
	dir, err := cacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "contracts.json"), nil
}

// LoadOrFetchContracts loads the valorant-api.com contract assets from local cache or fetches them remotely.
func LoadOrFetchContracts(clientVersion string) ([]models.ContractAsset, error) {
	path, err := contractsCachePath()
	if err != nil {
		return nil, err
	}

	// Try loading from cache
	if data, err := os.ReadFile(path); err == nil {
		var cc contractsCache
		if json.Unmarshal(data, &cc) == nil && cc.Version == clientVersion && len(cc.Contracts) > 0 {
			return cc.Contracts, nil
		}
	}

	// Cache miss — fetch fresh data
	contracts, err := fetchAllContracts()
	if err != nil {
		return nil, err
	}

	_ = SaveContracts(contracts, clientVersion)
	return contracts, nil
}

// SaveContracts writes contract definitions to disk cache.
func SaveContracts(contracts []models.ContractAsset, clientVersion string) error {
	path, err := contractsCachePath()
	if err != nil {
		return err
	}

	cc := contractsCache{
		Version:   clientVersion,
		Contracts: contracts,
		CachedAt:  time.Now(),
	}
	data, err := json.MarshalIndent(cc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func fetchAllContracts() ([]models.ContractAsset, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Get(contractsURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch contracts failed with status %d", resp.StatusCode)
	}

	var r models.ValorantAPIContractsResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return r.Data, nil
}

// FindBattlepassContractAsset finds the Battlepass ContractAsset corresponding to the active season ID.
func FindBattlepassContractAsset(contracts []models.ContractAsset, seasonID string) *models.ContractAsset {
	if seasonID == "" {
		return nil
	}
	for i := range contracts {
		if strings.EqualFold(contracts[i].Content.RelationType, "Season") &&
			strings.EqualFold(contracts[i].Content.RelationUUID, seasonID) {
			return &contracts[i]
		}
	}
	return nil
}

// DefaultBattlepassLevelXPs returns the standard 55-tier Valorant Battlepass XP requirement curve.
// Tiers 1–50 (Chapters 1–10): Tier 1 = 0 XP, Tiers 2–50 = (tier * 750) + 500 XP.
// Tiers 51–55 (Epilogue): 36,500 XP each.
func DefaultBattlepassLevelXPs() []int {
	xps := make([]int, 55)
	xps[0] = 0 // Tier 1: Free starting unlock
	for i := 1; i < 50; i++ {
		tier := i + 1
		xps[i] = (tier * 750) + 500
	}
	for i := 50; i < 55; i++ {
		xps[i] = 36500
	}
	return xps
}

// GetBattlepassLevelXPs extracts the per-tier XP curve from the contract asset across all chapters.
// If the contract has no levels or is nil, it falls back to DefaultBattlepassLevelXPs.
func GetBattlepassLevelXPs(asset *models.ContractAsset) []int {
	if asset == nil {
		return DefaultBattlepassLevelXPs()
	}
	var xps []int
	for _, chapter := range asset.Content.Chapters {
		for _, level := range chapter.Levels {
			xps = append(xps, level.XP)
		}
	}
	if len(xps) == 0 {
		return DefaultBattlepassLevelXPs()
	}
	return xps
}

// CalculateBattlepassTier computes the active tier, XP accumulated in that tier,
// XP required for that tier, and whether all tiers are completed.
func CalculateBattlepassTier(totalXP int, levelXPs []int) (currentTier, xpInTier, xpForTier int, completed bool) {
	if len(levelXPs) == 0 {
		levelXPs = DefaultBattlepassLevelXPs()
	}

	accumulated := 0
	for i, reqXP := range levelXPs {
		tierNum := i + 1
		if totalXP < accumulated+reqXP {
			// Player is currently working on this tier
			return tierNum, totalXP - accumulated, reqXP, false
		}
		accumulated += reqXP
	}

	// Completed all tiers (Tier 55)
	lastTierXP := levelXPs[len(levelXPs)-1]
	return len(levelXPs), lastTierXP, lastTierXP, true
}
