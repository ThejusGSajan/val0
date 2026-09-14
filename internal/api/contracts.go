package api

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/ThejusGSajan/val0/internal/cache"
	"github.com/ThejusGSajan/val0/internal/models"
)

// FetchContracts retrieves all player contracts (including Battlepass).
//
//	GET https://pd.{shard}.a.pvp.net/contracts/v1/contracts/{puuid}
func (c *Client) FetchContracts() (*models.ContractsResponse, error) {
	url := c.pdURL(fmt.Sprintf("/contracts/v1/contracts/%s", c.session.PUUID))
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch contracts: %w", err)
	}

	var cr models.ContractsResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decode contracts: %w", err)
	}
	return &cr, nil
}

// FetchContent retrieves the active season/act metadata.
//
//	GET https://shared.{shard}.a.pvp.net/content-service/v3/content
func (c *Client) FetchContent() (*models.ContentResponse, error) {
	url := c.sharedURL("/content-service/v3/content")
	body, err := c.doRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("fetch content: %w", err)
	}

	var cr models.ContentResponse
	if err := json.Unmarshal(body, &cr); err != nil {
		return nil, fmt.Errorf("decode content: %w", err)
	}
	return &cr, nil
}

// FindActiveBattlepass identifies the battlepass contract from the
// contracts list by cross-referencing with the active season or explicit contract definition ID.
// Returns the matching Contract and true, or zero-value and false.
func FindActiveBattlepass(
	contracts *models.ContractsResponse,
	content *models.ContentResponse,
	bpContractDefID ...string,
) (models.Contract, bool) {
	if contracts == nil || content == nil {
		return models.Contract{}, false
	}

	// 1. Explicit Contract Definition ID if resolved from asset cache
	var targetDefID string
	if len(bpContractDefID) > 0 && bpContractDefID[0] != "" {
		targetDefID = bpContractDefID[0]
	}

	// 2. Identify active Act ID
	var activeActID string
	for _, s := range content.Seasons {
		if s.IsActive && s.Type == "act" {
			activeActID = s.ID
			break
		}
	}

	// 3. Check KnownBattlepassContracts map if targetDefID not provided
	if targetDefID == "" && activeActID != "" {
		targetDefID = cache.KnownBattlepassContracts[activeActID]
	}

	// 4. Match against targetDefID
	if targetDefID != "" {
		for _, c := range contracts.Contracts {
			if strings.EqualFold(c.ContractDefinitionID, targetDefID) {
				return c, true
			}
		}
	}

	// 5. Direct activeActID match (legacy/test compatibility)
	if activeActID != "" {
		for _, c := range contracts.Contracts {
			if strings.EqualFold(c.ContractDefinitionID, activeActID) {
				return c, true
			}
		}
	}

	// 6. Safe Fallback: Exclude active agent special contract
	var best models.Contract
	var bestXP int
	for _, c := range contracts.Contracts {
		if contracts.ActiveSpecialContract != "" &&
			strings.EqualFold(c.ContractDefinitionID, contracts.ActiveSpecialContract) {
			continue
		}
		if xp := c.ContractProgression.TotalProgressionEarned; xp > bestXP {
			bestXP = xp
			best = c
		}
	}
	if bestXP > 0 {
		return best, true
	}
	return models.Contract{}, false
}
